// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package subscribe

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/consul/state"
	"github.com/dhiaayachi/consul/agent/consul/stream"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/proto/private/pbsubscribe"
)

// Server implements a StateChangeSubscriptionServer for accepting SubscribeRequests,
// and sending events to the subscription topic.
type Server struct {
	Backend Backend
	Logger  Logger
}

func NewServer(backend Backend, logger Logger) *Server {
	return &Server{Backend: backend, Logger: logger}
}

type Logger interface {
	Trace(msg string, args ...interface{})
	With(args ...interface{}) hclog.Logger
}

var _ pbsubscribe.StateChangeSubscriptionServer = (*Server)(nil)

type Backend interface {
	ResolveTokenAndDefaultMeta(token string, entMeta *acl.EnterpriseMeta, authzContext *acl.AuthorizerContext) (acl.Authorizer, error)
	Forward(info structs.RPCInfo, f func(*grpc.ClientConn) error) (handled bool, err error)
	Subscribe(req *stream.SubscribeRequest) (*stream.Subscription, error)
}

func (h *Server) Subscribe(req *pbsubscribe.SubscribeRequest, serverStream pbsubscribe.StateChangeSubscription_SubscribeServer) error {
	logger := newLoggerForRequest(h.Logger, req)
	handled, err := h.Backend.Forward(req, forwardToDC(req, serverStream, logger))
	if handled || err != nil {
		return err
	}

	logger.Trace("new subscription")
	defer logger.Trace("subscription closed")

	entMeta := req.EnterpriseMeta()
	authz, err := h.Backend.ResolveTokenAndDefaultMeta(req.Token, &entMeta, nil)
	if err != nil {
		return err
	}

	subReq, err := state.PBToStreamSubscribeRequest(req, entMeta)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	sub, err := h.Backend.Subscribe(subReq)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	ctx := serverStream.Context()
	elog := &eventLogger{logger: logger}
	for {
		event, err := sub.Next(ctx)
		switch {
		case errors.Is(err, stream.ErrSubForceClosed):
			logger.Trace("subscription reset by server")
			return status.Error(codes.Aborted, err.Error())
		case errors.Is(err, stream.ErrACLChanged):
			logger.Trace("ACL change occurred; re-authenticating")
			_, authzErr := h.Backend.ResolveTokenAndDefaultMeta(req.Token, &entMeta, nil)
			if authzErr != nil {
				return authzErr
			}
			// Otherwise, abort the stream so the client re-subscribes.
			return status.Error(codes.Aborted, err.Error())
		case err != nil:
			return err
		}

		if !event.Payload.HasReadPermission(authz) {
			continue
		}

		elog.Trace(event)

		// TODO: This conversion could be cached if needed
		e := event.Payload.ToSubscriptionEvent(event.Index)
		if err := serverStream.Send(e); err != nil {
			return err
		}
	}
}

func forwardToDC(
	req *pbsubscribe.SubscribeRequest,
	serverStream pbsubscribe.StateChangeSubscription_SubscribeServer,
	logger Logger,
) func(conn *grpc.ClientConn) error {
	return func(conn *grpc.ClientConn) error {
		logger.Trace("forwarding to another DC")
		defer logger.Trace("forwarded stream closed")

		client := pbsubscribe.NewStateChangeSubscriptionClient(conn)
		streamHandle, err := client.Subscribe(serverStream.Context(), req)
		if err != nil {
			return err
		}

		for {
			event, err := streamHandle.Recv()
			if err != nil {
				return err
			}
			if err := serverStream.Send(event); err != nil {
				return err
			}
		}
	}
}

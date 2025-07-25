// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/hashicorp/go-hclog"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/configentry"
	"github.com/dhiaayachi/consul/agent/consul/state"
	external "github.com/dhiaayachi/consul/agent/grpc-external"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/agent/xds/accesslogs"
	"github.com/dhiaayachi/consul/proto-public/pbdataplane"
)

func (s *Server) GetEnvoyBootstrapParams(ctx context.Context, req *pbdataplane.GetEnvoyBootstrapParamsRequest) (*pbdataplane.GetEnvoyBootstrapParamsResponse, error) {
	proxyID := req.ProxyId
	if req.GetServiceId() != "" {
		proxyID = req.GetServiceId()
	}
	logger := s.Logger.Named("get-envoy-bootstrap-params").With("proxy_id", proxyID, "request_id", external.TraceID())

	logger.Trace("Started processing request")
	defer logger.Trace("Finished processing request")

	options, err := external.QueryOptionsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var authzContext acl.AuthorizerContext
	entMeta := acl.NewEnterpriseMetaWithPartition(req.GetPartition(), req.GetNamespace())
	authz, err := s.ACLResolver.ResolveTokenAndDefaultMeta(options.Token, &entMeta, &authzContext)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	store := s.GetStore()

	_, svc, err := store.ServiceNode(req.GetNodeId(), req.GetNodeName(), proxyID, &entMeta, structs.DefaultPeerKeyword)
	if err != nil {
		logger.Error("Error looking up service", "error", err)
		if errors.Is(err, state.ErrNodeNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else if strings.Contains(err.Error(), "Node ID or name required") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			return nil, status.Error(codes.Internal, "Failure looking up service")
		}
	}
	if svc == nil {
		return nil, status.Error(codes.NotFound, "Service not found")
	}

	if err := authz.ToAllowAuthorizer().ServiceReadAllowed(svc.ServiceName, &authzContext); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	_, ns, err := configentry.MergeNodeServiceWithCentralConfig(
		nil,
		store,
		svc.ToNodeService(),
		logger,
	)
	if err != nil {
		logger.Error("Error merging with central config", "error", err)
		return nil, status.Errorf(codes.Unknown, "Error merging central config: %v", err)
	}

	bootstrapConfig, err := structpb.NewStruct(ns.Proxy.Config)
	if err != nil {
		logger.Error("Error creating the envoy boostrap params config", "error", err)
		return nil, status.Error(codes.Unknown, "Error creating the envoy boostrap params config")
	}

	// Inspect access logging
	// This is non-essential, and don't want to return an error unless there is a more serious issue
	var accessLogs []string
	if ns != nil {
		accessLogs = makeAccessLogs(&ns.Proxy.AccessLogs, logger)
	}

	// Build out the response
	var serviceName string
	if svc.ServiceKind == structs.ServiceKindConnectProxy {
		serviceName = svc.ServiceProxy.DestinationServiceName
	} else {
		serviceName = svc.ServiceName
	}

	return &pbdataplane.GetEnvoyBootstrapParamsResponse{
		Identity:   serviceName,
		Service:    serviceName,
		Partition:  svc.EnterpriseMeta.PartitionOrDefault(),
		Namespace:  svc.EnterpriseMeta.NamespaceOrDefault(),
		Config:     bootstrapConfig,
		Datacenter: s.Datacenter,
		NodeName:   svc.Node,
		AccessLogs: accessLogs,
	}, nil
}

func makeAccessLogs(logs *structs.AccessLogsConfig, logger hclog.Logger) []string {
	var accessLogs []string
	if logs.Enabled {
		envoyLoggers, err := accesslogs.MakeAccessLogs(logs, false)
		if err != nil {
			logger.Warn("Error creating the envoy access log config", "error", err)
		}

		accessLogs = make([]string, 0, len(envoyLoggers))

		for _, msg := range envoyLoggers {
			logConfig, err := protojson.Marshal(msg)
			if err != nil {
				logger.Warn("Error marshaling the envoy access log config", "error", err)
			}
			accessLogs = append(accessLogs, string(logConfig))
		}
	}

	return accessLogs
}

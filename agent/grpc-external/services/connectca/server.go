// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package connectca

import (
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-memdb"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/acl/resolver"
	"github.com/dhiaayachi/consul/agent/consul/stream"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/proto-public/pbconnectca"
)

type Server struct {
	Config
}

type Config struct {
	Publisher      EventPublisher
	GetStore       func() StateStore
	Logger         hclog.Logger
	ACLResolver    ACLResolver
	CAManager      CAManager
	ForwardRPC     func(structs.RPCInfo, func(*grpc.ClientConn) error) (bool, error)
	ConnectEnabled bool
}

type EventPublisher interface {
	Subscribe(*stream.SubscribeRequest) (*stream.Subscription, error)
}

type StateStore interface {
	CAConfig(memdb.WatchSet) (uint64, *structs.CAConfiguration, error)
	AbandonCh() <-chan struct{}
}

//go:generate mockery --name ACLResolver --inpackage
type ACLResolver interface {
	ResolveTokenAndDefaultMeta(token string, entMeta *acl.EnterpriseMeta, authzContext *acl.AuthorizerContext) (resolver.Result, error)
}

//go:generate mockery --name CAManager --inpackage
type CAManager interface {
	AuthorizeAndSignCertificate(csr *x509.CertificateRequest, authz acl.Authorizer) (*structs.IssuedCert, error)
}

func NewServer(cfg Config) *Server {
	return &Server{cfg}
}

func (s *Server) Register(registrar grpc.ServiceRegistrar) {
	pbconnectca.RegisterConnectCAServiceServer(registrar, s)
}

func (s *Server) requireConnect() error {
	if s.ConnectEnabled {
		return nil
	}
	return status.Error(codes.FailedPrecondition, "Connect must be enabled in order to use this endpoint")
}

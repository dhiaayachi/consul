// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package dataplane

import (
	"google.golang.org/grpc"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-memdb"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/acl/resolver"
	"github.com/dhiaayachi/consul/agent/configentry"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/proto-public/pbdataplane"
)

type Server struct {
	Config
}

type Config struct {
	GetStore    func() StateStore
	Logger      hclog.Logger
	ACLResolver ACLResolver
	// Datacenter of the Consul server this gRPC server is hosted on
	Datacenter string
}

type StateStore interface {
	ServiceNode(string, string, string, *acl.EnterpriseMeta, string) (uint64, *structs.ServiceNode, error)
	ReadResolvedServiceConfigEntries(memdb.WatchSet, string, *acl.EnterpriseMeta, []structs.ServiceID, structs.ProxyMode) (uint64, *configentry.ResolvedServiceConfigSet, error)
}

//go:generate mockery --name ACLResolver --inpackage
type ACLResolver interface {
	ResolveTokenAndDefaultMeta(string, *acl.EnterpriseMeta, *acl.AuthorizerContext) (resolver.Result, error)
}

func NewServer(cfg Config) *Server {
	return &Server{cfg}
}

var _ pbdataplane.DataplaneServiceServer = (*Server)(nil)

func (s *Server) Register(registrar grpc.ServiceRegistrar) {
	pbdataplane.RegisterDataplaneServiceServer(registrar, s)
}

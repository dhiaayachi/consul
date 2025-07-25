// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package proxycfgglue

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/consul/state"
	"github.com/dhiaayachi/consul/agent/proxycfg"
	"github.com/dhiaayachi/consul/agent/structs"
)

func TestServerInternalServiceDump(t *testing.T) {
	t.Run("remote queries are delegated to the remote source", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		var (
			req           = &structs.ServiceDumpRequest{Datacenter: "dc2"}
			correlationID = "correlation-id"
			ch            = make(chan<- proxycfg.UpdateEvent)
			result        = errors.New("KABOOM")
		)

		remoteSource := newMockInternalServiceDump(t)
		remoteSource.On("Notify", ctx, req, correlationID, ch).Return(result)

		dataSource := ServerInternalServiceDump(ServerDataSourceDeps{Datacenter: "dc1"}, remoteSource)
		err := dataSource.Notify(ctx, req, correlationID, ch)
		require.Equal(t, result, err)
	})

	t.Run("local queries are served from the state store", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		nextIndex := indexGenerator()

		store := state.NewStateStore(nil)

		services := []*structs.NodeService{
			{
				Service: "mgw",
				Kind:    structs.ServiceKindMeshGateway,
			},
			{
				Service: "web",
				Kind:    structs.ServiceKindTypical,
			},
			{
				Service: "web-deny",
				Kind:    structs.ServiceKindTypical,
			},
			{
				Service: "db",
				Kind:    structs.ServiceKindTypical,
			},
		}
		for idx, service := range services {
			require.NoError(t, store.EnsureRegistration(nextIndex(), &structs.RegisterRequest{
				Node:    fmt.Sprintf("node-%d", idx),
				Service: service,
			}))
		}

		policyAuth := policyAuthorizer(t, `
				service "mgw" { policy = "read" }
				service "web" { policy = "read" }
				service "web-deny" { policy = "deny" }
				service "db"  { policy = "read" }
				node_prefix "node-" { policy = "read" }
			`)
		authz := newStaticResolver(policyAuth)

		dataSource := ServerInternalServiceDump(ServerDataSourceDeps{
			GetStore:    func() Store { return store },
			ACLResolver: authz,
		}, nil)

		t.Run("filter by kind", func(t *testing.T) {
			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{
				ServiceKind:    structs.ServiceKindMeshGateway,
				UseServiceKind: true,
			}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Len(t, result.Nodes, 1)
			require.Equal(t, "mgw", result.Nodes[0].Service.Service)
		})

		t.Run("bexpr filtering", func(t *testing.T) {
			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{
				QueryOptions: structs.QueryOptions{Filter: `Service.Service == "web"`},
			}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Len(t, result.Nodes, 1)
			require.Equal(t, "web", result.Nodes[0].Service.Service)
		})

		t.Run("all services", func(t *testing.T) {
			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Len(t, result.Nodes, 3)
		})

		t.Run("access denied", func(t *testing.T) {
			authz.SwapAuthorizer(acl.DenyAll())

			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Empty(t, result.Nodes)
		})

		const (
			bexprMatchingUserTokenPermissions   = "Service.Service matches `web.*`"
			bexpNotMatchingUserTokenPermissions = "Service.Service matches `mgw.*`"
		)

		authz.SwapAuthorizer(policyAuthorizer(t, `
				service "mgw" { policy = "deny" }
				service "web" { policy = "read" }
				service "web-deny" { policy = "deny" }
				service "db"  { policy = "read" }
				node_prefix "node-" { policy = "read" }
			`))

		t.Run("request with filter that matches token permissions returns 1 result and ResultsFilteredByACLs equal to true", func(t *testing.T) {
			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{
				QueryOptions: structs.QueryOptions{Filter: bexprMatchingUserTokenPermissions},
			}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Len(t, result.Nodes, 1)
			require.Equal(t, "web", result.Nodes[0].Service.Service)
			require.True(t, result.ResultsFilteredByACLs)
		})

		t.Run("request with filter that does not match token permissions returns 0 results and ResultsFilteredByACLs equal to true", func(t *testing.T) {
			eventCh := make(chan proxycfg.UpdateEvent)
			require.NoError(t, dataSource.Notify(ctx, &structs.ServiceDumpRequest{
				QueryOptions: structs.QueryOptions{Filter: bexpNotMatchingUserTokenPermissions},
			}, "", eventCh))

			result := getEventResult[*structs.IndexedCheckServiceNodes](t, eventCh)
			require.Len(t, result.Nodes, 0)
			require.True(t, result.ResultsFilteredByACLs)
		})
	})
}

func newMockInternalServiceDump(t *testing.T) *mockInternalServiceDump {
	mock := &mockInternalServiceDump{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

type mockInternalServiceDump struct {
	mock.Mock
}

func (m *mockInternalServiceDump) Notify(ctx context.Context, req *structs.ServiceDumpRequest, correlationID string, ch chan<- proxycfg.UpdateEvent) error {
	return m.Called(ctx, req, correlationID, ch).Error(0)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package proxycfgglue

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/acl/resolver"
	"github.com/dhiaayachi/consul/agent/consul/state"
	"github.com/dhiaayachi/consul/agent/proxycfg"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/sdk/testutil"
)

func TestServerIntentions(t *testing.T) {
	nextIndex := indexGenerator()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store := state.NewStateStore(nil)

	const (
		serviceName = "web"
		index       = 1
	)
	require.NoError(t, store.SystemMetadataSet(1, &structs.SystemMetadataEntry{
		Key:   structs.SystemMetadataIntentionFormatKey,
		Value: structs.SystemMetadataIntentionFormatConfigValue,
	}))
	require.NoError(t, store.EnsureConfigEntry(nextIndex(), &structs.ServiceIntentionsConfigEntry{
		Name: serviceName,
		Sources: []*structs.SourceIntention{
			{
				Name:   "db",
				Action: structs.IntentionActionAllow,
			},
		},
	}))

	authz := policyAuthorizer(t, `
		service "web" { policy = "read" }
	`)

	logger := hclog.NewNullLogger()

	intentions := ServerIntentions(ServerDataSourceDeps{
		ACLResolver: newStaticResolver(authz),
		Logger:      logger,
		GetStore:    func() Store { return store },
	})

	eventCh := make(chan proxycfg.UpdateEvent)
	require.NoError(t, intentions.Notify(ctx, &structs.ServiceSpecificRequest{
		ServiceName:    serviceName,
		EnterpriseMeta: *acl.DefaultEnterpriseMeta(),
	}, "", eventCh))

	testutil.RunStep(t, "initial snapshot", func(t *testing.T) {
		result := getEventResult[structs.SimplifiedIntentions](t, eventCh)
		require.Len(t, result, 1)

		intention := result[0]
		require.Equal(t, intention.DestinationName, serviceName)
		require.Equal(t, intention.SourceName, "db")
	})

	testutil.RunStep(t, "updating an intention", func(t *testing.T) {
		require.NoError(t, store.EnsureConfigEntry(nextIndex(), &structs.ServiceIntentionsConfigEntry{
			Name: serviceName,
			Sources: []*structs.SourceIntention{
				{
					Name:   "api",
					Action: structs.IntentionActionAllow,
				},
				{
					Name:   "db",
					Action: structs.IntentionActionAllow,
				},
			},
		}))

		result := getEventResult[structs.SimplifiedIntentions](t, eventCh)
		require.Len(t, result, 2)

		for i, src := range []string{"api", "db"} {
			intention := result[i]
			require.Equal(t, intention.DestinationName, serviceName)
			require.Equal(t, intention.SourceName, src)
		}
	})

	testutil.RunStep(t, "publishing a delete event", func(t *testing.T) {
		require.NoError(t, store.DeleteConfigEntry(nextIndex(), structs.ServiceIntentions, serviceName, nil))

		result := getEventResult[structs.SimplifiedIntentions](t, eventCh)
		require.Len(t, result, 0)
	})
}

func TestServerIntentions_ACLDeny(t *testing.T) {
	nextIndex := indexGenerator()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store := state.NewStateStore(nil)

	const (
		serviceName = "web"
		index       = 1
	)
	require.NoError(t, store.SystemMetadataSet(1, &structs.SystemMetadataEntry{
		Key:   structs.SystemMetadataIntentionFormatKey,
		Value: structs.SystemMetadataIntentionFormatConfigValue,
	}))
	require.NoError(t, store.EnsureConfigEntry(nextIndex(), &structs.ServiceIntentionsConfigEntry{
		Name: serviceName,
		Sources: []*structs.SourceIntention{
			{
				Name:   "db",
				Action: structs.IntentionActionAllow,
			},
		},
	}))

	authz := policyAuthorizer(t, ``)

	logger := hclog.NewNullLogger()

	intentions := ServerIntentions(ServerDataSourceDeps{
		ACLResolver: newStaticResolver(authz),
		Logger:      logger,
		GetStore:    func() Store { return store },
	})

	eventCh := make(chan proxycfg.UpdateEvent)
	require.NoError(t, intentions.Notify(ctx, &structs.ServiceSpecificRequest{
		ServiceName:    serviceName,
		EnterpriseMeta: *acl.DefaultEnterpriseMeta(),
	}, "", eventCh))

	testutil.RunStep(t, "initial snapshot", func(t *testing.T) {
		result := getEventResult[structs.SimplifiedIntentions](t, eventCh)
		require.Len(t, result, 0)
	})
}

type staticResolver struct {
	mu         sync.Mutex
	authorizer acl.Authorizer
}

func newStaticResolver(authz acl.Authorizer) *staticResolver {
	resolver := new(staticResolver)
	resolver.SwapAuthorizer(authz)
	return resolver
}

func (r *staticResolver) SwapAuthorizer(authz acl.Authorizer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.authorizer = authz
}

func (r *staticResolver) ResolveTokenAndDefaultMeta(_ string, entMeta *acl.EnterpriseMeta, authzContext *acl.AuthorizerContext) (resolver.Result, error) {
	entMeta.FillAuthzContext(authzContext)

	r.mu.Lock()
	defer r.mu.Unlock()
	return resolver.Result{Authorizer: r.authorizer}, nil
}

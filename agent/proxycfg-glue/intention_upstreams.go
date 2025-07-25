// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package proxycfgglue

import (
	"context"

	"github.com/hashicorp/go-memdb"

	"github.com/dhiaayachi/consul/agent/cache"
	cachetype "github.com/dhiaayachi/consul/agent/cache-types"
	"github.com/dhiaayachi/consul/agent/consul"
	"github.com/dhiaayachi/consul/agent/consul/watch"
	"github.com/dhiaayachi/consul/agent/proxycfg"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/agent/structs/aclfilter"
)

// CacheIntentionUpstreams satisfies the proxycfg.IntentionUpstreams interface
// by sourcing upstreams for the given service, inferred from intentions, from
// the agent cache.
func CacheIntentionUpstreams(c *cache.Cache) proxycfg.IntentionUpstreams {
	return &cacheProxyDataSource[*structs.ServiceSpecificRequest]{c, cachetype.IntentionUpstreamsName}
}

// CacheIntentionUpstreamsDestination satisfies the proxycfg.IntentionUpstreams
// interface by sourcing upstreams for the given destination, inferred from
// intentions, from the agent cache.
func CacheIntentionUpstreamsDestination(c *cache.Cache) proxycfg.IntentionUpstreams {
	return &cacheProxyDataSource[*structs.ServiceSpecificRequest]{c, cachetype.IntentionUpstreamsDestinationName}
}

// ServerIntentionUpstreams satisfies the proxycfg.IntentionUpstreams interface
// by sourcing upstreams for the given service, inferred from intentions, from
// the server's state store.
func ServerIntentionUpstreams(deps ServerDataSourceDeps, defaultIntentionPolicy string) proxycfg.IntentionUpstreams {
	return serverIntentionUpstreams{deps, structs.IntentionTargetService, defaultIntentionPolicy}
}

// ServerIntentionUpstreamsDestination satisfies the proxycfg.IntentionUpstreams
// interface by sourcing upstreams for the given destination, inferred from
// intentions, from the server's state store.
func ServerIntentionUpstreamsDestination(deps ServerDataSourceDeps, defaultIntentionPolicy string) proxycfg.IntentionUpstreams {
	return serverIntentionUpstreams{deps, structs.IntentionTargetDestination, defaultIntentionPolicy}
}

type serverIntentionUpstreams struct {
	deps                   ServerDataSourceDeps
	target                 structs.IntentionTargetType
	defaultIntentionPolicy string
}

func (s serverIntentionUpstreams) Notify(ctx context.Context, req *structs.ServiceSpecificRequest, correlationID string, ch chan<- proxycfg.UpdateEvent) error {
	target := structs.NewServiceName(req.ServiceName, &req.EnterpriseMeta)

	return watch.ServerLocalNotify(ctx, correlationID, s.deps.GetStore,
		func(ws memdb.WatchSet, store Store) (uint64, *structs.IndexedServiceList, error) {
			authz, err := s.deps.ACLResolver.ResolveTokenAndDefaultMeta(req.Token, &req.EnterpriseMeta, nil)
			if err != nil {
				return 0, nil, err
			}

			defaultAllow := consul.DefaultIntentionAllow(authz, s.defaultIntentionPolicy)

			index, services, err := store.IntentionTopology(ws, target, false, defaultAllow, s.target)
			if err != nil {
				return 0, nil, err
			}

			result := &structs.IndexedServiceList{
				Services: services,
				QueryMeta: structs.QueryMeta{
					Index:   index,
					Backend: structs.QueryBackendBlocking,
				},
			}
			aclfilter.New(authz, s.deps.Logger).Filter(result)

			return index, result, nil
		},
		dispatchBlockingQueryUpdate[*structs.IndexedServiceList](ch),
	)
}

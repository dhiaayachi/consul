// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cachetype

import (
	"context"
	"fmt"

	"github.com/dhiaayachi/consul/agent/cache"
	"github.com/dhiaayachi/consul/agent/structs"
)

// Recommended name for registration.
const PeeredUpstreamsName = "peered-upstreams"

// PeeredUpstreams supports fetching imported upstream candidates of a given partition.
type PeeredUpstreams struct {
	RegisterOptionsBlockingRefresh
	RPC RPC
}

func (i *PeeredUpstreams) Fetch(opts cache.FetchOptions, req cache.Request) (cache.FetchResult, error) {
	var result cache.FetchResult

	reqReal, ok := req.(*structs.PartitionSpecificRequest)
	if !ok {
		return result, fmt.Errorf(
			"Internal cache failure: request wrong type: %T", req)
	}

	// Lightweight copy this object so that manipulating QueryOptions doesn't race.
	dup := *reqReal
	reqReal = &dup

	// Set the minimum query index to our current index so we block
	reqReal.QueryOptions.MinQueryIndex = opts.MinIndex
	reqReal.QueryOptions.MaxQueryTime = opts.Timeout

	// Always allow stale - there's no point in hitting leader if the request is
	// going to be served from cache and end up arbitrarily stale anyway. This
	// allows cached service-discover to automatically read scale across all
	// servers too.
	reqReal.AllowStale = true

	// Fetch
	var reply structs.IndexedPeeredServiceList
	if err := i.RPC.RPC(context.Background(), "Internal.PeeredUpstreams", reqReal, &reply); err != nil {
		return result, err
	}

	result.Value = &reply
	result.Index = reply.QueryMeta.Index
	return result, nil
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cachetype

import (
	"context"
	"fmt"

	"github.com/dhiaayachi/consul/agent/cache"
	"github.com/dhiaayachi/consul/agent/cacheshim"
	"github.com/dhiaayachi/consul/agent/structs"
)

// Recommended name for registration.
const ConnectCARootName = cacheshim.ConnectCARootName

// ConnectCARoot supports fetching the Connect CA roots. This is a
// straightforward cache type since it only has to block on the given
// index and return the data.
type ConnectCARoot struct {
	RegisterOptionsBlockingRefresh
	RPC RPC
}

func (c *ConnectCARoot) Fetch(opts cache.FetchOptions, req cache.Request) (cache.FetchResult, error) {
	var result cache.FetchResult

	// The request should be a DCSpecificRequest.
	reqReal, ok := req.(*structs.DCSpecificRequest)
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

	// Fetch
	var reply structs.IndexedCARoots
	if err := c.RPC.RPC(context.Background(), "ConnectCA.Roots", reqReal, &reply); err != nil {
		return result, err
	}

	result.Value = &reply
	result.Index = reply.QueryMeta.Index
	return result, nil
}

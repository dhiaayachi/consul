// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cachetype

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/dhiaayachi/consul/agent/cache"
	external "github.com/dhiaayachi/consul/agent/grpc-external"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/proto/private/pbpeering"
)

// Recommended name for registration.
const TrustBundleReadName = "peer-trust-bundle"

// TrustBundleReadRequest represents the combination of request payload
// and options that would normally be sent over headers.
type TrustBundleReadRequest struct {
	Request *pbpeering.TrustBundleReadRequest
	structs.QueryOptions
}

func (r *TrustBundleReadRequest) CacheInfo() cache.RequestInfo {
	info := cache.RequestInfo{
		Token:          r.Token,
		Datacenter:     "",
		MinIndex:       r.MinQueryIndex,
		Timeout:        r.MaxQueryTime,
		MaxAge:         r.MaxAge,
		MustRevalidate: r.MustRevalidate,
	}

	v, err := hashstructure.Hash([]interface{}{
		r.Request.Partition,
		r.Request.Name,
	}, nil)
	if err == nil {
		// If there is an error, we don't set the key. A blank key forces
		// no cache for this request so the request is forwarded directly
		// to the server.
		info.Key = strconv.FormatUint(v, 10)
	}

	return info
}

// TrustBundle supports fetching discovering service instances via prepared
// queries.
type TrustBundle struct {
	RegisterOptionsBlockingRefresh
	Client TrustBundleReader
}

//go:generate mockery --name TrustBundleReader --inpackage --filename mock_TrustBundleReader_test.go
type TrustBundleReader interface {
	TrustBundleRead(
		ctx context.Context, in *pbpeering.TrustBundleReadRequest, opts ...grpc.CallOption,
	) (*pbpeering.TrustBundleReadResponse, error)
}

func (t *TrustBundle) Fetch(opts cache.FetchOptions, req cache.Request) (cache.FetchResult, error) {
	var result cache.FetchResult

	// The request should be a TrustBundleReadRequest.
	// We do not need to make a copy of this request type like in other cache types
	// because the RequestInfo is synthetic.
	reqReal, ok := req.(*TrustBundleReadRequest)
	if !ok {
		return result, fmt.Errorf(
			"Internal cache failure: request wrong type: %T", req)
	}

	// Lightweight copy this object so that manipulating QueryOptions doesn't race.
	dup := *reqReal
	reqReal = &dup

	// Set the minimum query index to our current index, so we block
	reqReal.QueryOptions.MinQueryIndex = opts.MinIndex
	reqReal.QueryOptions.MaxQueryTime = opts.Timeout

	// Always allow stale - there's no point in hitting leader if the request is
	// going to be served from cache and end up arbitrarily stale anyway. This
	// allows cached service-discover to automatically read scale across all
	// servers too.
	reqReal.QueryOptions.SetAllowStale(true)

	// Fetch
	ctx, err := external.ContextWithQueryOptions(context.Background(), reqReal.QueryOptions)
	if err != nil {
		return result, err
	}

	var header metadata.MD
	reply, err := t.Client.TrustBundleRead(ctx, reqReal.Request, grpc.Header(&header))
	if err != nil {
		return result, err
	}

	// This first case is using the legacy index field
	// It should be removed in a future version in favor of the index from QueryMeta
	if reply.OBSOLETE_Index != 0 {
		result.Index = reply.OBSOLETE_Index
	} else {
		meta, err := external.QueryMetaFromGRPCMeta(header)
		if err != nil {
			return result, fmt.Errorf("could not convert gRPC metadata to query meta: %w", err)
		}
		result.Index = meta.GetIndex()
	}

	result.Value = reply

	return result, nil
}

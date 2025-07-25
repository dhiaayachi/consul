// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package resourcetest

import (
	"context"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/dhiaayachi/consul/internal/resource"
	"github.com/dhiaayachi/consul/internal/storage"
	"github.com/dhiaayachi/consul/proto-public/pbresource"
	"github.com/dhiaayachi/consul/sdk/testutil"
	"github.com/dhiaayachi/consul/sdk/testutil/retry"
)

type resourceBuilder struct {
	resource    *pbresource.Resource
	statuses    map[string]*pbresource.Status
	dontCleanup bool
}

func Resource(rtype *pbresource.Type, name string) *resourceBuilder {
	return &resourceBuilder{
		resource: &pbresource.Resource{
			Id: &pbresource.ID{
				Type: &pbresource.Type{
					Group:        rtype.Group,
					GroupVersion: rtype.GroupVersion,
					Kind:         rtype.Kind,
				},
				Name: name,
			},
		},
	}
}

func ResourceID(id *pbresource.ID) *resourceBuilder {
	return &resourceBuilder{
		resource: &pbresource.Resource{
			Id: id,
		},
	}
}

func (b *resourceBuilder) WithTenancy(tenant *pbresource.Tenancy) *resourceBuilder {
	b.resource.Id.Tenancy = tenant
	return b
}

func (b *resourceBuilder) WithVersion(version string) *resourceBuilder {
	b.resource.Version = version
	return b
}

func (b *resourceBuilder) WithData(t T, data protoreflect.ProtoMessage) *resourceBuilder {
	t.Helper()

	anyData, err := anypb.New(data)
	require.NoError(t, err)
	b.resource.Data = anyData
	return b
}

func (b *resourceBuilder) WithMeta(key string, value string) *resourceBuilder {
	if b.resource.Metadata == nil {
		b.resource.Metadata = make(map[string]string)
	}

	b.resource.Metadata[key] = value
	return b
}

func (b *resourceBuilder) WithOwner(id *pbresource.ID) *resourceBuilder {
	b.resource.Owner = id
	return b
}

func (b *resourceBuilder) WithStatus(key string, status *pbresource.Status) *resourceBuilder {
	if b.statuses == nil {
		b.statuses = make(map[string]*pbresource.Status)
	}
	b.statuses[key] = status
	return b
}

func (b *resourceBuilder) WithoutCleanup() *resourceBuilder {
	b.dontCleanup = true
	return b
}

func (b *resourceBuilder) WithGeneration(gen string) *resourceBuilder {
	b.resource.Generation = gen
	return b
}

func (b *resourceBuilder) Build() *pbresource.Resource {
	// clone the resource so we can add on status information
	res := proto.Clone(b.resource).(*pbresource.Resource)

	// fill in the generation if empty to make it look like
	// a real managed resource
	if res.Generation == "" {
		res.Generation = ulid.Make().String()
	}

	// Now create the status map
	if len(b.statuses) > 0 {
		res.Status = make(map[string]*pbresource.Status)
		for key, original := range b.statuses {
			status := &pbresource.Status{
				ObservedGeneration: res.Generation,
				Conditions:         original.Conditions,
			}
			res.Status[key] = status
		}
	}

	return res
}

func (b *resourceBuilder) ID() *pbresource.ID {
	return b.resource.Id
}

func (b *resourceBuilder) Reference(section string) *pbresource.Reference {
	return resource.Reference(b.ID(), section)
}

func (b *resourceBuilder) ReferenceNoSection() *pbresource.Reference {
	return resource.Reference(b.ID(), "")
}

func (b *resourceBuilder) Write(t T, client pbresource.ResourceServiceClient) *pbresource.Resource {
	t.Helper()

	var ctx context.Context
	rtestClient, ok := client.(*Client)
	if ok {
		ctx = rtestClient.Context(t)
	} else {
		ctx = testutil.TestContext(t)
		rtestClient = NewClient(client)
	}

	res := b.resource

	var rsp *pbresource.WriteResponse
	var err error

	// Retry any writes where the error is a UID mismatch and the UID was not specified. This is indicative
	// of using a follower to rewrite an object who is not perfectly in-sync with the leader.
	retry.Run(t, func(r *retry.R) {
		rsp, err = client.Write(ctx, &pbresource.WriteRequest{
			Resource: res,
		})

		if err == nil || res.Id.Uid != "" || status.Code(err) != codes.FailedPrecondition {
			if err != nil {
				t.Logf("write saw error: %v", err)
			}
			return
		}

		if strings.Contains(err.Error(), storage.ErrWrongUid.Error()) {
			r.Fatalf("resource write failed due to uid mismatch - most likely a transient issue when talking to a non-leader")
		} else {
			// other errors are unexpected and should cause an immediate failure
			r.Stop(err)
		}
	})

	require.NoError(t, err)
	require.NotNil(t, rsp)

	if !b.dontCleanup {
		id := proto.Clone(rsp.Resource.Id).(*pbresource.ID)
		id.Uid = ""
		t.Cleanup(func() {
			rtestClient.CleanupDelete(t, id)
		})
	}

	if len(b.statuses) == 0 {
		return rsp.Resource
	}

	for key, original := range b.statuses {
		status := &pbresource.Status{
			ObservedGeneration: rsp.Resource.Generation,
			Conditions:         original.Conditions,
		}
		_, err := client.WriteStatus(ctx, &pbresource.WriteStatusRequest{
			Id:     rsp.Resource.Id,
			Key:    key,
			Status: status,
		})
		require.NoError(t, err)
	}

	readResp, err := client.Read(ctx, &pbresource.ReadRequest{
		Id: rsp.Resource.Id,
	})

	require.NoError(t, err)

	return readResp.Resource
}

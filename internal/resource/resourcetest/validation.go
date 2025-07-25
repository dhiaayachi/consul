// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package resourcetest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/internal/resource"
	"github.com/dhiaayachi/consul/proto-public/pbresource"
)

func ValidateAndNormalize(t *testing.T, registry resource.Registry, res *pbresource.Resource) {
	t.Helper()
	typ := res.Id.Type

	typeInfo, ok := registry.Resolve(typ)
	if !ok {
		t.Fatalf("unhandled resource type: %q", resource.ToGVK(typ))
	}

	if typeInfo.Mutate != nil {
		require.NoError(t, typeInfo.Mutate(res), "failed to apply type mutation to resource")
	}

	if typeInfo.Validate != nil {
		require.NoError(t, typeInfo.Validate(res), "failed to validate resource")
	}
}

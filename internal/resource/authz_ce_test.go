// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package resource

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/proto-public/pbresource"
)

func TestAuthorizerContext_CE(t *testing.T) {
	t.Run("no peer", func(t *testing.T) {
		require.Equal(t,
			&acl.AuthorizerContext{},
			AuthorizerContext(&pbresource.Tenancy{
				Partition: "foo",
				Namespace: "bar",
			}),
		)
	})

	t.Run("with local peer", func(t *testing.T) {
		require.Equal(t,
			&acl.AuthorizerContext{},
			AuthorizerContext(&pbresource.Tenancy{
				Partition: "foo",
				Namespace: "bar",
			}),
		)
	})

	// TODO(peering/v2): add a test here for non-local peers
}

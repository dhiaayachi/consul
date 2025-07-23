// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package resource

import (
	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/proto-public/pbresource"
)

// AuthorizerContext builds an ACL AuthorizerContext for the given tenancy.
func AuthorizerContext(t *pbresource.Tenancy) *acl.AuthorizerContext {
	// TODO(peering/v2) handle non-local peers here
	return &acl.AuthorizerContext{}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package connect

import (
	"fmt"

	"github.com/dhiaayachi/consul/acl"
)

// GetEnterpriseMeta will synthesize an EnterpriseMeta struct from the SpiffeIDAgent.
// in CE this just returns an empty (but never nil) struct pointer
func (id SpiffeIDMeshGateway) GetEnterpriseMeta() *acl.EnterpriseMeta {
	return &acl.EnterpriseMeta{}
}

func (id SpiffeIDMeshGateway) uriPath() string {
	return fmt.Sprintf("/gateway/mesh/dc/%s", id.Datacenter)
}

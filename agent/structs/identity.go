// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package structs

import "github.com/dhiaayachi/consul/acl"

// Identity of some entity (ex: service, node, check).
//
// TODO: this type should replace ServiceID, ServiceName, and CheckID which all
// have roughly identical implementations.
type Identity struct {
	ID string
	acl.EnterpriseMeta
}

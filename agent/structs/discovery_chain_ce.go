// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package structs

import (
	"github.com/dhiaayachi/consul/acl"
)

func (t *DiscoveryTarget) GetEnterpriseMetadata() *acl.EnterpriseMeta {
	return DefaultEnterpriseMetaInDefaultPartition()
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package consul

import (
	"context"
	"fmt"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
)

func testIdentityForTokenEnterprise(string) (bool, structs.ACLIdentity, error) {
	return true, nil, acl.ErrNotFound
}

func testPolicyForIDEnterprise(string) (bool, *structs.ACLPolicy, error) {
	return true, nil, acl.ErrNotFound
}

func testRoleForIDEnterprise(string) (bool, *structs.ACLRole, error) {
	return true, nil, acl.ErrNotFound
}

// EnterpriseACLResolverTestDelegate stub
type EnterpriseACLResolverTestDelegate struct{}

// RPC stub for the EnterpriseACLResolverTestDelegate
func (d *EnterpriseACLResolverTestDelegate) RPC(context.Context, string, interface{}, interface{}) (bool, error) {
	return false, nil
}

func (d *EnterpriseACLResolverTestDelegate) UseTestLocalData(data []interface{}) {
	if len(data) > 0 {
		panic(fmt.Sprintf("unexpected data type: %T", data[0]))
	}
}
func (d *EnterpriseACLResolverTestDelegate) UseDefaultData() {}
func (d *EnterpriseACLResolverTestDelegate) Reset()          {}

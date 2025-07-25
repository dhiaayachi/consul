// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package pbcommon

import "github.com/dhiaayachi/consul/acl"

var DefaultEnterpriseMeta = &EnterpriseMeta{}

func NewEnterpriseMetaFromStructs(_ acl.EnterpriseMeta) *EnterpriseMeta {
	return &EnterpriseMeta{}
}
func EnterpriseMetaToStructs(s *EnterpriseMeta, t *acl.EnterpriseMeta) {
	if s == nil {
		return
	}
}
func EnterpriseMetaFromStructs(t *acl.EnterpriseMeta, s *EnterpriseMeta) {
	if s == nil {
		return
	}
}

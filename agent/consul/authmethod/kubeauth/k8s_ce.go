// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package kubeauth

import (
	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
)

type enterpriseConfig struct{}

func enterpriseValidation(method *structs.ACLAuthMethod, config *Config) error {
	return nil
}

func (v *Validator) k8sEntMetaFromFields(fields map[string]string) *acl.EnterpriseMeta {
	return nil
}

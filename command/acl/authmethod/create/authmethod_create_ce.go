// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package authmethodcreate

import "github.com/dhiaayachi/consul/api"

type enterpriseCmd struct {
}

func (c *cmd) initEnterpriseFlags() {}

func (c *cmd) enterprisePopulateAuthMethod(method *api.ACLAuthMethod) error {
	return nil
}

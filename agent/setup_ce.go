// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package agent

import (
	autoconf "github.com/dhiaayachi/consul/agent/auto-config"
	"github.com/dhiaayachi/consul/agent/config"
	"github.com/dhiaayachi/consul/agent/consul"
)

// initEnterpriseBaseDeps is responsible for initializing the enterprise dependencies that
// will be utilized throughout the whole Consul Agent.
func initEnterpriseBaseDeps(d BaseDeps, _ *config.RuntimeConfig) (BaseDeps, error) {
	return d, nil
}

// initEnterpriseAutoConfig is responsible for setting up auto-config for enterprise
func initEnterpriseAutoConfig(_ consul.EnterpriseDeps, _ *config.RuntimeConfig) autoconf.EnterpriseConfig {
	return autoconf.EnterpriseConfig{}
}

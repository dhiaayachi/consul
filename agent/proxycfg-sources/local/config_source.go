// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package local

import (
	"context"

	"github.com/dhiaayachi/consul/agent/grpc-external/limiter"
	"github.com/dhiaayachi/consul/agent/proxycfg"
	structs "github.com/dhiaayachi/consul/agent/structs"
)

// ConfigSource wraps a proxycfg.Manager to create watches on services
// local to the agent (pre-registered by Sync).
type ConfigSource struct {
	manager ConfigManager
}

// NewConfigSource builds a ConfigSource with the given proxycfg.Manager.
func NewConfigSource(cfgMgr ConfigManager) *ConfigSource {
	return &ConfigSource{cfgMgr}
}

func (m *ConfigSource) Watch(serviceID structs.ServiceID, nodeName string, _ string) (
	<-chan *proxycfg.ConfigSnapshot,
	limiter.SessionTerminatedChan,
	proxycfg.SrcTerminatedChan,
	context.CancelFunc,
	error,
) {
	watchCh, cancelWatch := m.manager.Watch(proxycfg.ProxyID{
		ServiceID: serviceID,
		NodeName:  nodeName,

		// Note: we *intentionally* don't set Token here. All watches on local
		// services use the same ACL token, regardless of whatever token is
		// presented in the xDS stream (the token presented to the xDS server
		// is checked before the watch is created).
		Token: "",
	})
	return watchCh, nil, nil, cancelWatch, nil
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package agent

import (
	"github.com/dhiaayachi/consul/api"
	autopilot "github.com/hashicorp/raft-autopilot"
)

func autopilotToAPIServerEnterprise(_ *autopilot.ServerState, _ *api.AutopilotServer) {
	// noop in ce
}

func autopilotToAPIStateEnterprise(state *autopilot.State, apiState *api.AutopilotState) {
	// without the enterprise features there is no different between these two and we don't want to
	// alarm anyone by leaving this as the zero value.
	apiState.OptimisticFailureTolerance = state.FailureTolerance
}

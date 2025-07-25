// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package acl

import (
	"fmt"
	"io"
	"testing"

	"github.com/dhiaayachi/consul/agent"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/testrpc"
	"github.com/stretchr/testify/require"
)

func Test_GetPolicyIDByName_Builtins(t *testing.T) {
	t.Parallel()

	a := agent.StartTestAgent(t,
		agent.TestAgent{
			LogOutput: io.Discard,
			HCL: `
				primary_datacenter = "dc1"
				acl {
					enabled = true
					tokens {
						initial_management = "root"
					}
				}
			`,
		},
	)

	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1", testrpc.WithToken("root"))

	client := a.Client()
	client.AddHeader("X-Consul-Token", "root")

	for _, policy := range structs.ACLBuiltinPolicies {
		name := fmt.Sprintf("%s policy", policy.Name)
		t.Run(name, func(t *testing.T) {
			id, err := GetPolicyIDByName(client, policy.Name)
			require.NoError(t, err)
			require.Equal(t, policy.ID, id)
		})
	}
}

func Test_GetPolicyIDByName_NotFound(t *testing.T) {
	t.Parallel()

	a := agent.StartTestAgent(t,
		agent.TestAgent{
			LogOutput: io.Discard,
			HCL: `
				primary_datacenter = "dc1"
				acl {
					enabled = true
					tokens {
						initial_management = "root"
					}
				}
			`,
		},
	)

	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1", testrpc.WithToken("root"))

	client := a.Client()
	client.AddHeader("X-Consul-Token", "root")

	id, err := GetPolicyIDByName(client, "not_found")
	require.Error(t, err)
	require.Equal(t, "", id)

}

func Test_GetPolicyIDFromPartial_Builtins(t *testing.T) {
	t.Parallel()

	a := agent.StartTestAgent(t,
		agent.TestAgent{
			LogOutput: io.Discard,
			HCL: `
				primary_datacenter = "dc1"
				acl {
					enabled = true
					tokens {
						initial_management = "root"
					}
				}
			`,
		},
	)

	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1", testrpc.WithToken("root"))

	client := a.Client()
	client.AddHeader("X-Consul-Token", "root")

	for _, policy := range structs.ACLBuiltinPolicies {
		name := fmt.Sprintf("%s policy", policy.Name)
		t.Run(name, func(t *testing.T) {
			id, err := GetPolicyIDFromPartial(client, policy.Name)
			require.NoError(t, err)
			require.Equal(t, policy.ID, id)
		})
	}
}

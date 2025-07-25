// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package consul

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	msgpackrpc "github.com/hashicorp/consul-net-rpc/net-rpc-msgpackrpc"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/serf/serf"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
	tokenStore "github.com/dhiaayachi/consul/agent/token"
	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/sdk/testutil"
	"github.com/dhiaayachi/consul/sdk/testutil/retry"
	"github.com/dhiaayachi/consul/testrpc"
)

func TestLeader_TombstoneGC_Reset(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()
	servers := []*Server{s1, s2, s3}

	// Try to join
	joinLAN(t, s2, s1)
	joinLAN(t, s3, s1)

	for _, s := range servers {
		retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 3)) })
	}

	var leader *Server
	for _, s := range servers {
		if s.IsLeader() {
			leader = s
			break
		}
	}
	if leader == nil {
		t.Fatalf("Should have a leader")
	}

	// Check that the leader has a pending GC expiration
	if !leader.tombstoneGC.PendingExpiration() {
		t.Fatalf("should have pending expiration")
	}

	// Kill the leader
	leader.Shutdown()
	time.Sleep(100 * time.Millisecond)

	// Wait for a new leader
	leader = nil
	retry.Run(t, func(r *retry.R) {
		for _, s := range servers {
			if s.IsLeader() {
				leader = s
				return
			}
		}
		r.Fatal("no leader")
	})

	retry.Run(t, func(r *retry.R) {
		if !leader.tombstoneGC.PendingExpiration() {
			r.Fatal("leader has no pending GC expiration")
		}
	})
}

func TestLeader_ReapTombstones(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
		c.ACLInitialManagementToken = "root"
		c.ACLResolverSettings.ACLDefaultPolicy = "deny"
		c.TombstoneTTL = 50 * time.Millisecond
		c.TombstoneTTLGranularity = 10 * time.Millisecond
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)

	testrpc.WaitForLeader(t, s1.RPC, "dc1")

	// Create a KV entry
	arg := structs.KVSRequest{
		Datacenter: "dc1",
		Op:         api.KVSet,
		DirEnt: structs.DirEntry{
			Key:   "test",
			Value: []byte("test"),
		},
		WriteRequest: structs.WriteRequest{
			Token: "root",
		},
	}
	var out bool
	if err := msgpackrpc.CallWithCodec(codec, "KVS.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Delete the KV entry (tombstoned).
	arg.Op = api.KVDelete
	if err := msgpackrpc.CallWithCodec(codec, "KVS.Apply", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make sure there's a tombstone.
	state := s1.fsm.State()
	retry.Run(t, func(r *retry.R) {
		snap := state.Snapshot()
		defer snap.Close()
		stones, err := snap.Tombstones()
		if err != nil {
			r.Fatalf("err: %s", err)
		}
		if stones.Next() == nil {
			r.Fatalf("missing tombstones")
		}
		if stones.Next() != nil {
			r.Fatalf("unexpected extra tombstones")
		}
	})

	// Check that the new leader has a pending GC expiration by
	// watching for the tombstone to get removed.
	retry.Run(t, func(r *retry.R) {
		snap := state.Snapshot()
		defer snap.Close()
		stones, err := snap.Tombstones()
		if err != nil {
			r.Fatal(err)
		}
		if stones.Next() != nil {
			r.Fatal("should have no tombstones")
		}
	})
}

func TestLeader_RollRaftServer(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = true
		c.Datacenter = "dc1"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
	})
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Try to join
	joinLAN(t, s2, s1)
	joinLAN(t, s3, s1)

	for _, s := range servers {
		retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 3)) })
	}

	// Kill the v1 server
	s2.Shutdown()

	for _, s := range []*Server{s1, s3} {
		retry.Run(t, func(r *retry.R) {
			// autopilot should force removal of the shutdown node
			r.Check(wantPeers(s, 2))
		})
	}

	// Replace the dead server with a new one
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()
	joinLAN(t, s4, s1)
	servers[1] = s4

	// Make sure the dead server is removed and we're back to 3 total peers
	for _, s := range servers {
		retry.Run(t, func(r *retry.R) {
			r.Check(wantPeers(s, 3))
		})
	}
}

func TestLeader_ChangeServerID(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	conf := func(c *Config) {
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 3
	}
	dir1, s1 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Try to join and wait for all servers to get promoted
	joinLAN(t, s2, s1)
	joinLAN(t, s3, s1)
	for _, s := range servers {
		testrpc.WaitForTestAgent(t, s.RPC, "dc1")
		retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 3)) })
	}

	// Shut down a server, freeing up its address/port
	s3.Shutdown()

	retry.Run(t, func(r *retry.R) {
		alive := 0
		for _, m := range s1.LANMembersInAgentPartition() {
			if m.Status == serf.StatusAlive {
				alive++
			}
		}
		if got, want := alive, 2; got != want {
			r.Fatalf("got %d alive members want %d", got, want)
		}
	})

	// Bring up a new server with s3's address that will get a different ID
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.Datacenter = "dc1"
		c.RaftConfig.ProtocolVersion = 3
		c.SerfLANConfig.MemberlistConfig = s3.config.SerfLANConfig.MemberlistConfig
		c.RPCAddr = s3.config.RPCAddr
		c.RPCAdvertise = s3.config.RPCAdvertise
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()

	joinLAN(t, s4, s1)
	testrpc.WaitForLeader(t, s4.RPC, "dc1")
	servers[2] = s4

	// While integrating #3327 it uncovered that this test was flaky. The
	// connection pool would use the same TCP connection to the old server
	// which would give EOF errors to the autopilot health check RPC call.
	// To make this more reliable we changed the connection pool to throw
	// away the connection if it sees an EOF error, since there's no way
	// that connection is going to work again. This made this test reliable
	// since it will make a new connection to s4.
	retry.Run(t, func(r *retry.R) {
		r.Check(wantRaft(servers))
		for _, s := range servers {
			// Make sure the dead server is removed and we're back below 4
			r.Check(wantPeers(s, 3))
		}
	})
}

func TestLeader_ChangeNodeID(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	dir2, s2 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerDCBootstrap(t, "dc1", false)
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	servers := []*Server{s1, s2, s3}

	// Try to join and wait for all servers to get promoted
	joinLAN(t, s2, s1)
	joinLAN(t, s3, s1)
	for _, s := range servers {
		testrpc.WaitForTestAgent(t, s.RPC, "dc1")
		retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 3)) })
	}

	// Shut down a server, freeing up its address/port
	s3.Shutdown()
	// wait for s1.LANMembersInAgentPartition() to show s3 as StatusFailed or StatusLeft on
	retry.Run(t, func(r *retry.R) {
		var gone bool
		for _, m := range s1.LANMembersInAgentPartition() {
			if m.Name == s3.config.NodeName && (m.Status == serf.StatusFailed || m.Status == serf.StatusLeft) {
				gone = true
			}
		}
		require.True(r, gone, "s3 has not been detected as failed or left after shutdown")
	})

	// Bring up a new server with s3's name that will get a different ID
	dir4, s4 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = false
		c.Datacenter = "dc1"
		c.NodeName = s3.config.NodeName
	})
	defer os.RemoveAll(dir4)
	defer s4.Shutdown()
	joinLAN(t, s4, s1)
	servers[2] = s4

	// Make sure the dead server is gone from both Raft and Serf and we're back to 3 total peers
	retry.Run(t, func(r *retry.R) {
		r.Check(wantRaft(servers))
		for _, s := range servers {
			r.Check(wantPeers(s, 3))
		}
	})

	retry.Run(t, func(r *retry.R) {
		for _, m := range s1.LANMembersInAgentPartition() {
			require.Equal(r, serf.StatusAlive, m.Status)
		}
	})
}

func TestLeader_ACL_Initialization(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	tests := []struct {
		name              string
		initialManagement string
		hcpManagement     string

		// canBootstrap tracks whether the ACL system can be bootstrapped
		// after the leader initializes ACLs. Bootstrapping is the act
		// of persisting a token with the Global Management policy.
		canBootstrap bool
	}{
		{
			name:              "bootstrap from initial management",
			initialManagement: "c9ad785a-420d-470d-9b4d-6d9f084bfa87",
			hcpManagement:     "",
			canBootstrap:      false,
		},
		{
			name:              "bootstrap from hcp management",
			initialManagement: "",
			hcpManagement:     "924bc0e1-a41b-4f3a-b5e8-0899502fc50e",
			canBootstrap:      false,
		},
		{
			name:              "bootstrap with both",
			initialManagement: "c9ad785a-420d-470d-9b4d-6d9f084bfa87",
			hcpManagement:     "924bc0e1-a41b-4f3a-b5e8-0899502fc50e",
			canBootstrap:      false,
		},
		{
			name:              "did not bootstrap",
			initialManagement: "",
			hcpManagement:     "",
			canBootstrap:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := func(c *Config) {
				c.Bootstrap = true
				c.Datacenter = "dc1"
				c.PrimaryDatacenter = "dc1"
				c.ACLsEnabled = true
				c.ACLInitialManagementToken = tt.initialManagement
				c.Cloud.ManagementToken = tt.hcpManagement
			}
			_, s1 := testServerWithConfig(t, conf)
			testrpc.WaitForTestAgent(t, s1.RPC, "dc1")

			// check that the builtin policies were created
			for _, builtinPolicy := range structs.ACLBuiltinPolicies {
				_, policy, err := s1.fsm.State().ACLPolicyGetByID(nil, builtinPolicy.ID, nil)
				require.NoError(t, err)
				require.NotNil(t, policy)
			}

			if tt.initialManagement != "" {
				_, initialManagement, err := s1.fsm.State().ACLTokenGetBySecret(nil, tt.initialManagement, nil)
				require.NoError(t, err)
				require.NotNil(t, initialManagement)
				require.Equal(t, tt.initialManagement, initialManagement.SecretID)
			}

			if tt.hcpManagement != "" {
				_, hcpManagement, err := s1.fsm.State().ACLTokenGetBySecret(nil, tt.hcpManagement, nil)
				require.NoError(t, err)
				require.NotNil(t, hcpManagement)
				require.Equal(t, tt.hcpManagement, hcpManagement.SecretID)
			}

			canBootstrap, _, err := s1.fsm.State().CanBootstrapACLToken()
			require.NoError(t, err)
			require.Equal(t, tt.canBootstrap, canBootstrap)

			_, anon, err := s1.fsm.State().ACLTokenGetBySecret(nil, anonymousToken, nil)
			require.NoError(t, err)
			require.NotNil(t, anon)

			serverToken, err := s1.GetSystemMetadata(structs.ServerManagementTokenAccessorID)
			require.NoError(t, err)
			require.NotEmpty(t, serverToken)

			_, err = uuid.ParseUUID(serverToken)
			require.NoError(t, err)
		})
	}
}

func TestLeader_ACL_Initialization_SecondaryDC(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = true
		c.Datacenter = "dc1"
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	testrpc.WaitForTestAgent(t, s1.RPC, "dc1")

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = true
		c.Datacenter = "dc2"
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()
	testrpc.WaitForTestAgent(t, s2.RPC, "dc2")

	// Check dc1's management token
	serverToken1, err := s1.GetSystemMetadata(structs.ServerManagementTokenAccessorID)
	require.NoError(t, err)
	require.NotEmpty(t, serverToken1)
	_, err = uuid.ParseUUID(serverToken1)
	require.NoError(t, err)

	// Check dc2's management token
	serverToken2, err := s2.GetSystemMetadata(structs.ServerManagementTokenAccessorID)
	require.NoError(t, err)
	require.NotEmpty(t, serverToken2)
	_, err = uuid.ParseUUID(serverToken2)
	require.NoError(t, err)

	// Ensure the tokens were not replicated between clusters.
	require.NotEqual(t, serverToken1, serverToken2)
}

func TestLeader_ACLUpgrade_IsStickyEvenIfSerfTagsRegress(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	// We test this by having two datacenters with one server each. They
	// initially come up and complete the migration, then we power them both
	// off. We leave the primary off permanently, and then we stand up the
	// secondary. Hopefully it should transition to ENABLED instead of being
	// stuck in LEGACY.

	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc1"
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
		c.ACLInitialManagementToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	waitForLeaderEstablishment(t, s1)

	dir2, s2 := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
		c.ACLTokenReplication = false
		c.ACLReplicationRate = 100
		c.ACLReplicationBurst = 100
		c.ACLReplicationApplyLimit = 1000000
	})
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()
	codec2 := rpcClient(t, s2)
	defer codec2.Close()

	s2.tokens.UpdateReplicationToken("root", tokenStore.TokenSourceConfig)

	testrpc.WaitForLeader(t, s2.RPC, "dc2")
	waitForLeaderEstablishment(t, s2)

	// Create the WAN link
	joinWAN(t, s2, s1)
	waitForLeaderEstablishment(t, s1)
	waitForLeaderEstablishment(t, s2)
	waitForNewACLReplication(t, s2, structs.ACLReplicatePolicies, 1, 0, 0)

	// Everybody has the builtin policies.
	retry.Run(t, func(r *retry.R) {
		for _, builtinPolicy := range structs.ACLBuiltinPolicies {
			_, policy1, err := s1.fsm.State().ACLPolicyGetByID(nil, builtinPolicy.ID, structs.DefaultEnterpriseMetaInDefaultPartition())
			require.NoError(r, err)
			require.NotNil(r, policy1)

			_, policy2, err := s2.fsm.State().ACLPolicyGetByID(nil, builtinPolicy.ID, structs.DefaultEnterpriseMetaInDefaultPartition())
			require.NoError(r, err)
			require.NotNil(r, policy2)
		}
	})

	// Shutdown s1 and s2.
	s1.Shutdown()
	s2.Shutdown()

	// Restart just s2

	dir2new, s2new := testServerWithConfig(t, func(c *Config) {
		c.Datacenter = "dc2"
		c.PrimaryDatacenter = "dc1"
		c.ACLsEnabled = true
		c.ACLTokenReplication = false
		c.ACLReplicationRate = 100
		c.ACLReplicationBurst = 100
		c.ACLReplicationApplyLimit = 1000000

		c.DataDir = s2.config.DataDir
		c.NodeName = s2.config.NodeName
		c.NodeID = s2.config.NodeID
	})
	defer os.RemoveAll(dir2new)
	defer s2new.Shutdown()

	waitForLeaderEstablishment(t, s2new)
}

func TestLeader_ConfigEntryBootstrap(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	global_entry_init := &structs.ProxyConfigEntry{
		Kind: structs.ProxyDefaults,
		Name: structs.ProxyConfigGlobal,
		Config: map[string]interface{}{
			"foo": "bar",
			"bar": int64(1),
		},
	}

	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Build = "1.5.0"
		c.ConfigEntryBootstrap = []structs.ConfigEntry{
			global_entry_init,
		}
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	testrpc.WaitForTestAgent(t, s1.RPC, "dc1")

	retry.Run(t, func(t *retry.R) {
		_, entry, err := s1.fsm.State().ConfigEntry(nil, structs.ProxyDefaults, structs.ProxyConfigGlobal, structs.DefaultEnterpriseMetaInDefaultPartition())
		require.NoError(t, err)
		require.NotNil(t, entry)
		global, ok := entry.(*structs.ProxyConfigEntry)
		require.True(t, ok)
		require.Equal(t, global_entry_init.Kind, global.Kind)
		require.Equal(t, global_entry_init.Name, global.Name)
		require.Equal(t, global_entry_init.Config, global.Config)
	})
}

func TestLeader_ConfigEntryBootstrap_Fail(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	type testcase struct {
		name          string
		entries       []structs.ConfigEntry
		serverCB      func(c *Config)
		expectMessage string
	}

	cases := []testcase{
		{
			name: "service-splitter without L7 protocol",
			entries: []structs.ConfigEntry{
				&structs.ServiceSplitterConfigEntry{
					Kind: structs.ServiceSplitter,
					Name: "web",
					Splits: []structs.ServiceSplit{
						{Weight: 100, Service: "web"},
					},
				},
			},
			expectMessage: `Failed to apply configuration entry "service-splitter" / "web": discovery chain "web" uses a protocol "tcp" that does not permit advanced routing or splitting behavior`,
		},
		{
			name: "service-intentions without migration",
			entries: []structs.ConfigEntry{
				&structs.ServiceIntentionsConfigEntry{
					Kind: structs.ServiceIntentions,
					Name: "web",
					Sources: []*structs.SourceIntention{
						{
							Name:   "debug",
							Action: structs.IntentionActionAllow,
						},
					},
				},
			},
			serverCB: func(c *Config) {
				c.OverrideInitialSerfTags = func(tags map[string]string) {
					tags["ft_si"] = "0"
				}
			},
			expectMessage: `Refusing to apply configuration entry "service-intentions" / "web" because intentions are still being migrated to config entries`,
		},
		{
			name: "service-intentions without Connect",
			entries: []structs.ConfigEntry{
				&structs.ServiceIntentionsConfigEntry{
					Kind: structs.ServiceIntentions,
					Name: "web",
					Sources: []*structs.SourceIntention{
						{
							Name:   "debug",
							Action: structs.IntentionActionAllow,
						},
					},
				},
			},
			serverCB: func(c *Config) {
				c.ConnectEnabled = false
			},
			expectMessage: `Refusing to apply configuration entry "service-intentions" / "web" because Connect must be enabled to bootstrap intentions`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pr, pw := io.Pipe()
			defer pw.Close()

			var (
				ch             = make(chan string, 1)
				applyErrorLine string
			)
			go func() {
				defer pr.Close()
				scan := bufio.NewScanner(pr)
				for scan.Scan() {
					line := scan.Text()
					lineJson := map[string]interface{}{}
					json.Unmarshal([]byte(line), &lineJson)

					if strings.Contains(line, "failed to establish leadership") {
						applyErrorLine = lineJson["error"].(string)
						ch <- ""
						return
					}
					if strings.Contains(line, "successfully established leadership") {
						ch <- "leadership should not have gotten here if config entries properly failed"
						return
					}
				}

				if scan.Err() != nil {
					ch <- fmt.Sprintf("ERROR: %v", scan.Err())
				} else {
					ch <- "should not get here"
				}
			}()

			_, config := testServerConfig(t)
			config.Build = "1.6.0"
			config.ConfigEntryBootstrap = tc.entries
			if tc.serverCB != nil {
				tc.serverCB(config)
			}

			logger := hclog.NewInterceptLogger(&hclog.LoggerOptions{
				Name:       config.NodeName,
				Level:      testutil.TestLogLevel,
				Output:     io.MultiWriter(pw, testutil.NewLogBuffer(t)),
				JSONFormat: true,
			})

			deps := newDefaultDeps(t, config)
			deps.Logger = logger

			srv, err := NewServer(config, deps, grpc.NewServer(), nil, logger)
			require.NoError(t, err)
			defer srv.Shutdown()

			select {
			case result := <-ch:
				require.Empty(t, result)
				if tc.expectMessage != "" {
					require.Contains(t, applyErrorLine, tc.expectMessage)
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for a result from tailing logs")
			}
		})
	}
}

func TestDatacenterSupportsFederationStates(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	addGateway := func(t *testing.T, srv *Server, dc, node string) {
		t.Helper()
		arg := structs.RegisterRequest{
			Datacenter: dc,
			Node:       node,
			Address:    "127.0.0.1",
			Service: &structs.NodeService{
				Kind:    structs.ServiceKindMeshGateway,
				ID:      "mesh-gateway",
				Service: "mesh-gateway",
				Port:    8080,
			},
		}

		var out struct{}
		require.NoError(t, srv.RPC(context.Background(), "Catalog.Register", &arg, &out))
	}

	t.Run("one node primary with old version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		updateSerfTags(s1, "ft_fs", "0")

		waitForLeaderEstablishment(t, s1)

		addGateway(t, s1, "dc1", "node1")

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 shouldn't activate fedstates")
			}
		})
	})

	t.Run("one node primary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		addGateway(t, s1, "dc1", "node1")

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 didn't activate fedstates")
			}
		})

		// Wait until after AE runs at least once.
		retry.Run(t, func(r *retry.R) {
			arg := structs.FederationStateQuery{
				Datacenter:       "dc1",
				TargetDatacenter: "dc1",
			}

			var out structs.FederationStateResponse
			require.NoError(r, s1.RPC(context.Background(), "FederationState.Get", &arg, &out))
			require.NotNil(r, out.State)
			require.Len(r, out.State.MeshGateways, 1)
		})
	})

	t.Run("two node primary with mixed versions", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		updateSerfTags(s1, "ft_fs", "0")

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.Bootstrap = false
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		// Put s1 last so we don't trigger a leader election.
		servers := []*Server{s2, s1}

		// Try to join
		joinLAN(t, s2, s1)
		for _, s := range servers {
			retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 2)) })
		}

		waitForLeaderEstablishment(t, s1)

		addGateway(t, s1, "dc1", "node1")

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 shouldn't activate fedstates")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if s2.DatacenterSupportsFederationStates() {
				r.Fatal("server 2 shouldn't activate fedstates")
			}
		})
	})

	t.Run("two node primary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.Bootstrap = false
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		// Put s1 last so we don't trigger a leader election.
		servers := []*Server{s2, s1}

		// Try to join
		joinLAN(t, s2, s1)
		for _, s := range servers {
			retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 2)) })
		}

		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s2.RPC, "dc1")

		addGateway(t, s1, "dc1", "node1")

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 didn't activate fedstates")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if !s2.DatacenterSupportsFederationStates() {
				r.Fatal("server 2 didn't activate fedstates")
			}
		})

		// Wait until after AE runs at least once.
		retry.Run(t, func(r *retry.R) {
			arg := structs.DCSpecificRequest{
				Datacenter: "dc1",
			}

			var out structs.IndexedFederationStates
			require.NoError(r, s1.RPC(context.Background(), "FederationState.List", &arg, &out))
			require.Len(r, out.States, 1)
			require.Len(r, out.States[0].MeshGateways, 1)
		})
	})

	t.Run("primary and secondary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc2"
			c.PrimaryDatacenter = "dc1"
			c.FederationStateReplicationRate = 100
			c.FederationStateReplicationBurst = 100
			c.FederationStateReplicationApplyLimit = 1000000
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		waitForLeaderEstablishment(t, s2)

		// Try to join
		joinWAN(t, s2, s1)
		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s1.RPC, "dc2")

		addGateway(t, s1, "dc1", "node1")
		addGateway(t, s2, "dc2", "node2")

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 didn't activate fedstates")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if !s2.DatacenterSupportsFederationStates() {
				r.Fatal("server 2 didn't activate fedstates")
			}
		})

		// Wait until after AE runs at least once for both.
		retry.Run(t, func(r *retry.R) {
			arg := structs.DCSpecificRequest{
				Datacenter: "dc1",
			}

			var out structs.IndexedFederationStates
			require.NoError(r, s1.RPC(context.Background(), "FederationState.List", &arg, &out))
			require.Len(r, out.States, 2)
			require.Len(r, out.States[0].MeshGateways, 1)
			require.Len(r, out.States[1].MeshGateways, 1)
		})

		// Wait until after replication runs for the secondary.
		retry.Run(t, func(r *retry.R) {
			arg := structs.DCSpecificRequest{
				Datacenter: "dc2",
			}

			var out structs.IndexedFederationStates
			require.NoError(r, s1.RPC(context.Background(), "FederationState.List", &arg, &out))
			require.Len(r, out.States, 2)
			require.Len(r, out.States[0].MeshGateways, 1)
			require.Len(r, out.States[1].MeshGateways, 1)
		})
	})

	t.Run("primary and secondary with mixed versions", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		updateSerfTags(s1, "ft_fs", "0")

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc2"
			c.PrimaryDatacenter = "dc1"
			c.FederationStateReplicationRate = 100
			c.FederationStateReplicationBurst = 100
			c.FederationStateReplicationApplyLimit = 1000000
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		waitForLeaderEstablishment(t, s2)

		// Try to join
		joinWAN(t, s2, s1)
		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s1.RPC, "dc2")

		addGateway(t, s1, "dc1", "node1")
		addGateway(t, s2, "dc2", "node2")

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsFederationStates() {
				r.Fatal("server 1 shouldn't activate fedstates")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if s2.DatacenterSupportsFederationStates() {
				r.Fatal("server 2 shouldn't activate fedstates")
			}
		})
	})
}

func TestDatacenterSupportsIntentionsAsConfigEntries(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	addLegacyIntention := func(srv *Server, dc, src, dest string, allow bool) error {
		ixn := &structs.Intention{
			SourceNS:        structs.IntentionDefaultNamespace,
			SourceName:      src,
			DestinationNS:   structs.IntentionDefaultNamespace,
			DestinationName: dest,
			SourceType:      structs.IntentionSourceConsul,
			Meta:            map[string]string{},
		}

		if allow {
			ixn.Action = structs.IntentionActionAllow
		} else {
			ixn.Action = structs.IntentionActionDeny
		}

		//nolint:staticcheck
		ixn.UpdatePrecedence()
		//nolint:staticcheck
		ixn.SetHash()

		arg := structs.IntentionRequest{
			Datacenter: dc,
			Op:         structs.IntentionOpCreate,
			Intention:  ixn,
		}

		var id string
		return srv.RPC(context.Background(), "Intention.Apply", &arg, &id)
	}

	getConfigEntry := func(srv *Server, dc, kind, name string) (structs.ConfigEntry, error) {
		arg := structs.ConfigEntryQuery{
			Datacenter: dc,
			Kind:       kind,
			Name:       name,
		}
		var reply structs.ConfigEntryResponse
		if err := srv.RPC(context.Background(), "ConfigEntry.Get", &arg, &reply); err != nil {
			return nil, err
		}
		reply.Entry.SetHash(0)
		return reply.Entry, nil
	}

	disableServiceIntentions := func(tags map[string]string) {
		tags["ft_si"] = "0"
	}

	defaultEntMeta := structs.DefaultEnterpriseMetaInDefaultPartition()

	t.Run("one node primary with old version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.OverrideInitialSerfTags = disableServiceIntentions
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 shouldn't activate service-intentions")
			}
		})

		testutil.RequireErrorContains(t,
			addLegacyIntention(s1, "dc1", "web", "api", true),
			ErrIntentionsNotUpgradedYet.Error(),
		)
	})

	t.Run("one node primary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 didn't activate service-intentions")
			}
		})

		// try to write a using the legacy API and it should work
		require.NoError(t, addLegacyIntention(s1, "dc1", "web", "api", true))

		// read it back as a config entry and that should work too
		raw, err := getConfigEntry(s1, "dc1", structs.ServiceIntentions, "api")
		require.NoError(t, err)
		require.NotNil(t, raw)

		got, ok := raw.(*structs.ServiceIntentionsConfigEntry)
		require.True(t, ok)
		require.Len(t, got.Sources, 1)

		expect := &structs.ServiceIntentionsConfigEntry{
			Kind:           structs.ServiceIntentions,
			Name:           "api",
			EnterpriseMeta: *defaultEntMeta,

			Sources: []*structs.SourceIntention{
				{
					Name:           "web",
					EnterpriseMeta: *defaultEntMeta,
					Action:         structs.IntentionActionAllow,
					Type:           structs.IntentionSourceConsul,
					Precedence:     9,
					LegacyMeta:     map[string]string{},
					LegacyID:       got.Sources[0].LegacyID,
					// steal
					LegacyCreateTime: got.Sources[0].LegacyCreateTime,
					LegacyUpdateTime: got.Sources[0].LegacyUpdateTime,
				},
			},

			RaftIndex: got.RaftIndex,
		}
		got.Hash = 0
		require.Equal(t, expect, got)
	})

	t.Run("two node primary with mixed versions", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.OverrideInitialSerfTags = disableServiceIntentions
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.Bootstrap = false
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		// Put s1 last so we don't trigger a leader election.
		servers := []*Server{s2, s1}

		// Try to join
		joinLAN(t, s2, s1)
		for _, s := range servers {
			retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 2)) })
		}

		waitForLeaderEstablishment(t, s1)

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 shouldn't activate service-intentions")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if s2.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 2 shouldn't activate service-intentions")
			}
		})

		testutil.RequireErrorContains(t,
			addLegacyIntention(s1, "dc1", "web", "api", true),
			ErrIntentionsNotUpgradedYet.Error(),
		)
		testutil.RequireErrorContains(t,
			addLegacyIntention(s2, "dc1", "web", "api", true),
			ErrIntentionsNotUpgradedYet.Error(),
		)
	})

	t.Run("two node primary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.Bootstrap = false
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		// Put s1 last so we don't trigger a leader election.
		servers := []*Server{s2, s1}

		// Try to join
		joinLAN(t, s2, s1)
		for _, s := range servers {
			retry.Run(t, func(r *retry.R) { r.Check(wantPeers(s, 2)) })
		}

		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s2.RPC, "dc1")

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 didn't activate service-intentions")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if !s2.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 2 didn't activate service-intentions")
			}
		})

		// try to write a using the legacy API and it should work from both sides
		require.NoError(t, addLegacyIntention(s1, "dc1", "web", "api", true))
		require.NoError(t, addLegacyIntention(s2, "dc1", "web2", "api", true))

		// read it back as a config entry and that should work too
		raw, err := getConfigEntry(s1, "dc1", structs.ServiceIntentions, "api")
		require.NoError(t, err)
		require.NotNil(t, raw)

		raw, err = getConfigEntry(s2, "dc1", structs.ServiceIntentions, "api")
		require.NoError(t, err)
		require.NotNil(t, raw)
	})

	t.Run("primary and secondary with new version", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc2"
			c.PrimaryDatacenter = "dc1"
			c.ConfigReplicationRate = 100
			c.ConfigReplicationBurst = 100
			c.ConfigReplicationApplyLimit = 1000000
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		waitForLeaderEstablishment(t, s2)

		// Try to join
		joinWAN(t, s2, s1)
		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s1.RPC, "dc2")

		retry.Run(t, func(r *retry.R) {
			if !s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 didn't activate service-intentions")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if !s2.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 2 didn't activate service-intentions")
			}
		})

		// try to write a using the legacy API
		require.NoError(t, addLegacyIntention(s1, "dc1", "web", "api", true))

		// read it back as a config entry and that should work too
		raw, err := getConfigEntry(s1, "dc1", structs.ServiceIntentions, "api")
		require.NoError(t, err)
		require.NotNil(t, raw)

		// Wait until after replication runs for the secondary.
		retry.Run(t, func(r *retry.R) {
			raw, err = getConfigEntry(s2, "dc1", structs.ServiceIntentions, "api")
			require.NoError(r, err)
			require.NotNil(r, raw)
		})
	})

	t.Run("primary and secondary with mixed versions", func(t *testing.T) {
		dir1, s1 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node1"
			c.Datacenter = "dc1"
			c.PrimaryDatacenter = "dc1"
			c.OverrideInitialSerfTags = disableServiceIntentions
		})
		defer os.RemoveAll(dir1)
		defer s1.Shutdown()

		waitForLeaderEstablishment(t, s1)

		dir2, s2 := testServerWithConfig(t, func(c *Config) {
			c.NodeName = "node2"
			c.Datacenter = "dc2"
			c.PrimaryDatacenter = "dc1"
			c.ConfigReplicationRate = 100
			c.ConfigReplicationBurst = 100
			c.ConfigReplicationApplyLimit = 1000000
		})
		defer os.RemoveAll(dir2)
		defer s2.Shutdown()

		waitForLeaderEstablishment(t, s2)

		// Try to join
		joinWAN(t, s2, s1)
		testrpc.WaitForLeader(t, s1.RPC, "dc1")
		testrpc.WaitForLeader(t, s1.RPC, "dc2")

		retry.Run(t, func(r *retry.R) {
			if s1.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 1 shouldn't activate service-intentions")
			}
		})
		retry.Run(t, func(r *retry.R) {
			if s2.DatacenterSupportsIntentionsAsConfigEntries() {
				r.Fatal("server 2 shouldn't activate service-intentions")
			}
		})

		testutil.RequireErrorContains(t,
			addLegacyIntention(s1, "dc1", "web", "api", true),
			ErrIntentionsNotUpgradedYet.Error(),
		)
		testutil.RequireErrorContains(t,
			addLegacyIntention(s2, "dc1", "web", "api", true),
			ErrIntentionsNotUpgradedYet.Error(),
		)
	})
}

func TestLeader_EnableVirtualIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	orig := virtualIPVersionCheckInterval
	virtualIPVersionCheckInterval = 50 * time.Millisecond
	t.Cleanup(func() { virtualIPVersionCheckInterval = orig })

	conf := func(c *Config) {
		c.Bootstrap = false
		c.BootstrapExpect = 3
		c.Datacenter = "dc1"
		c.Build = "1.11.2"
	}
	dir1, s1 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	dir2, s2 := testServerWithConfig(t, conf)
	defer os.RemoveAll(dir2)
	defer s2.Shutdown()

	dir3, s3 := testServerWithConfig(t, func(c *Config) {
		conf(c)
		c.Build = "1.10.0"
	})
	defer os.RemoveAll(dir3)
	defer s3.Shutdown()

	// Try to join and wait for all servers to get promoted
	joinLAN(t, s2, s1)
	joinLAN(t, s3, s1)
	testrpc.WaitForLeader(t, s1.RPC, "dc1")

	// Should have nothing stored.
	state := s1.fsm.State()
	_, entry, err := state.SystemMetadataGet(nil, structs.SystemMetadataVirtualIPsEnabled)
	require.NoError(t, err)
	require.Nil(t, entry)
	state = s1.fsm.State()
	_, entry, err = state.SystemMetadataGet(nil, structs.SystemMetadataTermGatewayVirtualIPsEnabled)
	require.NoError(t, err)
	require.Nil(t, entry)

	// Register a connect-native service and make sure we don't have a virtual IP yet.
	err = state.EnsureRegistration(10, &structs.RegisterRequest{
		Node:    "foo",
		Address: "127.0.0.1",
		Service: &structs.NodeService{
			Service: "api",
			Connect: structs.ServiceConnect{
				Native: true,
			},
		},
	})
	require.NoError(t, err)

	psn := structs.PeeredServiceName{ServiceName: structs.NewServiceName("api", nil)}
	vip, err := state.VirtualIPForService(psn)
	require.NoError(t, err)
	require.Equal(t, "", vip)

	// Register a terminating gateway.
	err = state.EnsureRegistration(11, &structs.RegisterRequest{
		Node:    "bar",
		Address: "127.0.0.2",
		Service: &structs.NodeService{
			Service: "tgate1",
			ID:      "tgate1",
			Kind:    structs.ServiceKindTerminatingGateway,
		},
	})
	require.NoError(t, err)

	err = state.EnsureConfigEntry(12, &structs.TerminatingGatewayConfigEntry{
		Kind: structs.TerminatingGateway,
		Name: "tgate1",
		Services: []structs.LinkedService{
			{
				Name: "bar",
			},
		},
	})
	require.NoError(t, err)

	// Make sure the service referenced in the terminating gateway config doesn't have
	// a virtual IP yet.
	psn = structs.PeeredServiceName{ServiceName: structs.NewServiceName("bar", nil)}
	vip, err = state.VirtualIPForService(psn)
	require.NoError(t, err)
	require.Equal(t, "", vip)

	// Leave s3 and wait for the version to get updated.
	require.NoError(t, s3.Leave())
	retry.Run(t, func(r *retry.R) {
		_, entry, err := state.SystemMetadataGet(nil, structs.SystemMetadataVirtualIPsEnabled)
		require.NoError(r, err)
		require.NotNil(r, entry)
		require.Equal(r, "true", entry.Value)
		_, entry, err = state.SystemMetadataGet(nil, structs.SystemMetadataTermGatewayVirtualIPsEnabled)
		require.NoError(r, err)
		require.NotNil(r, entry)
		require.Equal(r, "true", entry.Value)
	})

	// Update the connect-native service - now there should be a virtual IP assigned.
	err = state.EnsureRegistration(20, &structs.RegisterRequest{
		Node:    "foo",
		Address: "127.0.0.2",
		Service: &structs.NodeService{
			Service: "api",
			Connect: structs.ServiceConnect{
				Native: true,
			},
		},
	})
	require.NoError(t, err)
	psn = structs.PeeredServiceName{ServiceName: structs.NewServiceName("api", nil)}
	vip, err = state.VirtualIPForService(psn)
	require.NoError(t, err)
	require.Equal(t, "240.0.0.1", vip)

	// Update the terminating gateway config entry - now there should be a virtual IP assigned.
	err = state.EnsureConfigEntry(21, &structs.TerminatingGatewayConfigEntry{
		Kind: structs.TerminatingGateway,
		Name: "tgate1",
		Services: []structs.LinkedService{
			{
				Name: "api",
			},
			{
				Name: "baz",
			},
		},
	})
	require.NoError(t, err)

	_, node, err := state.NodeService(nil, "bar", "tgate1", nil, "")
	require.NoError(t, err)
	sn := structs.ServiceName{Name: "api"}
	key := structs.ServiceGatewayVirtualIPTag(sn)
	require.Contains(t, node.TaggedAddresses, key)
	require.Equal(t, node.TaggedAddresses[key].Address, "240.0.0.1")

	// Make sure the baz service (only referenced in the config entry so far)
	// has a virtual IP.
	psn = structs.PeeredServiceName{ServiceName: structs.NewServiceName("baz", nil)}
	vip, err = state.VirtualIPForService(psn)
	require.NoError(t, err)
	require.Equal(t, "240.0.0.2", vip)
}

func TestLeader_ACL_Initialization_AnonymousToken(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	dir1, s1 := testServerWithConfig(t, func(c *Config) {
		c.Bootstrap = true
		c.Datacenter = "dc1"
		c.ACLsEnabled = true
		c.ACLInitialManagementToken = "root"
	})
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()

	codec := rpcClient(t, s1)
	defer codec.Close()

	testrpc.WaitForTestAgent(t, s1.RPC, "dc1")

	retry.Run(t, func(r *retry.R) {
		_, anon, err := s1.fsm.State().ACLTokenGetBySecret(nil, anonymousToken, nil)
		require.NoError(r, err)
		require.NotNil(r, anon)
		require.Len(r, anon.Policies, 0)
	})

	reqToken := structs.ACLTokenSetRequest{
		Datacenter: "dc1",
		ACLToken: structs.ACLToken{
			AccessorID:  acl.AnonymousTokenID,
			SecretID:    anonymousToken,
			Description: "Anonymous Token",
			CreateTime:  time.Now(),
			Policies: []structs.ACLTokenPolicyLink{
				{
					ID: structs.ACLPolicyGlobalManagementID,
				},
			},
			EnterpriseMeta: *structs.DefaultEnterpriseMetaInDefaultPartition(),
		},
		WriteRequest: structs.WriteRequest{Token: "root"},
	}
	var respToken structs.ACLToken
	require.NoError(t, msgpackrpc.CallWithCodec(codec, "ACL.TokenSet", &reqToken, &respToken))

	// Restart the server to re-initialize ACLs when establishing leadership
	require.NoError(t, s1.Shutdown())
	dir2, newS1 := testServerWithConfig(t, func(c *Config) {
		// Keep existing data dir and node info since it's a restart
		c.DataDir = s1.config.DataDir
		c.NodeName = s1.config.NodeName
		c.NodeID = s1.config.NodeID
		c.Bootstrap = true
		c.Datacenter = "dc1"
		c.ACLsEnabled = true
	})
	defer os.RemoveAll(dir2)
	defer newS1.Shutdown()
	testrpc.WaitForTestAgent(t, newS1.RPC, "dc1")

	retry.Run(t, func(r *retry.R) {
		_, anon, err := newS1.fsm.State().ACLTokenGetBySecret(nil, anonymousToken, nil)
		require.NoError(r, err)
		require.NotNil(r, anon)

		// Existing token should not have been purged during ACL initialization
		require.Len(r, anon.Policies, 1)
		require.Equal(r, structs.ACLPolicyGlobalManagementID, anon.Policies[0].ID)
	})
}

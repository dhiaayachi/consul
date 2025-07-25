// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package ratelimit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/sdk/testutil"
	"github.com/dhiaayachi/consul/sdk/testutil/retry"
	"github.com/stretchr/testify/require"

	libcluster "github.com/dhiaayachi/consul/test/integration/consul-container/libs/cluster"
	libtopology "github.com/dhiaayachi/consul/test/integration/consul-container/libs/topology"
	"github.com/dhiaayachi/consul/test/integration/consul-container/libs/utils"
)

const (
	retryableErrorMsg    = "Unexpected response code: 429 (rate limit exceeded, try again later or against a different server)"
	nonRetryableErrorMsg = "Unexpected response code: 503 (rate limit exceeded for operation that can only be performed by the leader, try again later)"
)

// TestRateLimit
// This test validates
// - enforcing mode
//   - read_rate - returns 429 - was blocked and returns retryable error
//   - write_rate - returns 503 - was blocked and is not retryable
//   - on each
//     - fires metrics for exceeding
//     - logs for exceeding

func TestServerRequestRateLimit(t *testing.T) {

	type action struct {
		function           func(client *api.Client) error
		rateLimitOperation string
		rateLimitType      string // will become an array of strings
	}
	type operation struct {
		action            action
		expectedErrorMsg  string
		expectExceededLog bool
		expectMetric      bool
	}
	type testCase struct {
		description    string
		cmd            string
		operations     []operation
		mode           string
		enterpriseOnly bool
	}

	// getKV and putKV are net/RPC calls
	getKV := action{
		function: func(client *api.Client) error {
			_, _, err := client.KV().Get("foo", &api.QueryOptions{})
			return err
		},
		rateLimitOperation: "KVS.Get",
		rateLimitType:      "global/read",
	}
	putKV := action{
		function: func(client *api.Client) error {
			_, err := client.KV().Put(&api.KVPair{Key: "foo", Value: []byte("bar")}, &api.WriteOptions{})
			return err
		},
		rateLimitOperation: "KVS.Apply",
		rateLimitType:      "global/write",
	}

	// listPartition and putPartition are gRPC calls
	listPartition := action{
		function: func(client *api.Client) error {
			ctx := context.Background()
			_, _, err := client.Partitions().List(ctx, nil)
			return err
		},
		rateLimitOperation: "/partition.PartitionService/List",
		rateLimitType:      "global/read",
	}

	putPartition := action{
		function: func(client *api.Client) error {
			ctx := context.Background()
			p := api.Partition{
				Name: "ptest",
			}
			_, _, err := client.Partitions().Create(ctx, &p, nil)
			return err
		},
		rateLimitOperation: "/partition.PartitionService/Write",
		rateLimitType:      "global/write",
	}

	testCases := []testCase{
		// HTTP & net/RPC
		{
			description: "HTTP & net-RPC | Mode: disabled - errors: no | exceeded logs: no | metrics: no",
			cmd:         `-hcl=limits { request_limits { mode = "disabled" read_rate = 0 write_rate = 0 }}`,
			mode:        "disabled",
			operations: []operation{
				{
					action:            putKV,
					expectedErrorMsg:  "",
					expectExceededLog: false,
					expectMetric:      false,
				},
				{
					action:            getKV,
					expectedErrorMsg:  "",
					expectExceededLog: false,
					expectMetric:      false,
				},
			},
		},
		{
			description: "HTTP & net-RPC | Mode: permissive - errors: no | exceeded logs: yes | metrics: yes",
			cmd:         `-hcl=limits { request_limits { mode = "permissive" read_rate = 0 write_rate = 0 }}`,
			mode:        "permissive",
			operations: []operation{
				{
					action:            putKV,
					expectedErrorMsg:  "",
					expectExceededLog: true,
					expectMetric:      true,
				},
				{
					action:            getKV,
					expectedErrorMsg:  "",
					expectExceededLog: true,
					expectMetric:      true,
				},
			},
		},
		{
			description: "HTTP & net-RPC | Mode: enforcing - errors: yes | exceeded logs: yes | metrics: yes",
			cmd:         `-hcl=limits { request_limits { mode = "enforcing" read_rate = 0 write_rate = 0 }}`,
			mode:        "enforcing",
			operations: []operation{
				{
					action:            putKV,
					expectedErrorMsg:  nonRetryableErrorMsg,
					expectExceededLog: true,
					expectMetric:      true,
				},
				{
					action:            getKV,
					expectedErrorMsg:  retryableErrorMsg,
					expectExceededLog: true,
					expectMetric:      true,
				},
			},
		},
		// gRPC
		{
			description: "GRPC / Mode: disabled - errors: no / exceeded logs: no / metrics: no",
			cmd:         `-hcl=limits { request_limits { mode = "disabled" read_rate = 0 write_rate = 0 }}`,
			mode:        "disabled",
			operations: []operation{
				{
					action:            putPartition,
					expectedErrorMsg:  "",
					expectExceededLog: false,
					expectMetric:      false,
				},
				{
					action:            listPartition,
					expectedErrorMsg:  "",
					expectExceededLog: false,
					expectMetric:      false,
				},
			},
			enterpriseOnly: true,
		},
		{
			description: "GRPC / Mode: permissive - errors: no / exceeded logs: yes / metrics: no",
			cmd:         `-hcl=limits { request_limits { mode = "permissive" read_rate = 0 write_rate = 0 }}`,
			mode:        "permissive",
			operations: []operation{
				{
					action:            putPartition,
					expectedErrorMsg:  "",
					expectExceededLog: true,
					expectMetric:      true,
				},
				{
					action:            listPartition,
					expectedErrorMsg:  "",
					expectExceededLog: true,
					expectMetric:      true,
				},
			},
			enterpriseOnly: true,
		},
		{
			description: "GRPC / Mode: enforcing - errors: yes / exceeded logs: yes / metrics: yes",
			cmd:         `-hcl=limits { request_limits { mode = "enforcing" read_rate = 0 write_rate = 0 }}`,
			mode:        "enforcing",
			operations: []operation{
				{
					action:            putPartition,
					expectedErrorMsg:  nonRetryableErrorMsg,
					expectExceededLog: true,
					expectMetric:      true,
				},
				{
					action:            listPartition,
					expectedErrorMsg:  retryableErrorMsg,
					expectExceededLog: true,
					expectMetric:      true,
				},
			},
			enterpriseOnly: true,
		},
	}

	for _, tc := range testCases {
		if tc.enterpriseOnly && !utils.IsEnterprise() {
			continue
		}
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			clusterConfig := &libtopology.ClusterConfig{
				NumServers:  1,
				NumClients:  0,
				Cmd:         tc.cmd,
				LogConsumer: &libtopology.TestLogConsumer{},
				BuildOpts: &libcluster.BuildOptions{
					Datacenter:             "dc1",
					InjectAutoEncryption:   true,
					InjectGossipEncryption: true,
				},
				ApplyDefaultProxySettings: false,
			}

			cluster, client := setupClusterAndClient(t, clusterConfig, true)
			defer terminate(t, cluster)

			// perform actions and validate returned errors to client
			for _, op := range tc.operations {
				err := op.action.function(client)
				if len(op.expectedErrorMsg) > 0 {
					require.Error(t, err)
					require.Equal(t, op.expectedErrorMsg, err.Error())
				} else {
					require.NoError(t, err)
				}
			}

			// validate logs and metrics
			// doing this in a separate loop so we can perform actions, allow metrics
			// and logs to collect and then assert on each.
			for _, op := range tc.operations {
				timer := &retry.Timer{Timeout: 15 * time.Second, Wait: 500 * time.Millisecond}
				retry.RunWith(timer, t, func(r *retry.R) {
					checkForMetric(r, cluster, op.action.rateLimitOperation, op.action.rateLimitType, tc.mode, op.expectMetric)

					// validate logs
					// putting this last as there are cases where logs
					// were not present in consumer when assertion was made.
					checkLogsForMessage(r, clusterConfig.LogConsumer.Msgs,
						fmt.Sprintf("[DEBUG] agent.server.rpc-rate-limit: RPC exceeded allowed rate limit: rpc=%s", op.action.rateLimitOperation),
						op.action.rateLimitOperation, "exceeded", op.expectExceededLog)

				})
			}
		})
	}
}

func setupClusterAndClient(t *testing.T, config *libtopology.ClusterConfig, isServer bool) (*libcluster.Cluster, *api.Client) {
	cluster, _, _ := libtopology.NewCluster(t, config)

	client, err := cluster.GetClient(nil, isServer)
	require.NoError(t, err)

	return cluster, client
}

func checkForMetric(t testutil.TestingTB, cluster *libcluster.Cluster, operationName string, expectedLimitType string, expectedMode string, expectMetric bool) {
	// validate metrics
	server, err := cluster.GetClient(nil, true)
	require.NoError(t, err)
	metricsInfo, err := server.Agent().Metrics()
	// TODO(NET-1978): currently returns NaN error
	//			require.NoError(t, err)
	if metricsInfo != nil && err == nil {
		if expectMetric {
			const counterName = "consul.rpc.rate_limit.exceeded"

			var counter api.SampledValue
			for _, c := range metricsInfo.Counters {
				if c.Name == counterName {
					counter = c
					break
				}
			}
			require.NotEmptyf(t, counter.Name, "counter not found: %s", counterName)

			operation, ok := counter.Labels["op"]
			require.True(t, ok)

			limitType, ok := counter.Labels["limit_type"]
			require.True(t, ok)

			mode, ok := counter.Labels["mode"]
			require.True(t, ok)

			if operation == operationName {
				require.GreaterOrEqual(t, counter.Count, 1)
				require.Equal(t, expectedLimitType, limitType)
				require.Equal(t, expectedMode, mode)
			}
		}
	}
}

func checkLogsForMessage(t testutil.TestingTB, logs []string, msg string, operationName string, logType string, logShouldExist bool) {
	if logShouldExist {
		found := false
		for _, log := range logs {
			if strings.Contains(log, msg) {
				found = true
				break
			}
		}
		expectedLog := fmt.Sprintf("%s log check failed for: %s. Log expected: %t", logType, operationName, logShouldExist)
		require.Equal(t, logShouldExist, found, expectedLog)
	}
}

func terminate(t *testing.T, cluster *libcluster.Cluster) {
	err := cluster.Terminate()
	require.NoError(t, err)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package sprawl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/proto-public/pbresource"

	"github.com/dhiaayachi/consul/testing/deployer/sprawl/internal/secrets"
	"github.com/dhiaayachi/consul/testing/deployer/topology"
	"github.com/dhiaayachi/consul/testing/deployer/util"
)

func (s *Sprawl) getResourceClient(clusterName string) pbresource.ResourceServiceClient {
	return pbresource.NewResourceServiceClient(s.grpcConns[clusterName])
}

func (s *Sprawl) getManagementTokenContext(ctx context.Context, clusterName string) context.Context {
	mgmtToken := s.secrets.ReadGeneric(clusterName, secrets.BootstrapToken)
	//nolint:staticcheck
	return context.WithValue(ctx, "x-consul-token", mgmtToken)
}

func getLeader(client *api.Client) (string, error) {
	leaderAdd, err := client.Status().Leader()
	if err != nil {
		return "", fmt.Errorf("could not query leader: %w", err)
	}
	if leaderAdd == "" {
		return "", errors.New("no leader available")
	}
	return leaderAdd, nil
}

func (s *Sprawl) waitForLeader(cluster *topology.Cluster) {
	var (
		client = s.clients[cluster.Name]
		logger = s.logger.With("cluster", cluster.Name)
	)
	logger.Info("waiting for cluster to elect leader")
	for {
		leader, err := client.Status().Leader()
		if leader != "" && err == nil {
			logger.Info("cluster has leader", "leader_addr", leader)
			return
		}
		logger.Debug("cluster has no leader yet", "error", err)
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *Sprawl) rejoinAllConsulServers() error {
	// Join the servers together.
	for _, cluster := range s.topology.Clusters {
		if err := s.rejoinServers(cluster); err != nil {
			return fmt.Errorf("rejoinServers[%s]: %w", cluster.Name, err)
		}
		s.waitForLeader(cluster)
	}
	return nil
}

func (s *Sprawl) rejoinServers(cluster *topology.Cluster) error {
	var (
		// client = s.clients[cluster.Name]
		logger = s.logger.With("cluster", cluster.Name)
	)

	servers := cluster.ServerNodes()

	var recoveryToken string
	if servers[0].Images.GreaterThanVersion(topology.MinVersionAgentTokenPartition) {
		recoveryToken = s.secrets.ReadGeneric(cluster.Name, secrets.AgentRecovery)
	} else {
		recoveryToken = s.secrets.ReadGeneric(cluster.Name, secrets.BootstrapToken)
	}

	node0, rest := servers[0], servers[1:]
	client, err := util.ProxyNotPooledAPIClient(
		node0.LocalProxyPort(),
		node0.LocalAddress(),
		8500,
		recoveryToken,
	)
	if err != nil {
		return fmt.Errorf("could not get client for %q: %w", node0.ID(), err)
	}

	logger.Info("joining servers together",
		"from", node0.ID(),
		"rest", nodeSliceToNodeIDSlice(rest),
	)
	for _, node := range rest {
		for {
			err = client.Agent().Join(node.LocalAddress(), false)
			if err == nil {
				break
			}
			logger.Warn("could not join", "from", node0.ID(), "to", node.ID(), "error", err)
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

func nodeSliceToNodeIDSlice(nodes []*topology.Node) []topology.NodeID {
	var out []topology.NodeID
	for _, node := range nodes {
		out = append(out, node.ID())
	}
	return out
}

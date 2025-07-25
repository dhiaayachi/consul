// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package servers provides a Manager interface for Manager managed
// metadata.Servers objects.  The servers package manages servers from a Consul
// client's perspective (i.e. a list of servers that a client talks with for
// RPCs).  The servers package does not provide any API guarantees and should
// be called only by `hashicorp/consul`.
package router

import (
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/dhiaayachi/consul/agent/metadata"
	"github.com/dhiaayachi/consul/logging"
)

// ManagerSerfCluster is an interface wrapper around Serf in order to make this
// easier to unit test.
type ManagerSerfCluster interface {
	NumNodes() int
}

// Pinger is an interface wrapping client.ConnPool to prevent a cyclic import
// dependency.
type Pinger interface {
	Ping(dc, nodeName string, addr net.Addr) (bool, error)
}

// serverList is a local copy of the struct used to maintain the list of
// Consul servers used by Manager.
//
// NOTE(sean@): We are explicitly relying on the fact that serverList will
// be copied onto the stack.  Please keep this structure light.
type serverList struct {
	// servers tracks the locally known servers.  List membership is
	// maintained by Serf.
	servers []*metadata.Server
}

type Manager struct {
	// listValue manages the atomic load/store of a Manager's serverList
	listValue atomic.Value
	listLock  sync.Mutex

	// rebalanceTimer controls the duration of the rebalance interval
	rebalanceTimer *time.Timer

	// shutdownCh is a copy of the channel in consul.Client
	shutdownCh chan struct{}

	logger hclog.Logger

	// clusterInfo is used to estimate the approximate number of nodes in
	// a cluster and limit the rate at which it rebalances server
	// connections.  ManagerSerfCluster is an interface that wraps serf.
	clusterInfo ManagerSerfCluster

	// connPoolPinger is used to test the health of a server in the
	// connection pool.  Pinger is an interface that wraps
	// client.ConnPool.
	connPoolPinger Pinger

	rebalancer Rebalancer

	// serverName has the name of the managers's server. This is used to
	// short-circuit pinging to itself.
	serverName string

	// notifyFailedBarrier is acts as a barrier to prevent queuing behind
	// serverListLog and acts as a TryLock().
	notifyFailedBarrier int32

	// offline is used to indicate that there are no servers, or that all
	// known servers have failed the ping test.
	offline int32
}

// AddServer takes out an internal write lock and adds a new server.  If the
// server is not known, appends the server to the list.  The new server will
// begin seeing use after the rebalance timer fires or enough servers fail
// organically.  If the server is already known, merge the new server
// details.
func (m *Manager) AddServer(s *metadata.Server) {
	m.listLock.Lock()
	defer m.listLock.Unlock()
	l := m.getServerList()

	// Check if this server is known
	found := false
	for idx, existing := range l.servers {
		if existing.Name == s.Name {
			newServers := make([]*metadata.Server, len(l.servers))
			copy(newServers, l.servers)

			// Overwrite the existing server details in order to
			// possibly update metadata (e.g. server version)
			newServers[idx] = s

			l.servers = newServers
			found = true
			break
		}
	}

	// Add to the list if not known
	if !found {
		newServers := make([]*metadata.Server, len(l.servers), len(l.servers)+1)
		copy(newServers, l.servers)
		newServers = append(newServers, s)
		l.servers = newServers
	}

	// Assume we are no longer offline since we've just seen a new server.
	atomic.StoreInt32(&m.offline, 0)

	// Start using this list of servers.
	m.saveServerList(l)
}

// UpdateTLS updates the TLS setting for the servers in this manager
func (m *Manager) UpdateTLS(useTLS bool) {
	m.listLock.Lock()
	defer m.listLock.Unlock()

	list := m.getServerList()
	for _, server := range list.servers {
		server.UseTLS = useTLS
	}
	m.saveServerList(list)
}

// cycleServers returns a new list of servers that has dequeued the first
// server and enqueued it at the end of the list.  cycleServers assumes the
// caller is holding the listLock.  cycleServer does not test or ping
// the next server inline.  cycleServer may be called when the environment
// has just entered an unhealthy situation and blocking on a server test is
// less desirable than just returning the next server in the firing line.  If
// the next server fails, it will fail fast enough and cycleServer will be
// called again.
func (l *serverList) cycleServer() (servers []*metadata.Server) {
	numServers := len(l.servers)
	if numServers < 2 {
		return servers // No action required
	}

	newServers := make([]*metadata.Server, 0, numServers)
	newServers = append(newServers, l.servers[1:]...)
	newServers = append(newServers, l.servers[0])

	return newServers
}

// removeServerByKey performs an inline removal of the first matching server
func (l *serverList) removeServerByKey(targetKey *metadata.Key) {
	for i, s := range l.servers {
		if targetKey.Equal(s.Key()) {
			copy(l.servers[i:], l.servers[i+1:])
			l.servers[len(l.servers)-1] = nil
			l.servers = l.servers[:len(l.servers)-1]
			return
		}
	}
}

// shuffleServers shuffles the server list in place
func (l *serverList) shuffleServers() {
	for i := len(l.servers) - 1; i > 0; i-- {
		j := rand.Int31n(int32(i + 1))
		l.servers[i], l.servers[j] = l.servers[j], l.servers[i]
	}
}

// IsOffline checks to see if all the known servers have failed their ping
// test during the last rebalance.
func (m *Manager) IsOffline() bool {
	offline := atomic.LoadInt32(&m.offline)
	return offline == 1
}

// FindServer takes out an internal "read lock" and searches through the list
// of servers to find a "healthy" server.  If the server is actually
// unhealthy, we rely on Serf to detect this and remove the node from the
// server list.  If the server at the front of the list has failed or fails
// during an RPC call, it is rotated to the end of the list.  If there are no
// servers available, return nil.
func (m *Manager) FindServer() *metadata.Server {
	l := m.getServerList()
	numServers := len(l.servers)
	if numServers == 0 {
		m.logger.Warn("No servers available")
		return nil
	}

	// Return whatever is at the front of the list because it is
	// assumed to be the oldest in the server list (unless -
	// hypothetically - the server list was rotated right after a
	// server was added).
	return l.servers[0]
}

func (m *Manager) checkServers(fn func(srv *metadata.Server) bool) bool {
	if m == nil {
		return true
	}

	for _, srv := range m.getServerList().servers {
		if !fn(srv) {
			return false
		}
	}
	return true
}

func (m *Manager) CheckServers(fn func(srv *metadata.Server) bool) {
	_ = m.checkServers(fn)
}

// getServerList is a convenience method which hides the locking semantics
// of atomic.Value from the caller.
func (m *Manager) getServerList() serverList {
	if m == nil {
		return serverList{}
	}
	return m.listValue.Load().(serverList)
}

// saveServerList is a convenience method which hides the locking semantics
// of atomic.Value from the caller.
func (m *Manager) saveServerList(l serverList) {
	m.listValue.Store(l)
}

// New is the only way to safely create a new Manager struct.
func New(logger hclog.Logger, shutdownCh chan struct{}, clusterInfo ManagerSerfCluster, connPoolPinger Pinger, serverName string, rb Rebalancer) (m *Manager) {
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{})
	}

	m = new(Manager)
	m.logger = logger.Named(logging.Manager)
	m.clusterInfo = clusterInfo       // can't pass *consul.Client: import cycle
	m.connPoolPinger = connPoolPinger // can't pass *consul.ConnPool: import cycle
	m.rebalanceTimer = time.NewTimer(delayer.MinDelay)
	m.shutdownCh = shutdownCh
	m.rebalancer = rb
	m.serverName = serverName
	atomic.StoreInt32(&m.offline, 1)

	l := serverList{}
	l.servers = make([]*metadata.Server, 0)
	m.saveServerList(l)
	return m
}

// NotifyFailedServer marks the passed in server as "failed" by rotating it
// to the end of the server list.
func (m *Manager) NotifyFailedServer(s *metadata.Server) {
	l := m.getServerList()

	// If the server being failed is not the first server on the list,
	// this is a noop.  If, however, the server is failed and first on
	// the list, acquire the lock, retest, and take the penalty of moving
	// the server to the end of the list.

	// Only rotate the server list when there is more than one server
	if len(l.servers) > 1 && l.servers[0].Name == s.Name &&
		// Use atomic.CAS to emulate a TryLock().
		atomic.CompareAndSwapInt32(&m.notifyFailedBarrier, 0, 1) {
		defer atomic.StoreInt32(&m.notifyFailedBarrier, 0)

		// Grab a lock, retest, and take the hit of cycling the first
		// server to the end.
		m.listLock.Lock()
		defer m.listLock.Unlock()
		l = m.getServerList()

		if len(l.servers) > 1 && l.servers[0].Name == s.Name {
			l.servers = l.cycleServer()
			m.saveServerList(l)
			m.logger.Debug("cycled away from server", "server", s.String())
		}
	}
}

// NumServers takes out an internal "read lock" and returns the number of
// servers.  numServers includes both healthy and unhealthy servers.
func (m *Manager) NumServers() int {
	l := m.getServerList()
	return len(l.servers)
}

func (m *Manager) healthyServer(server *metadata.Server) bool {
	// Check to see if the manager is trying to ping itself. This
	// is a small optimization to avoid performing an unnecessary
	// RPC call.
	// If this is true, we know there are healthy servers for this
	// manager and we don't need to continue.
	if m.serverName != "" && server.Name == m.serverName {
		return true
	}
	if ok, err := m.connPoolPinger.Ping(server.Datacenter, server.ShortName, server.Addr); !ok {
		m.logger.Debug("pinging server failed",
			"server", server.String(),
			"error", err,
		)
		return false
	}
	return true
}

// RebalanceServers shuffles the list of servers on this metadata.  The server
// at the front of the list is selected for the next RPC.  RPC calls that
// fail for a particular server are rotated to the end of the list.  This
// method reshuffles the list periodically in order to redistribute work
// across all known consul servers (i.e. guarantee that the order of servers
// in the server list is not positively correlated with the age of a server
// in the Consul cluster).  Periodically shuffling the server list prevents
// long-lived clients from fixating on long-lived servers.
//
// Unhealthy servers are removed when serf notices the server has been
// deregistered.  Before the newly shuffled server list is saved, the new
// remote endpoint is tested to ensure its responsive.
func (m *Manager) RebalanceServers() {
	// Obtain a copy of the current serverList
	l := m.getServerList()

	// Shuffle servers so we have a chance of picking a new one.
	l.shuffleServers()

	// Iterate through the shuffled server list to find an assumed
	// healthy server.  NOTE: Do not iterate on the list directly because
	// this loop mutates the server list in-place.
	var foundHealthyServer bool
	for i := 0; i < len(l.servers); i++ {
		// Always test the first server. Failed servers are cycled
		// while Serf detects the node has failed.
		if m.healthyServer(l.servers[0]) {
			foundHealthyServer = true
			break
		}
		l.servers = l.cycleServer()
	}

	// If no healthy servers were found, sleep and wait for Serf to make
	// the world a happy place again. Update the offline status.
	if foundHealthyServer {
		atomic.StoreInt32(&m.offline, 0)
	} else {
		atomic.StoreInt32(&m.offline, 1)
		m.logger.Debug("No healthy servers during rebalance, aborting")
		return
	}

	// Verify that all servers are present
	if m.reconcileServerList(&l) {
		m.logger.Debug("Rebalanced servers, new active server",
			"number_of_servers", len(l.servers),
			"active_server", l.servers[0].String(),
		)
	}
	// else {
	// reconcileServerList failed because Serf removed the server
	// that was at the front of the list that had successfully
	// been Ping'ed.  Between the Ping and reconcile, a Serf
	// event had shown up removing the node.
	//
	// Instead of doing any heroics, "freeze in place" and
	// continue to use the existing connection until the next
	// rebalance occurs.
	// }
}

// reconcileServerList returns true when the first server in serverList
// exists in the receiver's serverList.  If true, the merged serverList is
// stored as the receiver's serverList.  Returns false if the first server
// does not exist in the list (i.e. was removed by Serf during a
// PingConsulServer() call.  Newly added servers are appended to the list and
// other missing servers are removed from the list.
func (m *Manager) reconcileServerList(l *serverList) bool {
	m.listLock.Lock()
	defer m.listLock.Unlock()

	// newServerCfg is a serverList that has been kept up to date with
	// Serf node join and node leave events.
	newServerCfg := m.getServerList()

	// If Serf has removed all nodes, or there is no selected server
	// (zero nodes in serverList), abort early.
	if len(newServerCfg.servers) == 0 || len(l.servers) == 0 {
		return false
	}

	type targetServer struct {
		server *metadata.Server

		//   'b' == both
		//   'o' == original
		//   'n' == new
		state byte
	}
	mergedList := make(map[metadata.Key]*targetServer, len(l.servers))
	for _, s := range l.servers {
		mergedList[*s.Key()] = &targetServer{server: s, state: 'o'}
	}
	for _, s := range newServerCfg.servers {
		k := s.Key()
		_, found := mergedList[*k]
		if found {
			mergedList[*k].state = 'b'
		} else {
			mergedList[*k] = &targetServer{server: s, state: 'n'}
		}
	}

	// Ensure the selected server has not been removed by Serf
	selectedServerKey := l.servers[0].Key()
	if v, found := mergedList[*selectedServerKey]; found && v.state == 'o' {
		return false
	}

	// Append any new servers and remove any old servers
	for k, v := range mergedList {
		switch v.state {
		case 'b':
			// Do nothing, server exists in both
		case 'o':
			// Servers has been removed
			l.removeServerByKey(&k)
		case 'n':
			// Servers added
			l.servers = append(l.servers, v.server)
		default:
			panic("unknown merge list state")
		}
	}

	m.saveServerList(*l)
	return true
}

// RemoveServer takes out an internal write lock and removes a server from
// the server list.
func (m *Manager) RemoveServer(s *metadata.Server) {
	m.listLock.Lock()
	defer m.listLock.Unlock()
	l := m.getServerList()

	// Remove the server if known
	for i := range l.servers {
		if l.servers[i].Name == s.Name {
			newServers := make([]*metadata.Server, 0, len(l.servers)-1)
			newServers = append(newServers, l.servers[:i]...)
			newServers = append(newServers, l.servers[i+1:]...)
			l.servers = newServers

			m.saveServerList(l)
			return
		}
	}
}

// ResetRebalanceTimer resets the rebalance timer.  This method exists for
// testing and should not be used directly.
func (m *Manager) ResetRebalanceTimer() {
	m.listLock.Lock()
	defer m.listLock.Unlock()
	m.rebalanceTimer.Reset(delayer.MinDelay)
}

// Run periodically shuffles the list of servers to evenly distribute load.
// Run exits when shutdownCh is closed.
//
// When a server fails it is moved to the end of the list, and new servers are
// appended to the end of the list. Run ensures that load is distributed evenly
// to all servers by randomly shuffling the list.
func (m *Manager) Run() {
	for {
		select {
		case <-m.rebalanceTimer.C:
			m.rebalancer()
			m.RebalanceServers()
			delay := delayer.Delay(len(m.getServerList().servers), m.clusterInfo.NumNodes())
			m.rebalanceTimer.Reset(delay)

		case <-m.shutdownCh:
			m.logger.Info("shutting down")
			return
		}
	}
}

// delayer is used to calculate the time to wait between calls to rebalance the
// servers. Rebalancing is necessary to ensure that load is balanced evenly
// across all the servers.
//
// The values used by delayer must balance perfectly distributed server load
// against the overhead of a client reconnecting to a server. Rebalancing on
// every request would cause a lot of unnecessary load as clients reconnect,
// where as never rebalancing would lead to situations where one or two servers
// handle a lot more requests than others.
//
// These values result in a minimum delay of 120-180s. Once the number of
// nodes/server exceeds 11520, the value will be determined by multiplying the
// node/server ratio by 15.625ms.
var delayer = rebalanceDelayer{
	MinDelay:  2 * time.Minute,
	MaxJitter: time.Minute,
	// Once the number of nodes/server exceeds 11520 this value is used to
	// increase the delay between rebalances to set a limit on the number of
	// reconnections per server in a given time frame.
	//
	// A higher value comes at the cost of increased recovery time after a
	// partition.
	//
	// For example, in a 100,000 node consul cluster with 5 servers, it will
	// take ~5min for all clients to rebalance their connections.  If
	// 99,995 agents are in the minority talking to only one server, it
	// will take ~26min for all clients to rebalance.  A 10K cluster in
	// the same scenario will take ~2.6min to rebalance.
	DelayPerNode: 15*time.Millisecond + 625*time.Microsecond,
}

type rebalanceDelayer struct {
	// MinDelay that may be returned by Delay
	MinDelay time.Duration
	// MaxJitter to add to MinDelay to ensure there is some randomness in the
	// delay.
	MaxJitter time.Duration
	// DelayPerNode is the duration to add to each node when calculating delay.
	// The value is divided by the number of servers to arrive at the final
	// delay value.
	DelayPerNode time.Duration
}

func (d *rebalanceDelayer) Delay(servers int, nodes int) time.Duration {
	min := d.MinDelay + time.Duration(rand.Int63n(int64(d.MaxJitter)))
	if servers == 0 {
		return min
	}

	delay := time.Duration(float64(nodes) * float64(d.DelayPerNode) / float64(servers))
	if delay < min {
		return min
	}
	return delay
}

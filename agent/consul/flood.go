// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package consul

import (
	"time"

	"github.com/dhiaayachi/consul/agent/router"
	"github.com/hashicorp/serf/serf"
)

// FloodNotify lets all the waiting Flood goroutines know that some change may
// have affected them.
func (s *Server) FloodNotify() {
	s.floodLock.RLock()
	defer s.floodLock.RUnlock()

	for _, ch := range s.floodCh {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Flood is a long-running goroutine that floods servers from the LAN to the
// given global Serf instance, such as the WAN. This will exit once either of
// the Serf instances are shut down.
func (s *Server) Flood(addrFn router.FloodAddrFn, dstSerf *serf.Serf) {
	s.floodLock.Lock()
	floodCh := make(chan struct{})
	s.floodCh = append(s.floodCh, floodCh)
	s.floodLock.Unlock()

	ticker := time.NewTicker(s.config.SerfFloodInterval)
	defer ticker.Stop()
	defer func() {
		s.floodLock.Lock()
		defer s.floodLock.Unlock()

		for i, ch := range s.floodCh {
			if ch == floodCh {
				s.floodCh = append(s.floodCh[:i], s.floodCh[i+1:]...)
				return
			}
		}
		panic("flood channels out of sync")
	}()

	for {
		select {
		case <-s.serfLAN.ShutdownCh():
			return

		case <-dstSerf.ShutdownCh():
			return

		case <-ticker.C:
			router.FloodJoins(s.logger, addrFn, s.config.Datacenter, s.serfLAN, dstSerf)

		case <-floodCh:
			router.FloodJoins(s.logger, addrFn, s.config.Datacenter, s.serfLAN, dstSerf)
		}

	}
}

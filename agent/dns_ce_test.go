// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package agent

import (
	"context"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/testrpc"
)

func TestDNS_CE_PeeredServices(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	a := StartTestAgent(t, TestAgent{HCL: ``, Overrides: `peering = { test_allow_peer_registrations = true } `})
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	makeReq := func() *structs.RegisterRequest {
		return &structs.RegisterRequest{
			PeerName:   "peer1",
			Datacenter: "dc1",
			Node:       "peernode1",
			Address:    "198.18.1.1",
			Service: &structs.NodeService{
				PeerName: "peer1",
				Kind:     structs.ServiceKindConnectProxy,
				Service:  "web-proxy",
				Address:  "199.0.0.1",
				Port:     12345,
				Proxy: structs.ConnectProxyConfig{
					DestinationServiceName: "peer-web",
				},
				EnterpriseMeta: *acl.DefaultEnterpriseMeta(),
			},
			EnterpriseMeta: *acl.DefaultEnterpriseMeta(),
		}
	}

	dnsQuery := func(t *testing.T, question string, typ uint16) *dns.Msg {
		m := new(dns.Msg)
		m.SetQuestion(question, typ)

		c := new(dns.Client)
		reply, _, err := c.Exchange(m, a.DNSAddr())
		require.NoError(t, err)
		require.Len(t, reply.Answer, 1, "zero valid records found for %q", question)
		return reply
	}

	assertARec := func(t *testing.T, rec dns.RR, expectName, expectIP string) {
		aRec, ok := rec.(*dns.A)
		require.True(t, ok, "Extra is not an A record: %T", rec)
		require.Equal(t, expectName, aRec.Hdr.Name)
		require.Equal(t, expectIP, aRec.A.String())
	}

	assertSRVRec := func(t *testing.T, rec dns.RR, expectName string, expectPort uint16) {
		srvRec, ok := rec.(*dns.SRV)
		require.True(t, ok, "Answer is not a SRV record: %T", rec)
		require.Equal(t, expectName, srvRec.Target)
		require.Equal(t, expectPort, srvRec.Port)
	}

	t.Run("srv-with-addr-reply", func(t *testing.T) {
		require.NoError(t, a.RPC(context.Background(), "Catalog.Register", makeReq(), &struct{}{}))
		q := dnsQuery(t, "web-proxy.service.peer1.peer.consul.", dns.TypeSRV)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 1)

		addr := "c7000001.addr.consul."
		assertSRVRec(t, q.Answer[0], addr, 12345)
		assertARec(t, q.Extra[0], addr, "199.0.0.1")

		// Query the addr to make sure it's also valid.
		q = dnsQuery(t, addr, dns.TypeA)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 0)
		assertARec(t, q.Answer[0], addr, "199.0.0.1")
	})

	t.Run("srv-with-node-reply", func(t *testing.T) {
		req := makeReq()
		// Clear service address to trigger node response
		req.Service.Address = ""
		require.NoError(t, a.RPC(context.Background(), "Catalog.Register", req, &struct{}{}))
		q := dnsQuery(t, "web-proxy.service.peer1.peer.consul.", dns.TypeSRV)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 1)

		nodeName := "peernode1.node.peer1.peer.consul."
		assertSRVRec(t, q.Answer[0], nodeName, 12345)
		assertARec(t, q.Extra[0], nodeName, "198.18.1.1")

		// Query the node to make sure it's also valid.
		q = dnsQuery(t, nodeName, dns.TypeA)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 0)
		assertARec(t, q.Answer[0], nodeName, "198.18.1.1")
	})

	t.Run("srv-with-fqdn-reply", func(t *testing.T) {
		req := makeReq()
		// Set non-ip address to trigger external response
		req.Address = "localhost"
		req.Service.Address = ""
		require.NoError(t, a.RPC(context.Background(), "Catalog.Register", req, &struct{}{}))
		q := dnsQuery(t, "web-proxy.service.peer1.peer.consul.", dns.TypeSRV)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 0)
		assertSRVRec(t, q.Answer[0], "localhost.", 12345)
	})

	t.Run("a-reply", func(t *testing.T) {
		require.NoError(t, a.RPC(context.Background(), "Catalog.Register", makeReq(), &struct{}{}))
		q := dnsQuery(t, "web-proxy.service.peer1.peer.consul.", dns.TypeA)
		require.Len(t, q.Answer, 1)
		require.Len(t, q.Extra, 0)
		assertARec(t, q.Answer[0], "web-proxy.service.peer1.peer.consul.", "199.0.0.1")
	})
}

func getTestCasesParseLocality() []testCaseParseLocality {
	testCases := []testCaseParseLocality{
		{
			name:                "test [.<datacenter>.dc]",
			labels:              []string{"test-dc", "dc"},
			enterpriseDNSConfig: enterpriseDNSConfig{},
			expectedResult: queryLocality{
				EnterpriseMeta: acl.EnterpriseMeta{},
				datacenter:     "test-dc",
			},
			expectedOK: true,
		},
		{
			name:                "test [.<peer>.peer]",
			labels:              []string{"test-peer", "peer"},
			enterpriseDNSConfig: enterpriseDNSConfig{},
			expectedResult: queryLocality{
				EnterpriseMeta: acl.EnterpriseMeta{},
				peer:           "test-peer",
			},
			expectedOK: true,
		},
		{
			name:                "test 1 label",
			labels:              []string{"test-peer"},
			enterpriseDNSConfig: enterpriseDNSConfig{},
			expectedResult: queryLocality{
				EnterpriseMeta:   acl.EnterpriseMeta{},
				peerOrDatacenter: "test-peer",
			},
			expectedOK: true,
		},
		{
			name:                "test 0 labels",
			labels:              []string{},
			enterpriseDNSConfig: enterpriseDNSConfig{},
			expectedResult:      queryLocality{},
			expectedOK:          true,
		},
		{
			name:                "test 3 labels returns not found",
			labels:              []string{"test-dc", "dc", "test-blah"},
			enterpriseDNSConfig: enterpriseDNSConfig{},
			expectedResult:      queryLocality{},
			expectedOK:          false,
		},
	}
	return testCases
}

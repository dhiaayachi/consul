// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/testrpc"
)

func TestConfig_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	reqs := []structs.ConfigEntryRequest{
		{
			Datacenter: "dc1",
			Entry: &structs.ServiceConfigEntry{
				Name: "foo",
			},
		},
		{
			Datacenter: "dc1",
			Entry: &structs.ServiceConfigEntry{
				Name: "bar",
			},
		},
		{
			Datacenter: "dc1",
			Entry: &structs.ProxyConfigEntry{
				Name: structs.ProxyConfigGlobal,
				Config: map[string]interface{}{
					"foo": "bar",
					"bar": 1,
				},
			},
		},
		{
			Datacenter: "dc1",
			Entry: &structs.MeshConfigEntry{
				TransparentProxy: structs.TransparentProxyMeshConfig{
					MeshDestinationsOnly: true,
				},
				Meta: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
	}
	for _, req := range reqs {
		out := false
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Apply", &req, &out))
	}

	t.Run("get a single service entry", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/config/service-defaults/foo", nil)
		resp := httptest.NewRecorder()
		obj, err := a.srv.Config(resp, req)
		require.NoError(t, err)

		value := obj.(structs.ConfigEntry)
		require.Equal(t, structs.ServiceDefaults, value.GetKind())
		entry := value.(*structs.ServiceConfigEntry)
		require.Equal(t, entry.Name, "foo")
	})
	t.Run("list both service entries", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/config/service-defaults", nil)
		resp := httptest.NewRecorder()
		obj, err := a.srv.Config(resp, req)
		require.NoError(t, err)

		value := obj.([]structs.ConfigEntry)
		require.Len(t, value, 2)
		require.Equal(t, value[0].(*structs.ServiceConfigEntry).Name, "bar")
		require.Equal(t, value[1].(*structs.ServiceConfigEntry).Name, "foo")
	})
	t.Run("get global proxy config", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/config/proxy-defaults/global", nil)
		resp := httptest.NewRecorder()
		obj, err := a.srv.Config(resp, req)
		require.NoError(t, err)

		value := obj.(structs.ConfigEntry)
		require.Equal(t, value.GetKind(), structs.ProxyDefaults)
		entry := value.(*structs.ProxyConfigEntry)
		require.Equal(t, structs.ProxyConfigGlobal, entry.Name)
		require.Contains(t, entry.Config, "foo")
		require.Equal(t, "bar", entry.Config["foo"])
	})
	t.Run("error on no arguments", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/config/", nil)
		resp := httptest.NewRecorder()
		_, err := a.srv.Config(resp, req)
		require.Error(t, errors.New("Must provide either a kind or both kind and name"), err)
	})
	t.Run("get the single mesh config", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/config/mesh/mesh", nil)
		resp := httptest.NewRecorder()
		obj, err := a.srv.Config(resp, req)
		require.NoError(t, err)

		ce, ok := obj.(*structs.MeshConfigEntry)
		require.True(t, ok, "wrong type %T", obj)
		// Set indexes and EnterpriseMeta to expected values for assertions
		ce.CreateIndex = 12
		ce.ModifyIndex = 13
		ce.Hash = 0
		ce.EnterpriseMeta = acl.EnterpriseMeta{}

		out, err := a.srv.marshalJSON(req, obj)
		require.NoError(t, err)

		expected := `
{
	"Kind": "mesh",
	"TransparentProxy": {
		"MeshDestinationsOnly": true
	},
	"Meta":{
		"key1": "value1",
		"key2": "value2"
	},
	"CreateIndex": 12,
	"ModifyIndex": 13
}
`
		require.JSONEq(t, expected, string(out))
	})
}

func TestConfig_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	reqs := []structs.ConfigEntryRequest{
		{
			Datacenter: "dc1",
			Entry: &structs.ServiceConfigEntry{
				Name: "foo",
			},
		},
		{
			Datacenter: "dc1",
			Entry: &structs.ServiceConfigEntry{
				Name: "bar",
			},
		},
	}
	for _, req := range reqs {
		out := false
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Apply", &req, &out))
	}

	// Delete an entry.
	{
		req, _ := http.NewRequest("DELETE", "/v1/config/service-defaults/bar", nil)
		resp := httptest.NewRecorder()
		_, err := a.srv.Config(resp, req)
		require.NoError(t, err)
	}
	// Get the remaining entry.
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.ServiceDefaults,
			Datacenter: "dc1",
		}
		var out structs.IndexedConfigEntries
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.List", &args, &out))
		require.Equal(t, structs.ServiceDefaults, out.Kind)
		require.Len(t, out.Entries, 1)
		entry := out.Entries[0].(*structs.ServiceConfigEntry)
		require.Equal(t, entry.Name, "foo")
	}
}

func TestConfig_Delete_CAS(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}
	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create a config entry.
	entry := &structs.ServiceConfigEntry{
		Kind: structs.ServiceDefaults,
		Name: "foo",
	}
	var created bool
	require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Apply", &structs.ConfigEntryRequest{
		Datacenter: "dc1",
		Entry:      entry,
	}, &created))
	require.True(t, created)

	// Read it back to get its ModifyIndex.
	var out structs.ConfigEntryResponse
	require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &structs.ConfigEntryQuery{
		Datacenter: "dc1",
		Kind:       entry.Kind,
		Name:       entry.Name,
	}, &out))
	require.NotNil(t, out.Entry)

	modifyIndex := out.Entry.GetRaftIndex().ModifyIndex

	t.Run("attempt to delete with an invalid index", func(t *testing.T) {
		req := httptest.NewRequest(
			"DELETE",
			fmt.Sprintf("/v1/config/%s/%s?cas=%d", entry.Kind, entry.Name, modifyIndex-1),
			nil,
		)
		rawRsp, err := a.srv.Config(httptest.NewRecorder(), req)
		require.NoError(t, err)

		deleted, isBool := rawRsp.(bool)
		require.True(t, isBool, "response should be a boolean")
		require.False(t, deleted, "entry should not have been deleted")

		// Verify it was not deleted.
		var out structs.ConfigEntryResponse
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &structs.ConfigEntryQuery{
			Datacenter: "dc1",
			Kind:       entry.Kind,
			Name:       entry.Name,
		}, &out))
		require.NotNil(t, out.Entry)
	})

	t.Run("attempt to delete with a valid index", func(t *testing.T) {
		req := httptest.NewRequest(
			"DELETE",
			fmt.Sprintf("/v1/config/%s/%s?cas=%d", entry.Kind, entry.Name, modifyIndex),
			nil,
		)
		rawRsp, err := a.srv.Config(httptest.NewRecorder(), req)
		require.NoError(t, err)

		deleted, isBool := rawRsp.(bool)
		require.True(t, isBool, "response should be a boolean")
		require.True(t, deleted, "entry should have been deleted")

		// Verify it was deleted.
		var out structs.ConfigEntryResponse
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &structs.ConfigEntryQuery{
			Datacenter: "dc1",
			Kind:       entry.Kind,
			Name:       entry.Name,
		}, &out))
		require.Nil(t, out.Entry)
	})
}

func TestConfig_Apply(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "service-defaults",
		"Name": "foo",
		"Protocol": "tcp"
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	if resp.Code != 200 {
		t.Fatalf(resp.Body.String())
	}

	// Get the remaining entry.
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.ServiceDefaults,
			Name:       "foo",
			Datacenter: "dc1",
		}
		var out structs.ConfigEntryResponse
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, &out))
		require.NotNil(t, out.Entry)
		entry := out.Entry.(*structs.ServiceConfigEntry)
		require.Equal(t, entry.Name, "foo")
	}
}

func TestConfig_Apply_TerminatingGateway(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "terminating-gateway",
		"Name": "west-gw-01",
		"Services": [
		  {
			"Name": "web",
			"CAFile": "/etc/web/ca.crt",
			"CertFile": "/etc/web/client.crt",
			"KeyFile": "/etc/web/tls.key"
		  },
		  {
			"Name": "api"
		  }
		]
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code, "!200 Response Code: %s", resp.Body.String())

	// List all entries, there should only be one
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.TerminatingGateway,
			Datacenter: "dc1",
		}
		var out structs.IndexedConfigEntries
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.List", &args, &out))
		require.NotNil(t, out)
		require.Len(t, out.Entries, 1)

		got := out.Entries[0].(*structs.TerminatingGatewayConfigEntry)
		expect := []structs.LinkedService{
			{
				Name:           "web",
				CAFile:         "/etc/web/ca.crt",
				CertFile:       "/etc/web/client.crt",
				KeyFile:        "/etc/web/tls.key",
				EnterpriseMeta: *structs.DefaultEnterpriseMetaInDefaultPartition(),
			},
			{
				Name:           "api",
				EnterpriseMeta: *structs.DefaultEnterpriseMetaInDefaultPartition(),
			},
		}
		require.Equal(t, expect, got.Services)
	}
}

func TestConfig_Apply_IngressGateway(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "ingress-gateway",
		"Name": "ingress",
		"Listeners": [
		  {
				"Port": 8080,
				"Services": [
					{ "Name": "web" }
				]
		  }
		]
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code, "!200 Response Code: %s", resp.Body.String())

	// List all entries, there should only be one
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.IngressGateway,
			Datacenter: "dc1",
		}
		var out structs.IndexedConfigEntries
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.List", &args, &out))
		require.NotNil(t, out)
		require.Len(t, out.Entries, 1)

		got := out.Entries[0].(*structs.IngressGatewayConfigEntry)
		// Ignore create and modify indices
		got.CreateIndex = 0
		got.ModifyIndex = 0
		got.Hash = 0

		expect := &structs.IngressGatewayConfigEntry{
			Name: "ingress",
			Kind: structs.IngressGateway,
			Listeners: []structs.IngressListener{
				{
					Port:     8080,
					Protocol: "tcp",
					Services: []structs.IngressService{
						{
							Name:           "web",
							EnterpriseMeta: *structs.DefaultEnterpriseMetaInDefaultPartition(),
						},
					},
				},
			},
			EnterpriseMeta: *structs.DefaultEnterpriseMetaInDefaultPartition(),
		}
		require.Equal(t, expect, got)
	}
}

func TestConfig_Apply_ProxyDefaultsMeshGateway(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "proxy-defaults",
		"Name": "global",
		"MeshGateway": {
			"Mode": "local"
		}
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code, "!200 Response Code: %s", resp.Body.String())

	// Get the remaining entry.
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.ProxyDefaults,
			Name:       "global",
			Datacenter: "dc1",
		}
		var out structs.ConfigEntryResponse
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, &out))
		require.NotNil(t, out.Entry)
		entry := out.Entry.(*structs.ProxyConfigEntry)
		require.Equal(t, structs.MeshGatewayModeLocal, entry.MeshGateway.Mode)
	}
}

func TestConfig_Apply_ProxyDefaultsProtocol(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	writeConf := func(body string) {
		req, _ := http.NewRequest("PUT", "/v1/config", bytes.NewBuffer([]byte(body)))
		resp := httptest.NewRecorder()
		_, err := a.srv.ConfigApply(resp, req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.Code, "non-200 Response Code: %s", resp.Body.String())
	}

	// Set the default protocol
	writeConf(`{
		"Kind": "proxy-defaults",
		"Name": "global",
		"Config": {
			"Protocol": "http"
		}
	}`)

	// Create a router that depends on the protocol
	writeConf(`{
		"Kind": "service-router",
		"Name": "route1"
	}`)

	// Ensure we can rewrite the proxy-defaults without a protocol-mismatch error.
	// This should be taken care of in the ProxyConfigEntry.Normalize() function.
	writeConf(`{
		"Kind": "proxy-defaults",
		"Name": "global",
		"Config": {
			"Protocol": "http",
			"some-field": "is_changed"
		}
	}`)

	// Rewrite the router that depends on the protocol
	writeConf(`{
		"Kind": "service-router",
		"Name": "route1"
	}`)
}

func TestConfig_Apply_CAS(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "service-defaults",
		"Name": "foo",
		"Protocol": "tcp"
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	if resp.Code != 200 {
		t.Fatalf(resp.Body.String())
	}

	// Get the entry remaining entry.
	args := structs.ConfigEntryQuery{
		Kind:       structs.ServiceDefaults,
		Name:       "foo",
		Datacenter: "dc1",
	}

	out := &structs.ConfigEntryResponse{}
	require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, out))
	require.NotNil(t, out.Entry)
	entry := out.Entry.(*structs.ServiceConfigEntry)

	body = bytes.NewBuffer([]byte(`
	{
		"Kind": "service-defaults",
		"Name": "foo",
		"Protocol": "http"
	}
	`))
	req, _ = http.NewRequest("PUT", "/v1/config?cas=0", body)
	resp = httptest.NewRecorder()
	writtenRaw, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	written, ok := writtenRaw.(bool)
	require.True(t, ok)
	require.False(t, written)
	require.EqualValues(t, 200, resp.Code, resp.Body.String())

	body = bytes.NewBuffer([]byte(`
	{
		"Kind": "service-defaults",
		"Name": "foo",
		"Protocol": "http"
	}
	`))
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/v1/config?cas=%d", entry.GetRaftIndex().ModifyIndex), body)
	resp = httptest.NewRecorder()
	writtenRaw, err = a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	written, ok = writtenRaw.(bool)
	require.True(t, ok)
	require.True(t, written)
	require.EqualValues(t, 200, resp.Code, resp.Body.String())

	// Get the entry remaining entry.
	args = structs.ConfigEntryQuery{
		Kind:       structs.ServiceDefaults,
		Name:       "foo",
		Datacenter: "dc1",
	}

	out = &structs.ConfigEntryResponse{}
	require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, out))
	require.NotNil(t, out.Entry)
	newEntry := out.Entry.(*structs.ServiceConfigEntry)
	require.NotEqual(t, entry.GetRaftIndex(), newEntry.GetRaftIndex())
}

func TestConfig_Apply_Decoding(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	t.Run("No Kind", func(t *testing.T) {
		body := bytes.NewBuffer([]byte(
			`{
			"Name": "foo",
			"Protocol": "tcp"
		}`))

		req, _ := http.NewRequest("PUT", "/v1/config", body)
		resp := httptest.NewRecorder()

		_, err := a.srv.ConfigApply(resp, req)
		require.Error(t, err)
		require.True(t, isHTTPBadRequest(err))
		require.Equal(t, "Request decoding failed: Payload does not contain a kind/Kind key at the top level", err.Error())
	})

	t.Run("Kind Not String", func(t *testing.T) {
		body := bytes.NewBuffer([]byte(
			`{
			"Kind": 123,
			"Name": "foo",
			"Protocol": "tcp"
		}`))

		req, _ := http.NewRequest("PUT", "/v1/config", body)
		resp := httptest.NewRecorder()

		_, err := a.srv.ConfigApply(resp, req)
		require.Error(t, err)
		require.True(t, isHTTPBadRequest(err))
		require.Equal(t, "Request decoding failed: Kind value in payload is not a string", err.Error())
	})

	t.Run("Lowercase kind", func(t *testing.T) {
		body := bytes.NewBuffer([]byte(
			`{
			"kind": "service-defaults",
			"Name": "foo",
			"Protocol": "tcp"
		}`))

		req, _ := http.NewRequest("PUT", "/v1/config", body)
		resp := httptest.NewRecorder()
		_, err := a.srv.ConfigApply(resp, req)
		require.NoError(t, err)
		require.EqualValues(t, 200, resp.Code, resp.Body.String())

		// Get the remaining entry.
		{
			args := structs.ConfigEntryQuery{
				Kind:       structs.ServiceDefaults,
				Name:       "foo",
				Datacenter: "dc1",
			}
			var out structs.ConfigEntryResponse
			require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, &out))
			require.NotNil(t, out.Entry)
			entry := out.Entry.(*structs.ServiceConfigEntry)
			require.Equal(t, entry.Name, "foo")
		}
	})
}

func TestConfig_Apply_ProxyDefaultsExpose(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()

	a := NewTestAgent(t, "")
	defer a.Shutdown()
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")

	// Create some config entries.
	body := bytes.NewBuffer([]byte(`
	{
		"Kind": "proxy-defaults",
		"Name": "global",
		"Expose": {
			"Checks": true,
			"Paths": [
				{
					"LocalPathPort": 8080,
					"ListenerPort": 21500,
					"Path": "/healthz",
					"Protocol": "http2"
				}
			]
		}
	}`))

	req, _ := http.NewRequest("PUT", "/v1/config", body)
	resp := httptest.NewRecorder()
	_, err := a.srv.ConfigApply(resp, req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code, "!200 Response Code: %s", resp.Body.String())

	// Get the remaining entry.
	{
		args := structs.ConfigEntryQuery{
			Kind:       structs.ProxyDefaults,
			Name:       "global",
			Datacenter: "dc1",
		}
		var out structs.ConfigEntryResponse
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Get", &args, &out))
		require.NotNil(t, out.Entry)
		entry := out.Entry.(*structs.ProxyConfigEntry)

		expose := structs.ExposeConfig{
			Checks: true,
			Paths: []structs.ExposePath{
				{
					LocalPathPort: 8080,
					ListenerPort:  21500,
					Path:          "/healthz",
					Protocol:      "http2",
				},
			},
		}
		require.Equal(t, expose, entry.Expose)
	}
}

func TestConfig_Exported_Services(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	a := NewTestAgent(t, "")
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")
	defer a.Shutdown()

	{
		// Register exported services
		args := &structs.ExportedServicesConfigEntry{
			Name: "default",
			Services: []structs.ExportedService{
				{
					Name: "api",
					Consumers: []structs.ServiceConsumer{
						{
							Peer: "east",
						},
						{
							Peer: "west",
						},
					},
				},
				{
					Name: "db",
					Consumers: []structs.ServiceConsumer{
						{
							Peer: "east",
						},
					},
				},
			},
		}
		req := structs.ConfigEntryRequest{
			Datacenter: "dc1",
			Entry:      args,
		}
		var configOutput bool
		require.NoError(t, a.RPC(context.Background(), "ConfigEntry.Apply", &req, &configOutput))
		require.True(t, configOutput)
	}

	t.Run("exported services", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/exported-services", nil)
		resp := httptest.NewRecorder()
		raw, err := a.srv.ExportedServices(resp, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.Code)

		services, ok := raw.([]api.ResolvedExportedService)
		require.True(t, ok)
		require.Len(t, services, 2)
		assertIndex(t, resp)

		entMeta := acl.DefaultEnterpriseMeta()

		expected := []api.ResolvedExportedService{
			{
				Service:   "api",
				Partition: entMeta.PartitionOrEmpty(),
				Namespace: entMeta.NamespaceOrEmpty(),
				Consumers: api.ResolvedConsumers{
					Peers: []string{"east", "west"},
				},
			},
			{
				Service:   "db",
				Partition: entMeta.PartitionOrEmpty(),
				Namespace: entMeta.NamespaceOrEmpty(),
				Consumers: api.ResolvedConsumers{
					Peers: []string{"east"},
				},
			},
		}
		require.Equal(t, expected, services)
	})
}

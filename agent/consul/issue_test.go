// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"

	consulfsm "github.com/dhiaayachi/consul/agent/consul/fsm"
	"github.com/dhiaayachi/consul/agent/consul/state"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/api"
)

func makeLog(buf []byte) *raft.Log {
	return &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  buf,
	}
}

// Testing for GH-300 and GH-279
func TestHealthCheckRace(t *testing.T) {
	t.Parallel()
	fsm := consulfsm.NewFromDeps(consulfsm.Deps{
		Logger: hclog.New(nil),
		NewStateStore: func() *state.Store {
			return state.NewStateStore(nil)
		},
		StorageBackend: consulfsm.NullStorageBackend,
	})
	state := fsm.State()

	req := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Node:      "foo",
			CheckID:   "db",
			Name:      "db connectivity",
			Status:    api.HealthPassing,
			ServiceID: "db",
		},
	}
	buf, err := structs.Encode(structs.RegisterRequestType, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	log := makeLog(buf)
	log.Index = 10
	resp := fsm.Apply(log)
	if resp != nil {
		t.Fatalf("resp: %v", resp)
	}

	// Verify the index
	idx, out1, err := state.CheckServiceNodes(nil, "db", nil, "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 10 {
		t.Fatalf("Bad index: %d", idx)
	}

	// Update the check state
	req.Check.Status = api.HealthCritical
	buf, err = structs.Encode(structs.RegisterRequestType, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	log = makeLog(buf)
	log.Index = 20
	resp = fsm.Apply(log)
	if resp != nil {
		t.Fatalf("resp: %v", resp)
	}

	// Verify the index changed
	idx, out2, err := state.CheckServiceNodes(nil, "db", nil, "")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 20 {
		t.Fatalf("Bad index: %d", idx)
	}

	if reflect.DeepEqual(out1, out2) {
		t.Fatalf("match: %#v %#v", *out1[0].Checks[0], *out2[0].Checks[0])
	}
}

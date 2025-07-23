package structs

import (
	"github.com/dhiaayachi/consul/acl"
)

type LLMAgentConfigEntry struct {
	// Kind of the config entry. This should be set to api.IngressGateway.
	Kind string

	// Name is used to match the config entry with its associated ingress gateway
	// service. This should match the name provided in the service definition.
	Name string

	// Partition is the partition the IngressGateway is associated with.
	// Partitioning is a Consul Enterprise feature.
	Partition string `json:",omitempty"`

	// Namespace is the namespace the IngressGateway is associated with.
	// Namespacing is a Consul Enterprise feature.
	Namespace string `json:",omitempty"`

	Meta map[string]string `json:",omitempty"`

	// CreateIndex is the Raft index this entry was created at. This is a
	// read-only field.
	CreateIndex uint64

	// ModifyIndex is used for the Check-And-Set operations and can also be fed
	// back into the WaitIndex of the QueryOptions in order to perform blocking
	// queries.
	ModifyIndex uint64

	Description string `json:",omitempty"`

	Servers []MCPServer `json:",omitempty"`

	Address string `json:",omitempty"`
}

type LLMAgentExternalServersConfigEntry struct {
	// Kind of the config entry. This should be set to api.IngressGateway.
	Kind string

	// Name is used to match the config entry with its associated ingress gateway
	// service. This should match the name provided in the service definition.
	Name string

	// Partition is the partition the IngressGateway is associated with.
	// Partitioning is a Consul Enterprise feature.
	Partition string `json:",omitempty"`

	// Namespace is the namespace the IngressGateway is associated with.
	// Namespacing is a Consul Enterprise feature.
	Namespace string `json:",omitempty"`

	Meta map[string]string `json:",omitempty"`

	// CreateIndex is the Raft index this entry was created at. This is a
	// read-only field.
	CreateIndex uint64

	// ModifyIndex is used for the Check-And-Set operations and can also be fed
	// back into the WaitIndex of the QueryOptions in order to perform blocking
	// queries.
	ModifyIndex uint64

	Servers []ExternalServer `json:",omitempty"`
}

func (L LLMAgentExternalServersConfigEntry) GetKind() string {
	return L.Kind
}

func (L LLMAgentExternalServersConfigEntry) GetName() string {
	return L.Name
}

func (L LLMAgentExternalServersConfigEntry) Normalize() error {
	return nil
}

func (L LLMAgentExternalServersConfigEntry) Validate() error {
	return nil
}

func (L LLMAgentExternalServersConfigEntry) CanRead(authorizer acl.Authorizer) error {
	return nil
}

func (L LLMAgentExternalServersConfigEntry) CanWrite(authorizer acl.Authorizer) error {
	return nil
}

func (L LLMAgentExternalServersConfigEntry) GetMeta() map[string]string {
	return L.Meta
}

func (L LLMAgentExternalServersConfigEntry) GetEnterpriseMeta() *acl.EnterpriseMeta {
	return &acl.EnterpriseMeta{}
}

func (L LLMAgentExternalServersConfigEntry) GetRaftIndex() *RaftIndex {
	return L.GetRaftIndex()
}

func (L LLMAgentExternalServersConfigEntry) GetHash() uint64 {
	return 0
}

func (L LLMAgentExternalServersConfigEntry) SetHash(h uint64) {
	return
}

type MCPServer struct {
	Name  string `json:",omitempty"`
	Tools []Tool `json:",omitempty"`
	Error string `json:",omitempty"`
}

type ExternalServer struct {
	Name   string `json:",omitempty"`
	Config string `json:",omitempty"`
}

type Tool struct {
	Name        string `json:",omitempty"`
	Description string `json:",omitempty"`
}

func (L LLMAgentConfigEntry) GetKind() string {
	return L.Kind
}

func (L LLMAgentConfigEntry) GetName() string {
	return L.Name
}

func (L LLMAgentConfigEntry) Normalize() error {
	return nil
}

func (L LLMAgentConfigEntry) Validate() error {
	return nil
}

func (L LLMAgentConfigEntry) CanRead(authorizer acl.Authorizer) error {
	return nil
}

func (L LLMAgentConfigEntry) CanWrite(authorizer acl.Authorizer) error {
	return nil
}

func (L LLMAgentConfigEntry) GetMeta() map[string]string {
	return L.Meta
}

func (L LLMAgentConfigEntry) GetEnterpriseMeta() *acl.EnterpriseMeta {
	return &acl.EnterpriseMeta{}
}

func (L LLMAgentConfigEntry) GetRaftIndex() *RaftIndex {
	return &RaftIndex{CreateIndex: L.CreateIndex, ModifyIndex: L.ModifyIndex}
}

func (L LLMAgentConfigEntry) GetHash() uint64 {
	return 0
}

func (L LLMAgentConfigEntry) SetHash(h uint64) {
	return
}

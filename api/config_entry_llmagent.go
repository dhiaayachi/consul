package api

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

type MCPServer struct {
	Name  string `json:",omitempty"`
	Tools []Tool `json:",omitempty"`
	Error string `json:",omitempty"`
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

func (L LLMAgentConfigEntry) GetPartition() string {
	return L.Partition
}

func (L LLMAgentConfigEntry) GetNamespace() string {
	return L.Namespace
}

func (L LLMAgentConfigEntry) GetMeta() map[string]string {
	return L.Meta
}

func (L LLMAgentConfigEntry) GetCreateIndex() uint64 {
	return L.CreateIndex
}

func (L LLMAgentConfigEntry) GetModifyIndex() uint64 {
	return L.ModifyIndex
}

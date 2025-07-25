// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package structs

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/hashstructure"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/cache"
	"github.com/dhiaayachi/consul/lib"

	"golang.org/x/crypto/blake2b"
)

const (
	// IntentionDefaultNamespace is the default namespace value.
	// NOTE(mitchellh): This is only meant to be a temporary constant.
	// When namespaces are introduced, we should delete this constant and
	// fix up all the places where this was used with the proper namespace
	// value.
	IntentionDefaultNamespace = "default"

	IntentionDefaultPolicyAllow = "allow"
	IntentionDefaultPolicyDeny  = "deny"
)

// Intention defines an intention for the Connect Service Graph. This defines
// the allowed or denied behavior of a connection between two services using
// Connect.
type Intention struct {
	// ID is the UUID-based ID for the intention, always generated by Consul.
	ID string `json:",omitempty"`

	// Description is a human-friendly description of this intention.
	// It is opaque to Consul and is only stored and transferred in API
	// requests.
	Description string `json:",omitempty"`

	// SourceNS, SourceName are the namespace and name, respectively, of
	// the source service. Either of these may be the wildcard "*", but only
	// the full value can be a wildcard. Partial wildcards are not allowed.
	// The source may also be a non-Consul service, as specified by SourceType.
	//
	// DestinationNS, DestinationName is the same, but for the destination
	// service. The same rules apply. The destination is always a Consul
	// service.
	SourceNS, SourceName           string
	DestinationNS, DestinationName string

	// SourcePartition and DestinationPartition cannot be wildcards "*" and
	// are not compatible with legacy intentions.
	SourcePartition      string `json:",omitempty"`
	DestinationPartition string `json:",omitempty"`

	// SourcePeer cannot be a wildcard "*" and is not compatible with legacy
	// intentions. Cannot be used with SourcePartition, as both represent the
	// same level of tenancy (partition is local to cluster, peer is remote).
	SourcePeer string `json:",omitempty"`

	// SourceSamenessGroup cannot be a wildcard "*" and is not compatible with legacy
	// intentions. Cannot be used with SourcePartition, as both represent the
	// same level of tenancy (sameness group includes both partitions and cluster peers).
	SourceSamenessGroup string `json:",omitempty"`

	// SourceType is the type of the value for the source.
	SourceType IntentionSourceType

	// Action is whether this is an allowlist or denylist intention.
	Action IntentionAction `json:",omitempty"`

	// Permissions is the list of additional L7 attributes that extend the
	// intention definition.
	//
	// NOTE: This field is not editable unless editing the underlying
	// service-intentions config entry directly.
	Permissions []*IntentionPermission `bexpr:"-" json:",omitempty"`

	// JWT specifies JWT authn that applies to incoming requests.
	JWT *IntentionJWTRequirement `bexpr:"-" json:",omitempty"`

	// DefaultAddr is not used.
	// Deprecated: DefaultAddr is not used and may be removed in a future version.
	DefaultAddr string `bexpr:"-" codec:",omitempty" json:",omitempty"`
	// DefaultPort is not used.
	// Deprecated: DefaultPort is not used and may be removed in a future version.
	DefaultPort int `bexpr:"-" codec:",omitempty" json:",omitempty"`

	// Meta is arbitrary metadata associated with the intention. This is
	// opaque to Consul but is served in API responses.
	Meta map[string]string `json:",omitempty"`

	// Precedence is the order that the intention will be applied, with
	// larger numbers being applied first. This is a read-only field, on
	// any intention update it is updated.
	Precedence int

	// CreatedAt and UpdatedAt keep track of when this record was created
	// or modified.
	CreatedAt, UpdatedAt time.Time `mapstructure:"-" bexpr:"-"`

	// Hash of the contents of the intention. This is only necessary for legacy
	// intention replication purposes.
	//
	// This is needed mainly for legacy replication purposes. When replicating
	// from one DC to another keeping the content Hash will allow us to detect
	// content changes more efficiently than checking every single field
	Hash []byte `bexpr:"-" json:",omitempty"`

	RaftIndex `bexpr:"-"`
}

func (t *Intention) Clone() *Intention {
	t2 := *t
	if len(t.Permissions) > 0 {
		t2.Permissions = make([]*IntentionPermission, 0, len(t.Permissions))
		for _, perm := range t.Permissions {
			t2.Permissions = append(t2.Permissions, perm.Clone())
		}
	}
	t2.Meta = cloneStringStringMap(t.Meta)
	t2.Hash = nil
	return &t2
}

func (t *Intention) ToExact() *IntentionQueryExact {
	return &IntentionQueryExact{
		SourcePartition:      t.SourcePartition,
		SourceNS:             t.SourceNS,
		SourceName:           t.SourceName,
		DestinationPartition: t.DestinationPartition,
		DestinationNS:        t.DestinationNS,
		DestinationName:      t.DestinationName,
	}
}

func (t *Intention) MarshalJSON() ([]byte, error) {
	type Alias Intention
	exported := &struct {
		CreatedAt, UpdatedAt *time.Time `json:",omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if !t.CreatedAt.IsZero() {
		exported.CreatedAt = &t.CreatedAt
	}
	if !t.UpdatedAt.IsZero() {
		exported.UpdatedAt = &t.UpdatedAt
	}
	return json.Marshal(exported)
}

func (t *Intention) UnmarshalJSON(data []byte) (err error) {
	type Alias Intention
	aux := &struct {
		Hash                 string
		CreatedAt, UpdatedAt string // effectively `json:"-"` on CreatedAt and UpdatedAt

		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err = lib.UnmarshalJSON(data, &aux); err != nil {
		return err
	}

	if aux.Hash != "" {
		t.Hash = []byte(aux.Hash)
	}
	return nil
}

// SetHash calculates Intention.Hash from any mutable "content" fields.
//
// The Hash is primarily used for legacy intention replication to determine if
// an intention has changed and should be updated locally.
//
// Deprecated: this is only used for legacy intention CRUD and replication
func (x *Intention) SetHash() {
	hash, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}

	// Write all the user set fields
	hash.Write([]byte(x.ID))
	hash.Write([]byte(x.Description))
	hash.Write([]byte(x.SourceNS))
	hash.Write([]byte(x.SourceName))
	hash.Write([]byte(x.DestinationNS))
	hash.Write([]byte(x.DestinationName))
	hash.Write([]byte(x.SourceType))
	hash.Write([]byte(x.Action))
	// hash.Write can not return an error, so the only way for binary.Write to
	// error is to pass it data with an invalid data type. Doing so would be a
	// programming error, so panic in that case.
	if err := binary.Write(hash, binary.LittleEndian, uint64(x.Precedence)); err != nil {
		panic(err)
	}

	// sort keys to ensure hash stability when meta is stored later
	var keys []string
	for k := range x.Meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		hash.Write([]byte(k))
		hash.Write([]byte(x.Meta[k]))
	}

	x.Hash = hash.Sum(nil)
}

// Validate returns an error if the intention is invalid for inserting
// or updating via the legacy APIs.
//
// Deprecated: this is only used for legacy intention CRUD
func (x *Intention) Validate() error {
	var result error

	// Empty values
	if x.SourceNS == "" {
		result = multierror.Append(result, fmt.Errorf("SourceNS must be set"))
	}
	if x.SourceName == "" {
		result = multierror.Append(result, fmt.Errorf("SourceName must be set"))
	}
	if x.DestinationNS == "" {
		result = multierror.Append(result, fmt.Errorf("DestinationNS must be set"))
	}
	if x.DestinationName == "" {
		result = multierror.Append(result, fmt.Errorf("DestinationName must be set"))
	}

	// Wildcard usage verification
	if x.SourceNS != WildcardSpecifier {
		if strings.Contains(x.SourceNS, WildcardSpecifier) {
			result = multierror.Append(result, fmt.Errorf(
				"SourceNS: wildcard character '*' cannot be used with partial values"))
		}
	}
	if x.SourceName != WildcardSpecifier {
		if strings.Contains(x.SourceName, WildcardSpecifier) {
			result = multierror.Append(result, fmt.Errorf(
				"SourceName: wildcard character '*' cannot be used with partial values"))
		}

		if x.SourceNS == WildcardSpecifier {
			result = multierror.Append(result, fmt.Errorf(
				"SourceName: exact value cannot follow wildcard namespace"))
		}
	}
	if x.DestinationNS != WildcardSpecifier {
		if strings.Contains(x.DestinationNS, WildcardSpecifier) {
			result = multierror.Append(result, fmt.Errorf(
				"DestinationNS: wildcard character '*' cannot be used with partial values"))
		}
	}
	if x.DestinationName != WildcardSpecifier {
		if strings.Contains(x.DestinationName, WildcardSpecifier) {
			result = multierror.Append(result, fmt.Errorf(
				"DestinationName: wildcard character '*' cannot be used with partial values"))
		}

		if x.DestinationNS == WildcardSpecifier {
			result = multierror.Append(result, fmt.Errorf(
				"DestinationName: exact value cannot follow wildcard namespace"))
		}
	}

	// Length of opaque values
	if len(x.Description) > metaValueMaxLength {
		result = multierror.Append(result, fmt.Errorf(
			"Description exceeds maximum length %d", metaValueMaxLength))
	}
	if len(x.Meta) > metaMaxKeyPairs {
		result = multierror.Append(result, fmt.Errorf(
			"Meta exceeds maximum element count %d", metaMaxKeyPairs))
	}
	for k, v := range x.Meta {
		if len(k) > metaKeyMaxLength {
			result = multierror.Append(result, fmt.Errorf(
				"Meta key %q exceeds maximum length %d", k, metaKeyMaxLength))
		}
		if len(v) > metaValueMaxLength {
			result = multierror.Append(result, fmt.Errorf(
				"Meta value for key %q exceeds maximum length %d", k, metaValueMaxLength))
		}
	}

	switch x.Action {
	case IntentionActionAllow, IntentionActionDeny:
	default:
		result = multierror.Append(result, fmt.Errorf(
			"Action must be set to 'allow' or 'deny'"))
	}

	if len(x.Permissions) > 0 {
		result = multierror.Append(result, fmt.Errorf(
			"Permissions must not be set when using the legacy APIs"))
	}

	switch x.SourceType {
	case IntentionSourceConsul:
	default:
		result = multierror.Append(result, fmt.Errorf(
			"SourceType must be set to 'consul'"))
	}

	return result
}

func (ixn *Intention) CanRead(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext

	// Read access on either end of the intention allows you to read the
	// complete intention. This is so that both ends can be aware of why
	// something does or does not work.

	// If SourcePeer is set, tenancy is irrelevant in the context of the local cluster
	// so we skip authorizing on the Source end.
	if ixn.SourceName != "" && ixn.SourcePeer == "" {
		ixn.FillAuthzContext(&authzContext, false)
		if authz.IntentionRead(ixn.SourceName, &authzContext) == acl.Allow {
			return true
		}
	}

	if ixn.DestinationName != "" {
		ixn.FillAuthzContext(&authzContext, true)
		if authz.IntentionRead(ixn.DestinationName, &authzContext) == acl.Allow {
			return true
		}
	}

	return false
}

func (ixn *Intention) CanWrite(authz acl.Authorizer) bool {
	if ixn.DestinationName == "" {
		// This is likely a strange form of legacy intention data validation
		// that happened within the authorization check, since intentions without
		// a destination cannot be written.
		// This may be able to be removed later.
		return false
	}

	var authzContext acl.AuthorizerContext
	ixn.FillAuthzContext(&authzContext, true)
	return authz.IntentionWrite(ixn.DestinationName, &authzContext) == acl.Allow
}

// UpdatePrecedence sets the Precedence value based on the fields of this
// structure.
//
// Deprecated: this is only used for legacy intention CRUD.
func (x *Intention) UpdatePrecedence() {
	// Max maintains the maximum value that the precedence can be depending
	// on the number of exact values in the destination.
	var max int
	switch x.countExact(x.DestinationNS, x.DestinationName) {
	case 2:
		max = 9
	case 1:
		max = 6
	case 0:
		max = 3
	default:
		// This shouldn't be possible, just set it to zero
		x.Precedence = 0
		return
	}

	// Given the maximum, the exact value is determined based on the
	// number of source exact values.
	countSrc := x.countExact(x.SourceNS, x.SourceName)
	x.Precedence = max - (2 - countSrc)
}

// countExact counts the number of exact values (not wildcards) in
// the given namespace and name.
func (x *Intention) countExact(ns, n string) int {
	// If NS is wildcard, it must be zero since wildcards only follow exact
	if ns == WildcardSpecifier {
		return 0
	}

	// Same reasoning as above, a wildcard can only follow an exact value
	// and an exact value cannot follow a wildcard, so if name is a wildcard
	// we must have exactly one.
	if n == WildcardSpecifier {
		return 1
	}

	return 2
}

// String returns a human-friendly string for this intention.
func (x *Intention) String() string {
	var idPart string
	if x.ID != "" {
		idPart = "ID: " + x.ID + ", "
	}

	// Cluster may be either partition (local) or peer (remote)
	var srcClusterPart string
	if x.SourcePartition != "" {
		srcClusterPart = x.SourcePartition + "/"
	}
	if x.SourcePeer != "" {
		srcClusterPart = "peer(" + x.SourcePeer + ")/"
	}
	if x.SourceSamenessGroup != "" {
		srcClusterPart = "sameness-group(" + x.SourceSamenessGroup + ")/"
	}

	var dstPartitionPart string
	if x.DestinationPartition != "" {
		dstPartitionPart = x.DestinationPartition + "/"
	}

	var detailPart string
	if len(x.Permissions) > 0 {
		detailPart = fmt.Sprintf("Permissions: %d", len(x.Permissions))
	} else {
		detailPart = "Action: " + strings.ToUpper(string(x.Action))
	}

	return fmt.Sprintf("%s%s/%s => %s%s/%s (%sPrecedence: %d, %s)",
		srcClusterPart, x.SourceNS, x.SourceName,
		dstPartitionPart, x.DestinationNS, x.DestinationName,
		idPart,
		x.Precedence,
		detailPart,
	)
}

// LegacyEstimateSize returns an estimate (in bytes) of the size of this structure when encoded.
//
// Deprecated: only exists for legacy intention replication during migration to 1.9.0+ cluster.
func (x *Intention) LegacyEstimateSize() int {
	// 56 = 36 (uuid) + 16 (RaftIndex) + 4 (Precedence)
	size := 56 + len(x.Description) + len(x.SourceNS) + len(x.SourceName) + len(x.DestinationNS) +
		len(x.DestinationName) + len(x.SourceType) + len(x.Action)

	for k, v := range x.Meta {
		size += len(k) + len(v)
	}

	return size
}

func (x *Intention) SourceServiceName() ServiceName {
	return NewServiceName(x.SourceName, x.SourceEnterpriseMeta())
}

func (x *Intention) DestinationServiceName() ServiceName {
	return NewServiceName(x.DestinationName, x.DestinationEnterpriseMeta())
}

// NOTE this is just used to manipulate user-provided data before an insert
// The RPC execution will do Normalize + Validate for us.
func (x *Intention) ToConfigEntry(legacy bool) *ServiceIntentionsConfigEntry {
	return &ServiceIntentionsConfigEntry{
		Kind:           ServiceIntentions,
		Name:           x.DestinationName,
		EnterpriseMeta: *x.DestinationEnterpriseMeta(),
		Sources:        []*SourceIntention{x.ToSourceIntention(legacy)},
	}
}

func (x *Intention) ToSourceIntention(legacy bool) *SourceIntention {
	ct := x.CreatedAt // copy
	ut := x.UpdatedAt

	src := &SourceIntention{
		Name:             x.SourceName,
		EnterpriseMeta:   *x.SourceEnterpriseMeta(),
		Peer:             x.SourcePeer,
		SamenessGroup:    x.SourceSamenessGroup,
		Action:           x.Action,
		Permissions:      nil, // explicitly not symmetric with the old APIs
		Precedence:       0,   // Ignore, let it be computed.
		LegacyID:         x.ID,
		Type:             x.SourceType,
		Description:      x.Description,
		LegacyMeta:       x.Meta,
		LegacyCreateTime: &ct,
		LegacyUpdateTime: &ut,
	}
	if !legacy {
		src.Permissions = x.Permissions
	}
	return src
}

// IntentionAction is the action that the intention represents. This
// can be "allow" or "deny".
type IntentionAction string

const (
	IntentionActionAllow IntentionAction = "allow"
	IntentionActionDeny  IntentionAction = "deny"
)

// IntentionSourceType is the type of the source within an intention.
type IntentionSourceType string

const (
	// IntentionSourceConsul is a service within the Consul catalog.
	IntentionSourceConsul IntentionSourceType = "consul"
)

type IntentionTargetType string

const (
	// IntentionTargetService is a service within the Consul catalog.
	IntentionTargetService IntentionTargetType = "service"
	// IntentionTargetDestination is a destination defined through a service-default config entry.
	IntentionTargetDestination IntentionTargetType = "destination"
)

// Intentions is a list of intentions.
type Intentions []*Intention

// IndexedIntentions represents a list of intentions for RPC responses.
type IndexedIntentions struct {
	Intentions Intentions

	// DataOrigin is used to indicate if this query was satisfied against the
	// old legacy intentions ("legacy") memdb table or via config entries
	// ("config"). This is really only of value for the legacy intention
	// replication routine to correctly detect that it should exit.
	DataOrigin string `json:"-"`
	QueryMeta
}

const (
	IntentionDataOriginLegacy        = "legacy"
	IntentionDataOriginConfigEntries = "config"
)

// IndexedIntentionMatches represents the list of matches for a match query.
type IndexedIntentionMatches struct {
	Matches []Intentions
	QueryMeta
}

// IntentionOp is the operation for a request related to intentions.
type IntentionOp string

const (
	IntentionOpCreate    IntentionOp = "create"
	IntentionOpUpdate    IntentionOp = "update"
	IntentionOpDelete    IntentionOp = "delete"
	IntentionOpDeleteAll IntentionOp = "delete-all" // NOTE: this is only accepted when it comes from the leader, RPCs will reject this
	IntentionOpUpsert    IntentionOp = "upsert"     // config-entry only
)

// IntentionRequest is used to create, update, and delete intentions.
type IntentionRequest struct {
	// Datacenter is the target for this request.
	Datacenter string

	// Op is the type of operation being requested.
	Op IntentionOp

	// Intention is the intention.
	//
	// This is mutually exclusive with the Mutation field.
	Intention *Intention

	// Mutation is a change to make to an Intention.
	//
	// This is mutually exclusive with the Intention field.
	//
	// This field is only set by the leader before writing to the raft log and
	// is not settable via the API or an RPC.
	Mutation *IntentionMutation

	// WriteRequest is a common struct containing ACL tokens and other
	// write-related common elements for requests.
	WriteRequest
}

type IntentionMutation struct {
	ID          string
	Destination ServiceName
	Source      ServiceName
	// TODO(peering): check if this needs peer field
	Value *SourceIntention
}

// RequestDatacenter returns the datacenter for a given request.
func (q *IntentionRequest) RequestDatacenter() string {
	return q.Datacenter
}

// IntentionMatchType is the target for a match request. For example,
// matching by source will look for all intentions that match the given
// source value.
type IntentionMatchType string

const (
	IntentionMatchSource      IntentionMatchType = "source"
	IntentionMatchDestination IntentionMatchType = "destination"
)

// IntentionQueryRequest is used to query intentions.
type IntentionQueryRequest struct {
	// Datacenter is the target this request is intended for.
	Datacenter string

	// IntentionID is the ID of a specific intention.
	IntentionID string

	// Match is non-nil if we're performing a match query. A match will
	// find intentions that "match" the given parameters. A match includes
	// resolving wildcards.
	Match *IntentionQueryMatch

	// Check is non-nil if we're performing a test query. A test will
	// return allowed/deny based on an exact match.
	Check *IntentionQueryCheck

	// Exact is non-nil if we're performing a lookup of an intention by its
	// unique name instead of its ID.
	Exact *IntentionQueryExact

	// Options for queries
	QueryOptions
}

// RequestDatacenter returns the datacenter for a given request.
func (q *IntentionQueryRequest) RequestDatacenter() string {
	return q.Datacenter
}

// CacheInfo implements cache.Request
func (q *IntentionQueryRequest) CacheInfo() cache.RequestInfo {
	info := cache.RequestInfo{
		Token:      q.Token,
		Datacenter: q.Datacenter,
		MinIndex:   q.MinQueryIndex,
		Timeout:    q.MaxQueryTime,
	}

	v, err := hashstructure.Hash(struct {
		IntentionID string
		Match       *IntentionQueryMatch
		Check       *IntentionQueryCheck
		Exact       *IntentionQueryExact
		Filter      string
	}{
		IntentionID: q.IntentionID,
		Check:       q.Check,
		Match:       q.Match,
		Exact:       q.Exact,
		Filter:      q.QueryOptions.Filter,
	}, nil)
	if err == nil {
		// If there is an error, we don't set the key. A blank key forces
		// no cache for this request so the request is forwarded directly
		// to the server.
		info.Key = strconv.FormatUint(v, 16)
	}

	return info
}

// IntentionQueryMatch are the parameters for performing a match request
// against the state store.
type IntentionQueryMatch struct {
	Type               IntentionMatchType
	Entries            []IntentionMatchEntry
	WithSamenessGroups bool
}

// IntentionMatchEntry is a single entry for matching an intention.
type IntentionMatchEntry struct {
	Partition string `json:",omitempty"`
	Namespace string
	Name      string
}

// IntentionQueryCheck are the parameters for performing a test request.
type IntentionQueryCheck struct {
	// SourceNS, SourceName, DestinationNS, and DestinationName are the
	// source and namespace, respectively, for the test. These must be
	// exact values.
	SourceNS, SourceName           string
	DestinationNS, DestinationName string

	// TODO(partitions): check query works with partitions
	SourcePartition      string `json:",omitempty"`
	DestinationPartition string `json:",omitempty"`

	// SourceType is the type of the value for the source.
	SourceType IntentionSourceType
}

// GetACLPrefix returns the prefix to look up the ACL policy for this
// request, and a boolean noting whether the prefix is valid to check
// or not. You must check the ok value before using the prefix.
func (q *IntentionQueryCheck) GetACLPrefix() (string, bool) {
	return q.DestinationName, q.DestinationName != ""
}

// IntentionQueryCheckResponse is the response for a test request.
type IntentionQueryCheckResponse struct {
	Allowed bool
}

// IntentionDecisionSummary contains a summary of a set of intentions between two services
// Currently contains:
// - Whether all actions are allowed
// - Whether the matching intention has L7 permissions attached
// - Whether the intention is managed by an external source like k8s
// - Whether there is an exact, or wildcard, intention referencing the two services
// - Whether intentions are in DefaultAllow mode
type IntentionDecisionSummary struct {
	Allowed        bool
	HasPermissions bool
	ExternalSource string
	HasExact       bool
	DefaultAllow   bool
}

// IntentionQueryExact holds the parameters for performing a lookup of an
// intention by its unique name instead of its ID.
type IntentionQueryExact struct {
	SourceNS, SourceName           string
	DestinationNS, DestinationName string

	// TODO(partitions): check query works with partitions
	SourcePartition      string `json:",omitempty"`
	DestinationPartition string `json:",omitempty"`

	SourcePeer          string `json:",omitempty"`
	SourceSamenessGroup string `json:",omitempty"`
}

// Validate is used to ensure all 4 required parameters are specified.
func (q *IntentionQueryExact) Validate() error {
	var err error
	if q.SourceNS == "" {
		err = multierror.Append(err, errors.New("SourceNS is missing"))
	}
	if q.SourceName == "" {
		err = multierror.Append(err, errors.New("SourceName is missing"))
	}
	if q.DestinationNS == "" {
		err = multierror.Append(err, errors.New("DestinationNS is missing"))
	}
	if q.DestinationName == "" {
		err = multierror.Append(err, errors.New("DestinationName is missing"))
	}
	return err
}

// TODO(peering): add support for listing peer
type IntentionListRequest struct {
	Datacenter         string
	Legacy             bool `json:"-"`
	acl.EnterpriseMeta `hcl:",squash" mapstructure:",squash"`
	QueryOptions
}

func (r *IntentionListRequest) RequestDatacenter() string {
	return r.Datacenter
}

// SimplifiedIntentions contains expanded sameness groups.
type SimplifiedIntentions Intentions

// IntentionPrecedenceSorter takes a list of intentions and sorts them
// based on the match precedence rules for intentions. The intentions
// closer to the head of the list have higher precedence. i.e. index 0 has
// the highest precedence.
type IntentionPrecedenceSorter Intentions

func (s IntentionPrecedenceSorter) Len() int { return len(s) }
func (s IntentionPrecedenceSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s IntentionPrecedenceSorter) Less(i, j int) bool {
	a, b := s[i], s[j]
	if a.Precedence != b.Precedence {
		return a.Precedence > b.Precedence
	}

	// Tie break on lexicographic order of the tuple in canonical form:
	//
	//   (SrcSamenessGroup, SrcPeer, SrcPxn, SrcNS, Src, DstPxn, DstNS, Dst)
	//
	// This is arbitrary but it keeps sorting deterministic which is a nice
	// property for consistency. It is arguably open to abuse if implementations
	// rely on this however by definition the order among same-precedence rules
	// is arbitrary and doesn't affect whether an allow or deny rule is acted on
	// since all applicable rules are checked.
	if a.SourceSamenessGroup != b.SourceSamenessGroup {
		return a.SourceSamenessGroup < b.SourceSamenessGroup
	}
	if a.SourcePeer != b.SourcePeer {
		return a.SourcePeer < b.SourcePeer
	}
	if a.SourcePartition != b.SourcePartition {
		return a.SourcePartition < b.SourcePartition
	}
	if a.SourceNS != b.SourceNS {
		return a.SourceNS < b.SourceNS
	}
	if a.SourceName != b.SourceName {
		return a.SourceName < b.SourceName
	}
	if a.DestinationPartition != b.DestinationPartition {
		return a.DestinationPartition < b.DestinationPartition
	}
	if a.DestinationNS != b.DestinationNS {
		return a.DestinationNS < b.DestinationNS
	}
	return a.DestinationName < b.DestinationName
}

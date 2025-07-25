// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package consul

import (
	"fmt"
	"sort"

	"github.com/armon/go-metrics"
	hashstructure_v2 "github.com/mitchellh/hashstructure/v2"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/configentry"
	"github.com/dhiaayachi/consul/agent/consul/state"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/hashicorp/go-bexpr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-memdb"
)

// Health endpoint is used to query the health information
type Health struct {
	srv    *Server
	logger hclog.Logger
}

// ChecksInState is used to get all the checks in a given state
func (h *Health) ChecksInState(args *structs.ChecksInStateRequest,
	reply *structs.IndexedHealthChecks) error {
	if done, err := h.srv.ForwardRPC("Health.ChecksInState", args, reply); done {
		return err
	}

	filter, err := bexpr.CreateFilter(args.Filter, nil, reply.HealthChecks)
	if err != nil {
		return err
	}

	_, err = h.srv.ResolveTokenAndDefaultMeta(args.Token, &args.EnterpriseMeta, nil)
	if err != nil {
		return err
	}

	if err := h.srv.validateEnterpriseRequest(&args.EnterpriseMeta, false); err != nil {
		return err
	}

	return h.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.Store) error {
			var index uint64
			var checks structs.HealthChecks
			var err error
			if len(args.NodeMetaFilters) > 0 {
				index, checks, err = state.ChecksInStateByNodeMeta(ws, args.State, args.NodeMetaFilters, &args.EnterpriseMeta, args.PeerName)
			} else {
				index, checks, err = state.ChecksInState(ws, args.State, &args.EnterpriseMeta, args.PeerName)
			}
			if err != nil {
				return err
			}
			reply.Index, reply.HealthChecks = index, checks

			// Note: we filter the results with ACLs *before* applying the user-supplied
			// bexpr filter to ensure that the user can only run expressions on data that
			// they have access to.  This is a security measure to prevent users from
			// running arbitrary expressions on data they don't have access to.
			// QueryMeta.ResultsFilteredByACLs being true already indicates to the user
			// that results they don't have access to have been removed.  If they were
			// also allowed to run the bexpr filter on the data, they could potentially
			// infer the specific attributes of data they don't have access to.
			if err := h.srv.filterACL(args.Token, reply); err != nil {
				return err
			}

			raw, err := filter.Execute(reply.HealthChecks)
			if err != nil {
				return err
			}
			reply.HealthChecks = raw.(structs.HealthChecks)

			return h.srv.sortNodesByDistanceFrom(args.Source, reply.HealthChecks)
		})
}

// NodeChecks is used to get all the checks for a node
func (h *Health) NodeChecks(args *structs.NodeSpecificRequest,
	reply *structs.IndexedHealthChecks) error {
	if done, err := h.srv.ForwardRPC("Health.NodeChecks", args, reply); done {
		return err
	}

	filter, err := bexpr.CreateFilter(args.Filter, nil, reply.HealthChecks)
	if err != nil {
		return err
	}

	_, err = h.srv.ResolveTokenAndDefaultMeta(args.Token, &args.EnterpriseMeta, nil)
	if err != nil {
		return err
	}

	if err := h.srv.validateEnterpriseRequest(&args.EnterpriseMeta, false); err != nil {
		return err
	}

	return h.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.Store) error {
			index, checks, err := state.NodeChecks(ws, args.Node, &args.EnterpriseMeta, args.PeerName)
			if err != nil {
				return err
			}
			reply.Index, reply.HealthChecks = index, checks

			// Note: we filter the results with ACLs *before* applying the user-supplied
			// bexpr filter to ensure that the user can only run expressions on data that
			// they have access to.  This is a security measure to prevent users from
			// running arbitrary expressions on data they don't have access to.
			// QueryMeta.ResultsFilteredByACLs being true already indicates to the user
			// that results they don't have access to have been removed.  If they were
			// also allowed to run the bexpr filter on the data, they could potentially
			// infer the specific attributes of data they don't have access to.
			if err := h.srv.filterACL(args.Token, reply); err != nil {
				return err
			}

			raw, err := filter.Execute(reply.HealthChecks)
			if err != nil {
				return err
			}
			reply.HealthChecks = raw.(structs.HealthChecks)

			return nil
		})
}

// ServiceChecks is used to get all the checks for a service
func (h *Health) ServiceChecks(args *structs.ServiceSpecificRequest,
	reply *structs.IndexedHealthChecks) error {

	// Reject if tag filtering is on
	if args.TagFilter {
		return fmt.Errorf("Tag filtering is not supported")
	}

	// Potentially forward
	if done, err := h.srv.ForwardRPC("Health.ServiceChecks", args, reply); done {
		return err
	}

	filter, err := bexpr.CreateFilter(args.Filter, nil, reply.HealthChecks)
	if err != nil {
		return err
	}

	_, err = h.srv.ResolveTokenAndDefaultMeta(args.Token, &args.EnterpriseMeta, nil)
	if err != nil {
		return err
	}

	if err := h.srv.validateEnterpriseRequest(&args.EnterpriseMeta, false); err != nil {
		return err
	}

	return h.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.Store) error {
			var index uint64
			var checks structs.HealthChecks
			var err error
			if len(args.NodeMetaFilters) > 0 {
				index, checks, err = state.ServiceChecksByNodeMeta(ws, args.ServiceName, args.NodeMetaFilters, &args.EnterpriseMeta, args.PeerName)
			} else {
				index, checks, err = state.ServiceChecks(ws, args.ServiceName, &args.EnterpriseMeta, args.PeerName)
			}
			if err != nil {
				return err
			}
			reply.Index, reply.HealthChecks = index, checks

			raw, err := filter.Execute(reply.HealthChecks)
			if err != nil {
				return err
			}
			reply.HealthChecks = raw.(structs.HealthChecks)

			// Note: we filter the results with ACLs *after* applying the user-supplied
			// bexpr filter, to ensure QueryMeta.ResultsFilteredByACLs does not include
			// results that would be filtered out even if the user did have permission.
			if err := h.srv.filterACL(args.Token, reply); err != nil {
				return err
			}

			return h.srv.sortNodesByDistanceFrom(args.Source, reply.HealthChecks)
		})
}

// ServiceNodes returns all the nodes registered as part of a service including health info
func (h *Health) ServiceNodes(args *structs.ServiceSpecificRequest, reply *structs.IndexedCheckServiceNodes) error {
	if done, err := h.srv.ForwardRPC("Health.ServiceNodes", args, reply); done {
		return err
	}

	// Verify the arguments
	if args.ServiceName == "" {
		return fmt.Errorf("Must provide service name")
	}

	// Determine the function we'll call
	var f func(memdb.WatchSet, *state.Store, *structs.ServiceSpecificRequest) (uint64, structs.CheckServiceNodes, error)
	switch {
	case args.Connect:
		f = h.serviceNodesConnect
	case args.TagFilter:
		f = h.serviceNodesTagFilter
	case args.Ingress:
		f = h.serviceNodesIngress
	default:
		f = h.serviceNodesDefault
	}

	authzContext := acl.AuthorizerContext{
		Peer: args.PeerName,
	}
	authz, err := h.srv.ResolveTokenAndDefaultMeta(args.Token, &args.EnterpriseMeta, &authzContext)
	if err != nil {
		return err
	}

	if err := h.srv.validateEnterpriseRequest(&args.EnterpriseMeta, false); err != nil {
		return err
	}

	// If we're doing a connect or ingress query, we need read access to the service
	// we're trying to find proxies for, so check that.
	if args.Connect || args.Ingress {
		if authz.ServiceRead(args.ServiceName, &authzContext) != acl.Allow {
			return acl.ErrPermissionDenied
		}
	}

	filter, err := bexpr.CreateFilter(args.Filter, nil, reply.Nodes)
	if err != nil {
		return err
	}

	var (
		priorMergeHash uint64
		ranMergeOnce   bool
	)

	err = h.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.Store) error {
			var thisReply structs.IndexedCheckServiceNodes

			sgIdx, sgArgs, err := h.getArgsForSamenessGroupMembers(args, ws, state)
			if err != nil {
				return err
			}

			for _, arg := range sgArgs {
				index, nodes, err := f(ws, state, arg)
				if err != nil {
					return err
				}

				resolvedNodes := nodes
				if arg.MergeCentralConfig {
					for _, node := range resolvedNodes {
						ns := node.Service
						if ns.IsSidecarProxy() || ns.IsGateway() {
							cfgIndex, mergedns, err := configentry.MergeNodeServiceWithCentralConfig(ws, state, ns, h.logger)
							if err != nil {
								return err
							}
							if cfgIndex > index {
								index = cfgIndex
							}
							*node.Service = *mergedns
						}
					}

					// Generate a hash of the resolvedNodes driving this response.
					// Use it to determine if the response is identical to a prior wakeup.
					newMergeHash, err := hashstructure_v2.Hash(resolvedNodes, hashstructure_v2.FormatV2, nil)
					if err != nil {
						return fmt.Errorf("error hashing reply for spurious wakeup suppression: %w", err)
					}
					if ranMergeOnce && priorMergeHash == newMergeHash {
						// the below assignment is not required as the if condition already validates equality,
						// but makes it more clear that prior value is being reset to the new hash on each run.
						priorMergeHash = newMergeHash
						reply.Index = index
						// NOTE: the prior response is still alive inside of *reply, which is desirable
						return errNotChanged
					} else {
						priorMergeHash = newMergeHash
						ranMergeOnce = true
					}

				}

				thisReply.Index, thisReply.Nodes = index, resolvedNodes

				if len(arg.NodeMetaFilters) > 0 {
					thisReply.Nodes = nodeMetaFilter(arg.NodeMetaFilters, thisReply.Nodes)
				}

				// Note: we filter the results with ACLs *before* applying the user-supplied
				// bexpr filter to ensure that the user can only run expressions on data that
				// they have access to.  This is a security measure to prevent users from
				// running arbitrary expressions on data they don't have access to.
				// QueryMeta.ResultsFilteredByACLs being true already indicates to the user
				// that results they don't have access to have been removed.  If they were
				// also allowed to run the bexpr filter on the data, they could potentially
				// infer the specific attributes of data they don't have access to.
				if err := h.srv.filterACL(arg.Token, &thisReply); err != nil {
					return err
				}

				raw, err := filter.Execute(thisReply.Nodes)
				if err != nil {
					return err
				}
				filteredNodes := raw.(structs.CheckServiceNodes)
				thisReply.Nodes = filteredNodes.Filter(structs.CheckServiceNodeFilterOptions{FilterType: arg.HealthFilterType})

				if err := h.srv.sortNodesByDistanceFrom(arg.Source, thisReply.Nodes); err != nil {
					return err
				}
				if len(thisReply.Nodes) > 0 {
					break
				}
			}

			// If sameness group was used, evaluate the index of the sameness group
			// and update the index of the response if it is greater.  If sameness group is not
			// used, the sgIdx will be 0 in this evaluation.
			if sgIdx > thisReply.Index {
				thisReply.Index = sgIdx
			}

			*reply = thisReply
			return nil
		})

	// Provide some metrics
	if err == nil {
		// For metrics, we separate Connect-based lookups from non-Connect
		key := "service"
		if args.Connect {
			key = "connect"
		}
		if args.Ingress {
			key = "ingress"
		}

		metrics.IncrCounterWithLabels([]string{"health", key, "query"}, 1,
			[]metrics.Label{{Name: "service", Value: args.ServiceName}})
		// DEPRECATED (singular-service-tag) - remove this when backwards RPC compat
		// with 1.2.x is not required.
		if args.ServiceTag != "" {
			metrics.IncrCounterWithLabels([]string{"health", key, "query-tag"}, 1,
				[]metrics.Label{{Name: "service", Value: args.ServiceName}, {Name: "tag", Value: args.ServiceTag}})
		}
		if len(args.ServiceTags) > 0 {
			// Sort tags so that the metric is the same even if the request
			// tags are in a different order
			sort.Strings(args.ServiceTags)

			labels := []metrics.Label{{Name: "service", Value: args.ServiceName}}
			for _, tag := range args.ServiceTags {
				labels = append(labels, metrics.Label{Name: "tag", Value: tag})
			}
			metrics.IncrCounterWithLabels([]string{"health", key, "query-tags"}, 1, labels)
		}
		if len(reply.Nodes) == 0 {
			metrics.IncrCounterWithLabels([]string{"health", key, "not-found"}, 1,
				[]metrics.Label{{Name: "service", Value: args.ServiceName}})
		}
	}
	return err
}

// The serviceNodes* functions below are the various lookup methods that
// can be used by the ServiceNodes endpoint.

func (h *Health) serviceNodesConnect(ws memdb.WatchSet, s *state.Store, args *structs.ServiceSpecificRequest) (uint64, structs.CheckServiceNodes, error) {
	return s.CheckConnectServiceNodes(ws, args.ServiceName, &args.EnterpriseMeta, args.PeerName)
}

func (h *Health) serviceNodesIngress(ws memdb.WatchSet, s *state.Store, args *structs.ServiceSpecificRequest) (uint64, structs.CheckServiceNodes, error) {
	return s.CheckIngressServiceNodes(ws, args.ServiceName, &args.EnterpriseMeta)
}

func (h *Health) serviceNodesTagFilter(ws memdb.WatchSet, s *state.Store, args *structs.ServiceSpecificRequest) (uint64, structs.CheckServiceNodes, error) {
	// DEPRECATED (singular-service-tag) - remove this when backwards RPC compat
	// with 1.2.x is not required.
	// Agents < v1.3.0 populate the ServiceTag field. In this case,
	// use ServiceTag instead of the ServiceTags field.
	if args.ServiceTag != "" {
		return s.CheckServiceTagNodes(ws, args.ServiceName, []string{args.ServiceTag}, &args.EnterpriseMeta, args.PeerName)
	}
	return s.CheckServiceTagNodes(ws, args.ServiceName, args.ServiceTags, &args.EnterpriseMeta, args.PeerName)
}

func (h *Health) serviceNodesDefault(ws memdb.WatchSet, s *state.Store, args *structs.ServiceSpecificRequest) (uint64, structs.CheckServiceNodes, error) {
	return s.CheckServiceNodes(ws, args.ServiceName, &args.EnterpriseMeta, args.PeerName)
}

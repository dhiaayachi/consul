// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package agent

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/internal/dnsutil"
)

const (
	serviceHealth = "service"
	connectHealth = "connect"
	ingressHealth = "ingress"
)

func (s *HTTPHandlers) HealthChecksInState(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Set default DC
	args := structs.ChecksInStateRequest{}
	if err := s.parseEntMeta(req, &args.EnterpriseMeta); err != nil {
		return nil, err
	}
	s.parseSource(req, &args.Source)
	args.NodeMetaFilters = s.parseMetaFilter(req)
	if done := s.parse(resp, req, &args.Datacenter, &args.QueryOptions); done {
		return nil, nil
	}

	// Pull out the service name
	args.State = strings.TrimPrefix(req.URL.Path, "/v1/health/state/")
	if args.State == "" {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "Missing check state"}
	}

	// Make the RPC request
	var out structs.IndexedHealthChecks
	defer setMeta(resp, &out.QueryMeta)
RETRY_ONCE:
	if err := s.agent.RPC(req.Context(), "Health.ChecksInState", &args, &out); err != nil {
		return nil, err
	}
	if args.QueryOptions.AllowStale && args.MaxStaleDuration > 0 && args.MaxStaleDuration < out.LastContact {
		args.AllowStale = false
		args.MaxStaleDuration = 0
		goto RETRY_ONCE
	}
	out.ConsistencyLevel = args.QueryOptions.ConsistencyLevel()

	// Use empty list instead of nil
	if out.HealthChecks == nil {
		out.HealthChecks = make(structs.HealthChecks, 0)
	}
	for i, c := range out.HealthChecks {
		if c.ServiceTags == nil {
			clone := *c
			clone.ServiceTags = make([]string, 0)
			out.HealthChecks[i] = &clone
		}
	}
	return out.HealthChecks, nil
}

func (s *HTTPHandlers) HealthNodeChecks(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Set default DC
	args := structs.NodeSpecificRequest{}
	if err := s.parseEntMeta(req, &args.EnterpriseMeta); err != nil {
		return nil, err
	}
	if done := s.parse(resp, req, &args.Datacenter, &args.QueryOptions); done {
		return nil, nil
	}

	// Pull out the service name
	args.Node = strings.TrimPrefix(req.URL.Path, "/v1/health/node/")
	if args.Node == "" {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "Missing node name"}
	}

	// Make the RPC request
	var out structs.IndexedHealthChecks
	defer setMeta(resp, &out.QueryMeta)
RETRY_ONCE:
	if err := s.agent.RPC(req.Context(), "Health.NodeChecks", &args, &out); err != nil {
		return nil, err
	}
	if args.QueryOptions.AllowStale && args.MaxStaleDuration > 0 && args.MaxStaleDuration < out.LastContact {
		args.AllowStale = false
		args.MaxStaleDuration = 0
		goto RETRY_ONCE
	}
	out.ConsistencyLevel = args.QueryOptions.ConsistencyLevel()

	// Use empty list instead of nil
	if out.HealthChecks == nil {
		out.HealthChecks = make(structs.HealthChecks, 0)
	}
	for i, c := range out.HealthChecks {
		if c.ServiceTags == nil {
			clone := *c
			clone.ServiceTags = make([]string, 0)
			out.HealthChecks[i] = &clone
		}
	}
	return out.HealthChecks, nil
}

func (s *HTTPHandlers) HealthServiceChecks(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Set default DC
	args := structs.ServiceSpecificRequest{}
	if err := s.parseEntMetaNoWildcard(req, &args.EnterpriseMeta); err != nil {
		return nil, err
	}
	s.parseSource(req, &args.Source)
	args.NodeMetaFilters = s.parseMetaFilter(req)
	if done := s.parse(resp, req, &args.Datacenter, &args.QueryOptions); done {
		return nil, nil
	}

	// Pull out the service name
	args.ServiceName = strings.TrimPrefix(req.URL.Path, "/v1/health/checks/")
	if args.ServiceName == "" {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "Missing service name"}
	}

	// Make the RPC request
	var out structs.IndexedHealthChecks
	defer setMeta(resp, &out.QueryMeta)
RETRY_ONCE:
	if err := s.agent.RPC(req.Context(), "Health.ServiceChecks", &args, &out); err != nil {
		return nil, err
	}
	if args.QueryOptions.AllowStale && args.MaxStaleDuration > 0 && args.MaxStaleDuration < out.LastContact {
		args.AllowStale = false
		args.MaxStaleDuration = 0
		goto RETRY_ONCE
	}
	out.ConsistencyLevel = args.QueryOptions.ConsistencyLevel()

	// Use empty list instead of nil
	if out.HealthChecks == nil {
		out.HealthChecks = make(structs.HealthChecks, 0)
	}
	for i, c := range out.HealthChecks {
		if c.ServiceTags == nil {
			clone := *c
			clone.ServiceTags = make([]string, 0)
			out.HealthChecks[i] = &clone
		}
	}
	return out.HealthChecks, nil
}

// HealthIngressServiceNodes should return "all the healthy ingress gateway instances
// that I can use to access this connect-enabled service without mTLS".
func (s *HTTPHandlers) HealthIngressServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return s.healthServiceNodes(resp, req, ingressHealth)
}

// HealthConnectServiceNodes should return "all healthy connect-enabled
// endpoints (e.g. could be side car proxies or native instances) for this
// service so I can connect with mTLS".
func (s *HTTPHandlers) HealthConnectServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return s.healthServiceNodes(resp, req, connectHealth)
}

// HealthServiceNodes should return "all the healthy instances of this service
// registered so I can connect directly to them".
func (s *HTTPHandlers) HealthServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return s.healthServiceNodes(resp, req, serviceHealth)
}

func (s *HTTPHandlers) healthServiceNodes(resp http.ResponseWriter, req *http.Request, healthType string) (interface{}, error) {
	// Set default DC
	args := structs.ServiceSpecificRequest{}
	if err := s.parseEntMetaNoWildcard(req, &args.EnterpriseMeta); err != nil {
		return nil, err
	}
	s.parseSource(req, &args.Source)
	args.NodeMetaFilters = s.parseMetaFilter(req)
	if done := s.parse(resp, req, &args.Datacenter, &args.QueryOptions); done {
		return nil, nil
	}

	s.parsePeerName(req, &args)
	s.parseSamenessGroup(req, &args)
	if args.SamenessGroup != "" && args.PeerName != "" {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "peer-name and sameness-group are mutually exclusive"}
	}

	// Check for tags
	params := req.URL.Query()
	if _, ok := params["tag"]; ok {
		args.ServiceTags = params["tag"]
		args.TagFilter = true
	}

	if _, ok := params["merge-central-config"]; ok {
		args.MergeCentralConfig = true
	}

	// Determine the prefix
	var prefix string
	switch healthType {
	case connectHealth:
		prefix = "/v1/health/connect/"
		args.Connect = true
	case ingressHealth:
		prefix = "/v1/health/ingress/"
		args.Ingress = true
	default:
		// serviceHealth is the default type
		prefix = "/v1/health/service/"
	}

	// Parse the service name from the query params
	args.ServiceName = strings.TrimPrefix(req.URL.Path, prefix)
	if args.ServiceName == "" {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "Missing service name"}
	}

	// Parse the passing flag from the query params and use to set the health filter type
	// to HealthFilterIncludeOnlyPassing if it is present.  Otherwise, do not filter by health.
	passing, err := getBoolQueryParam(params, api.HealthPassing)
	if err != nil {
		return nil, HTTPError{StatusCode: http.StatusBadRequest, Reason: "Invalid value for ?passing"}
	}
	healthFilterType := structs.HealthFilterIncludeAll
	if passing {
		healthFilterType = structs.HealthFilterIncludeOnlyPassing
	}
	args.HealthFilterType = healthFilterType

	out, md, err := s.agent.rpcClientHealth.ServiceNodes(req.Context(), args)
	if err != nil {
		return nil, err
	}

	if args.QueryOptions.UseCache {
		setCacheMeta(resp, &md)
	}
	out.QueryMeta.ConsistencyLevel = args.QueryOptions.ConsistencyLevel()
	_ = setMeta(resp, &out.QueryMeta)

	// Translate addresses after filtering so we don't waste effort.
	s.agent.TranslateAddresses(args.Datacenter, out.Nodes, dnsutil.TranslateAddressAcceptAny)

	// Use empty list instead of nil
	if out.Nodes == nil {
		out.Nodes = make(structs.CheckServiceNodes, 0)
	}
	for i := range out.Nodes {
		if out.Nodes[i].Checks == nil {
			out.Nodes[i].Checks = make(structs.HealthChecks, 0)
		}
		for j, c := range out.Nodes[i].Checks {
			if c.ServiceTags == nil {
				clone := *c
				clone.ServiceTags = make([]string, 0)
				out.Nodes[i].Checks[j] = &clone
			}
		}
		if out.Nodes[i].Service != nil && out.Nodes[i].Service.Tags == nil {
			clone := *out.Nodes[i].Service
			clone.Tags = make([]string, 0)
			out.Nodes[i].Service = &clone
		}
	}
	return out.Nodes, nil
}

func getBoolQueryParam(params url.Values, key string) (bool, error) {
	var param bool
	if _, ok := params[key]; ok {
		val := params.Get(key)
		// Orginally a comment declared this check should be removed after Consul
		// 0.10, to no longer support using ?passing without a value. However, I
		// think this is a reasonable experience for a user and so am keeping it
		// here.
		if val == "" {
			param = true
		} else {
			var err error
			param, err = strconv.ParseBool(val)
			if err != nil {
				return false, err
			}
		}
	}
	return param, nil
}

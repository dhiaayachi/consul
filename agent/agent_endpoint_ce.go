// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package agent

import (
	"net/http"

	"github.com/dhiaayachi/consul/acl"
)

func (s *HTTPHandlers) validateRequestPartition(_ http.ResponseWriter, _ *acl.EnterpriseMeta) bool {
	return true
}

func (s *HTTPHandlers) defaultMetaPartitionToAgent(entMeta *acl.EnterpriseMeta) {
}

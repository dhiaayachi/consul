// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package gateways

import (
	"testing"

	"github.com/dhiaayachi/consul/api"
)

func getOrCreateNamespace(_ *testing.T, _ *api.Client) string {
	return ""
}

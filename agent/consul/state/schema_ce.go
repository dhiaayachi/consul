// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package state

import (
	"fmt"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
)

func partitionedIndexEntryName(entry string, _ string) string {
	return entry
}

func partitionedAndNamespacedIndexEntryName(entry string, _ *acl.EnterpriseMeta) string {
	return entry
}

// peeredIndexEntryName returns the peered index key for an importable entity (e.g. checks, services, or nodes).
func peeredIndexEntryName(entry, peerName string) string {
	if peerName == "" {
		peerName = structs.LocalPeerKeyword
	}
	return fmt.Sprintf("peer.%s:%s", peerName, entry)
}

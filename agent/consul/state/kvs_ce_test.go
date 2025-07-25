// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package state

import (
	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
)

func testIndexerTableKVs() map[string]indexerTestCase {
	return map[string]indexerTestCase{
		indexID: {
			read: indexValue{
				source:   Query{Value: "TheKey"},
				expected: []byte("TheKey\x00"),
			},
			write: indexValue{
				source:   &structs.DirEntry{Key: "TheKey"},
				expected: []byte("TheKey\x00"),
			},
			prefix: []indexValue{
				{
					source:   "indexString",
					expected: []byte("indexString"),
				},
				{
					source:   acl.EnterpriseMeta{},
					expected: nil,
				},
				{
					source:   Query{Value: "TheKey"},
					expected: []byte("TheKey"),
				},
			},
		},
	}
}

func testIndexerTableTombstones() map[string]indexerTestCase {
	return map[string]indexerTestCase{
		indexID: {
			read: indexValue{
				source:   Query{Value: "TheKey"},
				expected: []byte("TheKey\x00"),
			},
			write: indexValue{
				source:   &Tombstone{Key: "TheKey"},
				expected: []byte("TheKey\x00"),
			},
			prefix: []indexValue{
				{
					source:   "indexString",
					expected: []byte("indexString"),
				},
				{
					source:   acl.EnterpriseMeta{},
					expected: nil,
				},
				{
					source:   Query{Value: "TheKey"},
					expected: []byte("TheKey"),
				},
			},
		},
	}
}

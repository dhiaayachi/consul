// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package impexp

import (
	"encoding/base64"

	"github.com/dhiaayachi/consul/api"
)

type Entry struct {
	Key       string `json:"key"`
	Flags     uint64 `json:"flags"`
	Value     string `json:"value"`
	Namespace string `json:"namespace,omitempty"`
	Partition string `json:"partition,omitempty"`
}

func ToEntry(pair *api.KVPair) *Entry {
	return &Entry{
		Key:       pair.Key,
		Flags:     pair.Flags,
		Value:     base64.StdEncoding.EncodeToString(pair.Value),
		Namespace: pair.Namespace,
		Partition: pair.Partition,
	}
}

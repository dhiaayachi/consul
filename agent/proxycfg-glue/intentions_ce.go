// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package proxycfgglue

import (
	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/proto/private/pbsubscribe"
)

func (s serverIntentions) buildSubjects(serviceName string, entMeta acl.EnterpriseMeta) []*pbsubscribe.NamedSubject {
	// Based on getIntentionPrecedenceMatchServiceNames in the state package.
	if serviceName == structs.WildcardSpecifier {
		return []*pbsubscribe.NamedSubject{
			{
				Key:       structs.WildcardSpecifier,
				Namespace: entMeta.NamespaceOrDefault(),
				Partition: entMeta.PartitionOrDefault(),
				PeerName:  structs.DefaultPeerKeyword,
			},
		}
	}

	return []*pbsubscribe.NamedSubject{
		{
			Key:       serviceName,
			Namespace: entMeta.NamespaceOrDefault(),
			Partition: entMeta.PartitionOrDefault(),
			PeerName:  structs.DefaultPeerKeyword,
		},
		{
			Key:       structs.WildcardSpecifier,
			Namespace: entMeta.NamespaceOrDefault(),
			Partition: entMeta.PartitionOrDefault(),
			PeerName:  structs.DefaultPeerKeyword,
		},
	}
}

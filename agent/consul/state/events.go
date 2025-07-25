// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package state

import (
	"errors"
	"fmt"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/agent/consul/stream"
	"github.com/dhiaayachi/consul/proto/private/pbsubscribe"
)

// PBToStreamSubscribeRequest takes a protobuf subscribe request and enterprise
// metadata to properly generate the matching stream subscribe request.
func PBToStreamSubscribeRequest(req *pbsubscribe.SubscribeRequest, entMeta acl.EnterpriseMeta) (*stream.SubscribeRequest, error) {
	var subject stream.Subject

	if req.GetWildcardSubject() {
		subject = stream.SubjectWildcard
	} else {
		named := req.GetNamedSubject()

		// Support the (deprecated) top-level Key, Partition, Namespace, and PeerName fields.
		if named == nil {
			named = &pbsubscribe.NamedSubject{
				Key:       req.Key,       // nolint:staticcheck // SA1019 intentional use of deprecated field
				Partition: req.Partition, // nolint:staticcheck // SA1019 intentional use of deprecated field
				Namespace: req.Namespace, // nolint:staticcheck // SA1019 intentional use of deprecated field
				PeerName:  req.PeerName,  // nolint:staticcheck // SA1019 intentional use of deprecated field
			}
		}

		if named.Key == "" {
			return nil, errors.New("either WildcardSubject or NamedSubject.Key is required")
		}

		switch req.Topic {
		case EventTopicServiceHealth, EventTopicServiceHealthConnect:
			subject = EventSubjectService{
				Key:            named.Key,
				EnterpriseMeta: entMeta,
				PeerName:       named.PeerName,
			}
		case EventTopicMeshConfig, EventTopicServiceResolver, EventTopicIngressGateway,
			EventTopicServiceIntentions, EventTopicServiceDefaults, EventTopicAPIGateway,
			EventTopicTCPRoute, EventTopicHTTPRoute, EventTopicJWTProvider, EventTopicInlineCertificate,
			EventTopicBoundAPIGateway, EventTopicSamenessGroup, EventTopicExportedServices,
			EventTopicFileSystemCertificate:
			subject = EventSubjectConfigEntry{
				Name:           named.Key,
				EnterpriseMeta: &entMeta,
			}
		case EventTopicServiceList:
			// Events on this topic are published to SubjectNone, but rather than
			// exposing this in (and further complicating) the streaming API we rely
			// on consumers passing WildcardSubject instead, which is functionally the
			// same for this purpose.
			return nil, fmt.Errorf("topic %s can only be consumed using WildcardSubject", EventTopicServiceList)
		default:
			return nil, fmt.Errorf("cannot construct subject for topic %s", req.Topic)
		}
	}

	return &stream.SubscribeRequest{
		Topic:   req.Topic,
		Subject: subject,
		Token:   req.Token,
		Index:   req.Index,
	}, nil
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package types

import (
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/internal/resource"
	pbmulticluster "github.com/hashicorp/consul/proto-public/pbmulticluster/v2beta1"
	"github.com/hashicorp/consul/proto-public/pbresource"
)

func RegisterPartitionExportedServices(r resource.Registry) {
	r.Register(resource.RegisterRequest{
		Type:     pbmulticluster.PartitionExportedServicesType,
		Proto:    &pbmulticluster.PartitionExportedServices{},
		Scope:    pbresource.Scope_SCOPE_PARTITION,
		Validate: ValidatePartitionExportedServices,
		ACLs: &resource.ACLHooks{
			Read:  aclReadHookPartitionExportedServices,
			Write: aclWriteHookPartitionExportedServices,
			List:  resource.NoOpACLListHook,
		},
	})
}

func aclReadHookPartitionExportedServices(authorizer acl.Authorizer, authzContext *acl.AuthorizerContext, id *pbresource.ID, res *pbresource.Resource) error {
	return authorizer.ToAllowAuthorizer().MeshReadAllowed(authzContext)
}

func aclWriteHookPartitionExportedServices(authorizer acl.Authorizer, authzContext *acl.AuthorizerContext, res *pbresource.Resource) error {
	return authorizer.ToAllowAuthorizer().MeshWriteAllowed(authzContext)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package multicluster

import (
	"github.com/dhiaayachi/consul/internal/controller"
	"github.com/dhiaayachi/consul/internal/multicluster/internal/controllers"
	"github.com/dhiaayachi/consul/internal/multicluster/internal/controllers/v1compat"
	"github.com/dhiaayachi/consul/internal/multicluster/internal/types"
	"github.com/dhiaayachi/consul/internal/resource"
)

// RegisterTypes adds all resource types within the "multicluster" API group
// to the given type registry
func RegisterTypes(r resource.Registry) {
	types.Register(r)
}

type CompatControllerDependencies = controllers.CompatDependencies

func DefaultCompatControllerDependencies(ac v1compat.AggregatedConfig) CompatControllerDependencies {
	return CompatControllerDependencies{
		ConfigEntryExports: ac,
	}
}

func RegisterCompatControllers(mgr *controller.Manager, deps CompatControllerDependencies) {
	controllers.RegisterCompat(mgr, deps)
}

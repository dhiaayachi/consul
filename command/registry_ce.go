// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !consulent
// +build !consulent

package command

import (
	mcli "github.com/mitchellh/cli"

	"github.com/hashicorp/consul/command/cli"
)

func registerEnterpriseCommands(_ cli.Ui, _ map[string]mcli.CommandFactory) {}

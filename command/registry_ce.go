// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package command

import (
	mcli "github.com/mitchellh/cli"

	"github.com/dhiaayachi/consul/command/cli"
)

func registerEnterpriseCommands(_ cli.Ui, _ map[string]mcli.CommandFactory) {}

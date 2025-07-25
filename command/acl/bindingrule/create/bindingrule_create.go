// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package bindingrulecreate

import (
	"flag"
	"fmt"
	"strings"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/command/acl"
	"github.com/dhiaayachi/consul/command/acl/bindingrule"
	"github.com/dhiaayachi/consul/command/flags"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}

type cmd struct {
	UI    cli.Ui
	flags *flag.FlagSet
	http  *flags.HTTPFlags
	help  string

	authMethodName string
	description    string
	selector       string
	bindType       string
	bindName       string
	bindVars       map[string]string

	showMeta bool
	format   string
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)

	c.flags.BoolVar(
		&c.showMeta,
		"meta",
		false,
		"Indicates that binding rule metadata such "+
			"as the raft indices should be shown for each entry.",
	)

	c.flags.StringVar(
		&c.authMethodName,
		"method",
		"",
		"The auth method's name for which this binding rule applies. "+
			"This flag is required.",
	)
	c.flags.StringVar(
		&c.description,
		"description",
		"",
		"A description of the binding rule.",
	)
	c.flags.StringVar(
		&c.selector,
		"selector",
		"",
		"Selector is an expression that matches against verified identity "+
			"attributes returned from the auth method during login.",
	)
	c.flags.StringVar(
		&c.bindType,
		"bind-type",
		string(api.BindingRuleBindTypeService),
		"Type of binding to perform (\"service\", \"role\", \"node\", \"policy\",  or \"templated-policy\").",
	)
	c.flags.Var(
		(*flags.FlagMapValue)(&c.bindVars),
		"bind-vars",
		"Templated policy variables. Can only be used when -bind-type is templated-policy."+
			" May be specified multiple times with different variables. Can use ${var} interpolation."+
			" Format is VariableName=Value",
	)
	c.flags.StringVar(
		&c.bindName,
		"bind-name",
		"",
		"Name to bind on match. Can use ${var} interpolation. "+
			"This flag is required.",
	)
	c.flags.StringVar(
		&c.format,
		"format",
		bindingrule.PrettyFormat,
		fmt.Sprintf("Output format {%s}", strings.Join(bindingrule.GetSupportedFormats(), "|")),
	)

	c.http = &flags.HTTPFlags{}
	flags.Merge(c.flags, c.http.ClientFlags())
	flags.Merge(c.flags, c.http.ServerFlags())
	flags.Merge(c.flags, c.http.MultiTenancyFlags())
	c.help = flags.Usage(help, c.flags)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if c.authMethodName == "" {
		c.UI.Error("Missing required '-method' flag")
		c.UI.Error(c.Help())
		return 1
	} else if c.bindType == "" {
		c.UI.Error("Missing required '-bind-type' flag")
		c.UI.Error(c.Help())
		return 1
	} else if c.bindName == "" {
		c.UI.Error("Missing required '-bind-name' flag")
		c.UI.Error(c.Help())
		return 1
	}

	if api.BindingRuleBindType(c.bindType) != api.BindingRuleBindTypeTemplatedPolicy && len(c.bindVars) > 0 {
		c.UI.Error("Cannot specify -bind-vars when -bind-type is not templated-policy")
		c.UI.Error(c.Help())
		return 1
	}

	processBindVars, err := acl.ExtractBindVars(c.bindVars)
	if err != nil {
		c.UI.Error("Failed to decode '-bind-vars'")
		c.UI.Error(c.Help())
		return 1
	}

	newRule := &api.ACLBindingRule{
		Description: c.description,
		AuthMethod:  c.authMethodName,
		BindType:    api.BindingRuleBindType(c.bindType),
		BindName:    c.bindName,
		BindVars:    processBindVars,
		Selector:    c.selector,
	}

	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	rule, _, err := client.ACL().BindingRuleCreate(newRule, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create new binding rule: %v", err))
		return 1
	}

	formatter, err := bindingrule.NewFormatter(c.format, c.showMeta)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	out, err := formatter.FormatBindingRule(rule)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	if out != "" {
		c.UI.Info(out)
	}

	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return flags.Usage(c.help, nil)
}

const synopsis = "Create an ACL binding rule"

const help = `
Usage: consul acl binding-rule create [options]

  Create a new binding rule:

    $ consul acl binding-rule create \
          -method=minikube \
          -bind-type=service \
          -bind-name='k8s-${serviceaccount.name}' \
          -selector='serviceaccount.namespace==default and serviceaccount.name==web'
`

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package rolecreate

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/command/acl"
	"github.com/dhiaayachi/consul/command/acl/role"
	"github.com/dhiaayachi/consul/command/flags"
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

	name                     string
	description              string
	policyIDs                []string
	policyNames              []string
	serviceIdents            []string
	nodeIdents               []string
	templatedPolicy          string
	templatedPolicyFile      string
	templatedPolicyVariables []string

	showMeta bool
	format   string
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.BoolVar(&c.showMeta, "meta", false, "Indicates that role metadata such "+
		"as the content hash and raft indices should be shown for each entry")
	c.flags.StringVar(&c.name, "name", "", "The new role's name. This flag is required.")
	c.flags.StringVar(&c.description, "description", "", "A description of the role")
	c.flags.Var((*flags.AppendSliceValue)(&c.policyIDs), "policy-id", "ID of a "+
		"policy to use for this role. May be specified multiple times")
	c.flags.Var((*flags.AppendSliceValue)(&c.policyNames), "policy-name", "Name of a "+
		"policy to use for this role. May be specified multiple times")
	c.flags.Var((*flags.AppendSliceValue)(&c.serviceIdents), "service-identity", "Name of a "+
		"service identity to use for this role. May be specified multiple times. Format is "+
		"the SERVICENAME or SERVICENAME:DATACENTER1,DATACENTER2,...")
	c.flags.Var((*flags.AppendSliceValue)(&c.nodeIdents), "node-identity", "Name of a "+
		"node identity to use for this role. May be specified multiple times. Format is "+
		"NODENAME:DATACENTER")
	c.flags.Var((*flags.AppendSliceValue)(&c.templatedPolicyVariables), "var", "Templated policy variables."+
		" Must be used in combination with -templated-policy flag to specify required variables."+
		" May be specified multiple times with different variables."+
		" Format is VariableName:Value")
	c.flags.StringVar(&c.templatedPolicy, "templated-policy", "", "The templated policy name. Use -var flag to specify variables when required.")
	c.flags.StringVar(&c.templatedPolicyFile, "templated-policy-file", "", "Path to a file containing templated policy names and variables.")
	c.flags.StringVar(
		&c.format,
		"format",
		role.PrettyFormat,
		fmt.Sprintf("Output format {%s}", strings.Join(role.GetSupportedFormats(), "|")),
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

	if c.name == "" {
		c.UI.Error("Missing required '-name' flag")
		c.UI.Error(c.Help())
		return 1
	}

	if len(c.policyNames) == 0 && len(c.policyIDs) == 0 && len(c.serviceIdents) == 0 && len(c.nodeIdents) == 0 &&
		len(c.templatedPolicy) == 0 && len(c.templatedPolicyFile) == 0 {
		c.UI.Error("Cannot create a role without specifying -policy-name, -policy-id, -service-identity, -node-identity, -templated-policy-file or -templated-policy at least once")
		return 1
	}

	if len(c.templatedPolicyFile) != 0 && len(c.templatedPolicy) != 0 {
		c.UI.Error("Cannot combine the use of templated-policy flag with templated-policy-file. " +
			"To create a role with a single templated policy and simple use case, use -templated-policy. " +
			"For multiple templated policies and more complicated use cases, use -templated-policy-file")
		return 1
	}

	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	newRole := &api.ACLRole{
		Name:        c.name,
		Description: c.description,
	}

	for _, policyName := range c.policyNames {
		// We could resolve names to IDs here but there isn't any reason why its would be better
		// than allowing the agent to do it.
		newRole.Policies = append(newRole.Policies, &api.ACLRolePolicyLink{Name: policyName})
	}

	for _, policyID := range c.policyIDs {
		policyID, err := acl.GetPolicyIDFromPartial(client, policyID)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error resolving policy ID %s: %v", policyID, err))
			return 1
		}
		newRole.Policies = append(newRole.Policies, &api.ACLRolePolicyLink{ID: policyID})
	}

	parsedServiceIdents, err := acl.ExtractServiceIdentities(c.serviceIdents)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	newRole.ServiceIdentities = parsedServiceIdents

	parsedNodeIdents, err := acl.ExtractNodeIdentities(c.nodeIdents)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	newRole.NodeIdentities = parsedNodeIdents

	parsedTemplatedPolicies, err := acl.ExtractTemplatedPolicies(c.templatedPolicy, c.templatedPolicyFile, c.templatedPolicyVariables)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	newRole.TemplatedPolicies = parsedTemplatedPolicies

	r, _, err := client.ACL().RoleCreate(newRole, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create new role: %v", err))
		return 1
	}

	formatter, err := role.NewFormatter(c.format, c.showMeta)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	out, err := formatter.FormatRole(r)
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

const synopsis = "Create an ACL role"

const help = `
Usage: consul acl role create -name NAME [options]

    Create a new role:

        $ consul acl role create -name "new-role" \
                                 -description "This is an example role" \
                                 -policy-id b52fc3de-5 \
                                 -policy-name "acl-replication" \
                                 -service-identity "web" \
                                 -service-identity "db:east,west" \
                                 -templated-policy "builtin/service" \
                                 -var "name:api"
`

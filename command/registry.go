// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dhiaayachi/consul/command/acl"
	aclagent "github.com/dhiaayachi/consul/command/acl/agenttokens"
	aclam "github.com/dhiaayachi/consul/command/acl/authmethod"
	aclamcreate "github.com/dhiaayachi/consul/command/acl/authmethod/create"
	aclamdelete "github.com/dhiaayachi/consul/command/acl/authmethod/delete"
	aclamlist "github.com/dhiaayachi/consul/command/acl/authmethod/list"
	aclamread "github.com/dhiaayachi/consul/command/acl/authmethod/read"
	aclamupdate "github.com/dhiaayachi/consul/command/acl/authmethod/update"
	aclbr "github.com/dhiaayachi/consul/command/acl/bindingrule"
	aclbrcreate "github.com/dhiaayachi/consul/command/acl/bindingrule/create"
	aclbrdelete "github.com/dhiaayachi/consul/command/acl/bindingrule/delete"
	aclbrlist "github.com/dhiaayachi/consul/command/acl/bindingrule/list"
	aclbrread "github.com/dhiaayachi/consul/command/acl/bindingrule/read"
	aclbrupdate "github.com/dhiaayachi/consul/command/acl/bindingrule/update"
	aclbootstrap "github.com/dhiaayachi/consul/command/acl/bootstrap"
	aclpolicy "github.com/dhiaayachi/consul/command/acl/policy"
	aclpcreate "github.com/dhiaayachi/consul/command/acl/policy/create"
	aclpdelete "github.com/dhiaayachi/consul/command/acl/policy/delete"
	aclplist "github.com/dhiaayachi/consul/command/acl/policy/list"
	aclpread "github.com/dhiaayachi/consul/command/acl/policy/read"
	aclpupdate "github.com/dhiaayachi/consul/command/acl/policy/update"
	aclrole "github.com/dhiaayachi/consul/command/acl/role"
	aclrcreate "github.com/dhiaayachi/consul/command/acl/role/create"
	aclrdelete "github.com/dhiaayachi/consul/command/acl/role/delete"
	aclrlist "github.com/dhiaayachi/consul/command/acl/role/list"
	aclrread "github.com/dhiaayachi/consul/command/acl/role/read"
	aclrupdate "github.com/dhiaayachi/consul/command/acl/role/update"
	acltp "github.com/dhiaayachi/consul/command/acl/templatedpolicy"
	acltplist "github.com/dhiaayachi/consul/command/acl/templatedpolicy/list"
	acltppreview "github.com/dhiaayachi/consul/command/acl/templatedpolicy/preview"
	acltpread "github.com/dhiaayachi/consul/command/acl/templatedpolicy/read"
	acltoken "github.com/dhiaayachi/consul/command/acl/token"
	acltclone "github.com/dhiaayachi/consul/command/acl/token/clone"
	acltcreate "github.com/dhiaayachi/consul/command/acl/token/create"
	acltdelete "github.com/dhiaayachi/consul/command/acl/token/delete"
	acltlist "github.com/dhiaayachi/consul/command/acl/token/list"
	acltread "github.com/dhiaayachi/consul/command/acl/token/read"
	acltupdate "github.com/dhiaayachi/consul/command/acl/token/update"
	"github.com/dhiaayachi/consul/command/agent"
	"github.com/dhiaayachi/consul/command/catalog"
	catlistdc "github.com/dhiaayachi/consul/command/catalog/list/dc"
	catlistnodes "github.com/dhiaayachi/consul/command/catalog/list/nodes"
	catlistsvc "github.com/dhiaayachi/consul/command/catalog/list/services"
	"github.com/dhiaayachi/consul/command/config"
	configdelete "github.com/dhiaayachi/consul/command/config/delete"
	configlist "github.com/dhiaayachi/consul/command/config/list"
	configread "github.com/dhiaayachi/consul/command/config/read"
	configwrite "github.com/dhiaayachi/consul/command/config/write"
	"github.com/dhiaayachi/consul/command/connect"
	"github.com/dhiaayachi/consul/command/connect/ca"
	caget "github.com/dhiaayachi/consul/command/connect/ca/get"
	caset "github.com/dhiaayachi/consul/command/connect/ca/set"
	"github.com/dhiaayachi/consul/command/connect/envoy"
	pipebootstrap "github.com/dhiaayachi/consul/command/connect/envoy/pipe-bootstrap"
	"github.com/dhiaayachi/consul/command/connect/expose"
	"github.com/dhiaayachi/consul/command/connect/proxy"
	"github.com/dhiaayachi/consul/command/connect/redirecttraffic"
	"github.com/dhiaayachi/consul/command/debug"
	"github.com/dhiaayachi/consul/command/event"
	"github.com/dhiaayachi/consul/command/exec"
	"github.com/dhiaayachi/consul/command/forceleave"
	"github.com/dhiaayachi/consul/command/info"
	"github.com/dhiaayachi/consul/command/intention"
	ixncheck "github.com/dhiaayachi/consul/command/intention/check"
	ixncreate "github.com/dhiaayachi/consul/command/intention/create"
	ixndelete "github.com/dhiaayachi/consul/command/intention/delete"
	ixnget "github.com/dhiaayachi/consul/command/intention/get"
	ixnlist "github.com/dhiaayachi/consul/command/intention/list"
	ixnmatch "github.com/dhiaayachi/consul/command/intention/match"
	"github.com/dhiaayachi/consul/command/join"
	"github.com/dhiaayachi/consul/command/keygen"
	"github.com/dhiaayachi/consul/command/keyring"
	"github.com/dhiaayachi/consul/command/kv"
	kvdel "github.com/dhiaayachi/consul/command/kv/del"
	kvexp "github.com/dhiaayachi/consul/command/kv/exp"
	kvget "github.com/dhiaayachi/consul/command/kv/get"
	kvimp "github.com/dhiaayachi/consul/command/kv/imp"
	kvput "github.com/dhiaayachi/consul/command/kv/put"
	"github.com/dhiaayachi/consul/command/leave"
	"github.com/dhiaayachi/consul/command/lock"
	"github.com/dhiaayachi/consul/command/login"
	"github.com/dhiaayachi/consul/command/logout"
	"github.com/dhiaayachi/consul/command/maint"
	"github.com/dhiaayachi/consul/command/members"
	"github.com/dhiaayachi/consul/command/monitor"
	"github.com/dhiaayachi/consul/command/operator"
	operauto "github.com/dhiaayachi/consul/command/operator/autopilot"
	operautoget "github.com/dhiaayachi/consul/command/operator/autopilot/get"
	operautoset "github.com/dhiaayachi/consul/command/operator/autopilot/set"
	operautostate "github.com/dhiaayachi/consul/command/operator/autopilot/state"
	operraft "github.com/dhiaayachi/consul/command/operator/raft"
	operraftlist "github.com/dhiaayachi/consul/command/operator/raft/listpeers"
	operraftremove "github.com/dhiaayachi/consul/command/operator/raft/removepeer"
	"github.com/dhiaayachi/consul/command/operator/raft/transferleader"
	"github.com/dhiaayachi/consul/command/operator/usage"
	"github.com/dhiaayachi/consul/command/operator/usage/instances"
	"github.com/dhiaayachi/consul/command/peering"
	peerdelete "github.com/dhiaayachi/consul/command/peering/delete"
	peerestablish "github.com/dhiaayachi/consul/command/peering/establish"
	peerexported "github.com/dhiaayachi/consul/command/peering/exportedservices"
	peergenerate "github.com/dhiaayachi/consul/command/peering/generate"
	peerlist "github.com/dhiaayachi/consul/command/peering/list"
	peerread "github.com/dhiaayachi/consul/command/peering/read"
	"github.com/dhiaayachi/consul/command/reload"
	"github.com/dhiaayachi/consul/command/resource"
	resourceapply "github.com/dhiaayachi/consul/command/resource/apply"
	resourceapplygrpc "github.com/dhiaayachi/consul/command/resource/apply-grpc"
	resourcedelete "github.com/dhiaayachi/consul/command/resource/delete"
	resourcedeletegrpc "github.com/dhiaayachi/consul/command/resource/delete-grpc"
	resourcelist "github.com/dhiaayachi/consul/command/resource/list"
	resourcelistgrpc "github.com/dhiaayachi/consul/command/resource/list-grpc"
	resourceread "github.com/dhiaayachi/consul/command/resource/read"
	resourcereadgrpc "github.com/dhiaayachi/consul/command/resource/read-grpc"
	"github.com/dhiaayachi/consul/command/rtt"
	"github.com/dhiaayachi/consul/command/services"
	svcsderegister "github.com/dhiaayachi/consul/command/services/deregister"
	svcsexport "github.com/dhiaayachi/consul/command/services/export"
	exportedservices "github.com/dhiaayachi/consul/command/services/exportedservices"
	svcsregister "github.com/dhiaayachi/consul/command/services/register"
	"github.com/dhiaayachi/consul/command/snapshot"
	snapdecode "github.com/dhiaayachi/consul/command/snapshot/decode"
	snapinspect "github.com/dhiaayachi/consul/command/snapshot/inspect"
	snaprestore "github.com/dhiaayachi/consul/command/snapshot/restore"
	snapsave "github.com/dhiaayachi/consul/command/snapshot/save"
	"github.com/dhiaayachi/consul/command/tls"
	tlsca "github.com/dhiaayachi/consul/command/tls/ca"
	tlscacreate "github.com/dhiaayachi/consul/command/tls/ca/create"
	tlscert "github.com/dhiaayachi/consul/command/tls/cert"
	tlscertcreate "github.com/dhiaayachi/consul/command/tls/cert/create"
	"github.com/dhiaayachi/consul/command/troubleshoot"
	troubleshootports "github.com/dhiaayachi/consul/command/troubleshoot/ports"
	troubleshootproxy "github.com/dhiaayachi/consul/command/troubleshoot/proxy"
	troubleshootupstreams "github.com/dhiaayachi/consul/command/troubleshoot/upstreams"
	"github.com/dhiaayachi/consul/command/validate"
	"github.com/dhiaayachi/consul/command/version"
	"github.com/dhiaayachi/consul/command/watch"

	mcli "github.com/mitchellh/cli"

	"github.com/dhiaayachi/consul/command/cli"
)

// RegisteredCommands returns a realized mapping of available CLI commands in a format that
// the CLI class can consume.
func RegisteredCommands(ui cli.Ui) map[string]mcli.CommandFactory {
	registry := map[string]mcli.CommandFactory{}
	registerCommands(ui, registry,
		entry{"acl", func(cli.Ui) (cli.Command, error) { return acl.New(), nil }},
		entry{"acl bootstrap", func(ui cli.Ui) (cli.Command, error) { return aclbootstrap.New(ui), nil }},
		entry{"acl policy", func(cli.Ui) (cli.Command, error) { return aclpolicy.New(), nil }},
		entry{"acl policy create", func(ui cli.Ui) (cli.Command, error) { return aclpcreate.New(ui), nil }},
		entry{"acl policy list", func(ui cli.Ui) (cli.Command, error) { return aclplist.New(ui), nil }},
		entry{"acl policy read", func(ui cli.Ui) (cli.Command, error) { return aclpread.New(ui), nil }},
		entry{"acl policy update", func(ui cli.Ui) (cli.Command, error) { return aclpupdate.New(ui), nil }},
		entry{"acl policy delete", func(ui cli.Ui) (cli.Command, error) { return aclpdelete.New(ui), nil }},
		entry{"acl set-agent-token", func(ui cli.Ui) (cli.Command, error) { return aclagent.New(ui), nil }},
		entry{"acl token", func(cli.Ui) (cli.Command, error) { return acltoken.New(), nil }},
		entry{"acl token create", func(ui cli.Ui) (cli.Command, error) { return acltcreate.New(ui), nil }},
		entry{"acl token clone", func(ui cli.Ui) (cli.Command, error) { return acltclone.New(ui), nil }},
		entry{"acl token list", func(ui cli.Ui) (cli.Command, error) { return acltlist.New(ui), nil }},
		entry{"acl token read", func(ui cli.Ui) (cli.Command, error) { return acltread.New(ui), nil }},
		entry{"acl token update", func(ui cli.Ui) (cli.Command, error) { return acltupdate.New(ui), nil }},
		entry{"acl token delete", func(ui cli.Ui) (cli.Command, error) { return acltdelete.New(ui), nil }},
		entry{"acl role", func(cli.Ui) (cli.Command, error) { return aclrole.New(), nil }},
		entry{"acl role create", func(ui cli.Ui) (cli.Command, error) { return aclrcreate.New(ui), nil }},
		entry{"acl role list", func(ui cli.Ui) (cli.Command, error) { return aclrlist.New(ui), nil }},
		entry{"acl role read", func(ui cli.Ui) (cli.Command, error) { return aclrread.New(ui), nil }},
		entry{"acl role update", func(ui cli.Ui) (cli.Command, error) { return aclrupdate.New(ui), nil }},
		entry{"acl role delete", func(ui cli.Ui) (cli.Command, error) { return aclrdelete.New(ui), nil }},
		entry{"acl auth-method", func(cli.Ui) (cli.Command, error) { return aclam.New(), nil }},
		entry{"acl auth-method create", func(ui cli.Ui) (cli.Command, error) { return aclamcreate.New(ui), nil }},
		entry{"acl auth-method list", func(ui cli.Ui) (cli.Command, error) { return aclamlist.New(ui), nil }},
		entry{"acl auth-method read", func(ui cli.Ui) (cli.Command, error) { return aclamread.New(ui), nil }},
		entry{"acl auth-method update", func(ui cli.Ui) (cli.Command, error) { return aclamupdate.New(ui), nil }},
		entry{"acl auth-method delete", func(ui cli.Ui) (cli.Command, error) { return aclamdelete.New(ui), nil }},
		entry{"acl binding-rule", func(cli.Ui) (cli.Command, error) { return aclbr.New(), nil }},
		entry{"acl binding-rule create", func(ui cli.Ui) (cli.Command, error) { return aclbrcreate.New(ui), nil }},
		entry{"acl binding-rule list", func(ui cli.Ui) (cli.Command, error) { return aclbrlist.New(ui), nil }},
		entry{"acl binding-rule read", func(ui cli.Ui) (cli.Command, error) { return aclbrread.New(ui), nil }},
		entry{"acl binding-rule update", func(ui cli.Ui) (cli.Command, error) { return aclbrupdate.New(ui), nil }},
		entry{"acl binding-rule delete", func(ui cli.Ui) (cli.Command, error) { return aclbrdelete.New(ui), nil }},
		entry{"acl templated-policy", func(cli.Ui) (cli.Command, error) { return acltp.New(), nil }},
		entry{"acl templated-policy list", func(ui cli.Ui) (cli.Command, error) { return acltplist.New(ui), nil }},
		entry{"acl templated-policy read", func(ui cli.Ui) (cli.Command, error) { return acltpread.New(ui), nil }},
		entry{"acl templated-policy preview", func(ui cli.Ui) (cli.Command, error) { return acltppreview.New(ui), nil }},
		entry{"agent", func(ui cli.Ui) (cli.Command, error) { return agent.New(ui), nil }},
		entry{"catalog", func(cli.Ui) (cli.Command, error) { return catalog.New(), nil }},
		entry{"catalog datacenters", func(ui cli.Ui) (cli.Command, error) { return catlistdc.New(ui), nil }},
		entry{"catalog nodes", func(ui cli.Ui) (cli.Command, error) { return catlistnodes.New(ui), nil }},
		entry{"catalog services", func(ui cli.Ui) (cli.Command, error) { return catlistsvc.New(ui), nil }},
		entry{"config", func(ui cli.Ui) (cli.Command, error) { return config.New(), nil }},
		entry{"config delete", func(ui cli.Ui) (cli.Command, error) { return configdelete.New(ui), nil }},
		entry{"config list", func(ui cli.Ui) (cli.Command, error) { return configlist.New(ui), nil }},
		entry{"config read", func(ui cli.Ui) (cli.Command, error) { return configread.New(ui), nil }},
		entry{"config write", func(ui cli.Ui) (cli.Command, error) { return configwrite.New(ui), nil }},
		entry{"connect", func(ui cli.Ui) (cli.Command, error) { return connect.New(), nil }},
		entry{"connect ca", func(ui cli.Ui) (cli.Command, error) { return ca.New(), nil }},
		entry{"connect ca get-config", func(ui cli.Ui) (cli.Command, error) { return caget.New(ui), nil }},
		entry{"connect ca set-config", func(ui cli.Ui) (cli.Command, error) { return caset.New(ui), nil }},
		entry{"connect proxy", func(ui cli.Ui) (cli.Command, error) { return proxy.New(ui, MakeShutdownCh()), nil }},
		entry{"connect envoy", func(ui cli.Ui) (cli.Command, error) { return envoy.New(ui), nil }},
		entry{"connect envoy pipe-bootstrap", func(ui cli.Ui) (cli.Command, error) { return pipebootstrap.New(ui), nil }},
		entry{"connect expose", func(ui cli.Ui) (cli.Command, error) { return expose.New(ui), nil }},
		entry{"connect redirect-traffic", func(ui cli.Ui) (cli.Command, error) { return redirecttraffic.New(ui), nil }},
		entry{"debug", func(ui cli.Ui) (cli.Command, error) { return debug.New(ui), nil }},
		entry{"event", func(ui cli.Ui) (cli.Command, error) { return event.New(ui), nil }},
		entry{"exec", func(ui cli.Ui) (cli.Command, error) { return exec.New(ui, MakeShutdownCh()), nil }},
		entry{"force-leave", func(ui cli.Ui) (cli.Command, error) { return forceleave.New(ui), nil }},
		entry{"info", func(ui cli.Ui) (cli.Command, error) { return info.New(ui), nil }},
		entry{"intention", func(ui cli.Ui) (cli.Command, error) { return intention.New(), nil }},
		entry{"intention check", func(ui cli.Ui) (cli.Command, error) { return ixncheck.New(ui), nil }},
		entry{"intention create", func(ui cli.Ui) (cli.Command, error) { return ixncreate.New(ui), nil }},
		entry{"intention delete", func(ui cli.Ui) (cli.Command, error) { return ixndelete.New(ui), nil }},
		entry{"intention get", func(ui cli.Ui) (cli.Command, error) { return ixnget.New(ui), nil }},
		entry{"intention list", func(ui cli.Ui) (cli.Command, error) { return ixnlist.New(ui), nil }},
		entry{"intention match", func(ui cli.Ui) (cli.Command, error) { return ixnmatch.New(ui), nil }},
		entry{"join", func(ui cli.Ui) (cli.Command, error) { return join.New(ui), nil }},
		entry{"keygen", func(ui cli.Ui) (cli.Command, error) { return keygen.New(ui), nil }},
		entry{"keyring", func(ui cli.Ui) (cli.Command, error) { return keyring.New(ui), nil }},
		entry{"kv", func(cli.Ui) (cli.Command, error) { return kv.New(), nil }},
		entry{"kv delete", func(ui cli.Ui) (cli.Command, error) { return kvdel.New(ui), nil }},
		entry{"kv export", func(ui cli.Ui) (cli.Command, error) { return kvexp.New(ui), nil }},
		entry{"kv get", func(ui cli.Ui) (cli.Command, error) { return kvget.New(ui), nil }},
		entry{"kv import", func(ui cli.Ui) (cli.Command, error) { return kvimp.New(ui), nil }},
		entry{"kv put", func(ui cli.Ui) (cli.Command, error) { return kvput.New(ui), nil }},
		entry{"leave", func(ui cli.Ui) (cli.Command, error) { return leave.New(ui), nil }},
		entry{"lock", func(ui cli.Ui) (cli.Command, error) { return lock.New(ui, MakeShutdownCh()), nil }},
		entry{"login", func(ui cli.Ui) (cli.Command, error) { return login.New(ui), nil }},
		entry{"logout", func(ui cli.Ui) (cli.Command, error) { return logout.New(ui), nil }},
		entry{"maint", func(ui cli.Ui) (cli.Command, error) { return maint.New(ui), nil }},
		entry{"members", func(ui cli.Ui) (cli.Command, error) { return members.New(ui), nil }},
		entry{"monitor", func(ui cli.Ui) (cli.Command, error) { return monitor.New(ui, MakeShutdownCh()), nil }},
		entry{"operator", func(cli.Ui) (cli.Command, error) { return operator.New(), nil }},
		entry{"operator autopilot", func(cli.Ui) (cli.Command, error) { return operauto.New(), nil }},
		entry{"operator autopilot get-config", func(ui cli.Ui) (cli.Command, error) { return operautoget.New(ui), nil }},
		entry{"operator autopilot set-config", func(ui cli.Ui) (cli.Command, error) { return operautoset.New(ui), nil }},
		entry{"operator autopilot state", func(ui cli.Ui) (cli.Command, error) { return operautostate.New(ui), nil }},
		entry{"operator raft", func(cli.Ui) (cli.Command, error) { return operraft.New(), nil }},
		entry{"operator raft list-peers", func(ui cli.Ui) (cli.Command, error) { return operraftlist.New(ui), nil }},
		entry{"operator raft remove-peer", func(ui cli.Ui) (cli.Command, error) { return operraftremove.New(ui), nil }},
		entry{"operator raft transfer-leader", func(ui cli.Ui) (cli.Command, error) { return transferleader.New(ui), nil }},
		entry{"operator usage", func(ui cli.Ui) (cli.Command, error) { return usage.New(), nil }},
		entry{"operator usage instances", func(ui cli.Ui) (cli.Command, error) { return instances.New(ui), nil }},
		entry{"peering", func(cli.Ui) (cli.Command, error) { return peering.New(), nil }},
		entry{"peering delete", func(ui cli.Ui) (cli.Command, error) { return peerdelete.New(ui), nil }},
		entry{"peering exported-services", func(ui cli.Ui) (cli.Command, error) { return peerexported.New(ui), nil }},
		entry{"peering generate-token", func(ui cli.Ui) (cli.Command, error) { return peergenerate.New(ui), nil }},
		entry{"peering establish", func(ui cli.Ui) (cli.Command, error) { return peerestablish.New(ui), nil }},
		entry{"peering list", func(ui cli.Ui) (cli.Command, error) { return peerlist.New(ui), nil }},
		entry{"peering read", func(ui cli.Ui) (cli.Command, error) { return peerread.New(ui), nil }},
		entry{"reload", func(ui cli.Ui) (cli.Command, error) { return reload.New(ui), nil }},
		entry{"resource", func(cli.Ui) (cli.Command, error) { return resource.New(), nil }},
		entry{"resource read", func(ui cli.Ui) (cli.Command, error) { return resourceread.New(ui), nil }},
		entry{"resource delete", func(ui cli.Ui) (cli.Command, error) { return resourcedelete.New(ui), nil }},
		entry{"resource apply", func(ui cli.Ui) (cli.Command, error) { return resourceapply.New(ui), nil }},
		// will be refactored to resource apply
		entry{"resource apply-grpc", func(ui cli.Ui) (cli.Command, error) { return resourceapplygrpc.New(ui), nil }},
		entry{"resource read-grpc", func(ui cli.Ui) (cli.Command, error) { return resourcereadgrpc.New(ui), nil }},
		entry{"resource list-grpc", func(ui cli.Ui) (cli.Command, error) { return resourcelistgrpc.New(ui), nil }},
		entry{"resource delete-grpc", func(ui cli.Ui) (cli.Command, error) { return resourcedeletegrpc.New(ui), nil }},
		entry{"resource list", func(ui cli.Ui) (cli.Command, error) { return resourcelist.New(ui), nil }},
		entry{"rtt", func(ui cli.Ui) (cli.Command, error) { return rtt.New(ui), nil }},
		entry{"services", func(cli.Ui) (cli.Command, error) { return services.New(), nil }},
		entry{"services register", func(ui cli.Ui) (cli.Command, error) { return svcsregister.New(ui), nil }},
		entry{"services deregister", func(ui cli.Ui) (cli.Command, error) { return svcsderegister.New(ui), nil }},
		entry{"services export", func(ui cli.Ui) (cli.Command, error) { return svcsexport.New(ui), nil }},
		entry{"services exported-services", func(ui cli.Ui) (cli.Command, error) { return exportedservices.New(ui), nil }},
		entry{"snapshot", func(cli.Ui) (cli.Command, error) { return snapshot.New(), nil }},
		entry{"snapshot decode", func(ui cli.Ui) (cli.Command, error) { return snapdecode.New(ui), nil }},
		entry{"snapshot inspect", func(ui cli.Ui) (cli.Command, error) { return snapinspect.New(ui), nil }},
		entry{"snapshot restore", func(ui cli.Ui) (cli.Command, error) { return snaprestore.New(ui), nil }},
		entry{"snapshot save", func(ui cli.Ui) (cli.Command, error) { return snapsave.New(ui), nil }},
		entry{"tls", func(ui cli.Ui) (cli.Command, error) { return tls.New(), nil }},
		entry{"tls ca", func(ui cli.Ui) (cli.Command, error) { return tlsca.New(), nil }},
		entry{"tls ca create", func(ui cli.Ui) (cli.Command, error) { return tlscacreate.New(ui), nil }},
		entry{"tls cert", func(ui cli.Ui) (cli.Command, error) { return tlscert.New(), nil }},
		entry{"tls cert create", func(ui cli.Ui) (cli.Command, error) { return tlscertcreate.New(ui), nil }},
		entry{"troubleshoot", func(ui cli.Ui) (cli.Command, error) { return troubleshoot.New(), nil }},
		entry{"troubleshoot proxy", func(ui cli.Ui) (cli.Command, error) { return troubleshootproxy.New(ui), nil }},
		entry{"troubleshoot upstreams", func(ui cli.Ui) (cli.Command, error) { return troubleshootupstreams.New(ui), nil }},
		entry{"troubleshoot ports", func(ui cli.Ui) (cli.Command, error) { return troubleshootports.New(ui), nil }},
		entry{"validate", func(ui cli.Ui) (cli.Command, error) { return validate.New(ui), nil }},
		entry{"version", func(ui cli.Ui) (cli.Command, error) { return version.New(ui), nil }},
		entry{"watch", func(ui cli.Ui) (cli.Command, error) { return watch.New(ui, MakeShutdownCh()), nil }},
	)
	registerEnterpriseCommands(ui, registry)
	return registry
}

// factory is a function that returns a new instance of a CLI-sub command.
type factory func(cli.Ui) (cli.Command, error)

// entry is a struct that contains a command's name and a factory for that command.
type entry struct {
	name string
	fn   factory
}

func registerCommands(ui cli.Ui, m map[string]mcli.CommandFactory, cmdEntries ...entry) {
	for _, ent := range cmdEntries {
		thisFn := ent.fn
		if _, ok := m[ent.name]; ok {
			panic(fmt.Sprintf("duplicate command: %q", ent.name))
		}
		m[ent.name] = func() (mcli.Command, error) {
			return thisFn(ui)
		}
	}
}

// MakeShutdownCh returns a channel that can be used for shutdown notifications
// for commands. This channel will send a message for every interrupt or SIGTERM
// received.
// Deprecated: use signal.NotifyContext
func MakeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}

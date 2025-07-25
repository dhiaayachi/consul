// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package exec

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/command/flags"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui, shutdownCh <-chan struct{}) *cmd {
	c := &cmd{UI: ui, shutdownCh: shutdownCh}
	c.init()
	return c
}

type cmd struct {
	UI    cli.Ui
	flags *flag.FlagSet
	http  *flags.HTTPFlags
	help  string

	shutdownCh <-chan struct{}
	conf       rExecConf
	apiclient  *api.Client
	sessionID  string
	stopCh     chan struct{}
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.conf.node, "node", "",
		"Regular expression to filter on node names.")
	c.flags.StringVar(&c.conf.service, "service", "",
		"Regular expression to filter on service instances.")
	c.flags.StringVar(&c.conf.tag, "tag", "",
		"Regular expression to filter on service tags. Must be used with -service.")
	c.flags.StringVar(&c.conf.prefix, "prefix", rExecPrefix,
		"Prefix in the KV store to use for request data.")
	c.flags.BoolVar(&c.conf.shell, "shell", true,
		"Use a shell to run the command.")
	c.flags.DurationVar(&c.conf.wait, "wait", rExecQuietWait,
		"Period to wait with no responses before terminating execution.")
	c.flags.DurationVar(&c.conf.replWait, "wait-repl", rExecReplicationWait,
		"Period to wait for replication before firing event. This is an optimization to allow stale reads to be performed.")
	c.flags.BoolVar(&c.conf.verbose, "verbose", false,
		"Enables verbose output.")

	c.http = &flags.HTTPFlags{}
	flags.Merge(c.flags, c.http.ClientFlags())
	flags.Merge(c.flags, c.http.ServerFlags())
	c.help = flags.Usage(help, c.flags)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	// Join the commands to execute
	c.conf.cmd = strings.Join(c.flags.Args(), " ")

	// If there is no command, read stdin for a script input
	if c.conf.cmd == "-" {
		if !c.conf.shell {
			c.UI.Error("Cannot configure -shell=false when reading from stdin")
			return 1
		}

		c.conf.cmd = ""
		var buf bytes.Buffer
		_, err := io.Copy(&buf, os.Stdin)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to read stdin: %v", err))
			c.UI.Error("")
			c.UI.Error(c.Help())
			return 1
		}
		c.conf.script = buf.Bytes()
	} else if !c.conf.shell {
		c.conf.cmd = ""
		c.conf.args = c.flags.Args()
	}

	// Ensure we have a command or script
	if c.conf.cmd == "" && len(c.conf.script) == 0 && len(c.conf.args) == 0 {
		c.UI.Error("Must specify a command to execute")
		c.UI.Error("")
		c.UI.Error(c.Help())
		return 1
	}

	// Validate the configuration
	if err := c.conf.validate(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	// Create and test the HTTP client
	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}
	info, err := client.Agent().Self()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}
	c.apiclient = client

	// Check if this is a foreign datacenter
	if c.http.Datacenter() != "" && c.http.Datacenter() != info["Config"]["Datacenter"] {
		if c.conf.verbose {
			c.UI.Info("Remote exec in foreign datacenter, using Session TTL")
		}
		c.conf.foreignDC = true
		c.conf.localDC = info["Config"]["Datacenter"].(string)
		c.conf.localNode = info["Config"]["NodeName"].(string)
	}

	// Create the job spec
	spec, err := c.makeRExecSpec()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create job spec: %s", err))
		return 1
	}

	// Create a session for this
	c.sessionID, err = c.createSession()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create session: %s", err))
		return 1
	}
	defer c.destroySession()
	if c.conf.verbose {
		c.UI.Info(fmt.Sprintf("Created remote execution session: %s", c.sessionID))
	}

	// Upload the payload
	if err := c.uploadPayload(spec); err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create job file: %s", err))
		return 1
	}
	defer c.destroyData()
	if c.conf.verbose {
		c.UI.Info(fmt.Sprintf("Uploaded remote execution spec"))
	}

	// Wait for replication. This is done so that when the event is
	// received, the job file can be read using a stale read. If the
	// stale read fails, we expect a consistent read to be done, so
	// largely this is a heuristic.
	select {
	case <-time.After(c.conf.replWait):
	case <-c.shutdownCh:
		return 1
	}

	// Fire the event
	id, err := c.fireEvent()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to fire event: %s", err))
		return 1
	}
	if c.conf.verbose {
		c.UI.Info(fmt.Sprintf("Fired remote execution event: %s", id))
	}

	// Wait for the job to finish now
	return c.waitForJob()
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return c.help
}

const synopsis = "Executes a command on Consul nodes"
const help = `
Usage: consul exec [options] [-|command...]

  Evaluates a command on remote Consul nodes. The nodes responding can
  be filtered using regular expressions on node name, service, and tag
  definitions. If a command is '-', stdin will be read until EOF
  and used as a script input.
`

// waitForJob is used to poll for results and wait until the job is terminated
func (c *cmd) waitForJob() int {
	// Although the session destroy is already deferred, we do it again here,
	// because invalidation of the session before destroyData() ensures there is
	// no race condition allowing an agent to upload data (the acquire will fail).
	defer c.destroySession()
	start := time.Now()
	ackCh := make(chan rExecAck, 128)
	heartCh := make(chan rExecHeart, 128)
	outputCh := make(chan rExecOutput, 128)
	exitCh := make(chan rExecExit, 128)
	doneCh := make(chan struct{})
	errCh := make(chan struct{}, 1)
	defer close(doneCh)
	go c.streamResults(doneCh, ackCh, heartCh, outputCh, exitCh, errCh)
	target := &TargetedUI{UI: c.UI}

	var ackCount, exitCount, badExit int
OUTER:
	for {
		// Determine wait time. We provide a larger window if we know about
		// nodes which are still working.
		waitIntv := c.conf.wait
		if ackCount > exitCount {
			waitIntv *= 2
		}

		select {
		case e := <-ackCh:
			ackCount++
			if c.conf.verbose {
				target.Target = e.Node
				target.Info("acknowledged")
			}

		case h := <-heartCh:
			if c.conf.verbose {
				target.Target = h.Node
				target.Info("heartbeat received")
			}

		case e := <-outputCh:
			target.Target = e.Node
			target.Output(string(e.Output))

		case e := <-exitCh:
			exitCount++
			target.Target = e.Node
			target.Info(fmt.Sprintf("finished with exit code %d", e.Code))
			if e.Code != 0 {
				badExit++
			}

		case <-time.After(waitIntv):
			c.UI.Info(fmt.Sprintf("%d / %d node(s) completed / acknowledged", exitCount, ackCount))
			if c.conf.verbose {
				c.UI.Info(fmt.Sprintf("Completed in %0.2f seconds",
					float64(time.Since(start))/float64(time.Second)))
			}
			if exitCount < ackCount {
				badExit++
			}
			break OUTER

		case <-errCh:
			return 1

		case <-c.shutdownCh:
			return 1
		}
	}

	if badExit > 0 {
		return 2
	}
	return 0
}

// streamResults is used to perform blocking queries against the KV endpoint and stream in
// notice of various events into waitForJob
func (c *cmd) streamResults(doneCh chan struct{}, ackCh chan rExecAck, heartCh chan rExecHeart,
	outputCh chan rExecOutput, exitCh chan rExecExit, errCh chan struct{}) {
	kv := c.apiclient.KV()
	opts := api.QueryOptions{WaitTime: c.conf.wait}
	dir := path.Join(c.conf.prefix, c.sessionID) + "/"
	seen := make(map[string]struct{})

	for {
		// Check if we've been signaled to exit
		select {
		case <-doneCh:
			return
		default:
		}

		// Block on waiting for new keys
		keys, qm, err := kv.Keys(dir, "", &opts)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to read results: %s", err))
			goto ERR_EXIT
		}

		// Fast-path the no-change case
		if qm.LastIndex == opts.WaitIndex {
			continue
		}
		opts.WaitIndex = qm.LastIndex

		// Handle each key
		for _, key := range keys {
			// Ignore if we've seen it
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			// Trim the directory
			full := key
			key = strings.TrimPrefix(key, dir)

			// Handle the key type
			switch {
			case key == rExecFileName:
				continue
			case strings.HasSuffix(key, rExecAckSuffix):
				ackCh <- rExecAck{Node: strings.TrimSuffix(key, rExecAckSuffix)}

			case strings.HasSuffix(key, rExecExitSuffix):
				pair, _, err := kv.Get(full, nil)
				if err != nil || pair == nil {
					c.UI.Error(fmt.Sprintf("Failed to read key '%s': %v", full, err))
					continue
				}
				code, err := strconv.ParseInt(string(pair.Value), 10, 32)
				if err != nil {
					c.UI.Error(fmt.Sprintf("Failed to parse exit code '%s': %v", pair.Value, err))
					continue
				}
				exitCh <- rExecExit{
					Node: strings.TrimSuffix(key, rExecExitSuffix),
					Code: int(code),
				}

			case strings.LastIndex(key, rExecOutputDivider) != -1:
				pair, _, err := kv.Get(full, nil)
				if err != nil || pair == nil {
					c.UI.Error(fmt.Sprintf("Failed to read key '%s': %v", full, err))
					continue
				}
				idx := strings.LastIndex(key, rExecOutputDivider)
				node := key[:idx]
				if len(pair.Value) == 0 {
					heartCh <- rExecHeart{Node: node}
				} else {
					outputCh <- rExecOutput{Node: node, Output: pair.Value}
				}

			default:
				c.UI.Error(fmt.Sprintf("Unknown key '%s', ignoring.", key))
			}
		}
	}

ERR_EXIT:
	select {
	case errCh <- struct{}{}:
	default:
	}
}

// validate checks that the configuration is reasonable
func (conf *rExecConf) validate() error {
	// Validate the filters
	if conf.node != "" {
		if _, err := regexp.Compile(conf.node); err != nil {
			return fmt.Errorf("Failed to compile node filter regexp: %v", err)
		}
	}
	if conf.service != "" {
		if _, err := regexp.Compile(conf.service); err != nil {
			return fmt.Errorf("Failed to compile service filter regexp: %v", err)
		}
	}
	if conf.tag != "" {
		if _, err := regexp.Compile(conf.tag); err != nil {
			return fmt.Errorf("Failed to compile tag filter regexp: %v", err)
		}
	}
	if conf.tag != "" && conf.service == "" {
		return fmt.Errorf("Cannot provide tag filter without service filter.")
	}
	return nil
}

// createSession is used to create a new session for this command
func (c *cmd) createSession() (string, error) {
	var id string
	var err error
	if c.conf.foreignDC {
		id, err = c.createSessionForeign()
	} else {
		id, err = c.createSessionLocal()
	}
	if err == nil {
		c.stopCh = make(chan struct{})
		go c.renewSession(id, c.stopCh)
	}
	return id, err
}

// createSessionLocal is used to create a new session in a local datacenter
// This is simpler since we can use the local agent to create the session.
func (c *cmd) createSessionLocal() (string, error) {
	session := c.apiclient.Session()
	se := api.SessionEntry{
		Name:     "Remote Exec",
		Behavior: api.SessionBehaviorDelete,
		TTL:      rExecTTL,
	}
	id, _, err := session.Create(&se, nil)
	return id, err
}

// createSessionLocal is used to create a new session in a foreign datacenter
// This is more complex since the local agent cannot be used to create
// a session, and we must associate with a node in the remote datacenter.
func (c *cmd) createSessionForeign() (string, error) {
	// Look for a remote node to bind to
	health := c.apiclient.Health()
	services, _, err := health.Service("consul", "", true, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to find Consul server in remote datacenter: %v", err)
	}
	if len(services) == 0 {
		return "", fmt.Errorf("Failed to find Consul server in remote datacenter")
	}
	node := services[0].Node.Node
	if c.conf.verbose {
		c.UI.Info(fmt.Sprintf("Binding session to remote node %s@%s", node, c.http.Datacenter()))
	}

	session := c.apiclient.Session()
	se := api.SessionEntry{
		Name:     fmt.Sprintf("Remote Exec via %s@%s", c.conf.localNode, c.conf.localDC),
		Node:     node,
		Checks:   []string{},
		Behavior: api.SessionBehaviorDelete,
		TTL:      rExecTTL,
	}
	id, _, err := session.CreateNoChecks(&se, nil)
	return id, err
}

// renewSession is a long running routine that periodically renews
// the session TTL. This is used for foreign sessions where we depend
// on TTLs.
func (c *cmd) renewSession(id string, stopCh chan struct{}) {
	session := c.apiclient.Session()
	for {
		select {
		case <-time.After(rExecRenewInterval):
			_, _, err := session.Renew(id, nil)
			if err != nil {
				c.UI.Error(fmt.Sprintf("Session renew failed: %v", err))
				return
			}
		case <-stopCh:
			return
		}
	}
}

// destroySession is used to destroy the associated session
func (c *cmd) destroySession() error {
	// Stop the session renew if any
	if c.stopCh != nil {
		close(c.stopCh)
		c.stopCh = nil
	}

	// Destroy the session explicitly
	session := c.apiclient.Session()
	_, err := session.Destroy(c.sessionID, nil)
	return err
}

// makeRExecSpec creates a serialized job specification
// that can be uploaded which will be parsed by agents to
// determine what to do.
func (c *cmd) makeRExecSpec() ([]byte, error) {
	spec := &rExecSpec{
		Command: c.conf.cmd,
		Args:    c.conf.args,
		Script:  c.conf.script,
		Wait:    c.conf.wait,
	}
	return json.Marshal(spec)
}

// uploadPayload is used to upload the request payload
func (c *cmd) uploadPayload(payload []byte) error {
	kv := c.apiclient.KV()
	pair := api.KVPair{
		Key:     path.Join(c.conf.prefix, c.sessionID, rExecFileName),
		Value:   payload,
		Session: c.sessionID,
	}
	ok, _, err := kv.Acquire(&pair, nil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failed to acquire key %s", pair.Key)
	}
	return nil
}

// destroyData is used to nuke all the data associated with
// this remote exec. We just do a recursive delete of our
// data directory.
func (c *cmd) destroyData() error {
	kv := c.apiclient.KV()
	dir := path.Join(c.conf.prefix, c.sessionID)
	_, err := kv.DeleteTree(dir, nil)
	return err
}

// fireEvent is used to fire the event that will notify nodes
// about the remote execution. Returns the event ID or error
func (c *cmd) fireEvent() (string, error) {
	// Create the user event payload
	msg := &rExecEvent{
		Prefix:  c.conf.prefix,
		Session: c.sessionID,
	}
	buf, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	// Format the user event
	event := c.apiclient.Event()
	params := &api.UserEvent{
		Name:          "_rexec",
		Payload:       buf,
		NodeFilter:    c.conf.node,
		ServiceFilter: c.conf.service,
		TagFilter:     c.conf.tag,
	}

	// Fire the event
	id, _, err := event.Fire(params, nil)
	return id, err
}

const (
	// rExecPrefix is the prefix in the KV store used to
	// store the remote exec data
	rExecPrefix = "_rexec"

	// rExecFileName is the name of the file we append to
	// the path, e.g. _rexec/session_id/job
	rExecFileName = "job"

	// rExecAck is the suffix added to an ack path
	rExecAckSuffix = "/ack"

	// rExecAck is the suffix added to an exit code
	rExecExitSuffix = "/exit"

	// rExecOutputDivider is used to namespace the output
	rExecOutputDivider = "/out/"

	// rExecReplicationWait is how long we wait for replication
	rExecReplicationWait = 200 * time.Millisecond

	// rExecQuietWait is how long we wait for no responses
	// before assuming the job is done.
	rExecQuietWait = 2 * time.Second

	// rExecTTL is how long we default the session TTL to
	rExecTTL = "15s"

	// rExecRenewInterval is how often we renew the session TTL
	// when doing an exec in a foreign DC.
	rExecRenewInterval = 5 * time.Second
)

// rExecConf is used to pass around configuration
type rExecConf struct {
	prefix string
	shell  bool

	foreignDC bool
	localDC   string
	localNode string

	node    string
	service string
	tag     string

	wait     time.Duration
	replWait time.Duration

	cmd    string
	args   []string
	script []byte

	verbose bool
}

// rExecEvent is the event we broadcast using a user-event
type rExecEvent struct {
	Prefix  string
	Session string
}

// rExecSpec is the file we upload to specify the parameters
// of the remote execution.
type rExecSpec struct {
	// Command is a single command to run directly in the shell
	Command string `json:",omitempty"`

	// Args is the list of arguments to run the subprocess directly
	Args []string `json:",omitempty"`

	// Script should be spilled to a file and executed
	Script []byte `json:",omitempty"`

	// Wait is how long we are waiting on a quiet period to terminate
	Wait time.Duration
}

// rExecAck is used to transmit an acknowledgement
type rExecAck struct {
	Node string
}

// rExecHeart is used to transmit a heartbeat
type rExecHeart struct {
	Node string
}

// rExecOutput is used to transmit a chunk of output
type rExecOutput struct {
	Node   string
	Output []byte
}

// rExecExit is used to transmit an exit code
type rExecExit struct {
	Node string
	Code int
}

// TargetedUI is a UI that wraps another UI implementation and modifies
// the output to indicate a specific target. Specifically, all Say output
// is prefixed with the target name. Message output is not prefixed but
// is offset by the length of the target so that output is lined up properly
// with Say output. Machine-readable output has the proper target set.
type TargetedUI struct {
	Target string
	UI     cli.Ui
}

func (u *TargetedUI) Ask(query string) (string, error) {
	return u.UI.Ask(u.prefixLines(true, query))
}

func (u *TargetedUI) Info(message string) {
	u.UI.Info(u.prefixLines(true, message))
}

func (u *TargetedUI) Output(message string) {
	u.UI.Output(u.prefixLines(false, message))
}

func (u *TargetedUI) Error(message string) {
	u.UI.Error(u.prefixLines(true, message))
}

func (u *TargetedUI) prefixLines(arrow bool, message string) string {
	arrowText := "==>"
	if !arrow {
		arrowText = strings.Repeat(" ", len(arrowText))
	}

	var result bytes.Buffer

	for _, line := range strings.Split(message, "\n") {
		result.WriteString(fmt.Sprintf("%s %s: %s\n", arrowText, u.Target, line))
	}

	return strings.TrimRightFunc(result.String(), unicode.IsSpace)
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package lib

import (
	"fmt"
	"runtime"

	"github.com/dhiaayachi/consul/version"
)

var (
	// projectURL is the project URL.
	projectURL = "https://developer.hashicorp.com/"

	// rt is the runtime - variable for tests.
	rt = runtime.Version()

	// versionFunc is the func that returns the current version. This is a
	// function to take into account the different build processes and distinguish
	// between enterprise and CE builds.
	versionFunc = func() string {
		return version.GetHumanVersion()
	}
)

// UserAgent returns the consistent user-agent string for Consul.
func UserAgent() string {
	return fmt.Sprintf("Consul/%s (+%s; %s)",
		versionFunc(), projectURL, rt)
}

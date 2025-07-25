// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package resourcehcl_test

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dhiaayachi/consul/internal/resource"
	"github.com/dhiaayachi/consul/internal/resource/demo"
	"github.com/dhiaayachi/consul/internal/resourcehcl"
	"github.com/dhiaayachi/consul/proto-public/pbresource"
	"github.com/dhiaayachi/consul/proto/private/prototest"
)

var update = flag.Bool("update", false, "update golden files")

func FuzzUnmarshall(f *testing.F) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(f, err)

	read := func(t *testing.F, path string) ([]byte, bool) {
		t.Helper()

		bytes, err := os.ReadFile(fmt.Sprintf("./testdata/%s", path))
		switch {
		case err == nil:
			return bytes, true
		case os.IsNotExist(err):
			return nil, false
		}

		t.Fatalf("failed to read file %s %v", path, err)
		return nil, false
	}
	for _, entry := range entries {
		ext := path.Ext(entry.Name())

		if ext != ".hcl" {
			continue
		}
		input, _ := read(f, entry.Name())
		f.Add(input)
	}
	registry := resource.NewRegistry()
	demo.RegisterTypes(registry)

	f.Fuzz(func(t *testing.T, input []byte) {
		got, err := resourcehcl.Unmarshal(input, registry)
		if err != nil {
			return
		}
		require.NotNil(t, got)
	})

}

func TestUnmarshal(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	read := func(t *testing.T, path string) ([]byte, bool) {
		t.Helper()

		bytes, err := os.ReadFile(fmt.Sprintf("./testdata/%s", path))
		switch {
		case err == nil:
			return bytes, true
		case os.IsNotExist(err):
			return nil, false
		}

		t.Fatalf("failed to read file %s %v", path, err)
		return nil, false
	}

	write := func(t *testing.T, path string, src []byte) {
		t.Helper()

		require.NoError(t, os.WriteFile(fmt.Sprintf("./testdata/%s", path), src, 0o600))
	}

	for _, entry := range entries {
		name := entry.Name()
		ext := path.Ext(name)

		if ext != ".hcl" {
			continue
		}

		base := name[0 : len(name)-len(ext)]

		t.Run(base, func(t *testing.T) {
			input, _ := read(t, name)

			registry := resource.NewRegistry()
			demo.RegisterTypes(registry)

			output, err := resourcehcl.UnmarshalOptions{SourceFileName: name}.
				Unmarshal(input, registry)

			if *update {
				if err == nil {
					json, err := protojson.Marshal(output)
					require.NoError(t, err)
					write(t, base+".golden", json)
				} else {
					write(t, base+".error", []byte(err.Error()))
				}
			}

			goldenJSON, haveGoldenJSON := read(t, base+".golden")
			goldenError, haveGoldenError := read(t, base+".error")

			if haveGoldenError && haveGoldenJSON {
				t.Fatalf("both %s.golden and %s.error exist, delete one", base, base)
			}

			if !haveGoldenError && !haveGoldenJSON && !*update {
				t.Fatalf("neither %s.golden or %s.error exist, run the tests again with the -update flag to create one", base, base)
			}

			if haveGoldenError {
				require.Error(t, err)
				require.Equal(t, string(goldenError), err.Error())
			}

			if haveGoldenJSON {
				require.NoError(t, err)

				var exp pbresource.Resource
				require.NoError(t, protojson.Unmarshal(goldenJSON, &exp))
				prototest.AssertDeepEqual(t, &exp, output)
			}
		})
	}
}

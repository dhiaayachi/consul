// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/dhiaayachi/consul/api"
	"github.com/dhiaayachi/consul/lib/decode"
	"github.com/hashicorp/go-multierror"
)

func LoadFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}
	return string(data), nil
}

func loadFromStdin(testStdin io.Reader) (string, error) {
	var stdin io.Reader = os.Stdin
	if testStdin != nil {
		stdin = testStdin
	}

	var b bytes.Buffer
	if _, err := io.Copy(&b, stdin); err != nil {
		return "", fmt.Errorf("Failed to read stdin: %v", err)
	}
	return b.String(), nil
}

func LoadDataSource(data string, testStdin io.Reader) (string, error) {
	// Handle empty quoted shell parameters
	if len(data) == 0 {
		return "", nil
	}

	switch data[0] {
	case '@':
		return LoadFromFile(data[1:])
	case '-':
		if len(data) > 1 {
			return data, nil
		}
		return loadFromStdin(testStdin)
	default:
		return data, nil
	}
}

func LoadDataSourceNoRaw(data string, testStdin io.Reader) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("Failed to load data: must specify a file path or '-' for stdin")
	}

	if data == "-" {
		return loadFromStdin(testStdin)
	}

	return LoadFromFile(data)
}

func ParseConfigEntry(data string) (api.ConfigEntry, error) {
	// parse the data
	var raw map[string]interface{}
	if err := hclDecode(&raw, data); err != nil {
		return nil, fmt.Errorf("Failed to decode config entry input: %v", err)
	}

	return newDecodeConfigEntry(raw)
}

// There is a 'structs' variation of this in
// agent/structs/config_entry.go:DecodeConfigEntry
func newDecodeConfigEntry(raw map[string]interface{}) (api.ConfigEntry, error) {
	var entry api.ConfigEntry

	kindVal, ok := raw["Kind"]
	if !ok {
		kindVal, ok = raw["kind"]
	}
	if !ok {
		return nil, fmt.Errorf("Payload does not contain a kind/Kind key at the top level")
	}

	if kindStr, ok := kindVal.(string); ok {
		newEntry, err := api.MakeConfigEntry(kindStr, "")
		if err != nil {
			return nil, err
		}
		entry = newEntry
	} else {
		return nil, fmt.Errorf("Kind value in payload is not a string")
	}

	var md mapstructure.Metadata
	decodeConf := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			decode.HookWeakDecodeFromSlice,
			decode.HookTranslateKeys,
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
		Metadata:         &md,
		Result:           &entry,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decodeConf)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(raw); err != nil {
		return nil, err
	}

	for _, k := range md.Unused {
		switch {
		case strings.ToLower(k) == "kind":
			// The kind field is used to determine the target, but doesn't need
			// to exist on the target.
			continue

		case strings.HasSuffix(strings.ToLower(k), "namespace"):
			err = multierror.Append(err, fmt.Errorf("invalid config key %q, namespaces are a consul enterprise feature", k))
		case strings.Contains(strings.ToLower(k), "jwt"):
			err = multierror.Append(err, fmt.Errorf("invalid config key %q, api-gateway jwt validation is a consul enterprise feature", k))
		default:
			err = multierror.Append(err, fmt.Errorf("invalid config key %q", k))
		}
	}
	if err != nil {
		return nil, err
	}

	return entry, nil
}

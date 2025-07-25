// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package prepared_query

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dhiaayachi/consul/agent/structs"
	"github.com/dhiaayachi/consul/types"
	"github.com/mitchellh/copystructure"
)

var (
	// bigBench is a test query that uses all the features of templates, not
	// in a realistic way, but in a complete way.
	bigBench = &structs.PreparedQuery{
		Name: "hello",
		Template: structs.QueryTemplateOptions{
			Type:            structs.QueryTemplateTypeNamePrefixMatch,
			Regexp:          "^hello-(.*)-(.*)$",
			RemoveEmptyTags: true,
		},
		Service: structs.ServiceQuery{
			Service: "${name.full}",
			Failover: structs.QueryFailoverOptions{
				Datacenters: []string{
					"${name.full}",
					"${name.prefix}",
					"${name.suffix}",
					"${match(0)}",
					"${match(1)}",
					"${match(2)}",
					"${agent.segment}",
				},
			},
			IgnoreCheckIDs: []types.CheckID{
				"${name.full}",
				"${name.prefix}",
				"${name.suffix}",
				"${match(0)}",
				"${match(1)}",
				"${match(2)}",
				"${agent.segment}",
			},
			Tags: []string{
				"${name.full}",
				"${name.prefix}",
				"${name.suffix}",
				"${match(0)}",
				"${match(1)}",
				"${match(2)}",
				"${agent.segment}",
			},
			NodeMeta: map[string]string{
				"foo": "${name.prefix}",
				"bar": "${match(0)}",
				"baz": "${match(1)}",
				"zoo": "${agent.segment}",
			},
		},
	}

	// smallBench is a small prepared query just for doing geo failover. This
	// is a minimal, useful configuration.
	smallBench = &structs.PreparedQuery{
		Name: "",
		Template: structs.QueryTemplateOptions{
			Type: structs.QueryTemplateTypeNamePrefixMatch,
		},
		Service: structs.ServiceQuery{
			Service: "${name.full}",
			Failover: structs.QueryFailoverOptions{
				Datacenters: []string{
					"dc1",
					"dc2",
					"dc3",
				},
			},
		},
	}
)

func compileBench(b *testing.B, query *structs.PreparedQuery) {
	for i := 0; i < b.N; i++ {
		_, err := Compile(query)
		if err != nil {
			b.Fatalf("err: %v", err)
		}
	}
}

func renderBench(b *testing.B, query *structs.PreparedQuery) {
	compiled, err := Compile(query)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := compiled.Render("hello-bench-mark", structs.QuerySource{})
		if err != nil {
			b.Fatalf("err: %v", err)
		}
	}
}

func BenchmarkTemplate_CompileSmall(b *testing.B) {
	compileBench(b, smallBench)
}

func BenchmarkTemplate_CompileBig(b *testing.B) {
	compileBench(b, bigBench)
}

func BenchmarkTemplate_RenderSmall(b *testing.B) {
	renderBench(b, smallBench)
}

func BenchmarkTemplate_RenderBig(b *testing.B) {
	renderBench(b, bigBench)
}

func TestTemplate_Compile(t *testing.T) {
	// Start with an empty query that's not even a template.
	query := &structs.PreparedQuery{}
	_, err := Compile(query)
	if err == nil || !strings.Contains(err.Error(), "Bad Template") {
		t.Fatalf("bad: %v", err)
	}
	if IsTemplate(query) {
		t.Fatalf("should not be a template")
	}

	// Make it a basic template, keeping a copy before we compile.
	query.Template.Type = structs.QueryTemplateTypeNamePrefixMatch
	query.Template.Regexp = "^(hello)there$"
	query.Service.Service = "${name.full}"
	query.Service.IgnoreCheckIDs = []types.CheckID{"${match(1)}", "${agent.segment}"}
	query.Service.Tags = []string{"${match(1)}", "${agent.segment}"}
	backup, err := copystructure.Copy(query)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	ct, err := Compile(query)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !IsTemplate(query) {
		t.Fatalf("should be a template")
	}

	// Do a sanity check render on it.
	actual, err := ct.Render("hellothere", structs.QuerySource{Segment: "segment-foo"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// See if it rendered correctly.
	expected := &structs.PreparedQuery{
		Template: structs.QueryTemplateOptions{
			Type:   structs.QueryTemplateTypeNamePrefixMatch,
			Regexp: "^(hello)there$",
		},
		Service: structs.ServiceQuery{
			Service: "hellothere",
			IgnoreCheckIDs: []types.CheckID{
				"hello",
				"segment-foo",
			},
			Tags: []string{
				"hello",
				"segment-foo",
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}

	// Prove that it didn't alter the definition we compiled.
	if !reflect.DeepEqual(query, backup.(*structs.PreparedQuery)) {
		t.Fatalf("bad: %#v", query)
	}

	// Try a bad HIL interpolation (syntax error).
	query.Service.Service = "${name.full"
	_, err = Compile(query)
	if err == nil || !strings.Contains(err.Error(), "Bad format") {
		t.Fatalf("bad: %v", err)
	}

	// Try a bad HIL interpolation (syntax ok but unknown variable).
	query.Service.Service = "${name.nope}"
	_, err = Compile(query)
	if err == nil || !strings.Contains(err.Error(), "unknown variable") {
		t.Fatalf("bad: %v", err)
	}

	// Try a bad regexp.
	query.Template.Regexp = "^(nope$"
	query.Service.Service = "${name.full}"
	_, err = Compile(query)
	if err == nil || !strings.Contains(err.Error(), "Bad Regexp") {
		t.Fatalf("bad: %v", err)
	}
}

func TestTemplate_Render(t *testing.T) {
	// Try a noop template that is all static.
	{
		query := &structs.PreparedQuery{
			Template: structs.QueryTemplateOptions{
				Type: structs.QueryTemplateTypeNamePrefixMatch,
			},
			Service: structs.ServiceQuery{
				Service: "hellothere",
			},
		}
		ct, err := Compile(query)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		actual, err := ct.Render("unused", structs.QuerySource{})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !reflect.DeepEqual(actual, query) {
			t.Fatalf("bad: %#v", actual)
		}
	}

	// Try all the variables and functions.
	query := &structs.PreparedQuery{
		Name: "hello-",
		Template: structs.QueryTemplateOptions{
			Type:   structs.QueryTemplateTypeNamePrefixMatch,
			Regexp: "^(.*?)-(.*?)-(.*)$",
		},
		Service: structs.ServiceQuery{
			Service: "${name.prefix} xxx ${name.full} xxx ${name.suffix} xxx ${agent.segment}",
			Tags: []string{
				"${match(-1)}",
				"${match(0)}",
				"${match(1)}",
				"${match(2)}",
				"${match(3)}",
				"${match(4)}",
				"${40 + 2}",
			},
			NodeMeta: map[string]string{"foo": "${match(1)}"},
		},
	}
	ct, err := Compile(query)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Run a case that matches the regexp.
	{
		actual, err := ct.Render("hello-foo-bar-none", structs.QuerySource{Segment: "segment-bar"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		expected := &structs.PreparedQuery{
			Name: "hello-",
			Template: structs.QueryTemplateOptions{
				Type:   structs.QueryTemplateTypeNamePrefixMatch,
				Regexp: "^(.*?)-(.*?)-(.*)$",
			},
			Service: structs.ServiceQuery{
				Service: "hello- xxx hello-foo-bar-none xxx foo-bar-none xxx segment-bar",
				Tags: []string{
					"",
					"hello-foo-bar-none",
					"hello",
					"foo",
					"bar-none",
					"",
					"42",
				},
				NodeMeta: map[string]string{"foo": "hello"},
			},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad: %#v", actual)
		}
	}

	// Run a case that doesn't match the regexp
	{
		actual, err := ct.Render("hello-nope", structs.QuerySource{Segment: "segment-bar"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		expected := &structs.PreparedQuery{
			Name: "hello-",
			Template: structs.QueryTemplateOptions{
				Type:   structs.QueryTemplateTypeNamePrefixMatch,
				Regexp: "^(.*?)-(.*?)-(.*)$",
			},
			Service: structs.ServiceQuery{
				Service: "hello- xxx hello-nope xxx nope xxx segment-bar",
				Tags: []string{
					"",
					"",
					"",
					"",
					"",
					"",
					"42",
				},
				NodeMeta: map[string]string{"foo": ""},
			},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad: %#v", actual)
		}
	}

	// Try all the variables and functions, removing empty tags.
	query = &structs.PreparedQuery{
		Name: "hello-",
		Template: structs.QueryTemplateOptions{
			Type:            structs.QueryTemplateTypeNamePrefixMatch,
			Regexp:          "^(.*?)-(.*?)-(.*)$",
			RemoveEmptyTags: true,
		},
		Service: structs.ServiceQuery{
			Service: "${name.prefix} xxx ${name.full} xxx ${name.suffix} xxx ${agent.segment}",
			Tags: []string{
				"${match(-1)}",
				"${match(0)}",
				"${match(1)}",
				"${match(2)}",
				"${match(3)}",
				"${match(4)}",
				"${40 + 2}",
			},
		},
	}
	ct, err = Compile(query)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Run a case that matches the regexp, removing empty tags.
	{
		actual, err := ct.Render("hello-foo-bar-none", structs.QuerySource{Segment: "segment-baz"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		expected := &structs.PreparedQuery{
			Name: "hello-",
			Template: structs.QueryTemplateOptions{
				Type:            structs.QueryTemplateTypeNamePrefixMatch,
				Regexp:          "^(.*?)-(.*?)-(.*)$",
				RemoveEmptyTags: true,
			},
			Service: structs.ServiceQuery{
				Service: "hello- xxx hello-foo-bar-none xxx foo-bar-none xxx segment-baz",
				Tags: []string{
					"hello-foo-bar-none",
					"hello",
					"foo",
					"bar-none",
					"42",
				},
			},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad:\n%#v\nexpected:\n%#v\n", actual, expected)
		}
	}

	// Run a case that doesn't match the regexp, removing empty tags.
	{
		actual, err := ct.Render("hello-nope", structs.QuerySource{Segment: "segment-baz"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		expected := &structs.PreparedQuery{
			Name: "hello-",
			Template: structs.QueryTemplateOptions{
				Type:            structs.QueryTemplateTypeNamePrefixMatch,
				Regexp:          "^(.*?)-(.*?)-(.*)$",
				RemoveEmptyTags: true,
			},
			Service: structs.ServiceQuery{
				Service: "hello- xxx hello-nope xxx nope xxx segment-baz",
				Tags: []string{
					"42",
				},
			},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad: %#v", actual)
		}
	}
}

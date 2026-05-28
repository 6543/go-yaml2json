// SPDX-FileCopyrightText: 2023 6543 <6543@obermui.de>
// SPDX-License-Identifier: MIT

package yaml2json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name, yaml, json string
	}{{
		name: "map list",
		yaml: `- name: Jack
- name: Jill
`,
		json: `[{"name":"Jack"},{"name":"Jill"}]`,
	}, {
		name: "single item map obj",
		yaml: `name: Jack`,
		json: `{"name":"Jack"}`,
	}, {
		name: "object as map",
		yaml: `name: Jack
job: Butcher
`,
		json: `{"job":"Butcher","name":"Jack"}`,
	}, {
		name: "object list",
		yaml: `- name: Jack
  job: Butcher
- name: Jill
  job: Cook
  obj:
    empty: false
    data: |
      some data 123
      with new line
`,
		json: `[{"job":"Butcher","name":"Jack"},{"job":"Cook","name":"Jill","obj":{"data":"some data 123\nwith new line\n","empty":false}}]`,
	}, {
		name: "advanced yaml with alias",
		yaml: `vars:
  - &node_image 'node:16-alpine'
  - &when_path
    # web source code
    - "web/**"
    - some

pipeline:
  deps:
    image: *node_image
    commands:
    - "cd web/"
    - yarn install
    when:
      path: *when_path
`,
		json: `{"pipeline":{"deps":{"commands":["cd web/","yarn install"],"image":"node:16-alpine","when":{"path":["web/**","some"]}}},"vars":["node:16-alpine",["web/**","some"]]}`,
	}, {
		name: "map merging",
		yaml: `
variables: &var
  target: dist
  recursive: false
  try: true
one:
  <<: *var
  name: awesome
two:
  <<: *var
  try: false
`,
		json: `{"one":{"name":"awesome","recursive":false,"target":"dist","try":true},"two":{"recursive":false,"target":"dist","try":false},"variables":{"recursive":false,"target":"dist","try":true}}`,
	}, {
		name: "map merging array",
		yaml: `one: &one
  name: awesome
two: &two
  try: false
comb:
  <<: [*one, *two]`,
		json: `{"comb":{"name":"awesome","try":false},"one":{"name":"awesome"},"two":{"try":false}}`,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Convert([]byte(tc.yaml))
			assert.NoError(t, err)
			assert.EqualValues(t, tc.json, string(result))
		})
	}
}

func TestStreamConvert(t *testing.T) {
	tests := []struct {
		name, yaml, json string
	}{{
		name: "empty doc",
		yaml: `---`,
		json: "null\n",
	}, {
		name: "values",
		yaml: `values:
  - int: 5
  - float: 6.8523015e+5
  - none: null
`,
		json: `{"values":[{"int":5},{"float":685230.15},{"none":null}]}` + "\n",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.yaml))
			w := new(strings.Builder)
			err := StreamConvert(r, w)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.json, w.String())
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		yaml  string
		error string
	}{{
		yaml:  ``,
		error: `EOF`,
	}}

	for _, tc := range tests {
		r := bytes.NewReader([]byte(tc.yaml))
		w := new(strings.Builder)
		err := StreamConvert(r, w)
		if assert.Error(t, err) {
			assert.EqualValues(t, tc.error, err.Error())
		}
	}

	// test max depth
	_, err := toJSON(nil, maxDepth)
	assert.ErrorIs(t, err, ErrMaxDepth)
}

// TestConvertInvalidYAML covers the yaml.Unmarshal error path in Convert.
func TestConvertInvalidYAML(t *testing.T) {
	// unbalanced flow mapping triggers a parser error
	_, err := Convert([]byte("{key: value, : }"))
	assert.Error(t, err)
}

// TestConvertNode covers ConvertNode being called directly (not via Convert).
func TestConvertNode(t *testing.T) {
	node := &yaml.Node{}
	assert.NoError(t, yaml.Unmarshal([]byte("foo: bar\n"), node))
	out, err := ConvertNode(node)
	assert.NoError(t, err)
	assert.EqualValues(t, `{"foo":"bar"}`, string(out))
}

// TestToJSONDocumentNode covers the yaml.DocumentNode branch of toJSON,
// reached when toJSON is called on a freshly unmarshaled node (before
// resolveMerges flattens it).
func TestToJSONDocumentNode(t *testing.T) {
	node := &yaml.Node{}
	assert.NoError(t, yaml.Unmarshal([]byte("answer: 42\n"), node))
	assert.Equal(t, yaml.DocumentNode, node.Kind)
	val, err := toJSON(node, 0)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"answer": int64(42)}, val)
}

// TestToJSONAliasNode covers the yaml.AliasNode branch of toJSON, reached
// by handing it an AliasNode directly (resolveMerges normally inlines
// aliases before toJSON ever sees one).
func TestToJSONAliasNode(t *testing.T) {
	target := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "aliased",
	}
	alias := &yaml.Node{Kind: yaml.AliasNode, Alias: target}
	val, err := toJSON(alias, 0)
	assert.NoError(t, err)
	assert.Equal(t, "aliased", val)
}

// TestToJSONBrokenMapping covers the ErrBrokenMappingNode path: a mapping
// node with an odd number of content entries (key without value).
func TestToJSONBrokenMapping(t *testing.T) {
	broken := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "lonely_key"},
		},
	}
	_, err := toJSON(broken, 0)
	assert.ErrorIs(t, err, ErrBrokenMappingNode)
}

// TestToJSONUnsupportedNode covers the ErrUnsupportedNode default branch
// via a zero-Kind node, which matches none of the switch cases.
func TestToJSONUnsupportedNode(t *testing.T) {
	bogus := &yaml.Node{Kind: 0}
	_, err := toJSON(bogus, 0)
	assert.ErrorIs(t, err, ErrUnsupportedNode)
}

// TestToJSONDeepRecursion drives toJSON close to its maxDepth cap with a
// real nested mapping, ensuring the depth counter increments through every
// recursive call.
func TestToJSONDeepRecursion(t *testing.T) {
	// build "a: a: a: ... : leaf" nested deeper than maxDepth
	var sb strings.Builder
	for i := 0; i < int(maxDepth)+10; i++ {
		sb.WriteString("a:\n")
		for j := 0; j <= i; j++ {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("leaf\n")
	_, err := Convert([]byte(sb.String()))
	assert.ErrorIs(t, err, ErrMaxDepth)
}

// TestStreamConvertBrokenMapping ensures the resolveMerges/toJSON error
// path inside StreamConvert is exercised end-to-end. A YAML alias used as
// a mapping key after merge resolution surfaces no error on its own, so
// instead we feed a value that fails int parsing inside a custom-tagged
// scalar.
func TestStreamConvertBadInt(t *testing.T) {
	// !!int tag with non-numeric value triggers strconv.ParseInt failure
	// inside toJSON's ScalarNode branch.
	r := bytes.NewReader([]byte("!!int notanumber\n"))
	w := new(strings.Builder)
	err := StreamConvert(r, w)
	assert.Error(t, err)
}

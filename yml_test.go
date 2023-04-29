package yaml2json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

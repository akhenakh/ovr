package action

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/tidwall/gjson"
	"go.yaml.in/yaml/v3"
)

// parseDict parses a JSON or YAML object from the input into a dict,
// JSON is tried first as it is a subset of YAML
func parseDict(in []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(in, &m); err == nil && m != nil {
		return m, nil
	}
	if err := yaml.Unmarshal(in, &m); err == nil && m != nil {
		return m, nil
	}
	return nil, fmt.Errorf("can't parse input as a JSON or YAML object")
}

var dictAction = New(Definition[[]byte, map[string]any]{
	Doc:          "Parse a JSON or YAML object from input into a dict",
	Names:        []string{"dict"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: DictFormat,
	Func: func(a Action, in []byte) (map[string]any, error) {
		return parseDict(in)
	},
})

var dictGetAction = New(Definition[map[string]any, []byte]{
	Doc:          `Get a value from a dict with a gjson path parameter, e.g. name.last, children.#, friends.#(last=="Murphy").first`,
	Names:        []string{"get", "gjson"},
	Type:         TransformAction,
	InputFormat:  DictFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "a gjson path, e.g. name.last"}},
	Func: func(a Action, in map[string]any) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("get parameter is not a string")
		}
		b, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}
		res := gjson.GetBytes(b, p)
		if !res.Exists() {
			return nil, fmt.Errorf("no result for gjson path %q", p)
		}
		if res.IsObject() || res.IsArray() {
			return []byte(res.Raw), nil
		}
		if res.Type == gjson.Null {
			return []byte("null"), nil
		}
		return []byte(res.String()), nil
	},
})

var dictToJSONAction = New(Definition[map[string]any, []byte]{
	Doc:          "Write a dict as pretty JSON",
	Names:        []string{"tojson"},
	Type:         TransformAction,
	InputFormat:  DictFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in map[string]any) ([]byte, error) {
		return json.MarshalIndent(in, "", "  ")
	},
})

var dictKeysAction = New(Definition[map[string]any, []string]{
	Doc:          "Get the keys of a dict as a sorted list",
	Names:        []string{"keys"},
	Type:         TransformAction,
	InputFormat:  DictFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in map[string]any) ([]string, error) {
		return slices.Sorted(maps.Keys(in)), nil
	},
})

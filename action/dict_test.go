package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_Dict(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TextFormat, "dict")

	out, err := a.Transform(NewDataText([]byte(`{"name":{"first":"Janet","last":"Prichard"},"age":47}`)))
	require.NoError(t, err)
	require.Equal(t, DictFormat, out.Format)
	m, ok := out.Value.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(47), m["age"])

	out, err = a.Transform(NewDataText([]byte("name:\n  first: Janet\n  last: Prichard\nage: 47\n")))
	require.NoError(t, err)
	require.Equal(t, DictFormat, out.Format)
	m, ok = out.Value.(map[string]any)
	require.True(t, ok)
	require.Equal(t, 47, m["age"])

	for _, in := range []string{"hello", "[1,2,3]", "{invalid", "null", "42"} {
		_, err := a.Transform(NewDataText([]byte(in)))
		require.Error(t, err, in)
	}
}

func TestAction_DictGet(t *testing.T) {
	r := NewRegistry()

	d := NewDataDict(map[string]any{
		"name":     map[string]any{"first": "Janet", "last": "Prichard"},
		"age":      47,
		"nothing":  nil,
		"children": []any{"Sara", "Alex", "Jack"},
	})

	a := r.MustActionByName(DictFormat, "get")
	require.NoError(t, a.SetInputParameters("name.last"))
	out, err := a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, "Prichard", string(out.RawValue))

	require.NoError(t, a.SetInputParameters("age"))
	out, err = a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, "47", string(out.RawValue))

	require.NoError(t, a.SetInputParameters("children.#"))
	out, err = a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, "3", string(out.RawValue))

	require.NoError(t, a.SetInputParameters("children"))
	out, err = a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, `["Sara","Alex","Jack"]`, string(out.RawValue))

	require.NoError(t, a.SetInputParameters("nothing"))
	out, err = a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, "null", string(out.RawValue))

	b := r.MustActionByName(DictFormat, "gjson")
	require.NoError(t, b.SetInputParameters("name.first"))
	out, err = b.Transform(d)
	require.NoError(t, err)
	require.Equal(t, "Janet", string(out.RawValue))

	require.NoError(t, a.SetInputParameters("nope.path"))
	_, err = a.Transform(d)
	require.Error(t, err)
}

func TestAction_DictToJSONKeys(t *testing.T) {
	r := NewRegistry()

	d := NewDataDict(map[string]any{"b": 2, "a": "x"})

	a := r.MustActionByName(DictFormat, "tojson")
	out, err := a.Transform(d)
	require.NoError(t, err)
	require.Equal(t, TextFormat, out.Format)
	require.Equal(t, "{\n  \"a\": \"x\",\n  \"b\": 2\n}", string(out.RawValue))

	k := r.MustActionByName(DictFormat, "keys")
	out, err = k.Transform(d)
	require.NoError(t, err)
	require.Equal(t, TextListFormat, out.Format)
	require.Equal(t, []string{"a", "b"}, out.Value)
}

func TestAction_DictString(t *testing.T) {
	d := NewDataDict(map[string]any{"b": 2, "a": "x"})
	require.Equal(t, "{\n  \"a\": \"x\",\n  \"b\": 2\n}", d.String())
}

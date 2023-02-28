package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// TextTransformer transforms a byte value into another
type TextTransformer interface {
	Transform([]byte) ([]byte, error)
}

type LineTransformer interface {
	Transform([]byte) ([]byte, error)
}

type DataTransformer interface {
	Transform(data Data, format string) (Data, error)
}

type Data struct {
	Format           Format
	UnstructuredData []any
	StructuredData   map[string]any
}

type ActionType uint16

const (
	TransformAction ActionType = iota
	ParseAction
)

type Action struct {
	Doc          string
	Names        []string // command and aliases
	Type         ActionType
	InputFormat  Format
	OutputFormat Format
	Func         func([]byte) ([]byte, error)
}

type Format struct {
	Name string
}

var (
	text = Format{"text"}
	bin  = Format{"bin"}
)

var upperAction = Action{
	Doc:          "transform text to uppercase",
	Names:        []string{"upper"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(strings.ToUpper(string(in))), nil
	},
}

var lowerAction = Action{
	Doc:          "transform text to lowercase",
	Names:        []string{"lower"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(strings.ToLower(string(in))), nil
	},
}

var fromHexAction = Action{
	Doc:          "decode hex text",
	Names:        []string{"hex"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return hex.DecodeString(string(in))
	},
}

var toHexAction = Action{
	Doc:          "encode binary to hex text",
	Names:        []string{"tohex"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(hex.EncodeToString(in)), nil
	},
}

type ActionRegistry struct {
	m map[string]*Action
}

func NewRegistry() *ActionRegistry {
	m := make(map[string]*Action)

	for _, action := range []Action{upperAction, lowerAction, toHexAction, fromHexAction} {
		a := action
		for _, name := range a.Names {
			m[name] = &a
		}
	}

	return &ActionRegistry{
		m: m,
	}
}

func (r *ActionRegistry) TextAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", action)
	}
	if a.InputFormat != text {
		return nil, fmt.Errorf("action %s does not take text input", action)
	}
	return a.Func(in)
}

func (r *ActionRegistry) BinAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist", action)
	}
	if a.InputFormat != bin {
		return nil, fmt.Errorf("action %s does not take binary input", action)
	}
	return a.Func(in)
}

func (a *Action) TextTransform(in []byte) ([]byte, error) {
	return a.Func(in)
}

func main() {
	r := NewRegistry()

	out, err := r.TextAction("upper", []byte("my text"))
	fmt.Println(string(out), err)
	out, err = r.TextAction("tohex", out)
	fmt.Println(out, err)
	out, err = r.TextAction("hex", out)
	fmt.Println(string(out), err)
}

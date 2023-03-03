package action

import (
	"fmt"
	"time"
)

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
	Func         func(any) (any, error)
}

type Data struct {
	Format           Format
	UnstructuredData []any
	StructuredData   map[string]any
}

type ActionType uint16

type Format struct {
	Name   string
	Prefix string
}

var (
	textFormat = Format{"text", "t"}
	binFormat  = Format{"bin", "b"}
	timeFormat = Format{"time", "T"}
)

func (a *Action) Transform(in any) (any, error) {
	switch {
	case a.InputFormat == textFormat && a.OutputFormat == textFormat:
		b, ok := in.([]byte)
		if !ok {
			return nil, fmt.Errorf("input is not []byte")
		}
		return a.textTransform(b)
	case a.InputFormat == textFormat && a.OutputFormat == timeFormat:
		t, ok := in.([]byte)
		if !ok {
			return nil, fmt.Errorf("input is not []byte")
		}
		return a.textTimeTransform(t)
	default:
		return nil, fmt.Errorf("unknwon format")
	}
}

func (a *Action) textTransform(in []byte) ([]byte, error) {
	at, err := a.Func(in)
	t, ok := at.([]byte)
	if !ok {
		return nil, fmt.Errorf("function does not return []byte")
	}
	return t, err
}

func (a *Action) textTimeTransform(in []byte) (time.Time, error) {
	at, err := a.Func(in)
	t, ok := at.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("function does not return a time.Time")
	}
	return t, err
}

func (a *Action) timeTransform(in time.Time) (time.Time, error) {
	at, err := a.Func(in)
	t, ok := at.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("function does not return a time.Time")
	}
	return t, err
}

func (a *Action) Title() string {
	return a.Names[0]
}

func (a *Action) Description() string {
	return a.Doc
}

func (a *Action) FilterValue() string {
	return a.Title()
}

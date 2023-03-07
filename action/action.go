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

type Format struct {
	Name   string
	Prefix string
}

type ActionType uint16

var (
	textFormat = Format{"text", "t"}
	binFormat  = Format{"bin", "b"}
	timeFormat = Format{"time", "T"}
)

func (a *Action) Transform(in *Data) (*Data, error) {
	switch {
	case a.InputFormat == textFormat && a.OutputFormat == textFormat:
		d, err := a.textTransform(in)
		return d, err
	case a.InputFormat == textFormat && a.OutputFormat == timeFormat:
		return a.textTimeTransform(in)
	case a.InputFormat == timeFormat && a.OutputFormat == timeFormat:
		return a.timeTransform(in)
	default:
		return nil, fmt.Errorf("unknwon format")
	}
}

func (a *Action) textTransform(in *Data) (*Data, error) {
	ab, err := a.Func(in.RawValue)
	b, ok := ab.([]byte)
	if !ok {
		return nil, fmt.Errorf("function does not return []byte")
	}
	return in.StoreTextValue(b, a), err
}

func (a *Action) textTimeTransform(in *Data) (*Data, error) {
	ab, err := a.Func(in.RawValue)
	b, ok := ab.(time.Time)
	if !ok {
		return nil, fmt.Errorf("function does not return a time.Time")
	}
	return in.StoreTimeValue(b, a), err
}

func (a *Action) timeTransform(in *Data) (*Data, error) {
	t, ok := in.Value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("input not a time.Time")
	}
	at, err := a.Func(t)
	ot, ok := at.(time.Time)
	if !ok {
		return nil, fmt.Errorf("function does not return a time.Time")
	}
	return in.StoreTimeValue(ot, a), err
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

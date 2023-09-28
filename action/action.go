package action

import (
	"fmt"
	"time"

	"github.com/peterstace/simplefeatures/geom"
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
	// change it to a variadic opts ...
	Func func(any) (any, error)
}

type Format struct {
	Name   string
	Prefix string
}

type ActionType uint16

var (
	textFormat     = Format{"text", "t"}
	binFormat      = Format{"bin", "b"}
	timeFormat     = Format{"time", "T"}
	jsonFormat     = Format{"json", "j"}
	geoFormat      = Format{"geometry", "g"}
	textListFormat = Format{"textList", "l"}
)

func (a *Action) Transform(in *Data) (*Data, error) {
	var data any
	var err error

	switch a.InputFormat {
	case textFormat:
		// the input format of the action needs to be applied to all
		// list members if tbe data is textListFormat
		if in.Format == textListFormat {
			l, ok := in.Value.([]string)
			if !ok {
				return nil, fmt.Errorf("input not a list of string")
			}

			resp := make([]string, len(l))
			for i, s := range l {
				v, err := a.Func([]byte(s))
				if err != nil {
					return nil, err
				}
				resp[i] = string(v.([]byte))
			}
			data = resp
			a.OutputFormat = textListFormat
		} else {
			if len(in.RawValue) == 0 {
				return nil, fmt.Errorf("value is empty")
			}
			data, err = a.Func(in.RawValue)
			if err != nil {
				return nil, err
			}
		}

	case textListFormat:
		_, ok := in.Value.([]string)
		if !ok {
			return nil, fmt.Errorf("input not a list of string")
		}
		data, err = a.Func(in.Value)
		if err != nil {
			return nil, err
		}
	case geoFormat:
		_, ok := in.Value.(geom.Geometry)
		if !ok {
			return nil, fmt.Errorf("input not a geometry")
		}
		data, err = a.Func(in.Value)
		if !ok {
			return nil, err
		}

	case timeFormat:
		_, ok := in.Value.(time.Time)
		if !ok {
			return nil, fmt.Errorf("input not a time.Time")
		}
		data, err = a.Func(in.Value)
		if !ok {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown input format")
	}

	switch a.OutputFormat {
	case textFormat:
		b, ok := data.([]byte)
		if !ok {
			return nil, fmt.Errorf("function does not return []byte")
		}
		return in.StoreTextValue(b, a), err
	case textListFormat:
		l, ok := data.([]string)
		if !ok {
			return nil, fmt.Errorf("function does not return a []string")
		}
		return in.StoreTextListValue(l, a), err
	case timeFormat:
		b, ok := data.(time.Time)
		if !ok {
			return nil, fmt.Errorf("function does not return a time.Time")
		}
		return in.StoreTimeValue(b, a), err
	case geoFormat:
		g, ok := data.(geom.Geometry)
		if !ok {
			return nil, fmt.Errorf("function does not return a geom")
		}
		return in.StoreGeomValue(g, a), err

	default:
		return nil, fmt.Errorf("unknown output format")
	}
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

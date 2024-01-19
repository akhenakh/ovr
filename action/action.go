package action

import (
	"fmt"
	"strings"
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

type Actions []*Action

type Format struct {
	Name   string
	Prefix string
}

type ActionType uint16

var (
	TextFormat     = Format{"text", "t"}
	BinFormat      = Format{"bin", "b"}
	TimeFormat     = Format{"time", "T"}
	JSONFormat     = Format{"json", "j"}
	GeoFormat      = Format{"geometry", "g"}
	TextListFormat = Format{"textList", "l"}
)

func (a *Action) Transform(in *Data) (*Data, error) {
	var data any
	var err error

	switch a.InputFormat {
	case TextFormat:
		// the input format of the action needs to be applied to all
		// list members if tbe data is textListFormat
		if in.Format == TextListFormat {
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
			a.OutputFormat = TextListFormat
		} else {
			if len(in.RawValue) == 0 {
				return nil, fmt.Errorf("value is empty")
			}
			data, err = a.Func(in.RawValue)
			if err != nil {
				return nil, err
			}
		}

	case TextListFormat:
		_, ok := in.Value.([]string)
		if !ok {
			return nil, fmt.Errorf("input not a list of string")
		}
		data, err = a.Func(in.Value)
		if err != nil {
			return nil, err
		}
	case GeoFormat:
		_, ok := in.Value.(geom.Geometry)
		if !ok {
			return nil, fmt.Errorf("input not a geometry")
		}
		data, err = a.Func(in.Value)
		if !ok {
			return nil, err
		}

	case TimeFormat:
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
	case TextFormat:
		b, ok := data.([]byte)
		if !ok {
			return nil, fmt.Errorf("function does not return []byte")
		}
		return in.StoreTextValue(b, a), err
	case TextListFormat:
		l, ok := data.([]string)
		if !ok {
			return nil, fmt.Errorf("function does not return a []string")
		}
		return in.StoreTextListValue(l, a), err
	case TimeFormat:
		b, ok := data.(time.Time)
		if !ok {
			return nil, fmt.Errorf("function does not return a time.Time")
		}
		return in.StoreTimeValue(b, a), err
	case GeoFormat:
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
	return strings.Title(a.Names[0])
}

func (a *Action) Description() string {
	return a.Doc
}

func (a *Action) FilterValue() string {
	return a.Title()
}

func (a *Action) FullDescription() string {
	return a.Title() + ": " + a.Doc
}

func (actions Actions) Len() int {
	return len(actions)
}

// String returns a full description + name
// used for display
func (actions Actions) String(idx int) string {
	return actions[idx].Description()
}

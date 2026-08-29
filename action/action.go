package action

import (
	"fmt"
	"time"
	"unicode"

	"github.com/peterstace/simplefeatures/geom"
)

const (
	TransformAction ActionType = iota
	ParseAction
)

type ActionType uint16

type Format struct {
	Name   string
	Prefix string
}

var (
	TextFormat     = Format{"text", "t"}
	BinFormat      = Format{"bin", "b"}
	TimeFormat     = Format{"time", "T"}
	JSONFormat     = Format{"json", "j"}
	GeoFormat      = Format{"geometry", "g"}
	TextListFormat = Format{"textList", "l"}
	TableFormat    = Format{"table", "r"}
)

type ActionParameter struct {
	ActionParameterType
	Doc string
}

type ActionParameterType int

const (
	IntParameter ActionParameterType = iota
	FloatParameter
	StringParameter
)

// Func is a typed transformation function.
type Func[I, O any] func(a Action, in I) (O, error)

// Definition describes an action with typed input and output.
type Definition[I, O any] struct {
	Doc          string
	Names        []string
	Type         ActionType
	InputFormat  Format
	OutputFormat Format
	Parameters   []ActionParameter
	Interactive  bool
	Func         Func[I, O]
}

// Action is a type erased transformation usable in registries and data stacks.
type Action interface {
	Doc() string
	Names() []string
	Type() ActionType
	InputFormat() Format
	OutputFormat() Format
	Parameters() []ActionParameter
	InputParameters() []any
	Interactive() bool

	SetInputParameters(values ...any) error

	Transform(in *Data) (*Data, error)

	Title() string
	Description() string
	FullDescription() string
	FilterValue() string

	clone() Action
}

// New creates a type erased Action from a typed Definition.
// The definition formats must be compatible with the input and output types.
func New[I, O any](d Definition[I, O]) Action {
	return &erased[I, O]{
		def: d,
		in:  inputFunc[I](d.InputFormat),
		out: outputFunc[O](d.OutputFormat),
	}
}

type erased[I, O any] struct {
	def    Definition[I, O]
	in     func(*Data) (I, error)
	out    func(O, Action, *Data) (*Data, error)
	params []any
}

func (e *erased[I, O]) Doc() string                   { return e.def.Doc }
func (e *erased[I, O]) Names() []string               { return e.def.Names }
func (e *erased[I, O]) Type() ActionType              { return e.def.Type }
func (e *erased[I, O]) InputFormat() Format           { return e.def.InputFormat }
func (e *erased[I, O]) OutputFormat() Format          { return e.def.OutputFormat }
func (e *erased[I, O]) Parameters() []ActionParameter { return e.def.Parameters }
func (e *erased[I, O]) InputParameters() []any        { return e.params }
func (e *erased[I, O]) Interactive() bool             { return e.def.Interactive }

func (e *erased[I, O]) clone() Action {
	c := *e
	c.params = nil
	return &c
}

func (e *erased[I, O]) SetInputParameters(values ...any) error {
	if len(values) != len(e.def.Parameters) {
		return fmt.Errorf("%s expects %d parameters, got %d", e.Title(), len(e.def.Parameters), len(values))
	}

	for i, p := range e.def.Parameters {
		switch p.ActionParameterType {
		case IntParameter:
			if _, ok := values[i].(int); !ok {
				return fmt.Errorf("%s parameter at position %d is not an int: %T", e.Title(), i, values[i])
			}
		case FloatParameter:
			if _, ok := values[i].(float64); !ok {
				return fmt.Errorf("%s parameter at position %d is not a float: %T", e.Title(), i, values[i])
			}
		case StringParameter:
			if _, ok := values[i].(string); !ok {
				return fmt.Errorf("%s parameter at position %d is not a string: %T", e.Title(), i, values[i])
			}
		}
	}

	e.params = values
	return nil
}

func (e *erased[I, O]) Transform(in *Data) (*Data, error) {
	if len(e.params) != len(e.def.Parameters) {
		return nil, fmt.Errorf("input parameters required for %s", e.Title())
	}

	// the input format of the action needs to be applied to all
	// list members if the data is a text list
	if in.Format == TextListFormat && e.def.InputFormat == TextFormat && e.def.OutputFormat == TextFormat && isTextPair[I, O]() {
		l, ok := in.Value.([]string)
		if !ok {
			return nil, fmt.Errorf("input not a list of string")
		}

		resp := make([]string, len(l))
		for i, s := range l {
			v, err := e.def.Func(e, any([]byte(s)).(I))
			if err != nil {
				return nil, err
			}
			b, ok := any(v).([]byte)
			if !ok {
				return nil, fmt.Errorf("function does not return []byte")
			}
			resp[i] = string(b)
		}
		return in.StoreTextListValue(resp, e), nil
	}

	d, err := e.in(in)
	if err != nil {
		return nil, err
	}

	out, err := e.def.Func(e, d)
	if err != nil {
		return nil, err
	}

	return e.out(out, e, in)
}

func (e *erased[I, O]) Title() string {
	return capitalize(e.def.Names[0])
}

func (e *erased[I, O]) FullDescription() string {
	return e.Title() + ": " + e.def.Doc
}

func (e *erased[I, O]) Description() string {
	return e.def.Doc
}

func (e *erased[I, O]) FilterValue() string {
	return e.Title()
}

// Actions is a list of actions for display purposes.
type Actions []Action

func (actions Actions) Len() int {
	return len(actions)
}

// String returns a full description + name
// used for display
func (actions Actions) String(idx int) string {
	return actions[idx].FullDescription()
}

func isTextPair[I, O any]() bool {
	var i I
	var o O
	_, okI := any(&i).(*[]byte)
	_, okO := any(&o).(*[]byte)
	return okI && okO
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func inputFunc[I any](f Format) func(*Data) (I, error) {
	return func(d *Data) (I, error) {
		var v I
		switch p := any(&v).(type) {
		case *[]byte:
			switch f {
			case TextFormat:
				if d.Format != TextFormat {
					return v, fmt.Errorf("input is not text")
				}
				if len(d.RawValue) == 0 {
					return v, fmt.Errorf("value is empty")
				}
				*p = d.RawValue
			case BinFormat:
				if d.Format != BinFormat {
					return v, fmt.Errorf("input is not binary")
				}
				*p = d.RawValue
			default:
				return v, fmt.Errorf("%s is not a valid input format for text", f.Name)
			}
		case *[]string:
			if f != TextListFormat || d.Format != TextListFormat {
				return v, fmt.Errorf("input is not a text list")
			}
			l, ok := d.Value.([]string)
			if !ok {
				return v, fmt.Errorf("input not a list of string")
			}
			*p = l
		case *[][]string:
			if f != TableFormat || d.Format != TableFormat {
				return v, fmt.Errorf("input is not a table")
			}
			t, ok := d.Value.([][]string)
			if !ok {
				return v, fmt.Errorf("input not a table")
			}
			*p = t
		case *time.Time:
			if f != TimeFormat || d.Format != TimeFormat {
				return v, fmt.Errorf("input is not a time")
			}
			t, ok := d.Value.(time.Time)
			if !ok {
				return v, fmt.Errorf("input not a time.Time")
			}
			*p = t
		case *geom.Geometry:
			if f != GeoFormat || d.Format != GeoFormat {
				return v, fmt.Errorf("input is not a geometry")
			}
			g, ok := d.Value.(geom.Geometry)
			if !ok {
				return v, fmt.Errorf("input not a geometry")
			}
			*p = g
		default:
			return v, fmt.Errorf("unsupported input type %T", v)
		}
		return v, nil
	}
}

func outputFunc[O any](f Format) func(O, Action, *Data) (*Data, error) {
	return func(v O, a Action, d *Data) (*Data, error) {
		switch p := any(&v).(type) {
		case *[]byte:
			if f != TextFormat {
				return nil, fmt.Errorf("%s is not a valid output format for text", f.Name)
			}
			return d.StoreTextValue(*p, a), nil
		case *[]string:
			if f != TextListFormat {
				return nil, fmt.Errorf("%s is not a valid output format for a text list", f.Name)
			}
			return d.StoreTextListValue(*p, a), nil
		case *[][]string:
			if f != TableFormat {
				return nil, fmt.Errorf("%s is not a valid output format for a table", f.Name)
			}
			return d.StoreTableValue(*p, a), nil
		case *time.Time:
			if f != TimeFormat {
				return nil, fmt.Errorf("%s is not a valid output format for time", f.Name)
			}
			return d.StoreTimeValue(*p, a), nil
		case *geom.Geometry:
			if f != GeoFormat {
				return nil, fmt.Errorf("%s is not a valid output format for geometry", f.Name)
			}
			return d.StoreGeomValue(*p, a), nil
		default:
			return nil, fmt.Errorf("unsupported output type %T", v)
		}
	}
}

package action

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/peterstace/simplefeatures/geom"
)

// Data to hold the current state of the input and the stack of applied transformations
type Data struct {
	RawValue       []byte
	Value          any
	Format         Format
	StructuredData map[string]any
	Stack          []*Action
}

var ErrEmptyStack = errors.New("empty stack")

func NewDataText(v []byte) *Data {
	return &Data{RawValue: v, Format: TextFormat}
}

func (d *Data) StoreTextValue(v []byte, a *Action) *Data {
	return &Data{RawValue: v, Format: TextFormat, Stack: append(d.Stack, a)}
}

func (d *Data) StoreTextListValue(l []string, a *Action) *Data {
	return &Data{Value: l, Format: TextListFormat, Stack: append(d.Stack, a)}
}

func (d *Data) StoreTimeValue(t time.Time, a *Action) *Data {
	return &Data{Value: t, Stack: append(d.Stack, a), Format: TimeFormat}
}

func (d *Data) StoreGeomValue(g geom.Geometry, a *Action) *Data {
	return &Data{Value: g, Stack: append(d.Stack, a), Format: GeoFormat}
}

func (d *Data) StoreJSONValue(t time.Time, a *Action) *Data {
	return &Data{Value: t, Stack: append(d.Stack, a), Format: JSONFormat}
}

// Undo removed the last actions if any
// Reapply the stack with input
func (d *Data) Undo(in []byte) (*Data, *Action, error) {
	if len(d.Stack) == 0 {
		return nil, nil, ErrEmptyStack
	}
	var oa *Action

	oa, d.Stack = d.Stack[len(d.Stack)-1], d.Stack[:len(d.Stack)-1]

	nd := NewDataText(in)

	for _, a := range d.Stack {
		out, err := a.Transform(nd)
		if err != nil {
			return nil, nil, err
		}
		nd = out
	}

	return nd, oa, nil
}

func (d *Data) String() string {
	switch d.Format {
	case TextFormat:
		return string(d.RawValue)
	case TimeFormat:
		t := d.Value.(time.Time)
		return t.String()
	default:
		return fmt.Sprintf("%v", d.Value)
	}
}

func (d *Data) StackString() string {
	names := make([]string, len(d.Stack))
	for i, a := range d.Stack {
		names[i] = a.Title()
	}
	return strings.Join(names, ",")
}

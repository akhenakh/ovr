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
	Stack          []Action
}

var ErrEmptyStack = errors.New("empty stack")

func NewDataText(v []byte) *Data {
	return &Data{RawValue: v, Format: TextFormat}
}

func NewDataBin(v []byte) *Data {
	return &Data{RawValue: v, Format: BinFormat}
}

func NewDataTextList(l []string) *Data {
	return &Data{Value: l, Format: TextListFormat}
}

func NewDataTable(t [][]string) *Data {
	return &Data{Value: t, Format: TableFormat}
}

func NewDataTime(t time.Time) *Data {
	return &Data{Value: t, Format: TimeFormat}
}

func NewDataGeom(g geom.Geometry) *Data {
	return &Data{Value: g, Format: GeoFormat}
}

func (d *Data) StoreTextValue(v []byte, a Action) *Data {
	return &Data{RawValue: v, Format: TextFormat, Stack: append(d.Stack, a)}
}

func (d *Data) StoreTextListValue(l []string, a Action) *Data {
	return &Data{Value: l, Format: TextListFormat, Stack: append(d.Stack, a)}
}

func (d *Data) StoreTableValue(t [][]string, a Action) *Data {
	return &Data{Value: t, Format: TableFormat, Stack: append(d.Stack, a)}
}

func (d *Data) StoreTimeValue(t time.Time, a Action) *Data {
	return &Data{Value: t, Stack: append(d.Stack, a), Format: TimeFormat}
}

func (d *Data) StoreGeomValue(g geom.Geometry, a Action) *Data {
	return &Data{Value: g, Stack: append(d.Stack, a), Format: GeoFormat}
}

func (d *Data) StoreJSONValue(v any, a Action) *Data {
	return &Data{Value: v, Stack: append(d.Stack, a), Format: JSONFormat}
}

// Undo removed the last actions if any
// Reapply the stack with input
func (d *Data) Undo(in []byte) (*Data, Action, error) {
	if len(d.Stack) == 0 {
		return nil, nil, ErrEmptyStack
	}
	var oa Action

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
	case TextFormat, BinFormat:
		return string(d.RawValue)
	case TimeFormat:
		t := d.Value.(time.Time)
		return t.String()
	case GeoFormat:
		g := d.Value.(geom.Geometry)
		return g.AsText()
	case TableFormat:
		t, ok := d.Value.([][]string)
		if !ok {
			return fmt.Sprintf("%v", d.Value)
		}
		rows := make([]string, len(t))
		for i, r := range t {
			rows[i] = strings.Join(r, ",")
		}
		return strings.Join(rows, "\n")
	case TextListFormat:
		l, ok := d.Value.([]string)
		if !ok {
			return fmt.Sprintf("%v", d.Value)
		}
		return strings.Join(l, "\n")
	default:
		return fmt.Sprintf("%v", d.Value)
	}
}

func (d *Data) StackString() string {
	names := make([]string, len(d.Stack))
	for i, a := range d.Stack {
		names[i] = a.Title()
		if params := a.InputParameters(); len(params) > 0 {
			strs := make([]string, len(params))
			for j, p := range params {
				if s, ok := p.(string); ok {
					strs[j] = fmt.Sprintf("%q", s)
				} else {
					strs[j] = fmt.Sprintf("%v", p)
				}
			}
			names[i] += "(" + strings.Join(strs, ", ") + ")"
		}
	}
	return strings.Join(names, ",")
}

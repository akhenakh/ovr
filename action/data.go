package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/peterstace/simplefeatures/geom"
)

// Data to hold the current state of the input and the stack of applied transformations
type Data struct {
	RawValue       []byte
	Value          any
	Format         Format
	StructuredData map[string]any
	Stack          []Action

	strOnce sync.Once
	str     string
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

func NewDataDict(m map[string]any) *Data {
	return &Data{Value: m, Format: DictFormat}
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

func (d *Data) StoreDictValue(m map[string]any, a Action) *Data {
	return &Data{Value: m, Stack: append(d.Stack, a), Format: DictFormat}
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

// String returns the entry as its string representation, the result is
// computed once and memoized, transforming returns new Data values
func (d *Data) String() string {
	d.strOnce.Do(func() {
		d.str = d.computeString()
	})
	return d.str
}

// Preview returns a single line string representation of the entry,
// whitespace collapsed and truncated to maxRunes runes,
// it is meant for titles and status lines, big entries must not
// end up rendered or styled on every frame
func (d *Data) Preview(maxRunes int) string {
	s := d.String()
	limit := maxRunes * utf8.UTFMax
	if len(s) > limit {
		s = s[:limit]
		for len(s) > 0 && !utf8.RuneStart(s[len(s)-1]) {
			s = s[:len(s)-1]
		}
	}
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > maxRunes {
		return string(r[:maxRunes]) + "…"
	}
	return s
}

func (d *Data) computeString() string {
	switch d.Format {
	case TextFormat, BinFormat:
		return string(d.RawValue)
	case TimeFormat:
		t := d.Value.(time.Time)
		return t.String()
	case DictFormat:
		m, ok := d.Value.(map[string]any)
		if !ok {
			return fmt.Sprintf("%v", d.Value)
		}
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", d.Value)
		}
		return string(b)
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

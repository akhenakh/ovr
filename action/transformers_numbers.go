package action

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var intBaseAction = New(Definition[[]byte, []byte]{
	Doc:          "Convert an integer between bases, 0 auto-detects a 0x/0b/0o prefix, e.g. from 16 to 10 turns ff into 255",
	Names:        []string{"intbase"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters: []ActionParameter{
		{IntParameter, "source base, 2 to 36, 0 to auto-detect a 0x/0b/0o prefix"},
		{IntParameter, "target base, 2 to 36"},
	},
	Func: func(a Action, in []byte) ([]byte, error) {
		params := a.InputParameters()
		from, ok := params[0].(int)
		if !ok {
			return nil, fmt.Errorf("source base parameter is not an int")
		}
		to, ok := params[1].(int)
		if !ok {
			return nil, fmt.Errorf("target base parameter is not an int")
		}
		if from != 0 && (from < 2 || from > 36) {
			return nil, fmt.Errorf("source base must be 0 or between 2 and 36, got %d", from)
		}
		if to < 2 || to > 36 {
			return nil, fmt.Errorf("target base must be between 2 and 36, got %d", to)
		}

		n, err := parseBaseInt(string(in), from)
		if err != nil {
			return nil, err
		}
		return []byte(n.Text(to)), nil
	},
})

// parseBaseInt parses a possibly signed integer in base from, 0 auto-detects
// the 0x/0b/0o prefixes like Go literals.
func parseBaseInt(s string, from int) (*big.Int, error) {
	s = strings.ReplaceAll(strings.TrimSpace(s), "_", "")
	if s == "" {
		return nil, fmt.Errorf("empty input")
	}

	neg := false
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		neg = s[0] == '-'
		s = s[1:]
	}

	lower := strings.ToLower(s)
	switch {
	case from == 0 && strings.HasPrefix(lower, "0x"):
		from, s = 16, s[2:]
	case from == 0 && strings.HasPrefix(lower, "0b"):
		from, s = 2, s[2:]
	case from == 0 && strings.HasPrefix(lower, "0o"):
		from, s = 8, s[2:]
	case from == 0:
		from = 10
	case from == 16 && strings.HasPrefix(lower, "0x"):
		s = s[2:]
	case from == 2 && strings.HasPrefix(lower, "0b"):
		s = s[2:]
	case from == 8 && strings.HasPrefix(lower, "0o"):
		s = s[2:]
	}

	n, ok := new(big.Int).SetString(s, from)
	if !ok {
		return nil, fmt.Errorf("%q is not a valid base %d integer", s, from)
	}
	if neg {
		n.Neg(n)
	}
	return n, nil
}

// unit is a measurement unit with its factor to the base unit of its family
// and, for temperatures, an offset applied before scaling.
type unit struct {
	family string
	factor float64
	offset float64
}

// units holds every unit the convert action understands.
var units = map[string]unit{
	// length, base meter
	"mm":  {"length", 0.001, 0},
	"cm":  {"length", 0.01, 0},
	"m":   {"length", 1, 0},
	"km":  {"length", 1000, 0},
	"in":  {"length", 0.0254, 0},
	"ft":  {"length", 0.3048, 0},
	"yd":  {"length", 0.9144, 0},
	"mi":  {"length", 1609.344, 0},
	"nmi": {"length", 1852, 0},

	// mass, base kilogram
	"mg": {"mass", 1e-6, 0},
	"g":  {"mass", 0.001, 0},
	"kg": {"mass", 1, 0},
	"t":  {"mass", 1000, 0},
	"oz": {"mass", 0.028349523125, 0},
	"lb": {"mass", 0.45359237, 0},

	// data, base byte
	"b":   {"data", 1, 0},
	"kb":  {"data", 1e3, 0},
	"mb":  {"data", 1e6, 0},
	"gb":  {"data", 1e9, 0},
	"tb":  {"data", 1e12, 0},
	"kib": {"data", 1 << 10, 0},
	"mib": {"data", 1 << 20, 0},
	"gib": {"data", 1 << 30, 0},
	"tib": {"data", 1 << 40, 0},

	// time, base second
	"ms":  {"time", 0.001, 0},
	"s":   {"time", 1, 0},
	"min": {"time", 60, 0},
	"h":   {"time", 3600, 0},
	"d":   {"time", 86400, 0},
	"w":   {"time", 604800, 0},

	// temperature, base celsius
	"c": {"temp", 1, 0},
	"f": {"temp", 5.0 / 9.0, -32},
	"k": {"temp", 1, -273.15},
}

var convertInputRe = regexp.MustCompile(`^\s*([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)\s*([a-zA-Z]*)\s*$`)

var convertAction = New(Definition[[]byte, []byte]{
	Doc:          "Convert a measurement, the unit is in the input or given as 'from>to', e.g. 12 km with to mi, or 12 with km>mi; supports length, mass, data, time and temperature",
	Names:        []string{"convert"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "the target unit, or 'from>to', e.g. mi or km>mi"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("unit parameter is not a string")
		}
		fromUnit, toUnit := "", ""
		if before, after, found := strings.Cut(p, ">"); found {
			fromUnit, toUnit = strings.TrimSpace(before), strings.TrimSpace(after)
		} else {
			toUnit = strings.TrimSpace(p)
		}

		m := convertInputRe.FindStringSubmatch(string(in))
		if m == nil {
			return nil, fmt.Errorf("can't parse %q as a number with a unit", strings.TrimSpace(string(in)))
		}
		value, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return nil, fmt.Errorf("can't parse %q as a number: %w", m[1], err)
		}
		if m[2] != "" {
			fromUnit = m[2]
		}
		if fromUnit == "" || toUnit == "" {
			return nil, fmt.Errorf("a source and a target unit are required, e.g. 12km mi or 12 km>mi")
		}

		out, err := convertValue(value, fromUnit, toUnit)
		if err != nil {
			return nil, err
		}
		return []byte(strconv.FormatFloat(out, 'g', 6, 64) + " " + strings.ToLower(toUnit)), nil
	},
})

// convertValue converts value from one unit to another of the same family.
func convertValue(value float64, fromUnit, toUnit string) (float64, error) {
	from, ok := units[strings.ToLower(fromUnit)]
	if !ok {
		return 0, fmt.Errorf("unknown unit %q", fromUnit)
	}
	to, ok := units[strings.ToLower(toUnit)]
	if !ok {
		return 0, fmt.Errorf("unknown unit %q", toUnit)
	}
	if from.family != to.family {
		return 0, fmt.Errorf("can't convert %s (%s) to %s (%s)", fromUnit, from.family, toUnit, to.family)
	}

	if from.family == "temp" {
		// normalize to celsius, then to the target scale
		celsius := (value + from.offset) * from.factor
		return celsius/to.factor - to.offset, nil
	}
	return value * from.factor / to.factor, nil
}

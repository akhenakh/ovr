//go:generate go run ../tools/gen-tznames

package action

import (
	"fmt"
	"strings"
	"time"

	// embed the tz database so timezone actions work on systems without one
	_ "time/tzdata"
)

type timezoneDef struct {
	names    []string
	location string
	doc      string
}

var timezones = []timezoneDef{
	{[]string{"utc"}, "UTC", "Change time to UTC timezone"},
	{[]string{"et", "est"}, "America/New_York", "Change time to US Eastern timezone"},
	{[]string{"ct", "cst"}, "America/Chicago", "Change time to US Central timezone"},
	{[]string{"mt", "mst"}, "America/Denver", "Change time to US Mountain timezone"},
	{[]string{"pt", "pst"}, "America/Los_Angeles", "Change time to US Pacific timezone"},
	{[]string{"brt"}, "America/Sao_Paulo", "Change time to Brazil timezone"},
	{[]string{"hst"}, "Pacific/Honolulu", "Change time to Hawaii timezone"},
	{[]string{"gmt"}, "Europe/London", "Change time to GMT timezone"},
	{[]string{"cet"}, "Europe/Paris", "Change time to Central European timezone"},
	{[]string{"eet"}, "Europe/Athens", "Change time to Eastern European timezone"},
	{[]string{"msk"}, "Europe/Moscow", "Change time to Moscow timezone"},
	{[]string{"bst"}, "Europe/London", "Change time to UK timezone"},
	{[]string{"ist"}, "Asia/Kolkata", "Change time to India timezone"},
	{[]string{"sgt"}, "Asia/Singapore", "Change time to Singapore timezone"},
	{[]string{"hkt"}, "Asia/Hong_Kong", "Change time to Hong Kong timezone"},
	{[]string{"jst"}, "Asia/Tokyo", "Change time to Japan timezone"},
	{[]string{"kst"}, "Asia/Seoul", "Change time to Korea timezone"},
	{[]string{"aest"}, "Australia/Sydney", "Change time to Australian Eastern timezone"},
	{[]string{"nzst"}, "Pacific/Auckland", "Change time to New Zealand timezone"},
}

func newTimezoneActions() []Action {
	out := make([]Action, 0, len(timezones))
	for _, tz := range timezones {
		location := tz.location
		out = append(out, New(Definition[time.Time, time.Time]{
			Doc:          tz.doc,
			Names:        tz.names,
			Type:         TransformAction,
			InputFormat:  TimeFormat,
			OutputFormat: TimeFormat,
			Func: func(a Action, in time.Time) (time.Time, error) {
				loc, err := loadLocation(location)
				if err != nil {
					return time.Time{}, fmt.Errorf("unknown timezone %s: %w", location, err)
				}
				return in.In(loc), nil
			},
		}))
	}
	return out
}

// tzNamesFold maps lowercase IANA names to their canonical spelling.
var tzNamesFold = func() map[string]string {
	m := make(map[string]string, len(tzNames))
	for _, n := range tzNames {
		m[strings.ToLower(n)] = n
	}
	return m
}()

// canonicalTZName returns the canonical spelling of a timezone name,
// matching case-insensitively, or "" when the name is not a known IANA zone.
func canonicalTZName(name string) string {
	return tzNamesFold[strings.ToLower(name)]
}

// loadLocation loads a timezone by name, accepting common case mistakes
// like america/montreal for America/Montreal. The canonical spelling is
// resolved against the embedded name list first: on case-insensitive
// filesystems (macOS) time.LoadLocation would happily load the raw name
// and report it back unnormalized.
func loadLocation(name string) (*time.Location, error) {
	if canonical := canonicalTZName(name); canonical != "" {
		name = canonical
	}
	loc, err := time.LoadLocation(name)
	if err == nil {
		return loc, nil
	}
	// fallback for zones newer than the generated name list
	for _, candidate := range locationCaseVariants(name) {
		if loc, err := time.LoadLocation(candidate); err == nil {
			return loc, nil
		}
	}
	return nil, fmt.Errorf("unknown timezone %s: %w", name, err)
}

// locationCaseVariants returns the case variants of a timezone name to try
// when the exact name is not found: America/New_York, AMERICA/NEW_YORK...
func locationCaseVariants(name string) []string {
	segments := strings.Split(name, "/")
	titled := make([]string, len(segments))
	for i, s := range segments {
		words := strings.Split(s, "_")
		for j, w := range words {
			r := []rune(w)
			if len(r) == 0 {
				continue
			}
			words[j] = strings.ToUpper(string(r[0])) + strings.ToLower(string(r[1:]))
		}
		titled[i] = strings.Join(words, "_")
	}

	candidates := []string{
		strings.Join(titled, "/"),
		strings.ToUpper(name),
		strings.ToLower(name),
	}

	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c != name {
			out = append(out, c)
		}
	}
	return out
}

var timezoneActions = newTimezoneActions()

var tzTimeAction = New(Definition[time.Time, time.Time]{
	Doc:          "Change time to a timezone, provide an IANA timezone name like America/New_York, case insensitive",
	Names:        []string{"tz"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TimeFormat,
	Parameters:   []ActionParameter{{StringParameter, "IANA timezone name"}},
	Func: func(a Action, in time.Time) (time.Time, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return time.Time{}, fmt.Errorf("tz parameter is not a string")
		}
		loc, err := loadLocation(p)
		if err != nil {
			return time.Time{}, fmt.Errorf("unknown timezone %s: %w", p, err)
		}
		return in.In(loc), nil
	},
})

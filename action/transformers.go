package action

import (
	"bytes"
	"cmp"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"go.yaml.in/yaml/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var upperAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input with all Unicode letters mapped to their upper case",
	Names:        []string{"upper"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		caser := cases.Upper(language.Und)
		upper := caser.String(string(in))
		return []byte(upper), nil
	},
})

var lowerAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input with all Unicode letters mapped to their lower case",
	Names:        []string{"lower"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		caser := cases.Lower(language.Und)
		lower := caser.String(string(in))
		return []byte(lower), nil
	},
})

var titleAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input title",
	Names:        []string{"title"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		caser := cases.Title(language.Und)
		titleStr := caser.String(string(in))
		return []byte(titleStr), nil
	},
})

var camelAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input to camelCase, e.g. hello world, hello_world, helloWorld to helloWorld",
	Names:        []string{"camel"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		words := splitWords(string(in))
		for i := range words {
			if i > 0 {
				words[i] = capitalize(strings.ToLower(words[i]))
			} else {
				words[i] = strings.ToLower(words[i])
			}
		}
		return []byte(strings.Join(words, "")), nil
	},
})

var snakeAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input to snake_case, e.g. hello world, helloWorld to hello_world",
	Names:        []string{"snake"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		words := splitWords(string(in))
		for i := range words {
			words[i] = strings.ToLower(words[i])
		}
		return []byte(strings.Join(words, "_")), nil
	},
})

var kebabAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input to kebab-case, e.g. hello world, helloWorld to hello-world",
	Names:        []string{"kebab"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		words := splitWords(string(in))
		for i := range words {
			words[i] = strings.ToLower(words[i])
		}
		return []byte(strings.Join(words, "-")), nil
	},
})

var pascalAction = New(Definition[[]byte, []byte]{
	Doc:          "Transforms input to PascalCase, e.g. hello world, hello_world to HelloWorld",
	Names:        []string{"pascal"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		words := splitWords(string(in))
		for i := range words {
			words[i] = capitalize(strings.ToLower(words[i]))
		}
		return []byte(strings.Join(words, "")), nil
	},
})

// splitWords splits text into words on whitespace, separators and case
// boundaries, e.g. "helloWorld", "hello_world" and "hello world" all
// split into the words hello and World, acronyms are kept, e.g.
// "HTTPServer" splits into HTTP and Server
func splitWords(s string) []string {
	var words []string
	runes := []rune(s)
	start := 0
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if i > start {
				words = append(words, string(runes[start:i]))
			}
			start = i + 1
			continue
		}
		if i > start && unicode.IsUpper(r) {
			prev := runes[i-1]
			acronymEnd := unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || acronymEnd {
				words = append(words, string(runes[start:i]))
				start = i
			}
		}
	}
	if start < len(runes) {
		words = append(words, string(runes[start:]))
	}
	return words
}

var trimSpaceAction = New(Definition[[]byte, []byte]{
	Doc:          "Trim spaces from input",
	Names:        []string{"trimspace"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(strings.TrimSpace(string(in))), nil
	},
})

var quoteAction = New(Definition[[]byte, []byte]{
	Doc:          "Quotes string with escape characters",
	Names:        []string{"quote"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(strconv.Quote(string(in))), nil
	},
})

var unquoteAction = New(Definition[[]byte, []byte]{
	Doc:          "Removes quotes from escaped characters",
	Names:        []string{"unquote"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		unescape, err := strconv.Unquote(string(in))
		return []byte(unescape), err
	},
})

var hmacSha256Action = New(Definition[[]byte, []byte]{
	Doc:          "HMAC SHA256 of the data with the key parameter, to hex string",
	Names:        []string{"hmac"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "the HMAC key"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("hmac parameter is not a string")
		}
		mac := hmac.New(sha256.New, []byte(p))
		mac.Write(in)
		return []byte(hex.EncodeToString(mac.Sum(nil))), nil
	},
})

var crc32HashAction = New(Definition[[]byte, []byte]{
	Doc:          "CRC32 checksum of the data to hex string",
	Names:        []string{"crc32"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(fmt.Sprintf("%08x", crc32.ChecksumIEEE(in))), nil
	},
})

var md5HashAction = New(Definition[[]byte, []byte]{
	Doc:          "MD5 checksum of the data to hex string",
	Names:        []string{"md5"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		h := md5.New()
		io.WriteString(h, string(in))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
})

var sha1HashAction = New(Definition[[]byte, []byte]{
	Doc:          "SHA1 checksum of the data to hex string",
	Names:        []string{"sha1"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		h := sha1.New()
		io.WriteString(h, string(in))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
})

var sha256HashAction = New(Definition[[]byte, []byte]{
	Doc:          "SHA256 checksum of the data to hex string",
	Names:        []string{"sha256"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		h := sha256.New()
		io.WriteString(h, string(in))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
})

var sha512HashAction = New(Definition[[]byte, []byte]{
	Doc:          "SHA512 checksum of the data to hex string",
	Names:        []string{"sha512"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		h := sha512.New()
		io.WriteString(h, string(in))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
})

var fromBase64StringAction = New(Definition[[]byte, []byte]{
	Doc:          "Returns the bytes represented by the base64 of the input",
	Names:        []string{"base64"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return base64.StdEncoding.DecodeString(string(in))
	},
})

var parseDateStringAction = New(Definition[[]byte, time.Time]{
	Doc:          "Parse a date from input, supports ISO 8601/JSON dates, Go time strings and unix date output",
	Names:        []string{"date", "jsondate"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in []byte) (time.Time, error) {
		s := strings.TrimSpace(string(in))
		for _, layout := range []string{
			time.RFC3339Nano,
			time.DateTime + " -0700 MST", // Go time.String() format, e.g. 2026-08-28 21:59:57 -0400 EDT
			time.DateTime + " -0700",
			time.DateTime,
			time.DateOnly,
			"Mon 02 Jan 2006 03:04:05 PM MST", // unix date, e.g. Sun 13 Sep 2026 09:08:54 PM EDT
			"Mon Jan 02 15:04:05 MST 2006",    // unix date, e.g. Sun Sep 13 21:08:54 EDT 2026
		} {
			if t, err := time.Parse(layout, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("can't parse %q as a date", s)
	},
})

var jsonCompactAction = New(Definition[[]byte, []byte]{
	Doc:          "Minify/compact JSON from input",
	Names:        []string{"jsoncompact", "minify"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		dst := &bytes.Buffer{}
		if err := json.Compact(dst, in); err != nil {
			return nil, err
		}
		return dst.Bytes(), nil
	},
})

var jsonPrettifyAction = New(Definition[[]byte, []byte]{
	Doc:          "Reformat/prettify JSON from input",
	Names:        []string{"jsonpretty", "unminify", "reformat"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		dst := &bytes.Buffer{}
		if err := json.Indent(dst, in, "", "  "); err != nil {
			return nil, err
		}
		return dst.Bytes(), nil
	},
})

var toBase64StringAction = New(Definition[[]byte, []byte]{
	Doc:          "Returns the base64 encoding of input",
	Names:        []string{"tobase64"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(in)), nil
	},
})

var fromHexStringAction = New(Definition[[]byte, []byte]{
	Doc:          "Returns the bytes represented by the hexadecimal input, expects that input contains only hexadecimal",
	Names:        []string{"hex"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return hex.DecodeString(strings.ReplaceAll(string(in), " ", ""))
	},
})

var toHexStringAction = New(Definition[[]byte, []byte]{
	Doc:          "Returns the hexadecimal encoding of the input",
	Names:        []string{"tohex"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(hex.EncodeToString(in)), nil
	},
})

var isoTimeAction = New(Definition[time.Time, []byte]{
	Doc:          "time to ISO RFC3339 text",
	Names:        []string{"iso"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in time.Time) ([]byte, error) {
		return []byte(in.Format(time.RFC3339)), nil
	},
})

var humanizeTimeAction = New(Definition[time.Time, []byte]{
	Doc:          "time to relative humanized text, e.g. '3 days ago'",
	Names:        []string{"humanize", "ago"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in time.Time) ([]byte, error) {
		return []byte(humanize.Time(in)), nil
	},
})

// parseDuration parses durations like Go time.ParseDuration (1s, 2h30m, 100ms)
// and additionally days (2d) and weeks (3w). A leading - or per-segment
// negative values subtract time.
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty duration")
	}

	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}

	var total time.Duration
	i := 0
	for i < len(s) {
		start := i
		for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
			i++
		}
		if start == i {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		num, err := strconv.ParseFloat(s[start:i], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}

		useg := i
		for i < len(s) && !unicode.IsDigit(rune(s[i])) && s[i] != '.' {
			i++
		}
		unit := s[useg:i]
		if unit == "" {
			return 0, fmt.Errorf("missing unit in duration %q", s)
		}

		var d time.Duration
		switch unit {
		case "ns":
			d = time.Duration(num * float64(time.Nanosecond))
		case "us", "µs":
			d = time.Duration(num * float64(time.Microsecond))
		case "ms":
			d = time.Duration(num * float64(time.Millisecond))
		case "s":
			d = time.Duration(num * float64(time.Second))
		case "m":
			d = time.Duration(num * float64(time.Minute))
		case "h":
			d = time.Duration(num * float64(time.Hour))
		case "d":
			d = time.Duration(num * 24 * float64(time.Hour))
		case "w":
			d = time.Duration(num * 7 * 24 * float64(time.Hour))
		default:
			return 0, fmt.Errorf("unknown unit %q in duration %q", unit, s)
		}
		total += d
	}

	if neg {
		total = -total
	}
	return total, nil
}

// humanizeDuration renders a duration in human readable units like
// "2 days 3 hours", zero units are skipped
func humanizeDuration(d time.Duration) string {
	if d == 0 {
		return "0 seconds"
	}
	neg := d < 0
	if neg {
		d = -d
	}
	units := []struct {
		size time.Duration
		name string
	}{
		{7 * 24 * time.Hour, "week"},
		{24 * time.Hour, "day"},
		{time.Hour, "hour"},
		{time.Minute, "minute"},
		{time.Second, "second"},
		{time.Millisecond, "millisecond"},
		{time.Microsecond, "microsecond"},
		{time.Nanosecond, "nanosecond"},
	}
	var parts []string
	for _, u := range units {
		if d >= u.size {
			n := d / u.size
			d %= u.size
			if n == 1 {
				parts = append(parts, fmt.Sprintf("1 %s", u.name))
			} else {
				parts = append(parts, fmt.Sprintf("%d %ss", n, u.name))
			}
		}
	}
	s := strings.Join(parts, " ")
	if neg {
		s = "-" + s
	}
	return s
}

var addDurationTimeAction = New(Definition[time.Time, time.Time]{
	Doc:          "Add a duration to time, accepts Go durations like 1s, 2h30m and days like 2d, negative values subtract",
	Names:        []string{"adddur"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TimeFormat,
	Parameters:   []ActionParameter{{StringParameter, "duration to add, e.g. 1s, 2h30m, 2d, negative to subtract"}},
	Func: func(a Action, in time.Time) (time.Time, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return time.Time{}, fmt.Errorf("adddur parameter is not a string")
		}
		d, err := parseDuration(p)
		if err != nil {
			return time.Time{}, err
		}
		return in.Add(d), nil
	},
})

var toJSONDateStringAction = New(Definition[time.Time, []byte]{
	Doc:          "time to JSON ISO 8601 date string",
	Names:        []string{"tojsondate"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in time.Time) ([]byte, error) {
		return []byte(in.Format(time.RFC3339Nano)), nil
	},
})

var timeEpochAction = New(Definition[time.Time, []byte]{
	Doc:          "time to Epoch",
	Names:        []string{"epoch"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in time.Time) ([]byte, error) {
		return []byte(fmt.Sprintf("%d", in.Unix())), nil
	},
})

var epochTimeAction = New(Definition[[]byte, time.Time]{
	Doc:          "Parse Epoch time from input",
	Names:        []string{"epoch"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in []byte) (time.Time, error) {
		ts, err := strconv.Atoi(string(in))
		if err != nil {
			return time.Time{}, err
		}

		return time.Unix(int64(ts), 0), nil
	},
})

var commaTextListAction = New(Definition[[]byte, []string]{
	Doc:          "Parse a text input as a list separated by ,",
	Names:        []string{"comma"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []byte) ([]string, error) {
		l := strings.Split(string(in), ",")
		if len(l) <= 1 {
			return []string{}, errors.New("can't split using ,")
		}

		return l, nil
	},
})

var spaceTextListAction = New(Definition[[]byte, []string]{
	Doc:          "Parse a text input as a list separated by whitespace",
	Names:        []string{"space"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []byte) ([]string, error) {
		l := strings.Fields(string(in))
		if len(l) <= 1 {
			return []string{}, errors.New("can't split using space")
		}

		return l, nil
	},
})

var splitTextListAction = New(Definition[[]byte, []string]{
	Doc:          "Parse a text input as a list separated by a provided char",
	Names:        []string{"split"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextListFormat,
	Parameters:   []ActionParameter{{StringParameter, "a string to split"}},
	Func: func(a Action, in []byte) ([]string, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("split parameter is not a string")
		}
		l := strings.Split(string(in), p)
		if len(l) <= 1 {
			return []string{}, fmt.Errorf("can't split using %s", p)
		}

		return l, nil
	},
})

var pipeTextListAction = New(Definition[[]byte, []string]{
	Doc:          "Parse a text input as a list separated by |",
	Names:        []string{"pipe"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []byte) ([]string, error) {
		l := strings.Split(string(in), "|")
		if len(l) <= 1 {
			return []string{}, errors.New("can't split using |")
		}

		return l, nil
	},
})

var jwtTextListAction = New(Definition[[]byte, []string]{
	Doc:          "Parse a JWT and show the 3 JSON parts,",
	Names:        []string{"jwt"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []byte) ([]string, error) {
		l := strings.Split(string(in), ".")
		if len(l) != 3 {
			return []string{}, errors.New("not a valid JWT")
		}

		out := make([]string, 2)
		for i, t := range l[0:2] {
			j, err := base64.StdEncoding.DecodeString(t)
			if err != nil {
				// The shorter version (67 characters) is probably just missing a padding character (=) to be correct Base64.
				j, err = base64.RawStdEncoding.DecodeString(t)
				if err != nil {
					return nil, fmt.Errorf("can't decode base64 part of the JWT: %w", err)
				}
			}
			out[i] = string(j)
		}
		return out, nil
	},
})

var textListJoinCommaAction = New(Definition[[]string, []byte]{
	Doc:          "Join a list with a comma ,",
	Names:        []string{"comma"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return []byte(strings.Join(in, ",")), nil
	},
})

var textListJoinNewLineAction = New(Definition[[]string, []byte]{
	Doc:          "Join a list with new lines",
	Names:        []string{"line"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return []byte(strings.Join(in, "\n")), nil
	},
})

var textListToJSONAction = New(Definition[[]string, []byte]{
	Doc:          "Write a list as a JSON array",
	Names:        []string{"tojson"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return json.Marshal(in)
	},
})

var textListToYAMLAction = New(Definition[[]string, []byte]{
	Doc:          "Write a list as YAML",
	Names:        []string{"toyaml"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return yaml.Marshal(in)
	},
})

var textListCharJoinAction = New(Definition[[]string, []byte]{
	Doc:          "Join a list with a provided char",
	Names:        []string{"join"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "a string to join"}},
	Func: func(a Action, in []string) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("join parameter is not a string")
		}
		return []byte(strings.Join(in, p)), nil
	},
})

var textListSortAction = New(Definition[[]string, []string]{
	Doc:          "Sort a list alphabetically",
	Names:        []string{"sort"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []string) ([]string, error) {
		out := make([]string, len(in))
		copy(out, in)
		slices.Sort(out)
		return out, nil
	},
})

var textListReverseAction = New(Definition[[]string, []string]{
	Doc:          "Reverse the order of a list",
	Names:        []string{"reverse"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []string) ([]string, error) {
		out := make([]string, len(in))
		copy(out, in)
		slices.Reverse(out)
		return out, nil
	},
})

var textListUniqueAction = New(Definition[[]string, []string]{
	Doc:          "Deduplicate a list keeping the first occurrence of every element, preserving order",
	Names:        []string{"unique", "dedupe"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []string) ([]string, error) {
		seen := make(map[string]struct{}, len(in))
		out := make([]string, 0, len(in))
		for _, s := range in {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
		return out, nil
	},
})

var textListShuffleAction = New(Definition[[]string, []string]{
	Doc:          "Shuffle a list randomly",
	Names:        []string{"shuffle"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextListFormat,
	Func: func(a Action, in []string) ([]string, error) {
		out := make([]string, len(in))
		copy(out, in)
		rand.Shuffle(len(out), func(i, j int) {
			out[i], out[j] = out[j], out[i]
		})
		return out, nil
	},
})

var textListFilterAction = New(Definition[[]string, []string]{
	Doc:          "Filter a list keeping the lines matching a regexp parameter",
	Names:        []string{"filter"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextListFormat,
	Parameters:   []ActionParameter{{StringParameter, "a regular expression, matching lines are kept"}},
	Func: func(a Action, in []string) ([]string, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("filter parameter is not a string")
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		var out []string
		for _, s := range in {
			if re.MatchString(s) {
				out = append(out, s)
			}
		}
		return out, nil
	},
})

var textListCountAction = New(Definition[[]string, []byte]{
	Doc:          "Count the number of elements in a list",
	Names:        []string{"count"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return []byte(strconv.Itoa(len(in))), nil
	},
})

var textListFirstAction = New(Definition[[]string, []byte]{
	Doc:          "Select the first element of a list",
	Names:        []string{"first"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return []byte(in[0]), nil
	},
})

var textListLastAction = New(Definition[[]string, []byte]{
	Doc:          "Select the last element of a list",
	Names:        []string{"last"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []string) ([]byte, error) {
		return []byte(in[len(in)-1]), nil
	},
})

var textListIndexAction = New(Definition[[]string, []byte]{
	Doc:          "Select the element from a list at index parameter",
	Names:        []string{"index"},
	Type:         TransformAction,
	InputFormat:  TextListFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{IntParameter, "select the item at index"}},
	Func: func(a Action, in []string) ([]byte, error) {
		p, ok := a.InputParameters()[0].(int)
		if !ok {
			return nil, fmt.Errorf("index parameter is not an int")
		}
		if p < 0 || p > len(in)-1 {
			return nil, fmt.Errorf("index is out of list limits")
		}
		return []byte(in[p]), nil
	},
})

var unescapeTextAction = New(Definition[[]byte, []byte]{
	Doc:          "Unescape \\n and \\t from input",
	Names:        []string{"unescape"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(strings.ReplaceAll(strings.ReplaceAll(string(in), "\\n", "\n"), "\\t", "\t")), nil
	},
})

var stripWhitespaceAction = New(Definition[[]byte, []byte]{
	Doc:          "Removes all spaces, carriage returns, line feeds, tabs and form feeds from input",
	Names:        []string{"strip"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		r := strings.NewReplacer(" ", "", "\r", "", "\n", "", "\t", "", "\f", "")
		return []byte(r.Replace(string(in))), nil
	},
})

var textCountAction = New(Definition[[]byte, []byte]{
	Doc:          "Count the number of characters in the input",
	Names:        []string{"count"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(strconv.Itoa(utf8.RuneCountInString(string(in)))), nil
	},
})

var textCountLinesAction = New(Definition[[]byte, []byte]{
	Doc:          "Count the number of lines in the input",
	Names:        []string{"countlines"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		s := string(in)
		n := strings.Count(s, "\n")
		if !strings.HasSuffix(s, "\n") {
			n++
		}
		return []byte(strconv.Itoa(n)), nil
	},
})

var humanizeDurationTextAction = New(Definition[[]byte, []byte]{
	Doc:          "Humanize a duration from input, e.g. 2h30m, 2d, 500ms to '2 hours 30 minutes'",
	Names:        []string{"humanize", "humandur"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		d, err := parseDuration(string(in))
		if err != nil {
			return nil, err
		}
		return []byte(humanizeDuration(d)), nil
	},
})

var humanBytesAction = New(Definition[[]byte, []byte]{
	Doc:          "Humanize a byte count from input, e.g. 1048576 to '1.0 MB'",
	Names:        []string{"humanbytes"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		n, err := strconv.ParseUint(strings.TrimSpace(string(in)), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("can't parse %q as a byte count", string(in))
		}
		return []byte(humanize.Bytes(n)), nil
	},
})

var parseCSVAction = New(Definition[[]byte, [][]string]{
	Doc:          "Parse CSV text input into a table of rows",
	Names:        []string{"csv"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TableFormat,
	Func: func(a Action, in []byte) ([][]string, error) {
		r := csv.NewReader(bytes.NewReader(in))
		r.FieldsPerRecord = -1
		r.ReuseRecord = false
		t, err := r.ReadAll()
		if err != nil {
			return nil, err
		}
		if len(t) == 0 {
			return nil, errors.New("no CSV records found")
		}
		return t, nil
	},
})

var tableSortColumnAction = New(Definition[[][]string, [][]string]{
	Doc:          "Sort table rows by the column index parameter",
	Names:        []string{"sortcol"},
	Type:         TransformAction,
	InputFormat:  TableFormat,
	OutputFormat: TableFormat,
	Parameters:   []ActionParameter{{IntParameter, "the column index to sort by"}},
	Func: func(a Action, in [][]string) ([][]string, error) {
		p, ok := a.InputParameters()[0].(int)
		if !ok {
			return nil, fmt.Errorf("sortcol parameter is not an int")
		}
		if p < 0 {
			return nil, fmt.Errorf("column index is negative")
		}

		out := make([][]string, len(in))
		copy(out, in)

		for _, r := range out {
			if p > len(r)-1 {
				return nil, fmt.Errorf("column index %d is out of row limits", p)
			}
		}

		slices.SortStableFunc(out, func(x, y []string) int {
			xf, xerr := strconv.ParseFloat(x[p], 64)
			yf, yerr := strconv.ParseFloat(y[p], 64)
			if xerr == nil && yerr == nil {
				return cmp.Compare(xf, yf)
			}
			return strings.Compare(x[p], y[p])
		})

		return out, nil
	},
})

var tableToCSVAction = New(Definition[[][]string, []byte]{
	Doc:          "Write a table as CSV text",
	Names:        []string{"tocsv"},
	Type:         TransformAction,
	InputFormat:  TableFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in [][]string) ([]byte, error) {
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		if err := w.WriteAll(in); err != nil {
			return nil, err
		}
		w.Flush()
		return b.Bytes(), nil
	},
})

var pipeCommandAction = New(Definition[[]byte, []byte]{
	Doc:          "Pipe the input into a shell command, the command receives the input on stdin, output is the command stdout",
	Names:        []string{"exec", "sh"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "the command to run"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("exec parameter is not a string")
		}

		cmd := exec.Command("sh", "-c", p) //nolint:gosec
		cmd.Stdin = bytes.NewReader(in)

		var out bytes.Buffer
		var errBuf bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			msg := strings.TrimSpace(errBuf.String())
			if msg != "" {
				return nil, fmt.Errorf("%w: %s", err, msg)
			}
			return nil, err
		}

		return out.Bytes(), nil
	},
})

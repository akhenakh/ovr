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
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

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

var parseJSONDateStringAction = New(Definition[[]byte, time.Time]{
	Doc:          "Parse JSON ISO 8601 from input",
	Names:        []string{"jsondate"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in []byte) (time.Time, error) {
		return time.Parse(time.RFC3339, string(in))
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

// newTzTimeAction builds a timezone conversion action for the given IANA location
func newTzTimeAction(name, locName string) Action {
	return New(Definition[time.Time, time.Time]{
		Doc:          "Change time to " + name + " timezone (" + locName + ")",
		Names:        []string{name},
		Type:         TransformAction,
		InputFormat:  TimeFormat,
		OutputFormat: TimeFormat,
		Func: func(a Action, in time.Time) (time.Time, error) {
			loc, err := time.LoadLocation(locName)
			if err != nil {
				return time.Time{}, fmt.Errorf("unknown timezone %s: %w", locName, err)
			}
			return in.In(loc), nil
		},
	})
}

var estTimeAction = newTzTimeAction("est", "EST")
var etTimeAction = newTzTimeAction("et", "America/New_York")
var utcTimeAction = newTzTimeAction("utc", "UTC")

var timezoneActions = []Action{
	newTzTimeAction("pt", "America/Los_Angeles"),
	newTzTimeAction("pst", "America/Los_Angeles"),
	newTzTimeAction("mst", "America/Denver"),
	newTzTimeAction("cst", "America/Chicago"),
	newTzTimeAction("ct", "America/Chicago"),
	newTzTimeAction("brt", "America/Sao_Paulo"),
	newTzTimeAction("gmt", "Europe/London"),
	newTzTimeAction("cet", "Europe/Paris"),
	newTzTimeAction("eet", "Europe/Athens"),
	newTzTimeAction("msk", "Europe/Moscow"),
	newTzTimeAction("ist", "Asia/Kolkata"),
	newTzTimeAction("sgt", "Asia/Singapore"),
	newTzTimeAction("hkt", "Asia/Hong_Kong"),
	newTzTimeAction("jst", "Asia/Tokyo"),
	newTzTimeAction("kst", "Asia/Seoul"),
	newTzTimeAction("aest", "Australia/Sydney"),
	newTzTimeAction("nzst", "Pacific/Auckland"),
	newTzTimeAction("hst", "Pacific/Honolulu"),
}

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

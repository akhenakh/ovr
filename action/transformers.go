package action

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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
		return time.Parse("2006-01-02T15:04:05Z0700", string(in))
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

var estTimeAction = New(Definition[time.Time, time.Time]{
	Doc:          "Change time to EST timezone",
	Names:        []string{"est"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in time.Time) (time.Time, error) {
		est, _ := time.LoadLocation("EST")
		return in.In(est), nil
	},
})

var etTimeAction = New(Definition[time.Time, time.Time]{
	Doc:          "Change time to ET timezone",
	Names:        []string{"et"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in time.Time) (time.Time, error) {
		est, _ := time.LoadLocation("ET")
		return in.In(est), nil
	},
})

var utcTimeAction = New(Definition[time.Time, time.Time]{
	Doc:          "Change time to UTC timezone",
	Names:        []string{"utc"},
	Type:         TransformAction,
	InputFormat:  TimeFormat,
	OutputFormat: TimeFormat,
	Func: func(a Action, in time.Time) (time.Time, error) {
		est, _ := time.LoadLocation("UTC")
		return in.In(est), nil
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

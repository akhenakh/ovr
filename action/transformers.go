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

var upperAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their upper case",
	Names:        []string{"upper"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		caser := cases.Upper(language.Und)
		upper := caser.String(string(in.([]byte)))
		return []byte(upper), nil
	},
}

var lowerAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their lower case",
	Names:        []string{"lower"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		caser := cases.Lower(language.Und)
		lower := caser.String(string(in.([]byte)))
		return []byte(lower), nil
	},
}

var titleAction = Action{
	Doc:          "Transforms input title",
	Names:        []string{"title"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		caser := cases.Title(language.Und)
		titleStr := caser.String(string(in.([]byte)))
		return []byte(titleStr), nil
	},
}

var trimSpaceAction = Action{
	Doc:          "Trim spaces from input",
	Names:        []string{"trimspace"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(strings.TrimSpace(string(in.([]byte)))), nil
	},
}

var quoteAction = Action{
	Doc:          "Quotes string with escape characters",
	Names:        []string{"quote"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(strconv.Quote(string(in.([]byte)))), nil
	},
}

var unquoteAction = Action{
	Doc:          "Removes quotes from escaped characters",
	Names:        []string{"unquote"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		unescape, err := strconv.Unquote(string(in.([]byte)))
		return []byte(unescape), err
	},
}

var md5HashAction = Action{
	Doc:          "MD5 checksum of the data to hex string",
	Names:        []string{"md5"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		h := md5.New()
		io.WriteString(h, string(in.([]byte)))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
}

var sha1HashAction = Action{
	Doc:          "SHA1 checksum of the data to hex string",
	Names:        []string{"sha1"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		h := sha1.New()
		io.WriteString(h, string(in.([]byte)))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
}

var sha256HashAction = Action{
	Doc:          "SHA256 checksum of the data to hex string",
	Names:        []string{"sha256"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		h := sha256.New()
		io.WriteString(h, string(in.([]byte)))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
}

var sha512HashAction = Action{
	Doc:          "SHA512 checksum of the data to hex string",
	Names:        []string{"sha512"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		h := sha512.New()
		io.WriteString(h, string(in.([]byte)))
		return []byte(hex.EncodeToString(h.Sum(nil))), nil
	},
}

var fromBase64StringAction = Action{
	Doc:          "Returns the bytes represented by the base64 of the input",
	Names:        []string{"base64"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return base64.StdEncoding.DecodeString(string(in.([]byte)))
	},
}

var parseJSONDateStringAction = Action{
	Doc:          "Parse JSON ISO 8601 from input",
	Names:        []string{"jsondate"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: timeFormat,
	Func: func(in any) (any, error) {
		return time.Parse("2006-01-02T15:04:05Z0700", string(in.([]byte)))
	},
}

var jsonCompactAction = Action{
	Doc:          "Minify/compact JSON from input",
	Names:        []string{"jsoncompact", "minify"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		dst := &bytes.Buffer{}
		if err := json.Compact(dst, in.([]byte)); err != nil {
			return nil, err
		}
		return dst.Bytes(), nil
	},
}

var toBase64StringAction = Action{
	Doc:          "Returns the base64 encoding of input",
	Names:        []string{"tobase64"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(base64.StdEncoding.EncodeToString(in.([]byte))), nil
	},
}

var fromHexStringAction = Action{
	Doc:          "Returns the bytes represented by the hexadecimal input, expects that input contains only hexadecimal",
	Names:        []string{"hex"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return hex.DecodeString(strings.ReplaceAll(string(in.([]byte)), " ", ""))
	},
}

var toHexStringAction = Action{
	Doc:          "Returns the hexadecimal encoding of the input",
	Names:        []string{"tohex"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(hex.EncodeToString(in.([]byte))), nil
	},
}

var estTimeAction = Action{
	Doc:          "Change time to EST timezone",
	Names:        []string{"est"},
	Type:         TransformAction,
	InputFormat:  timeFormat,
	OutputFormat: timeFormat,
	Func: func(in any) (any, error) {
		est, _ := time.LoadLocation("EST")
		return in.(time.Time).In(est), nil
	},
}

var etTimeAction = Action{
	Doc:          "Change time to ET timezone",
	Names:        []string{"et"},
	Type:         TransformAction,
	InputFormat:  timeFormat,
	OutputFormat: timeFormat,
	Func: func(in any) (any, error) {
		est, _ := time.LoadLocation("ET")
		return in.(time.Time).In(est), nil
	},
}

var utcTimeAction = Action{
	Doc:          "Change time to UTC timezone",
	Names:        []string{"utc"},
	Type:         TransformAction,
	InputFormat:  timeFormat,
	OutputFormat: timeFormat,
	Func: func(in any) (any, error) {
		est, _ := time.LoadLocation("UTC")
		return in.(time.Time).In(est), nil
	},
}

var isoTimeAction = Action{
	Doc:          "time to ISO RFC3339 text",
	Names:        []string{"iso"},
	Type:         TransformAction,
	InputFormat:  timeFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(in.(time.Time).Format(time.RFC3339)), nil
	},
}

var timeEpochAction = Action{
	Doc:          "time to Epoch",
	Names:        []string{"epoch"},
	Type:         TransformAction,
	InputFormat:  timeFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(fmt.Sprintf("%d", in.(time.Time).Unix())), nil
	},
}

var epochTimeAction = Action{
	Doc:          "Parse Epoch time from input",
	Names:        []string{"epoch"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: timeFormat,
	Func: func(in any) (any, error) {
		ts, err := strconv.Atoi(string(in.([]byte)))
		if err != nil {
			return nil, err
		}

		return time.Unix(int64(ts), 0), nil
	},
}

var commaTextListAction = Action{
	Doc:          "Parse a text input as a list separated by ,",
	Names:        []string{"comma"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textListFormat,
	Func: func(in any) (any, error) {
		l := strings.Split(string(in.([]byte)), ",")
		if len(l) <= 1 {
			return []string{}, errors.New("can't split using ,")
		}

		return l, nil
	},
}

var jwtTextListAction = Action{
	Doc:          "Parse a JWT and show the 3 JSON parts,",
	Names:        []string{"jwt"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textListFormat,
	Func: func(in any) (any, error) {
		l := strings.Split(string(in.([]byte)), ".")
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
}

var textListJoinCommaAction = Action{
	Doc:          "Join a list separated by ,",
	Names:        []string{"comma"},
	Type:         TransformAction,
	InputFormat:  textListFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		l := in.([]string)
		return []byte(strings.Join(l, ",")), nil
	},
}

var textListFirstAction = Action{
	Doc:          "Select the first element of a list",
	Names:        []string{"first"},
	Type:         TransformAction,
	InputFormat:  textListFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		l := in.([]string)
		return []byte(l[0]), nil
	},
}

var textListLastAction = Action{
	Doc:          "Select the last element of a list",
	Names:        []string{"last"},
	Type:         TransformAction,
	InputFormat:  textListFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		l := in.([]string)
		return []byte(l[len(l)-1]), nil
	},
}

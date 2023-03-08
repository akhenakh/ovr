package action

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
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
		return hex.DecodeString(string(in.([]byte)))
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

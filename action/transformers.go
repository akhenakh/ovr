package action

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"
)

var upperAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their upper case",
	Names:        []string{"upper"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(strings.ToUpper(string(in.([]byte)))), nil
	},
}

var lowerAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their lower case",
	Names:        []string{"lower"},
	Type:         TransformAction,
	InputFormat:  textFormat,
	OutputFormat: textFormat,
	Func: func(in any) (any, error) {
		return []byte(strings.ToLower(string(in.([]byte)))), nil
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

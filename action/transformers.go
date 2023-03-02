package action

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
)

var upperAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their upper case",
	Names:        []string{"upper"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(strings.ToUpper(string(in))), nil
	},
}

var lowerAction = Action{
	Doc:          "Transforms input with all Unicode letters mapped to their lower case",
	Names:        []string{"lower"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(strings.ToLower(string(in))), nil
	},
}

var fromBase64StringAction = Action{
	Doc:          "Returns the bytes represented by the base64 of the input",
	Names:        []string{"base64"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return base64.StdEncoding.DecodeString(string(in))
	},
}

var toBase64StringAction = Action{
	Doc:          "Returns the base64 encoding of input",
	Names:        []string{"tobase64"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(base64.StdEncoding.EncodeToString(in)), nil
	},
}

var fromHexStringAction = Action{
	Doc:          "Returns the bytes represented by the hexadecimal input, expects that input contains only hexadecimal",
	Names:        []string{"hex"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return hex.DecodeString(string(in))
	},
}

var toHexStringAction = Action{
	Doc:          "Returns the hexadecimal encoding of the input",
	Names:        []string{"tohex"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(hex.EncodeToString(in)), nil
	},
}

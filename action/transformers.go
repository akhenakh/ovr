package action

import (
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
	Doc:          "encode binary to hex string",
	Names:        []string{"tohex"},
	Type:         TransformAction,
	InputFormat:  text,
	OutputFormat: text,
	Func: func(in []byte) ([]byte, error) {
		return []byte(hex.EncodeToString(in)), nil
	},
}

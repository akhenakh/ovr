package action

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
)

var xmlPrettyAction = New(Definition[[]byte, []byte]{
	Doc:          "Indent and reformat XML input with two spaces per level",
	Names:        []string{"xmlpretty"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return prettyXML(in)
	},
})

// prettyXML re-encodes the document with indentation, dropping the
// whitespace-only text nodes that made the input hard to read.
func prettyXML(in []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(in))
	dec.Strict = true

	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid XML: %w", err)
		}
		if cd, ok := tok.(xml.CharData); ok {
			s := string(cd)
			if strings.TrimSpace(s) == "" && strings.Contains(s, "\n") {
				continue
			}
		}
		if err := enc.EncodeToken(tok); err != nil {
			return nil, fmt.Errorf("invalid XML: %w", err)
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var markdownAction = New(Definition[[]byte, []byte]{
	Doc:          "Render Markdown input as HTML",
	Names:        []string{"markdown", "md"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		var buf bytes.Buffer
		if err := goldmark.Convert(in, &buf); err != nil {
			return nil, fmt.Errorf("can't render markdown: %w", err)
		}
		return bytes.TrimRight(buf.Bytes(), "\n"), nil
	},
})

package main

import (
	"fmt"

	"github.com/akhenakh/ovr/action"
)

// TextTransformer transforms a byte value into another
type TextTransformer interface {
	Transform([]byte) ([]byte, error)
}

type LineTransformer interface {
	Transform([]byte) ([]byte, error)
}

type DataTransformer interface {
	Transform(data action.Data, format string) (action.Data, error)
}

func main() {
	r := action.NewRegistry()

	out, err := r.TextAction("upper", []byte("my text"))
	fmt.Println(string(out), err)
	out, err = r.TextAction("tohex", out)
	fmt.Println(out, err)
	out, err = r.TextAction("hex", out)
	fmt.Println(string(out), err)
}

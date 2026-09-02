// Package clipboard wraps golang.design/x/clipboard, which panics in
// CGO-free builds instead of returning errors. Every function here
// recovers those panics so callers can degrade gracefully.
package clipboard

import (
	"errors"
	"fmt"

	goclipboard "golang.design/x/clipboard"
)

var available bool

// Format and FmtText re-export the golang.design/x/clipboard API.
type Format = goclipboard.Format

const FmtText = goclipboard.FmtText

// Init initializes the clipboard. It returns an error instead of
// panicking when built with CGO_ENABLED=0 or when no backend exists.
func Init() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
			available = false
		}
	}()
	err = goclipboard.Init()
	available = err == nil
	return err
}

// Available reports whether the clipboard was initialized successfully.
func Available() bool {
	return available
}

// Read returns the clipboard content, or nil when unavailable.
func Read(t goclipboard.Format) (buf []byte) {
	if !available {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			available = false
			buf = nil
		}
	}()
	return goclipboard.Read(t)
}

// Write stores buf in the clipboard, no-op when unavailable.
func Write(t goclipboard.Format, buf []byte) {
	if !available {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			available = false
		}
	}()
	goclipboard.Write(t, buf)
}

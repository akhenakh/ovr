//go:build js

package clipboard

import "errors"

// On js/wasm builds Go has no system clipboard access; the shirei backend
// handles copy/paste through the browser (RequestTextCopy and the paste
// request are delivered by the host). This stub keeps the desktop API
// shape so callers compile unchanged and degrade to the paste popup.

var errUnavailable = errors.New("clipboard unavailable on the web")

// Format mirrors the golang.design/x/clipboard API used by desktop builds.
type Format int

// FmtText mirrors golang.design/x/clipboard.FmtText.
const FmtText Format = 0

// Init always fails on the web.
func Init() error { return errUnavailable }

// Available always reports false on the web.
func Available() bool { return false }

// Read returns nil on the web.
func Read(Format) []byte { return nil }

// Write is a no-op on the web.
func Write(Format, []byte) {}

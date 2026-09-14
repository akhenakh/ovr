//go:build js

package action

// The browser has no external processes: no $EDITOR and no shell, so the
// edit and exec actions are not registered on wasm builds.

var editorActions = []Action{}

var execActions = []Action{}

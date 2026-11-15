//go:build js

package action

// The browser has no shell, so exec is not registered on wasm builds. The
// edit action is: ovrui provides the in-app document editor in place of
// $EDITOR, so the action is usable in the browser too.

var editorActions = []Action{editAction}

var execActions = []Action{}

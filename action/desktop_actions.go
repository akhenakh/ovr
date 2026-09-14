//go:build !js

package action

// Actions that spawn external processes ($EDITOR, shell commands) only
// exist on desktop builds; they are excluded from the wasm/web registry.

// editorActions edit the data with an external editor, they convert any
// data to its string representation first
var editorActions = []Action{editAction}

var execActions = []Action{pipeCommandAction}

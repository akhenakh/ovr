//go:build js

package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSRegistry_HasInAppEditorNoShell(t *testing.T) {
	r := NewRegistry()

	// the in-app editor makes the edit action usable in the browser
	a, ok := r.ActionByName(TextFormat, "edit")
	require.True(t, ok, "edit should be registered on wasm")
	_, isEditor := a.(Editor)
	require.True(t, isEditor, "edit should implement the Editor hook")

	// there is no shell in the browser
	_, ok = r.ActionByName(TextFormat, "exec")
	require.False(t, ok, "exec must not be registered on wasm")
}

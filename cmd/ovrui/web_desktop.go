//go:build !js

package main

// onWeb is true when ovrui is compiled for the browser
// (GOOS=js GOARCH=wasm, eg. `task web`). Filesystem paths and piped stdin
// do not exist there; the paste popup is the input source.
const onWeb = false

package main

import (
	_ "embed"
	"fmt"
	"os"

	. "go.hasen.dev/shirei"
)

//go:embed iosevskanerdfont.ttf
var iosevkaBytes []byte

const fontFamily = "Iosevka Nerd Font Mono"

func init() {
	// register synchronously so text renders deterministically in headless
	// renders (--png) without waiting for the system font scan
	if err := UseFontBytes(iosevkaBytes); err != nil {
		fmt.Fprintln(os.Stderr, "font registration failed:", err)
	}
}

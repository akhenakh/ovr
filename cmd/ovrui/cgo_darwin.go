//go:build darwin

package main

// Force cmd/ovrui to be a cgo package on darwin so the C shim in
// availability_shim.c is compiled and linked into the binary.

import "C"

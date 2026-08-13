//go:build !js || !wasm

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "zkp-prover-wasm is a test-only js/wasm compatibility command")
	os.Exit(2)
}

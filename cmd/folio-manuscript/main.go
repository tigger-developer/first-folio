// ABOUTME: Command entry point for prose manuscript rendering.
// ABOUTME: Delegates argument parsing and rendering to internal/manuscript.
package main

import (
	"fmt"
	"os"

	"github.com/tadg-paul/first-folio/internal/manuscript"
)

func main() {
	if err := manuscript.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

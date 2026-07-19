// ABOUTME: Process entry point for the single First Folio Go executable.
// ABOUTME: Delegates all command behaviour to the testable internal application package.
package main

import (
	"os"

	"github.com/tigger-developer/first-folio/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

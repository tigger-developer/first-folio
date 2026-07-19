// ABOUTME: Dispatches the public First Folio command without terminating the process.
// ABOUTME: Keeps CLI output and status testable through injected streams.
package app

import (
	"fmt"
	"io"
	"strings"

	folio "github.com/tigger-developer/first-folio"
	"github.com/tigger-developer/first-folio/internal/manuscript"
)

var Version = "0.4.10"

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	_ = stdin
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Error: no subcommand given.")
		fmt.Fprintln(stderr, "Usage: folio <command> [options]")
		fmt.Fprintln(stderr, "Run 'folio --help' for available commands.")
		return 1
	}
	if hasVersion(args) {
		fmt.Fprintf(stdout, "folio %s\n", Version)
		return 0
	}

	switch args[0] {
	case "-h", "--help":
		return writeHelp(stdout, "docs/folio-help.md")
	case "convert", "letter", "manuscript":
		if hasHelp(args[1:]) {
			name := args[0]
			return writeHelp(stdout, "docs/folio-"+name+"-help.md")
		}
		if args[0] == "convert" {
			if err := runConvert(args[1:], stdout, stderr); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return 1
			}
			return 0
		}
		if args[0] == "letter" {
			if err := runLetter(args[1:], stdout); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return 1
			}
			return 0
		}
		if err := manuscript.RunWithIO(args[1:], stdout); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown subcommand '%s'.\n", args[0])
		fmt.Fprintln(stderr, "Run 'folio --help' for available commands.")
		return 1
	}
}

func hasVersion(args []string) bool {
	for _, arg := range args {
		if arg == "--version" {
			return true
		}
	}
	return false
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func writeHelp(dst io.Writer, name string) int {
	content, err := folio.Assets.ReadFile(name)
	if err != nil {
		fmt.Fprintf(dst, "Error: cannot load help: %v\n", err)
		return 1
	}
	if _, err := io.Copy(dst, strings.NewReader(string(content))); err != nil {
		return 1
	}
	return 0
}

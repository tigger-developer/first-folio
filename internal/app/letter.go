// ABOUTME: Implements the public cover-letter subcommand in Go.
// ABOUTME: Handles filtering, stable filenames, shared configuration, and PDF output.
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/letter"
)

func runLetter(args []string, stdout io.Writer) error {
	to, outputDir, prefix, source, err := parseLetterOptions(args)
	if err != nil {
		return err
	}
	raw, err := readTextInput(source)
	if err != nil {
		return err
	}
	letters, err := letter.ParseOrg(string(raw))
	if err != nil {
		return err
	}
	if err := letter.Validate(letters, source); err != nil {
		return err
	}
	if outputDir == "" {
		outputDir = filepath.Dir(source)
	}
	if prefix == "" {
		prefix = "letter"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: home, LocalDir: filepath.Dir(source)})
	if err != nil {
		return err
	}
	var filter *regexp.Regexp
	if to != "" {
		filter, err = regexp.Compile("(?i)" + to)
		if err != nil {
			return fmt.Errorf("invalid recipient filter %q: %w", to, err)
		}
	}
	count := 0
	for _, item := range letters {
		if filter != nil && !filter.MatchString(item.Organization) && !filter.MatchString(item.Recipient) {
			continue
		}
		slug := letter.Slug(item.Organization)
		if slug == "" {
			slug = letter.Slug(item.Recipient)
		}
		if slug == "" {
			slug = fmt.Sprintf("letter-%d", count)
		}
		output := filepath.Join(outputDir, prefix+"-"+slug+".pdf")
		typst, err := letter.RenderTypst(item, cfg)
		if err != nil {
			return err
		}
		if err := letter.CompilePDF(typst, output); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Generated: %s\n", output)
		count++
	}
	if count == 0 && to != "" {
		return fmt.Errorf("no recipients matching '%s' found", to)
	}
	if count > 1 {
		fmt.Fprintf(stdout, "%d cover letter(s) generated.\n", count)
	}
	return nil
}

func parseLetterOptions(args []string) (string, string, string, string, error) {
	values := map[string]string{}
	var positional []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			key, value, consumed, err := optionValue(args, i)
			if err != nil {
				return "", "", "", "", err
			}
			i += consumed
			if key != "to" && key != "dir" && key != "prefix" {
				return "", "", "", "", fmt.Errorf("unknown letter option --%s", key)
			}
			values[key] = value
		} else {
			positional = append(positional, args[i])
		}
	}
	if len(positional) == 0 {
		return "", "", "", "", fmt.Errorf("no source file specified\nUsage: folio letter <source.org> [--to RECIPIENT]")
	}
	if len(positional) > 1 {
		return "", "", "", "", fmt.Errorf("too many letter source files")
	}
	return values["to"], values["dir"], values["prefix"], positional[0], nil
}

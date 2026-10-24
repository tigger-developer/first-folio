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
	"github.com/tigger-developer/first-folio/internal/office"
)

func runLetter(args []string, stdout, stderr io.Writer) error {
	args, format, err := letterOutputFormat(args)
	if err != nil {
		return err
	}
	to, outputDir, prefix, source, err := parseLetterOptions(args)
	if err != nil {
		return err
	}
	var raw []byte
	if format == "pdf" {
		raw, err = readTextInput(source)
	} else {
		var sources [][]byte
		sources, err = office.ReadSources([]string{source})
		if err == nil {
			raw = sources[0]
		}
	}
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
	for _, warning := range cfg.Warnings() {
		if _, err := fmt.Fprintln(stderr, "Warning:", warning); err != nil {
			return err
		}
	}
	var filter *regexp.Regexp
	if to != "" {
		filter, err = regexp.Compile("(?i)" + to)
		if err != nil {
			return fmt.Errorf("invalid recipient filter %q: %w", to, err)
		}
	}
	if format != "pdf" {
		return exportOfficeLetters(letters, cfg, filter, source, outputDir, prefix, format, stdout, stderr)
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

func letterOutputFormat(args []string) ([]string, string, error) {
	format := "pdf"
	var remaining []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--format" || strings.HasPrefix(args[i], "--format=") {
			_, value, consumed, err := optionValue(args, i)
			if err != nil {
				return nil, "", fmt.Errorf("letter --format requires pdf, docx or odt: %w", err)
			}
			i += consumed
			format = strings.ToLower(value)
			if format != "pdf" && format != "docx" && format != "odt" {
				return nil, "", fmt.Errorf("invalid letter format %q; expected pdf, docx or odt", value)
			}
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return remaining, format, nil
}

func exportOfficeLetters(letters []letter.Letter, cfg config.Config, filter *regexp.Regexp, source, dir, prefix, format string, stdout, stderr io.Writer) error {
	if strings.ContainsAny(prefix, `/\`) || prefix == "." || prefix == ".." {
		return fmt.Errorf("office letter prefix must not contain path separators or traversal")
	}
	type target struct {
		item letter.Letter
		path string
	}
	var selected []target
	seen := map[string]bool{}
	for _, item := range letters {
		if filter != nil && !filter.MatchString(item.Organization) && !filter.MatchString(item.Recipient) {
			continue
		}
		slug := letter.Slug(item.Organization)
		if slug == "" {
			slug = letter.Slug(item.Recipient)
		}
		if slug == "" {
			slug = fmt.Sprintf("letter-%d", len(selected))
		}
		path := filepath.Join(dir, prefix+"-"+slug+"."+format)
		key := strings.ToLower(filepath.Clean(path))
		if seen[key] {
			return fmt.Errorf("duplicate office letter target %s", path)
		}
		seen[key] = true
		if err := office.CheckTarget([]string{source}, path); err != nil {
			return err
		}
		selected = append(selected, target{item, path})
	}
	if len(selected) == 0 && filter != nil {
		return fmt.Errorf("no recipients matching '%s' found", filter.String())
	}
	for _, target := range selected {
		plan, err := letter.Office(target.item, cfg)
		if err != nil {
			return fmt.Errorf("converting %s: %w", target.path, err)
		}
		raw, warnings, err := plan.Export(format)
		if err != nil {
			return fmt.Errorf("converting %s: %w", target.path, err)
		}
		for _, warning := range warnings {
			if _, err := fmt.Fprintln(stderr, "Warning:", warning); err != nil {
				return err
			}
		}
		if err := office.Publish(target.path, raw); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "Generated: %s\n", target.path); err != nil {
			return err
		}
	}
	if len(selected) > 1 {
		_, err := fmt.Fprintf(stdout, "%d cover letter(s) generated.\n", len(selected))
		return err
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

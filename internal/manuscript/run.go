// ABOUTME: Orchestrates manuscript CLI execution from args to output artefact.
// ABOUTME: Wires input resolution, config loading, parsing, Typst rendering, and compilation.
package manuscript

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	folio "github.com/tigger-developer/first-folio"
)

const Version = "0.4.7"

func Run(args []string) error {
	return RunWithIO(args, os.Stdout)
}

func RunWithIO(args []string, stdout io.Writer) error {
	opts, inputs, err := parseArgs(args)
	if err != nil {
		return err
	}
	if opts.ShowHelp {
		fmt.Fprint(stdout, Usage())
		return nil
	}
	if opts.ShowVersion {
		fmt.Fprintf(stdout, "folio-manuscript %s\n", Version)
		return nil
	}
	if opts.Output == "" {
		return fmt.Errorf("no output target specified")
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no input files specified")
	}

	inputSet, err := ResolveInputs(inputs)
	if err != nil {
		return err
	}

	sourceDir := filepath.Dir(inputSet.Paths[0])
	cfg, err := LoadConfig(sourceDir, opts)
	if err != nil {
		return err
	}

	text, err := ReadJoined(inputSet.Paths)
	if err != nil {
		return err
	}

	doc, err := Parse(inputSet.Format, text)
	if err != nil {
		return err
	}
	applyMetadataOverrides(&doc.Metadata, opts, cfg)
	// #16 AC16.4: when chapter.number-reset is "never", renumber chapters
	// continuously across all parts. Parser defaults to per-part reset.
	if cfg.Folio.Manuscript.Chapter.NumberReset == "never" {
		continuous := 0
		for i := range doc.Blocks {
			if doc.Blocks[i].Kind == "chapter" {
				continuous++
				doc.Blocks[i].Number = continuous
			}
		}
	}

	if opts.DryRun {
		printDryRun(stdout, inputSet, opts, cfg)
		return nil
	}

	typst, err := RenderTypst(doc, cfg)
	if err != nil {
		return err
	}

	if strings.HasSuffix(strings.ToLower(opts.Output), ".typ") {
		return os.WriteFile(opts.Output, []byte(typst), 0o644)
	}
	if !strings.HasSuffix(strings.ToLower(opts.Output), ".pdf") {
		return fmt.Errorf("manuscript output must end in .pdf or .typ")
	}
	return compileTypst(typst, opts.Output)
}

func parseArgs(args []string) (Options, []string, error) {
	var opts Options
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			opts.ShowHelp = true
			return opts, nil, nil
		}
		if arg == "--dry-run" {
			opts.DryRun = true
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}
		key := strings.TrimPrefix(arg, "--")
		if key == "version" {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				opts.ShowVersion = true
				return opts, nil, nil
			}
		}
		i++
		if i >= len(args) {
			return opts, nil, fmt.Errorf("missing value for --%s", key)
		}
		setOption(&opts, key, args[i])
	}

	if len(positional) < 2 {
		return opts, nil, fmt.Errorf("usage: folio manuscript <input>... <target>")
	}
	opts.Output = positional[len(positional)-1]
	return opts, positional[:len(positional)-1], nil
}

func setOption(opts *Options, key string, value string) {
	switch key {
	case "style":
		opts.Style = value
	case "title":
		opts.Title = value
	case "subtitle":
		opts.Subtitle = value
	case "author":
		opts.Author = value
	case "attribution", "author-attribution":
		opts.AuthorAttribution = value
	case "date":
		opts.Date = value
	case "version":
		opts.Version = value
	case "wordcount":
		opts.WordCount = value
	case "contact-name":
		opts.ContactName = value
	}
}

func usageError() error {
	return fmt.Errorf("usage: folio manuscript <input>... <target> [--style british|us]")
}

func Usage() string {
	raw, err := folio.Assets.ReadFile("docs/folio-manuscript-help.md")
	if err != nil {
		return "Usage: folio manuscript <input>... <target> [options]\n"
	}
	return string(raw)
}

func printDryRun(stdout io.Writer, inputSet InputSet, opts Options, cfg Config) {
	fmt.Fprintf(stdout, "format: %s\n", inputSet.Format)
	fmt.Fprintf(stdout, "output: %s\n", opts.Output)
	fmt.Fprintf(stdout, "style: %s\n", cfg.Folio.Manuscript.Style)
	fmt.Fprintf(stdout, "page: %s\n", cfg.Folio.Manuscript.Page)
	fmt.Fprintf(stdout, "margin: %s\n", cfg.Folio.Manuscript.Margin)
	fmt.Fprintln(stdout, "inputs:")
	for _, path := range inputSet.Paths {
		fmt.Fprintf(stdout, "  - %s\n", path)
	}
}

func applyMetadataOverrides(meta *Metadata, opts Options, cfg Config) {
	overrideString(&meta.Title, opts.Title, cfg.Title)
	overrideString(&meta.Subtitle, opts.Subtitle, cfg.Subtitle)
	overrideString(&meta.Author, opts.Author, cfg.Author)
	overrideString(&meta.AuthorAttribution, opts.AuthorAttribution, cfg.Attribution, cfg.AuthorAttribution, cfg.Folio.Manuscript.Attribution, cfg.Folio.Manuscript.AuthorAttribution)
	overrideString(&meta.Date, opts.Date, cfg.Date)
	// #21: default folio.date to today when otherwise unset, so copyright-year
	// derivation and any other year-of-publication tokens always resolve.
	if meta.Date == "" {
		meta.Date = time.Now().Format("2006-01-02")
	}
	overrideString(&meta.Version, opts.Version, cfg.Version)
	overrideString(&meta.WordCount, opts.WordCount, cfg.WordCount)
	overrideString(&meta.ContactName, opts.ContactName, cfg.ContactName)
	overrideString(&meta.Address, cfg.Address)
	overrideString(&meta.Phone, cfg.Phone)
	overrideString(&meta.Email, cfg.Email)
	overrideString(&meta.Website, cfg.Website)
}

func overrideString(target *string, values ...string) {
	for _, value := range values {
		if value != "" {
			*target = value
		}
	}
}

func compileTypst(source string, output string) error {
	tmp, err := os.CreateTemp("", "folio-manuscript-*.typ")
	if err != nil {
		return fmt.Errorf("creating temporary Typst source: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString(source); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temporary Typst source: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temporary Typst source: %w", err)
	}

	cmd := exec.Command("typst", "compile", tmpPath, output)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("typst compile failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

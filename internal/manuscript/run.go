// ABOUTME: Orchestrates manuscript CLI execution from args to output artefact.
// ABOUTME: Wires input resolution, config loading, parsing, Typst rendering, and compilation.
package manuscript

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Version = "0.4.7"

func Run(args []string) error {
	opts, inputs, err := parseArgs(args)
	if err != nil {
		return err
	}
	if opts.ShowHelp {
		fmt.Print(Usage())
		return nil
	}
	if opts.ShowVersion {
		fmt.Printf("folio-manuscript %s\n", Version)
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

	if opts.DryRun {
		printDryRun(inputSet, opts, cfg)
		return nil
	}

	canonicalMarkdown := RenderMarkdown(doc)
	canonicalDoc, err := Parse("markdown", canonicalMarkdown)
	if err != nil {
		return err
	}
	typst, err := RenderTypst(canonicalDoc, cfg)
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
	case "author-attribution":
		opts.AuthorAttribution = value
	case "date":
		opts.Date = value
	case "version":
		opts.Version = value
	case "wordcount":
		opts.WordCount = value
	}
}

func usageError() error {
	return fmt.Errorf("usage: folio manuscript <input>... <target> [--style british|us]")
}

func Usage() string {
	lines := []string{
		"Usage: folio manuscript <input>... <target> [options]",
		"",
		"Render Markdown or org-mode prose manuscript chapters to .typ or .pdf.",
		"",
		"Options:",
		"  --style british|us          Manuscript preset, default british",
		"  --title TITLE               Override manuscript title",
		"  --subtitle SUBTITLE         Override manuscript subtitle",
		"  --author AUTHOR             Override author name",
		"  --author-attribution TEXT   Override author attribution, default by",
		"  --date DATE                 Override manuscript date",
		"  --version [VERSION]         Show command version, or override manuscript version when VALUE is supplied",
		"  --wordcount WORDS           Override manuscript word count",
		"  --dry-run                   Validate inputs and print the render plan",
		"  -h, --help                  Show this help message",
		"",
	}
	return strings.Join(lines, "\n")
}

func printDryRun(inputSet InputSet, opts Options, cfg Config) {
	fmt.Printf("format: %s\n", inputSet.Format)
	fmt.Printf("output: %s\n", opts.Output)
	fmt.Printf("style: %s\n", cfg.Folio.Manuscript.Style)
	fmt.Printf("page: %s\n", cfg.Folio.Manuscript.Page)
	fmt.Printf("margin: %s\n", cfg.Folio.Manuscript.Margin)
	fmt.Println("inputs:")
	for _, path := range inputSet.Paths {
		fmt.Printf("  - %s\n", path)
	}
}

func applyMetadataOverrides(meta *Metadata, opts Options, cfg Config) {
	overrideString(&meta.Title, opts.Title, cfg.Title)
	overrideString(&meta.Subtitle, opts.Subtitle, cfg.Subtitle)
	overrideString(&meta.Author, opts.Author, cfg.Author)
	overrideString(&meta.Date, opts.Date, cfg.Date)
	overrideString(&meta.Version, opts.Version, cfg.Version)
	overrideString(&meta.WordCount, opts.WordCount, cfg.WordCount)
	overrideString(&meta.Address, cfg.Address)
	overrideString(&meta.Phone, cfg.Phone)
	overrideString(&meta.Email, cfg.Email)
	overrideString(&meta.Website, cfg.Website)
	if opts.AuthorAttribution != "" {
		meta.AuthorAttribution = opts.AuthorAttribution
	}
	if meta.AuthorAttribution == "" {
		meta.AuthorAttribution = cfg.Folio.Manuscript.AuthorAttribution
	}
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

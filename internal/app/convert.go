// ABOUTME: Implements the public stage-play conversion subcommand in Go.
// ABOUTME: Owns option parsing, validation, configuration, filtering, and text output.
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

type convertOptions struct {
	to     string
	force  bool
	cli    map[string]any
	source string
	target string
}

func runConvert(args []string, stdout io.Writer, stderr io.Writer) error {
	opts, err := parseConvertOptions(args)
	if err != nil {
		return err
	}
	raw, err := readTextInput(opts.source)
	if err != nil {
		return err
	}
	sourceFormat, err := play.FormatFromPath(opts.source)
	if err != nil {
		return err
	}
	if !sourceFormat.Readable() {
		return fmt.Errorf("cannot read from %s format. PDF is write-only", sourceFormat)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}
	cfg, err := config.Load(config.Options{
		Mode:     config.ModeScript,
		Home:     home,
		LocalDir: filepath.Dir(opts.source),
		CLI:      opts.cli,
	})
	if err != nil {
		return err
	}

	targetFormat, toStdout, err := convertTarget(opts, cfg)
	if err != nil {
		return err
	}
	doc, parseWarnings, err := play.Parse(sourceFormat, string(raw), opts.source)
	if err != nil {
		return err
	}
	for _, warning := range parseWarnings {
		fmt.Fprintln(stderr, warning)
	}
	applyScriptConfig(&doc, cfg)
	if targetFormat == play.FormatPDF {
		return renderPlayDocument(doc, cfg, opts.target, toStdout, opts.force, stdout)
	}
	result, err := play.Emit(doc, targetFormat)
	if err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(stderr, warning)
	}
	if toStdout {
		_, err = io.WriteString(stdout, result.Text)
		return err
	}
	if err := os.WriteFile(opts.target, []byte(result.Text), 0o644); err != nil {
		return fmt.Errorf("cannot write to %s: %w", opts.target, err)
	}
	return nil
}

func parseConvertOptions(args []string) (convertOptions, error) {
	opts := convertOptions{cli: map[string]any{}}
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--force":
			opts.force = true
		case "--direction-italic":
			opts.cli["direction-italic"] = true
		case "--no-direction-italic":
			opts.cli["direction-italic"] = false
		case "--direction-center", "--direction-centre":
			opts.cli["direction-center"] = true
		case "--no-direction-center", "--no-direction-centre":
			opts.cli["direction-center"] = false
		default:
			if strings.HasPrefix(arg, "--") {
				key, value, consumed, err := optionValue(args, i)
				if err != nil {
					return opts, err
				}
				i += consumed
				if key == "to" {
					opts.to = value
				} else {
					opts.cli[key] = value
				}
			} else {
				positional = append(positional, arg)
			}
		}
	}
	if len(positional) == 0 {
		return opts, fmt.Errorf("no source file specified\nUsage: folio convert <source> [target] [--to FORMAT]")
	}
	if len(positional) > 2 {
		return opts, fmt.Errorf("too many positional arguments")
	}
	opts.source = positional[0]
	if len(positional) == 2 {
		opts.target = positional[1]
	}
	return opts, nil
}

func optionValue(args []string, index int) (string, string, int, error) {
	arg := strings.TrimPrefix(args[index], "--")
	if key, value, ok := strings.Cut(arg, "="); ok {
		return key, value, 0, nil
	}
	if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
		return "", "", 0, fmt.Errorf("missing value for --%s", arg)
	}
	return arg, args[index+1], 1, nil
}

func readTextInput(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	probe := raw
	if len(probe) > 8192 {
		probe = probe[:8192]
	}
	if strings.IndexByte(string(probe), 0) >= 0 {
		return nil, fmt.Errorf("%s appears to be a binary file, not a text document", path)
	}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("%s has invalid encoding: expected UTF-8", path)
	}
	return raw, nil
}

func convertTarget(opts convertOptions, cfg config.Config) (play.Format, bool, error) {
	if opts.target != "" {
		format, err := play.FormatFromPath(opts.target)
		return format, false, err
	}
	name := opts.to
	if name == "" {
		name = cfg.String("folio.default-format", "")
		if name == "" {
			return "", false, fmt.Errorf("no target file or --to format specified, and no default-format configured")
		}
	}
	format, err := play.ParseFormat(name)
	return format, true, err
}

func applyScriptConfig(doc *play.Document, cfg config.Config) {
	for _, key := range []string{"title", "subtitle", "author", "date", "version"} {
		if value := cfg.String(key, ""); value != "" {
			doc.Metadata[key] = value
		}
	}
	filtered := doc.Events[:0]
	for _, event := range doc.Events {
		keep := true
		switch event.Kind {
		case play.EventStageDirection:
			keep = cfg.Bool("render.stage-directions", true)
		case play.EventCharacterTableStart, play.EventCharacterTableRow, play.EventCharacterTableEnd:
			keep = cfg.Bool("render.character-table", true)
		case play.EventFootnote:
			keep = cfg.Bool("render.footnotes", true)
		case play.EventTransition:
			keep = cfg.Bool("render.transitions", true)
		case play.EventIntroHeader, play.EventIntroText:
			keep = cfg.Bool("render.frontmatter", true)
		}
		if keep {
			filtered = append(filtered, event)
		}
	}
	doc.Events = filtered
}

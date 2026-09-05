// ABOUTME: Regression tests for configurable manuscript blockquote typography, spacing, and indentation.
// ABOUTME: Covers whole-block indentation and compatibility with precise code-block spacing overrides.
package manuscript

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManuscriptBlockSpacingAndQuoteFontCanBeConfigured(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    quoted-block-spacing: 1.25em",
		"    code-block-spacing: 1.75em",
		"    quote-block-indent: 2em",
		"    code-block-indent: 3em",
		"    quoted-block:",
		"      font:",
		`        family: Libertinus "Serif"`,
		"        size: 11pt",
		"        weight: semibold",
		"        stretch: 125%",
		"        style: italic",
		"        letter-spacing: 0.03em",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "chapter.md"), strings.Join([]string{
		"## Chapter 1",
		"",
		"Before the quotation.",
		"",
		"> This is a quote.",
		"",
		"After the quotation.",
		"",
		"```",
		"Code block",
		"```",
		"",
	}, "\n"))

	output := filepath.Join(dir, "manuscript.typ")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `#quote(block: true)[`)
	assertContains(t, typst, strings.Join([]string{
		"#show quote.where(block: true): it => block(",
		"  above: 1.25em,",
		"  below: 1.25em,",
		")[",
		"  #pad(left: 2em)[#text(",
		`    font: "Libertinus \"Serif\"",`,
		"    size: 11pt,",
		`    weight: "semibold",`,
		"    stretch: 125%,",
		`    style: "italic",`,
		"    tracking: 0.03em,",
	}, "\n"))
	assertContains(t, typst, strings.Join([]string{
		"#show raw.where(block: true): it => block(",
		"  above: 1.75em,",
		"  below: 1.75em,",
		")[#pad(left: 3em)[#text(",
	}, "\n"))
}

func TestInvalidQuotedBlockConfigurationFailsBeforeOutput(t *testing.T) {
	tests := []struct {
		name       string
		configLine string
		wantPath   string
	}{
		{"spacing", "    quoted-block-spacing: close", "folio.manuscript.quoted-block-spacing"},
		{"code spacing", "    code-block-spacing: close", "folio.manuscript.code-block-spacing"},
		{"empty spacing", `    quoted-block-spacing: ""`, "folio.manuscript.quoted-block-spacing"},
		{"empty code spacing", `    code-block-spacing: ""`, "folio.manuscript.code-block-spacing"},
		{"mapped spacing", "    quoted-block-spacing: {bad: value}", "folio.manuscript.quoted-block-spacing"},
		{"quote indent", "    quote-block-indent: inward", "folio.manuscript.quote-block-indent"},
		{"code indent", "    code-block-indent: inward", "folio.manuscript.code-block-indent"},
		{"empty quote indent", `    quote-block-indent: ""`, "folio.manuscript.quote-block-indent"},
		{"empty code indent", `    code-block-indent: ""`, "folio.manuscript.code-block-indent"},
		{"size", "        size: enormous", "folio.manuscript.quoted-block.font.size"},
		{"weight", "        weight: supreme", "folio.manuscript.quoted-block.font.weight"},
		{"stretch", "        stretch: extended", "folio.manuscript.quoted-block.font.stretch"},
		{"NaN stretch", "        stretch: NaN", "folio.manuscript.quoted-block.font.stretch"},
		{"infinite stretch", "        stretch: Inf", "folio.manuscript.quoted-block.font.stretch"},
		{"style", "        style: bold", "folio.manuscript.quoted-block.font.style"},
		{"letter spacing", "        letter-spacing: loose", "folio.manuscript.quoted-block.font.letter-spacing"},
		{"empty family", `        family: ""`, "folio.manuscript.quoted-block.font.family"},
		{"empty size", `        size: ""`, "folio.manuscript.quoted-block.font.size"},
		{"empty weight", `        weight: ""`, "folio.manuscript.quoted-block.font.weight"},
		{"empty stretch", `        stretch: ""`, "folio.manuscript.quoted-block.font.stretch"},
		{"empty style", `        style: ""`, "folio.manuscript.quoted-block.font.style"},
		{"empty letter spacing", `        letter-spacing: ""`, "folio.manuscript.quoted-block.font.letter-spacing"},
		{"null style", "        style: null", "folio.manuscript.quoted-block.font.style"},
		{"unknown font property", "        decoration: underline", "folio.manuscript.quoted-block.font.decoration"},
		{"mapped font size", "        size: {bad: value}", "folio.manuscript.quoted-block.font.size"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			lines := []string{"folio:", "  manuscript:"}
			if strings.HasPrefix(test.configLine, "        ") {
				lines = append(lines, "    quoted-block:", "      font:")
			}
			lines = append(lines, test.configLine, "")
			writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join(lines, "\n"))
			writeFile(t, filepath.Join(dir, "chapter.md"), "## Chapter 1\n\n> Quoted text.\n")
			output := filepath.Join(dir, "manuscript.typ")
			var stdout bytes.Buffer

			err := RunWithIO([]string{filepath.Join(dir, "chapter.md"), output}, &stdout)
			if err == nil {
				t.Fatalf("expected invalid configuration to fail")
			}
			assertContains(t, err.Error()+stdout.String(), test.wantPath)
			if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
				t.Fatalf("invalid configuration wrote output %s", output)
			}
		})
	}
}

func TestQuotedBlockFontPropertiesInheritWhenOmitted(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		wantStyle  string
	}{
		{"font block omitted", "", `style: "normal",`},
		{"style only", "folio:\n  manuscript:\n    quoted-block:\n      font:\n        style: italic\n", `style: "italic",`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typst := renderIssue15Manuscript(t, test.configYAML)
			assertContains(t, typst, strings.Join([]string{
				"#show quote.where(block: true): it => block(",
				"  above: 0.5em,",
				"  below: 0.5em,",
				")[",
				"  #pad(left: 0em)[#text(",
				`    font: "Libertinus Serif",`,
				"    size: 12pt,",
				`    weight: "regular",`,
				"    stretch: 100%,",
				"    " + test.wantStyle,
				"    tracking: 0em,",
			}, "\n"))
		})
	}
}

func TestBlockSpacingDefaultsWithoutPresetValues(t *testing.T) {
	var cfg Config
	normalizeConfig(&cfg)
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("defaulted configuration is invalid: %v", err)
	}
	if got := cfg.Folio.Manuscript.QuotedBlockSpacing.Value; got != "0.5em" {
		t.Fatalf("quoted-block spacing = %q, want 0.5em", got)
	}
	if got := cfg.Folio.Manuscript.CodeBlockSpacing.Value; got != "0.5em" {
		t.Fatalf("code-block spacing = %q, want 0.5em", got)
	}
	if got := cfg.Folio.Manuscript.QuoteBlockIndent.Value; got != "0em" {
		t.Fatalf("quote-block indent = %q, want 0em", got)
	}
	if got := cfg.Folio.Manuscript.CodeBlockIndent.Value; got != "0em" {
		t.Fatalf("code-block indent = %q, want 0em", got)
	}
}

func TestQuotedBlockFontAcceptsNumericWeightAndPlainStretch(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    quoted-block:",
		"      font:",
		"        weight: 250",
		"        stretch: 125",
		"",
	}, "\n"))

	assertContains(t, typst, "    weight: 250,")
	assertContains(t, typst, "    stretch: 125%,")
}

func TestPreciseCodeBlockSpacingOverridesEqualSpacing(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    code-block-spacing: 1.75em",
		"    code-block:",
		"      space-before: 0.25em",
		"",
	}, "\n"))

	assertContains(t, typst, strings.Join([]string{
		"#show raw.where(block: true): it => block(",
		"  above: 0.25em,",
		"  below: 1.75em,",
	}, "\n"))
}

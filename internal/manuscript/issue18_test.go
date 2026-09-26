// ABOUTME: Regression tests for issue #18 semantic authoring of parts/chapters.
// ABOUTME: Covers RT-18.1..RT-18.26 across parser derivation, config schema, rendering, and alt-format.
package manuscript

import (
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// AC18.1 -- position-derived numbering from source order
// -----------------------------------------------------------------------------

// RT-18.1: three H1s (`# One`, `# Two`, `# Three`) yield three part blocks with
// semantic-name field One, Two, Three and derived numbers 1, 2, 3.
func TestRT_18_1_PartsDerivedFromSourceOrder(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---",
		"title: Test",
		"---",
		"",
		"# One",
		"",
		"body one",
		"",
		"# Two",
		"",
		"body two",
		"",
		"# Three",
		"",
		"body three",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseMarkdown: %v", err)
	}
	var parts []Block
	for _, b := range doc.Blocks {
		if b.Kind == "part" {
			parts = append(parts, b)
		}
	}
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	for i, expected := range []struct {
		name   string
		number int
	}{{"One", 1}, {"Two", 2}, {"Three", 3}} {
		if parts[i].Name != expected.name {
			t.Errorf("part %d: name %q, want %q", i, parts[i].Name, expected.name)
		}
		if parts[i].Number != expected.number {
			t.Errorf("part %d: number %d, want %d", i, parts[i].Number, expected.number)
		}
	}
}

// RT-18.2: multi-file part numbering is derived from joined source order.
func TestRT_18_2_MultiFilePartNumbering(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "01.md"), strings.Join([]string{
		"---", "title: T", "---", "", "# One", "", "body A",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "02.md"), strings.Join([]string{
		"# Two", "", "body B",
	}, "\n"))
	output := filepath.Join(dir, "out.typ")
	runManuscriptDirect(t, filepath.Join(dir, "0?.md"), output)
	typst := readFile(t, output)
	// The two parts land as folio-part calls with name= "One" then name= "Two".
	assertContains(t, typst, `name: "One"`)
	assertContains(t, typst, `name: "Two"`)
	// Numbers 1 and 2.
	assertContains(t, typst, `number: "1"`)
	assertContains(t, typst, `number: "2"`)
}

// RT-18.3: chapter counter resets at each new part.
func TestRT_18_3_ChapterCounterResetsPerPart(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
		"", "## Second", "", "body",
		"", "# Beta", "", "## First Again", "", "body",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseMarkdown: %v", err)
	}
	var chapters []Block
	for _, b := range doc.Blocks {
		if b.Kind == "chapter" {
			chapters = append(chapters, b)
		}
	}
	if len(chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(chapters))
	}
	expected := []struct {
		name   string
		number int
	}{
		{"First", 1},
		{"Second", 2},
		{"First Again", 1}, // reset at Beta
	}
	for i, e := range expected {
		if chapters[i].Name != e.name || chapters[i].Number != e.number {
			t.Errorf("chapter %d: got name=%q number=%d, want name=%q number=%d",
				i, chapters[i].Name, chapters[i].Number, e.name, e.number)
		}
	}
}

// -----------------------------------------------------------------------------
// AC18.2 -- parser strips Part/Chapter prefixes and captures semantic name
// -----------------------------------------------------------------------------

// RT-18.4: `# PART ONE: UNBELIEVED` parses to name UNBELIEVED, source number ONE.
func TestRT_18_4_PartPrefixStrippedUnbelieved(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---", "title: T", "---", "",
		"# PART ONE: UNBELIEVED", "", "body",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseMarkdown: %v", err)
	}
	if len(doc.Blocks) == 0 || doc.Blocks[0].Kind != "part" {
		t.Fatalf("expected first block to be a part, got %#v", doc.Blocks)
	}
	part := doc.Blocks[0]
	if part.Name != "UNBELIEVED" {
		t.Errorf("Name = %q, want UNBELIEVED", part.Name)
	}
	if part.Number != 1 {
		t.Errorf("Number = %d, want 1", part.Number)
	}
	if part.SourceNumber != "ONE" {
		t.Errorf("SourceNumber = %q, want ONE", part.SourceNumber)
	}
	if part.SourceSeparator != ":" {
		t.Errorf("SourceSeparator = %q, want :", part.SourceSeparator)
	}
}

// RT-18.5: `## Chapter 12: The Watch` (first chapter in part) gets derived number 1
// with source number 12 ignored under default explicit-numbering: derived.
func TestRT_18_5_ChapterSourceNumberIgnoredByDefault(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---", "title: T", "---", "",
		"## Chapter 12: The Watch", "", "body",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseMarkdown: %v", err)
	}
	// First (and only) chapter.
	var chapter Block
	for _, b := range doc.Blocks {
		if b.Kind == "chapter" {
			chapter = b
			break
		}
	}
	if chapter.Kind != "chapter" {
		t.Fatalf("no chapter block in parse output: %#v", doc.Blocks)
	}
	if chapter.Name != "The Watch" {
		t.Errorf("Name = %q, want The Watch", chapter.Name)
	}
	if chapter.Number != 1 {
		t.Errorf("derived Number = %d, want 1 (source 12 must be ignored under default)", chapter.Number)
	}
	if chapter.SourceNumber != "12" {
		t.Errorf("SourceNumber = %q, want 12", chapter.SourceNumber)
	}
}

// RT-18.6: with chapter.explicit-numbering: source, the same `## Chapter 12: The Watch`
// renders with number 12 in the composed heading.
func TestRT_18_6_ChapterExplicitNumberingSource(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      explicit-numbering: source",
		"      prefix: \"Chapter \"",
		"      separator: \": \"",
		"      show-number: true",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Part One", "",
		"## Chapter 12: The Watch", "", "body",
	}, "\n"))
	// Composed heading with source number 12.
	assertContains(t, typst, `full: "Chapter 12: The Watch"`)
}

// RT-18.7: a heading not matching the prefix pattern (e.g. `# Unbelieved`) is used verbatim.
func TestRT_18_7_HeadingWithoutPrefixVerbatim(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---", "title: T", "---", "",
		"# Unbelieved", "", "body",
	}, "\n"))
	if err != nil {
		t.Fatalf("parseMarkdown: %v", err)
	}
	part := doc.Blocks[0]
	if part.Name != "Unbelieved" {
		t.Errorf("Name = %q, want Unbelieved", part.Name)
	}
	if part.Number != 1 {
		t.Errorf("Number = %d, want 1", part.Number)
	}
	if part.SourceNumber != "" {
		t.Errorf("SourceNumber = %q, want empty", part.SourceNumber)
	}
}

// -----------------------------------------------------------------------------
// AC18.3 -- prefix / number-format / separator / suffix / show-* config keys
// -----------------------------------------------------------------------------

// RT-18.8: part.prefix + number-format + separator produces "PART 1: <name>".
func TestRT_18_8_PartComposedHeading(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      prefix: \"PART \"",
		"      number-format: \"1\"",
		"      separator: \": \"",
		"      show-number: true",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Unbelieved", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "PART 1: Unbelieved"`)
}

// RT-18.9: chapter.number-format: "I" renders chapter numbers as roman upper.
func TestRT_18_9_ChapterNumberFormatRomanUpper(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"Chapter \"",
		"      number-format: \"I\"",
		"      separator: \": \"",
		"      show-number: true",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Book", "",
		"## First", "", "body", "",
		"## Second", "", "body", "",
		"## Third", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Chapter I: First"`)
	assertContains(t, typst, `full: "Chapter II: Second"`)
	assertContains(t, typst, `full: "Chapter III: Third"`)
}

// RT-18.10: chapter.number-format: "i" renders chapter numbers as roman lower.
func TestRT_18_10_ChapterNumberFormatRomanLower(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"Chapter \"",
		"      number-format: \"i\"",
		"      separator: \": \"",
		"      show-number: true",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Book", "",
		"## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Chapter i: First"`)
}

// RT-18.11: chapter.show-number: false omits the number entirely.
func TestRT_18_11_ChapterShowNumberFalse(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"Chapter: \"",
		"      show-number: false",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Book", "",
		"## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Chapter: First"`)
}

// RT-18.12: chapter.show-name: false omits the semantic name.
func TestRT_18_12_ChapterShowNameFalse(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"Chapter \"",
		"      show-number: true",
		"      show-name: false",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Book", "",
		"## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Chapter 1"`)
}

// RT-18.13: unconfigured manuscript renders parts/chapters as bare semantic names.
func TestRT_18_13_DefaultsRenderSemanticNameOnly(t *testing.T) {
	typst := renderIssue18Manuscript(t, "", strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	// Default: no prefix, show-number defaults false, show-name defaults true.
	// So the composed heading is just the semantic name.
	assertContains(t, typst, `full: "Alpha"`)
	assertContains(t, typst, `full: "First"`)
}

// -----------------------------------------------------------------------------
// AC18.4 -- extended placeholders in header/footer format
// -----------------------------------------------------------------------------

// RT-18.14: [chapter] returns semantic name only in the running header.
func TestRT_18_14_ChapterPlaceholderReturnsSemanticName(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[page] > [chapter] > [part]\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "author: A", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `state("folio-current-chapter-name").get()`)
	assertContains(t, typst, `state("folio-current-part-name").get()`)
}

// RT-18.15: [chapter-number] returns the formatted number.
func TestRT_18_15_ChapterNumberPlaceholder(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[chapter-number] * [chapter]\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `state("folio-current-chapter-number").get()`)
}

// RT-18.16: [chapter-prefix] returns the configured prefix.
func TestRT_18_16_ChapterPrefixPlaceholder(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[chapter-prefix] * [chapter]\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `state("folio-current-chapter-prefix").get()`)
}

// RT-18.17: [chapter-full] returns the fully rendered chapter heading.
func TestRT_18_17_ChapterFullPlaceholder(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[page] * [chapter-full]\"",
		"    chapter:",
		"      prefix: \"Chapter \"",
		"      separator: \": \"",
		"      show-number: true",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `state("folio-current-chapter-full").get()`)
	// The state stores the fully composed heading.
	assertContains(t, typst, `full: "Chapter 1: First"`)
}

// RT-18.18: unknown bracket tokens still render verbatim (regression for AC15.1).
func TestRT_18_18_UnknownBracketTokenRenders(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[undefined]\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `\[undefined\]`)
}

// -----------------------------------------------------------------------------
// AC18.5 -- case-transform and name-case
// -----------------------------------------------------------------------------

// RT-18.19: part.case-transform: upper uppercases the composed heading.
func TestRT_18_19_CaseTransformUpper(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      prefix: \"Part \"",
		"      separator: \": \"",
		"      show-number: true",
		"      case-transform: upper",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Unbelieved", "", "body",
	}, "\n"))
	// caseTransform("upper") uppercases the full composed heading. The `full: "..."`
	// argument on the folio-part call stores the untransformed composed heading; the
	// uppercased form is emitted as the body content.
	assertContains(t, typst, `[PART 1: UNBELIEVED]`)
}

// RT-18.20: part.name-case: upper uppercases just the name segment.
func TestRT_18_20_NameCaseUpper(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      prefix: \"Part \"",
		"      separator: \": \"",
		"      show-number: true",
		"      name-case: upper",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Unbelieved", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Part 1: UNBELIEVED"`)
}

// RT-18.21: part.name-case: title renders "# the watch" as "Part 1: The Watch".
func TestRT_18_21_NameCaseTitle(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      prefix: \"Part \"",
		"      separator: \": \"",
		"      show-number: true",
		"      name-case: title",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# the watch", "", "body",
	}, "\n"))
	assertContains(t, typst, `full: "Part 1: The Watch"`)
}

// -----------------------------------------------------------------------------
// AC18.6 -- alt-format for right (recto) pages
// -----------------------------------------------------------------------------

// RT-18.22: page-header with format + alt-format emits a #context if calc.odd branch.
func TestRT_18_22_HeaderAltFormatEmitsBranching(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"L-format\"",
		"      alt-format: \"R-alt-format\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `L-format`)
	assertContains(t, typst, `R-alt-format`)
	assertContains(t, typst, `if calc.odd(pg)`)
}

// RT-18.23: page-header with only format (no alt-format) emits neither the branch nor an alt.
func TestRT_18_23_HeaderAltUnsetNoBranch(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"only-format\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `only-format`)
	// No if calc.odd branch was emitted for the header (footer still may have one, but we
	// use a distinctive format text here to focus on the header line).
	body := extractBodyPageBlock(t, typst)
	// The header text region in the body-page setup does not contain a calc.odd branch.
	// (There is a calc.odd(pg) elsewhere for footer skip-list membership; not this one.)
	if strings.Contains(body, `if calc.odd(pg) { [`) {
		t.Fatalf("expected no alt-format branch when alt-format is unset; got:\n%s", body)
	}
}

// RT-18.24: page-footer.alt-format works symmetrically.
func TestRT_18_24_FooterAltFormatSymmetric(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-L\"",
		"      alt-format: \"footer-R\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	assertContains(t, typst, `footer-L`)
	assertContains(t, typst, `footer-R`)
}

// RT-18.25: alt-format accepts the full placeholder set (including AC18.4 placeholders).
func TestRT_18_25_AltFormatAcceptsExtendedPlaceholders(t *testing.T) {
	typst := renderIssue18Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      format: \"[chapter]\"",
		"      alt-format: \"[chapter-full] * [part-number]\"",
	}, "\n"), strings.Join([]string{
		"---", "title: T", "---", "",
		"# Alpha", "", "## First", "", "body",
	}, "\n"))
	// The alt-format substitution wires both chapter-full and part-number states.
	assertContains(t, typst, `state("folio-current-chapter-full").get()`)
	assertContains(t, typst, `state("folio-current-part-number").get()`)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// renderIssue18Manuscript writes a script.yaml and a chapter file to a temp directory
// and returns the generated Typst source.
func renderIssue18Manuscript(t *testing.T, scriptYAML string, chapterMD string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	enableFixtureHeader(t)
	dir := t.TempDir()
	if scriptYAML != "" {
		writeFile(t, filepath.Join(dir, "script.yaml"), scriptYAML)
	}
	writeFile(t, filepath.Join(dir, "ch01.md"), chapterMD)
	output := filepath.Join(dir, "out.typ")
	runManuscriptDirect(t, filepath.Join(dir, "ch01.md"), output)
	return readFile(t, output)
}

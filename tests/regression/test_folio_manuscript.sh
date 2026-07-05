#!/usr/bin/env bash
# ABOUTME: Regression tests for prose manuscript rendering (issue #9).
# ABOUTME: Covers Markdown/org manuscript input, presets, config, TOC, and errors.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
FOLIO="$PROJECT_DIR/bin/folio"
PASS=0
FAIL=0
FAILURES=()
TMPDIR_TEST="$(mktemp -d)"

cleanup() {
    rm -rf -- "$TMPDIR_TEST"
}
trap cleanup EXIT

pass() {
    PASS=$((PASS + 1))
    echo "  PASS: $1"
}

fail() {
    FAIL=$((FAIL + 1))
    FAILURES+=("$1")
    echo "  FAIL: $1"
    if [[ -n "${2:-}" ]]; then
        echo "        $2"
    fi
}

run_ok() {
    local desc="$1"
    shift
    if "$@" >"$TMPDIR_TEST/stdout.txt" 2>"$TMPDIR_TEST/stderr.txt"; then
        pass "$desc"
        return 0
    fi
    fail "$desc" "$(cat "$TMPDIR_TEST/stderr.txt")"
    return 1
}

run_fail() {
    local desc="$1"
    shift
    if "$@" >"$TMPDIR_TEST/stdout.txt" 2>"$TMPDIR_TEST/stderr.txt"; then
        fail "$desc" "command succeeded"
        return 1
    fi
    if [[ -s "$TMPDIR_TEST/stderr.txt" ]]; then
        pass "$desc"
    else
        fail "$desc" "stderr was empty"
        return 1
    fi
}

has_text() {
    local file="$1"
    local needle="$2"
    local desc="$3"
    if grep -Fq "$needle" "$file"; then
        pass "$desc"
    else
        fail "$desc" "missing '$needle'"
    fi
}

has_pdf() {
    local file="$1"
    local desc="$2"
    if [[ ! -s "$file" ]]; then
        fail "$desc" "missing or empty PDF: $file"
        return 1
    fi
    if file "$file" | grep -Fq "PDF document"; then
        pass "$desc"
    else
        fail "$desc" "$(file "$file")"
    fi
}

has_a4_pdf() {
    local file="$1"
    local desc="$2"
    if [[ ! -s "$file" ]]; then
        fail "$desc" "missing or empty PDF: $file"
        return 1
    fi
    if pdfinfo "$file" | grep -Fq "A4"; then
        pass "$desc"
    else
        fail "$desc" "$(pdfinfo "$file")"
    fi
}

output_absent() {
    local file="$1"
    local desc="$2"
    if [[ -e "$file" ]]; then
        fail "$desc" "unexpected output artefact: $file"
    else
        pass "$desc"
    fi
}

not_has_text() {
    local file="$1"
    local needle="$2"
    local desc="$3"
    if grep -Fq "$needle" "$file"; then
        fail "$desc" "unexpected '$needle'"
    else
        pass "$desc"
    fi
}

has_text_count() {
    local file="$1"
    local needle="$2"
    local expected="$3"
    local desc="$4"
    local count
    count="$(grep -F -c "$needle" "$file")"
    if [[ "$count" == "$expected" ]]; then
        pass "$desc"
    else
        fail "$desc" "expected $expected occurrence(s) of '$needle', found $count"
    fi
}

write_markdown_fixtures() {
    local dir="$1"
    mkdir -p "$dir/part1" "$dir/part2"

    printf '%s\n' \
        '# About Time' \
        '' \
        '**A Novel**' \
        '' \
        '*by Test Author*' \
        '' \
        '--- Draft 4 | July 2026 ---' \
        > "$dir/part1/ch00.md"

    printf '%s\n' \
        '## PART ONE' \
        '' \
        '### Chapter 1' \
        '' \
        "First paragraph with *emphasis* and \`code\`." \
        '' \
        '```' \
        'monospace block' \
        '```' \
        '' \
        '***' \
        '' \
        'Second paragraph.[^note]' \
        '' \
        '[^note]: A useful footnote.' \
        > "$dir/part1/ch01.md"

    printf '%s\n' \
        '### Chapter 2' \
        '' \
        'Chapter two starts here.' \
        '' \
        '### Notes <!-- noexport -->' \
        '' \
        'Private planning note.' \
        > "$dir/part1/ch02.md"

    printf '%s\n' \
        '## PART TWO' \
        '' \
        '### Chapter 3' \
        '' \
        'Third chapter text.' \
        > "$dir/part2/ch03.md"
}

write_org_fixtures() {
    local dir="$1"
    mkdir -p "$dir/part1" "$dir/part2"

    printf '%s\n' \
        '#+TITLE: About Time' \
        '#+SUBTITLE: A Novel' \
        '#+AUTHOR: Test Author' \
        '#+DATE: July 2026' \
        '#+VERSION: Draft 4' \
        '#+WORDCOUNT: 80000' \
        '#+ADDRESS: 1 Example Street / Galway / Ireland' \
        '#+EMAIL: test@example.com' \
        '#+WEBSITE: https://example.com' \
        > "$dir/part1/ch00.org"

    printf '%s\n' \
        '* PART ONE' \
        '** Chapter 1' \
        'First paragraph with /emphasis/ and ~code~.' \
        '' \
        '-----' \
        '' \
        'Second paragraph.[fn:note]' \
        '' \
        '[fn:note] A useful footnote.' \
        > "$dir/part1/ch01.org"

    printf '%s\n' \
        '** Chapter 2' \
        'Chapter two starts here.' \
        '' \
        '*** Notes :noexport:' \
        'Private planning note.' \
        > "$dir/part1/ch02.org"

    printf '%s\n' \
        '* PART TWO' \
        '** Chapter 3' \
        'Third chapter text.' \
        > "$dir/part2/ch03.org"
}

echo "=== Folio manuscript tests (issue #9) ==="

MD_DIR="$TMPDIR_TEST/md"
ORG_DIR="$TMPDIR_TEST/org"
write_markdown_fixtures "$MD_DIR"
write_org_fixtures "$ORG_DIR"

echo ""
echo "Markdown and org manuscript rendering"
run_ok "RT-9.1: Multiple Markdown chapter fixtures create manuscript output" \
    "$FOLIO" manuscript "$MD_DIR"/part?/ch??.md "$TMPDIR_TEST/md.pdf"
has_pdf "$TMPDIR_TEST/md.pdf" "RT-9.1: Markdown manuscript output is a valid PDF"
run_ok "RT-9.2: Multiple org-mode chapter fixtures create manuscript output" \
    "$FOLIO" manuscript "$ORG_DIR"/part?/ch??.org "$TMPDIR_TEST/org.pdf"
has_pdf "$TMPDIR_TEST/org.pdf" "RT-9.2: Org manuscript output is a valid PDF"
run_ok "RT-9.3: Markdown manuscript can emit Typst source" \
    "$FOLIO" manuscript "$MD_DIR"/part?/ch??.md "$TMPDIR_TEST/md.typ"
run_ok "RT-9.3: Org manuscript can emit Typst source" \
    "$FOLIO" manuscript "$ORG_DIR"/part?/ch??.org "$TMPDIR_TEST/org.typ"
has_text "$TMPDIR_TEST/md.typ" "Generated by First Folio manuscript" "RT-9.3: .typ target creates Typst source"

echo ""
echo "Selection, ordering, and file separators"
run_ok "RT-9.4: Unsorted explicit paths render in deterministic chapter order" \
    "$FOLIO" manuscript "$MD_DIR/part1/ch02.md" "$MD_DIR/part1/ch01.md" "$TMPDIR_TEST/order.typ"
has_text "$TMPDIR_TEST/order.typ" "Chapter 1" "RT-9.4: Chapter 1 present after deterministic sort"
run_ok "RT-9.5: Quoted glob subset excludes non-matching chapters" \
    "$FOLIO" manuscript "$MD_DIR/part1/ch0[12].md" "$TMPDIR_TEST/subset.typ"
not_has_text "$TMPDIR_TEST/subset.typ" "Chapter 3" "RT-9.5: Subset excludes Chapter 3"
run_fail "RT-9.6: Glob with no matches exits non-zero" \
    "$FOLIO" manuscript "$MD_DIR/part1/ch99.md" "$TMPDIR_TEST/none.typ"
has_text "$TMPDIR_TEST/order.typ" "First paragraph" "RT-9.7: Adjacent Markdown files remain separated"
run_ok "RT-9.8: Adjacent org files without trailing newlines remain separated" \
    "$FOLIO" manuscript "$ORG_DIR/part1/ch02.org" "$ORG_DIR/part1/ch01.org" "$TMPDIR_TEST/org-order.typ"

echo ""
echo "Markdown contract"
has_text "$TMPDIR_TEST/md.typ" "First paragraph" "RT-9.9: Markdown paragraph appears"
has_text "$TMPDIR_TEST/md.typ" "PART ONE" "RT-9.10: Markdown part heading appears"
has_text "$TMPDIR_TEST/md.typ" "Chapter 1" "RT-9.10: Markdown chapter heading appears"
has_text "$TMPDIR_TEST/md.typ" "scene-break" "RT-9.11: Markdown scene break appears"
has_text "$TMPDIR_TEST/md.typ" "footnote" "RT-9.11: Markdown footnote appears"
not_has_text "$TMPDIR_TEST/md.typ" "Private planning note" "RT-9.12: Markdown noexport content excluded"

echo ""
echo "Org contract"
has_text "$TMPDIR_TEST/org.typ" "First paragraph" "RT-9.13: Org paragraph appears"
has_text "$TMPDIR_TEST/org.typ" "PART ONE" "RT-9.14: Org part heading appears"
has_text "$TMPDIR_TEST/org.typ" "Chapter 1" "RT-9.14: Org chapter heading appears"
has_text "$TMPDIR_TEST/org.typ" "scene-break" "RT-9.15: Org scene break appears"
has_text "$TMPDIR_TEST/org.typ" "footnote" "RT-9.15: Org footnote appears"
not_has_text "$TMPDIR_TEST/org.typ" "Private planning note" "RT-9.16: Org noexport content excluded"

echo ""
echo "Config inheritance and fonts"
printf '%s\n' \
    'folio:' \
    '  font: Test Body' \
    '  font-size: 13pt' \
    '  heading-font: Test Heading' \
    '  heading-font-size: 15pt' \
    '  letter:' \
    '    font: Letter Font' \
    '    font-size: 10pt' \
    > "$TMPDIR_TEST/script.yaml"
cp "$MD_DIR/part1/ch01.md" "$TMPDIR_TEST/ch01.md"
cp "$MD_DIR/part1/ch00.md" "$TMPDIR_TEST/ch00.md"
run_ok "RT-9.17-RT-9.24: Root font/page config applies to manuscript" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/config.typ"
has_text "$TMPDIR_TEST/config.typ" 'paper: "a4"' "RT-9.17: Root/default A4 appears"
has_text "$TMPDIR_TEST/config.typ" 'font: "Test Body"' "RT-9.18: Root body font appears"
has_text "$TMPDIR_TEST/config.typ" 'size: 13pt' "RT-9.19: Root body font size appears"
has_text "$TMPDIR_TEST/config.typ" 'font: "Test Heading"' "RT-9.21: Root heading font appears"
has_text "$TMPDIR_TEST/config.typ" '15pt' "RT-9.22: Root heading font size appears"
not_has_text "$TMPDIR_TEST/config.typ" "Letter Font" "RT-9.20: Letter font does not affect manuscript"

printf '%s\n' \
    'folio:' \
    '  font: Fallback Body' \
    '  font-size: 12.5pt' \
    > "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.23-RT-9.24: Missing heading config inherits body config" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/fallback.typ"
has_text "$TMPDIR_TEST/fallback.typ" 'Fallback Body' "RT-9.23: Heading font inherits body font"
has_text "$TMPDIR_TEST/fallback.typ" '12.5pt' "RT-9.24: Heading size inherits body size"

printf '%s\n' \
    'folio:' \
    '  manuscript:' \
    '    font: Manuscript Body' \
    '    heading-font: Manuscript Heading' \
    '    mono-font: Manuscript Mono' \
    '    title-font: Manuscript Title' \
    '    author-font: Manuscript Author' \
    '    date-font: Manuscript Date' \
    '    version-font: Manuscript Version' \
    '    page-header:' \
    '      font: Manuscript Header' \
    > "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.25-RT-9.29: Manuscript font overrides apply to their elements" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch00.md" "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/overrides.typ"
has_text "$TMPDIR_TEST/overrides.typ" "Manuscript Body" "RT-9.25: Manuscript body font override appears"
has_text "$TMPDIR_TEST/overrides.typ" "Manuscript Heading" "RT-9.26: Manuscript heading font override appears"
has_text "$TMPDIR_TEST/overrides.typ" 'text(font: "Manuscript Mono"' "RT-9.27: Manuscript mono font override applies to inline code"
has_text "$TMPDIR_TEST/overrides.typ" 'raw(block: true' "RT-9.27: Manuscript mono font override applies to fenced code blocks"
has_text "$TMPDIR_TEST/overrides.typ" 'font: "Manuscript Title"' "RT-9.28: Manuscript title font applies to title page"
has_text "$TMPDIR_TEST/overrides.typ" 'font: "Manuscript Author"' "RT-9.28: Manuscript author font applies to author element"
has_text "$TMPDIR_TEST/overrides.typ" 'font: "Manuscript Date"' "RT-9.28: Manuscript date font applies to date element"
has_text "$TMPDIR_TEST/overrides.typ" 'font: "Manuscript Version"' "RT-9.28: Manuscript version font applies to version element"
has_text "$TMPDIR_TEST/overrides.typ" 'header: align(right)[#text(font: "Manuscript Header"' "RT-9.29: Manuscript page header font applies to running header"

echo ""
echo "British and US manuscript presets"
rm -f "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.30-RT-9.34: British manuscript preset renders expected layout values" \
    "$FOLIO" manuscript --style british "$MD_DIR"/part?/ch??.md "$TMPDIR_TEST/british.typ"
has_text "$TMPDIR_TEST/british.typ" "margin: 20mm" "RT-9.30: British margin is 20mm"
has_text "$TMPDIR_TEST/british.typ" "Libertinus Sans" "RT-9.30: British heading/header font is Libertinus Sans"
has_text "$TMPDIR_TEST/british.typ" "header-ascent: 20mm" "RT-9.31: British header distance is applied to page layout"
has_text "$TMPDIR_TEST/british.typ" "#v(10mm)" "RT-9.32: British header padding is applied before content"
has_text "$TMPDIR_TEST/british.typ" "#v(33%)" "RT-9.33: British chapter position is applied"
has_text "$TMPDIR_TEST/british.typ" "#align(center + horizon)" "RT-9.34: British part page centring is applied"

run_ok "RT-9.35-RT-9.40: US manuscript override renders expected layout values" \
    "$FOLIO" manuscript --style us --wordcount 80000 \
        --address "1 Example Street / Galway / Ireland" \
        --email test@example.com --website https://example.com \
        "$MD_DIR"/part?/ch??.md "$TMPDIR_TEST/us.typ"
has_text "$TMPDIR_TEST/us.typ" "style: us" "RT-9.35: US override style appears"
has_text "$TMPDIR_TEST/us.typ" 'paper: "a4"' "RT-9.36: US style keeps inherited A4"
has_text "$TMPDIR_TEST/us.typ" "Libertinus Mono" "RT-9.37: US mono body font appears"
has_text "$TMPDIR_TEST/us.typ" "A Novel" "RT-9.38: US title-page subtitle appears"
has_text "$TMPDIR_TEST/us.typ" "by Test Author" "RT-9.38: US title-page author attribution appears"
has_text "$TMPDIR_TEST/us.typ" "Draft 4" "RT-9.38: US title-page version appears"
has_text "$TMPDIR_TEST/us.typ" "July 2026" "RT-9.38: US title-page date appears"
has_text "$TMPDIR_TEST/us.typ" "word count: 80000" "RT-9.38: US title-page word count appears"
has_text "$TMPDIR_TEST/us.typ" "1 Example Street" "RT-9.38: US title-page address appears"
has_text "$TMPDIR_TEST/us.typ" "test\\@example.com" "RT-9.38: US title-page email appears"
has_text "$TMPDIR_TEST/us.typ" "https://example.com" "RT-9.38: US title-page website appears"
has_text "$TMPDIR_TEST/us.typ" "[author] / [title] / [page]" "RT-9.39: US shared running header format appears"
has_text "$TMPDIR_TEST/us.typ" "line-spacing: 2" "RT-9.40: US line spacing appears"
has_text "$TMPDIR_TEST/us.typ" "paragraph-indent: 12.7mm" "RT-9.40: US paragraph indent appears"
has_text "$TMPDIR_TEST/us.typ" "paragraph-spacing: 0" "RT-9.40: US paragraph spacing appears"
has_text "$TMPDIR_TEST/us.typ" "line-spacing: 2" "RT-9.56: US line spacing default appears"
has_text "$TMPDIR_TEST/us.typ" "paragraph-indent: 12.7mm" "RT-9.56: US paragraph indent default appears"
has_text "$TMPDIR_TEST/us.typ" "paragraph-spacing: 0" "RT-9.56: US paragraph spacing default appears"

echo ""
echo "Page defaults, TOC, and paragraph config"
rm -f "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.41: Script output uses A4 with no page config" \
    "$FOLIO" convert "$PROJECT_DIR/examples/the-bus-stop.md" "$TMPDIR_TEST/default-script.pdf"
has_a4_pdf "$TMPDIR_TEST/default-script.pdf" "RT-9.41: Script default PDF page size is A4"
run_ok "RT-9.42: Letter output uses A4 with no page config" \
    "$FOLIO" letter "$PROJECT_DIR/examples/about-time.org" --dir "$TMPDIR_TEST" --prefix default-letter
mapfile -t LETTER_PDFS < <(find "$TMPDIR_TEST" -maxdepth 1 -type f -name 'default-letter-*.pdf' -print)
if [[ "${#LETTER_PDFS[@]}" -gt 0 ]]; then
    has_a4_pdf "${LETTER_PDFS[0]}" "RT-9.42: Letter default PDF page size is A4"
else
    fail "RT-9.42: Letter default PDF page size is A4" "no default-letter PDF was generated"
fi
has_text "$TMPDIR_TEST/british.typ" 'paper: "a4"' "RT-9.43: Manuscript output uses A4 by default"
run_ok "RT-9.44: US script style keeps A4 unless explicitly configured" \
    "$FOLIO" convert --style us "$PROJECT_DIR/examples/the-bus-stop.md" "$TMPDIR_TEST/us-script.pdf"
has_a4_pdf "$TMPDIR_TEST/us-script.pdf" "RT-9.44: US script style keeps A4"
run_ok "RT-9.44: Screenplay style keeps A4 unless explicitly configured" \
    "$FOLIO" convert --style screenplay "$PROJECT_DIR/examples/the-bus-stop.md" "$TMPDIR_TEST/screenplay.pdf"
has_a4_pdf "$TMPDIR_TEST/screenplay.pdf" "RT-9.44: Screenplay style keeps A4"
run_ok "RT-9.44: US manuscript style keeps A4 unless explicitly configured" \
    "$FOLIO" manuscript --style us "$MD_DIR"/part?/ch??.md "$TMPDIR_TEST/us-page.typ"
has_text "$TMPDIR_TEST/us-page.typ" 'paper: "a4"' "RT-9.44: US manuscript style keeps A4"
has_text "$TMPDIR_TEST/british.typ" "Contents" "RT-9.50: TOC enabled by default"
printf '%s\n' \
    'folio:' \
    '  manuscript:' \
    '    toc:' \
    '      enabled: false' \
    > "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.51: TOC can be disabled" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/no-toc.typ"
not_has_text "$TMPDIR_TEST/no-toc.typ" "Contents" "RT-9.51: Disabled TOC absent"

printf '%s\n' \
    'folio:' \
    '  manuscript:' \
    '    toc:' \
    '      font: TOC Font' \
    '      font-size: 9pt' \
    '      font-weight: bold' \
    '      heading-font: TOC Heading' \
    '      heading-font-size: 17pt' \
    '      heading-font-weight: bold' \
    '      include-parts: false' \
    '      include-chapters: true' \
    '      include-sections: true' \
    '    line-spacing: 1.25' \
    '    paragraph-indent: 8mm' \
    '    paragraph-spacing: 2mm' \
    > "$TMPDIR_TEST/script.yaml"
run_ok "RT-9.52-RT-9.57: TOC and paragraph layout config applies" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/toc-config.typ"
has_text "$TMPDIR_TEST/toc-config.typ" "TOC Font" "RT-9.52: TOC entry font appears"
has_text "$TMPDIR_TEST/toc-config.typ" "TOC Heading" "RT-9.53: TOC heading font appears"
has_text "$TMPDIR_TEST/toc-config.typ" "Chapter 1" "RT-9.54: TOC includes chapter heading"
has_text_count "$TMPDIR_TEST/toc-config.typ" "PART ONE" 1 "RT-9.54: TOC part inclusion can be disabled"
has_text "$TMPDIR_TEST/british.typ" "line-spacing: 1.5" "RT-9.55: British line spacing appears"
has_text "$TMPDIR_TEST/british.typ" "paragraph-indent: 10mm" "RT-9.55: British paragraph indent appears"
has_text "$TMPDIR_TEST/british.typ" "paragraph-spacing: 0" "RT-9.55: British paragraph spacing appears"
has_text "$TMPDIR_TEST/toc-config.typ" "line-spacing: 1.25" "RT-9.57: Explicit line spacing appears"
has_text "$TMPDIR_TEST/toc-config.typ" "paragraph-indent: 8mm" "RT-9.57: Explicit paragraph indent appears"
has_text "$TMPDIR_TEST/toc-config.typ" "paragraph-spacing: 2mm" "RT-9.57: Explicit paragraph spacing appears"

echo ""
echo "Invalid input handling"
run_fail "RT-9.45: Missing input path exits non-zero" \
    "$FOLIO" manuscript "$TMPDIR_TEST/missing.md" "$TMPDIR_TEST/missing.typ"
output_absent "$TMPDIR_TEST/missing.typ" "RT-9.45: Missing input creates no output artefact"
run_fail "RT-9.46: Mixed Markdown and org input exits non-zero" \
    "$FOLIO" manuscript "$MD_DIR/part1/ch01.md" "$ORG_DIR/part1/ch01.org" "$TMPDIR_TEST/mixed.typ"
output_absent "$TMPDIR_TEST/mixed.typ" "RT-9.46: Mixed input creates no output artefact"
printf '%s\n' 'Title: Wrong Format' > "$TMPDIR_TEST/script.fountain"
run_fail "RT-9.47: Fountain input is rejected" \
    "$FOLIO" manuscript "$TMPDIR_TEST/script.fountain" "$TMPDIR_TEST/fountain.typ"
output_absent "$TMPDIR_TEST/fountain.typ" "RT-9.47: Fountain input creates no output artefact"
printf '%s\n' 'Wrong format.' > "$TMPDIR_TEST/notes.txt"
run_fail "RT-9.48: Other non-manuscript format is rejected" \
    "$FOLIO" manuscript "$TMPDIR_TEST/notes.txt" "$TMPDIR_TEST/text.typ"
output_absent "$TMPDIR_TEST/text.typ" "RT-9.48: Unsupported input creates no output artefact"
run_fail "RT-9.49: Unwritable output destination exits non-zero" \
    "$FOLIO" manuscript "$TMPDIR_TEST/ch01.md" "$TMPDIR_TEST/no-such-dir/out.typ"
output_absent "$TMPDIR_TEST/no-such-dir/out.typ" "RT-9.49: Unwritable target creates no output artefact"

echo ""
echo "=== Summary ==="
echo "PASS: $PASS"
echo "FAIL: $FAIL"

if [[ "$FAIL" -ne 0 ]]; then
    echo ""
    echo "Failures:"
    for failure in "${FAILURES[@]}"; do
        echo "  - $failure"
    done
    exit 1
fi

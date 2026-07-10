#!/usr/bin/env bash
# ABOUTME: Regression tests for cover-letter generation.
# ABOUTME: Covers issue #12 inline org markup rendering in letter bodies.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
FOLIO="$PROJECT_DIR/bin/folio"
PASS=0
FAIL=0
FAILURES=()
mkdir -p "$PROJECT_DIR/.agent/tmp"
TMPDIR_TEST="$(mktemp -d "$PROJECT_DIR/.agent/tmp/folio-letter-test.XXXXXX")"

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

assert_contains() {
    local content="$1"
    local needle="$2"
    local desc="$3"

    if [[ "$content" == *"$needle"* ]]; then
        pass "$desc"
    else
        fail "$desc" "Missing: $needle"
    fi
}

assert_not_contains() {
    local content="$1"
    local needle="$2"
    local desc="$3"

    if [[ "$content" == *"$needle"* ]]; then
        fail "$desc" "Unexpected: $needle"
    else
        pass "$desc"
    fi
}

latest_typst_source() {
    local typst_files=("$TMPDIR_TEST"/typst/cover-*.typ)

    if [[ ! -e "${typst_files[0]}" ]]; then
        echo "No kept Typst source found" >&2
        return 1
    fi

    printf '%s\n' "${typst_files[0]}"
}

create_letter_fixture() {
    cat > "$TMPDIR_TEST/letter.org" <<'ORG'
#+TITLE: Test Play
#+AUTHOR: Test Author
#+EMAIL: test@example.invalid

* Cover letters :letter:
** Sender / Address :sender:
** Submission :subject:
This sentence has =inline code= in it.

This line keeps *bold*, /italic/, and _underline_ markup.
*** Recipient / Address :to:
ORG
}

render_letter() {
    mkdir -p "$TMPDIR_TEST/typst"
    TMPDIR="$TMPDIR_TEST/typst" FOLIO_KEEP_TYPST=1 \
        "$FOLIO" letter "$TMPDIR_TEST/letter.org" \
        --dir "$TMPDIR_TEST" \
        --prefix issue12
}

echo "=== Cover letter tests ==="

create_letter_fixture
if render_letter; then
    typst_path="$(latest_typst_source)"
    typst_content="$(< "$typst_path")"

    assert_contains \
        "$typst_content" \
        '#text(font: "Libertinus Mono")[inline code]' \
        "RT-12.1: org inline code renders as monospace Typst text"
    assert_not_contains \
        "$typst_content" \
        '=inline code=' \
        "RT-12.1: org inline code delimiters are not visible in generated Typst"

    assert_contains "$typst_content" '*bold*' \
        "RT-12.2: org bold still renders as Typst strong markup"
    assert_contains "$typst_content" '_italic_' \
        "RT-12.2: org italic still renders as Typst emphasis markup"
    assert_contains "$typst_content" '#underline[underline]' \
        "RT-12.2: org underline still renders as Typst underline markup"
else
    fail "RT-12.1: folio letter command succeeds" "folio letter exited non-zero"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
if [[ "$FAIL" -gt 0 ]]; then
    printf 'Failures:\n'
    printf '  - %s\n' "${FAILURES[@]}"
    exit 1
fi

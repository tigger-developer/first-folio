---
title: Folio manuscript DOCX semantic-limit failure
version: 1
last-updated: 2026-10-19
---

# Finding

A direct DOCX export of the assembled About Time manuscript fails with exit
status 1 and produces no DOCX file. The failure reproduces when invoking Folio
directly, without the `render-manuscript` wrapper:

```text
Error: Pandoc semantic block, inline or nesting limit exceeded
```

The exact limit, threshold and responsible implementation have not been
identified. The message alone does not establish whether Pandoc itself failed
or Folio rejected a parsed representation. Folio source code was not inspected;
investigation of that implementation belongs to its owning project.

## Environment and input

- Executable: `/Users/tigger/.local/bin/folio`.
- Reported version: `folio 0.4.16`.
- Installed executable SHA-256, recorded on 2026-10-19:
  `6d4143b9764cd7716c33df6b18ea5dc56d03b35d1dc5d172e719772ecea236f2`.
- Wrapper: `/Users/tigger/bin/render-manuscript`, version 1.2, commit
  `83d32a94` (shared rendering loop for all output formats).
- Wrapper SHA-256:
  `1bfc6590cdda36bf6325050f8ab3aece8adb66d24ffabc66ec636b3d532692e3`.
- Manuscript root: `/Users/tigger/notes/about-time-novel`.
- Reported chapter word count: 109,119, excluding front matter.
- Failing input: the original edition, assembled as a 614,842-byte Markdown file.
- Selected variant: `script-6x9.yaml`, exposed to Folio through a staging
  directory's `script.yaml` symlink.

The version string remained `0.4.16` after an installation refresh. The binary
hash identifies the installed artefact observed during this investigation;
the version string alone does not distinguish the earlier and refreshed builds.

## Observed sequence

The supplied fresh-shell transcript identifies `render-manuscript` as
`/Users/tigger/bin/render-manuscript`. Its output detects `script-6x9.yaml` and
`script-a4.yaml`, with both original and quoted editions selected.

The wrapper renders each selected input through one `folio manuscript` call
inside a `pdf docx odt` loop. Only the output extension changes between formats.

For the run beginning at `2026-10-19_03-31-56`:

1. The original 6x9 PDF was written. The file exists and is 3,564,190 bytes.
2. The next command, the original 6x9 DOCX export, reported the semantic-limit
   error and returned failure.
3. The wrapper stopped before the original ODT, the quoted edition's exports
   and the A4 variant. Assembled Markdown and staging files were retained.

Retained input:

```text
/Users/tigger/notes/about-time-novel/rendered/.folio.wnU1yM/2026-10-19_03-31-56-a-thing-that-could-break.md
```

Configuration symlink:

```text
/Users/tigger/notes/about-time-novel/rendered/.folio.wnU1yM/script.yaml
  -> /Users/tigger/notes/about-time-novel/script-6x9.yaml
```

Existing PDF from that run:

```text
/Users/tigger/notes/about-time-novel/rendered/2026-10-19_03-31-56-a-thing-that-could-break-6x9.pdf
```

## Direct reproduction

The direct diagnostic used the retained staged input and a newly allocated
operating-system temporary directory for its target and temporary files.
It did not change the manuscript, project configuration or wrapper.

Equivalent Fish commands follow. Run them in order; capture the exit status
immediately after the Folio invocation. The retained input must still exist.

```fish
set -l folio_probe_dir (mktemp -d)
```

```fish
env TMPDIR="$folio_probe_dir" /Users/tigger/.local/bin/folio manuscript \
    /Users/tigger/notes/about-time-novel/rendered/.folio.wnU1yM/2026-10-19_03-31-56-a-thing-that-could-break.md \
    "$folio_probe_dir/probe.docx"
```

```fish
printf 'Exit status: %s\n' $status
```

Observed result: exit status 1; `probe.docx` was not created. Stderr was:

```text
Warning: loaded YAML file "/Users/tigger/notes/about-time-novel/rendered/.folio.wnU1yM/script.yaml" into manuscript configuration cascade.
Error: Pandoc semantic block, inline or nesting limit exceeded
```

The direct diagnostic's logs are retained at:

```text
/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-docx-failure-20261019-44241-x6hbn5/stdout.log
/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-docx-failure-20261019-44241-x6hbn5/stderr.log
```

These are temporary evidence paths and may disappear during later cleanup.

## Pandoc dependency distinction

The refreshed executable's `folio manuscript --help` states that both office
formats use Pandoc for source parsing, DOCX generation is native, and ODT uses
the configured Pandoc writer. A native DOCX writer therefore does not establish
that the DOCX export path is independent of Pandoc parsing or semantic limits.

The current wrapper invokes Folio for every format. It does not invoke
`pandocx`, `pandodt` or Pandoc directly.

## Evidence boundaries and next investigation

- Direct reproduction establishes that the wrapper is not necessary to trigger
  the failure. It does not identify the failing internal function.
- The manuscript's length is relevant test data, not a demonstrated cause.
  Block count, inline count, nesting depth and configured thresholds remain
  unmeasured.
- There is no evidence that the YAML-load diagnostic or quotation warnings
  caused this failure.
- Small synthetic manuscripts exported successfully during wrapper validation.
  Those checks did not establish full-novel export capability.
- The exact internal Pandoc invocation, parser options and returned data were
  not captured. The original 6x9 DOCX is the only failed format reproduced
  directly; ODT and A4 results cannot be inferred from it.
- The owning project should locate the exact error string, identify each
  guarded quantity and threshold, and measure this input against them before
  selecting a correction. Arbitrarily raising or disabling limits is not
  justified by the current evidence.

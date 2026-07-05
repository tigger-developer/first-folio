<!-- Version: 1.0 | Last updated: 2026-07-06 -->
Usage: folio manuscript <input.md|input.org...> <output.typ|output.pdf> [options]

Generate a prose manuscript from one or more Markdown or org-mode files.
Inputs are sorted deterministically and joined with a blank line separator.

Options:
  --style STYLE              Manuscript preset: british (default) or us
  --font FONT                Body font family
  --font-size SIZE           Body font size
  --heading-font FONT        Heading font family
  --heading-font-size SIZE   Heading font size
  --mono-font FONT           Monospace/code font family
  --page SIZE                Page size (default: a4)
  --margin SIZE              Page margins
  --line-spacing VALUE       Paragraph leading multiplier
  --paragraph-indent SIZE    First-line paragraph indent
  --paragraph-spacing SIZE   Space between paragraphs

Metadata overrides:
  --title TEXT
  --subtitle TEXT
  --author TEXT
  --author-attribution TEXT
  --date TEXT
  --wordcount TEXT
  --version TEXT
  --address TEXT
  --email TEXT
  --website TEXT

Config: script.yaml in source dir or ~/.config/first-folio/script.yaml

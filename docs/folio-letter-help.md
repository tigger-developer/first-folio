<!-- Version: 1.0 | Last updated: 2026-07-06 -->
Usage: folio letter <source.org> [options]

Generate letters from :letter: tagged sections in an org file.
One PDF is generated per recipient.

Options:
  --to RECIPIENT     Generate only for matching recipients (regex substring)
  --dir DIR          Output directory (default: same as source file)
  --prefix PREFIX    Output filename prefix (default: "letter")
  -h, --help         Show this help message

Placeholders:
  Any [keyword] in the letter body is replaced from two sources:

  1. Recipient tags (H4 under :to:) -- highest priority:
       **** The Abbey                :org:       -> [org]
       **** Dave                     :moniker:   -> [moniker]
       **** Friday                   :deadline:  -> [deadline]
     Any tag name works. Define custom tags as needed.

  2. Document front matter -- fallback:
       #+TITLE:    -> [title]      #+AUTHOR:   -> [author]
       #+SUBTITLE: -> [subtitle]   #+DATE:     -> [date]
       #+VERSION:  -> [version]

  Unresolved placeholders pass through as [name].

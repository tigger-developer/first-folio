Usage: folio convert <source> [target] [options]

Convert a play from one format to another.
Format is deduced from file extensions (.org, .md, .fountain, .ftn, .pdf).

When no target file is given, output goes to stdout.
Use --to to specify the output format for stdout mode.

Options:
  --to FORMAT              Output format (org, md, fountain, pdf)
  --style STYLE            Layout style: british (default) or us
  --force                  Force binary output to terminal

PDF options ignored for non-PDF output:
  --font FONT              Body font family
  --font-size SIZE         Body font size
  --margin SIZE            Page margins
  --page SIZE              Page size (a4, letter, etc.)
  --indent SIZE            Dialogue indent depth
  --dialogue-spacing SIZE  Space before dialogue blocks
  --direction-spacing SIZE Space before stage directions
  --[no-]direction-italic  Italicize stage directions
  --[no-]direction-centre  Centre stage directions

Config: script.yaml in source dir or ~/.config/first-folio/script.yaml

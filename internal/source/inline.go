// ABOUTME: Parses nested source-aware inline spans while protecting literal/code contents.
// ABOUTME: Keeps contiguous prose in one text node and bounds source nesting and expansion.
package source

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type parser struct {
	format                            string
	script, literalTypography, breaks bool
	notes                             map[string]string
	resolving                         map[string]bool
	blocks, inlines                   int
}

func (p *parser) addInline(out *[]Node, n Node) error {
	p.inlines++
	if p.inlines > 100000 {
		return fmt.Errorf("native source inline limit exceeded (100000)")
	}
	*out = append(*out, n)
	return nil
}

// ScriptInline shares native recognition while retaining script typography and syntax.
func ScriptInline(format, text string, notes map[string]string) ([]Node, error) {
	p := parser{format: format, script: true, literalTypography: true, notes: notes, resolving: map[string]bool{}}
	return p.inline(text, 0)
}

func (p *parser) inline(text string, depth int) ([]Node, error) {
	if depth > 128 {
		return nil, fmt.Errorf("native inline nesting exceeds 128")
	}
	var out []Node
	var plain strings.Builder
	flush := func() error {
		if plain.Len() == 0 {
			return nil
		}
		value := plain.String()
		plain.Reset()
		if !p.literalTypography {
			value = smart(value)
		}
		return p.addInline(&out, node("Str", value))
	}
	for i := 0; i < len(text); {
		if p.breaks && text[i] == '\n' {
			if err := flush(); err != nil {
				return nil, err
			}
			if err := p.addInline(&out, node("LineBreak", nil)); err != nil {
				return nil, err
			}
			i++
			continue
		}
		if text[i] == '\\' && i+1 < len(text) {
			plain.WriteByte(text[i+1])
			i += 2
			continue
		}
		if text[i] == '`' || (p.format == "org" || p.script) && (text[i] == '=' || text[i] == '~' && !p.script) {
			mark := string(text[i])
			end := i + 1
			if text[i] == '`' {
				for end < len(text) && text[end] == '`' {
					end++
				}
				mark = text[i:end]
			}
			if at := strings.Index(text[end:], mark); at >= 0 {
				if err := flush(); err != nil {
					return nil, err
				}
				if err := p.addInline(&out, node("Code", []any{attr(), text[end : end+at]})); err != nil {
					return nil, err
				}
				i = end + at + len(mark)
				continue
			}
		}
		if !p.script && p.format == "org" && strings.HasPrefix(text[i:], "[[") {
			if end := strings.Index(text[i+2:], "]]"); end >= 0 {
				inside := text[i+2 : i+2+end]
				target, label, hasLabel := strings.Cut(inside, "][")
				if !hasLabel {
					label = target
				}
				if imageTarget(target) {
					return nil, fmt.Errorf("source images are unsupported")
				}
				if err := flush(); err != nil {
					return nil, err
				}
				children, err := p.inline(label, depth+1)
				if err != nil {
					return nil, err
				}
				if err := p.addInline(&out, node("Link", []any{attr(), children, []string{target, ""}})); err != nil {
					return nil, err
				}
				i += end + 4
				continue
			}
		}
		if text[i] == '[' {
			if end := strings.IndexByte(text[i:], ']'); end > 0 {
				label := text[i+1 : i+end]
				id := ""
				if (p.format == "org" || p.script) && strings.HasPrefix(label, "fn:") {
					id = strings.TrimPrefix(label, "fn:")
				}
				if (p.format != "org" || p.script) && strings.HasPrefix(label, "^") {
					id = strings.TrimPrefix(label, "^")
				}
				if note, ok := p.notes[id]; ok && id != "" {
					if p.script {
						if err := flush(); err != nil {
							return nil, err
						}
						if err := p.addInline(&out, node("NoteReference", id)); err != nil {
							return nil, err
						}
						i += end + 1
						continue
					}
					if p.resolving[id] {
						return nil, fmt.Errorf("recursive footnote %q", id)
					}
					if err := flush(); err != nil {
						return nil, err
					}
					p.resolving[id] = true
					children, err := p.inline(note, depth+1)
					delete(p.resolving, id)
					if err != nil {
						return nil, err
					}
					noteBlocks := []Node{}
					if err := p.appendBlock(&noteBlocks, node("Para", children)); err != nil {
						return nil, err
					}
					if err := p.addInline(&out, node("Note", noteBlocks)); err != nil {
						return nil, err
					}
					i += end + 1
					continue
				}
				if !p.script && p.format != "org" && strings.HasPrefix(text[i+end+1:], "(") {
					if close := strings.IndexByte(text[i+end+2:], ')'); close >= 0 {
						if err := flush(); err != nil {
							return nil, err
						}
						children, err := p.inline(label, depth+1)
						if err != nil {
							return nil, err
						}
						target := strings.TrimSpace(text[i+end+2 : i+end+2+close])
						target = strings.SplitN(target, " \"", 2)[0]
						if err := p.addInline(&out, node("Link", []any{attr(), children, []string{target, ""}})); err != nil {
							return nil, err
						}
						i += end + close + 3
						continue
					}
				}
			}
		}
		delimiters := []struct{ mark, kind string }{{"***", "StrongEmph"}, {"**", "Strong"}, {"~~", "Strikeout"}, {"*", "Emph"}}
		if p.format == "org" {
			delimiters = []struct{ mark, kind string }{{"*", "Strong"}, {"/", "Emph"}, {"_", "Underline"}, {"+", "Strikeout"}}
		}
		if p.script {
			star := "Emph"
			if p.format == "org" {
				star = "Strong"
			}
			delimiters = []struct{ mark, kind string }{{"***", "StrongEmph"}, {"**", "Strong"}, {"/", "Emph"}, {"_", "Underline"}, {"*", star}}
		}
		matched := false
		for _, d := range delimiters {
			if !strings.HasPrefix(text[i:], d.mark) || i+len(d.mark) >= len(text) {
				continue
			}
			if p.format == "org" && !p.script && i > 0 {
				r, _ := utf8.DecodeLastRuneInString(text[:i])
				if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '/' {
					continue
				}
			}
			end := findClosing(text, i+len(d.mark), d.mark)
			if end < 0 {
				continue
			}
			if err := flush(); err != nil {
				return nil, err
			}
			children, err := p.inline(text[i+len(d.mark):end], depth+1)
			if err != nil {
				return nil, err
			}
			kind := d.kind
			if kind == "StrongEmph" {
				kind = "Strong"
				children = []Node{node("Emph", children)}
			}
			if err := p.addInline(&out, node(kind, children)); err != nil {
				return nil, err
			}
			i = end + len(d.mark)
			matched = true
			break
		}
		if matched {
			continue
		}
		plain.WriteByte(text[i])
		i++
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []Node{}
	}
	return out, nil
}

func findClosing(text string, start int, mark string) int {
	for start < len(text) {
		at := strings.Index(text[start:], mark)
		if at < 0 {
			return -1
		}
		at += start
		if at > start && text[at-1] != '\\' && !strings.ContainsAny(text[at-1:at], " \t\n") {
			return at
		}
		start = at + len(mark)
	}
	return -1
}

func smart(text string) string {
	text = strings.NewReplacer("---", "—", "--", "–", "...", "…").Replace(text)
	var out strings.Builder
	for i, r := range text {
		if r == '\'' || r == '"' {
			open := i == 0
			if i > 0 {
				before, _ := utf8.DecodeLastRuneInString(text[:i])
				open = unicode.IsSpace(before) || strings.ContainsRune("([{:—–", before)
			}
			if r == '\'' {
				if open {
					r = '‘'
				} else {
					r = '’'
				}
			} else {
				if open {
					r = '“'
				} else {
					r = '”'
				}
			}
		}
		out.WriteRune(r)
	}
	return out.String()
}

func imageTarget(target string) bool {
	lower := strings.ToLower(target)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"} {
		if strings.HasPrefix(lower, "file:") && strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

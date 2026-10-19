// ABOUTME: Retains source-parser state across the composed paragraphs of one letter.
// ABOUTME: Keeps per-recipient resource ownership distinct from letter layout roles.
package source

import "fmt"

// LetterParser owns one recipient's inline parsing state.
type LetterParser struct {
	state parser
	bytes int
}

func NewLetterParser() *LetterParser {
	return &LetterParser{state: parser{format: "org", literalTypography: true, breaks: true, notes: map[string]string{}, resolving: map[string]bool{}}}
}

// Parse consumes one composed paragraph without resetting the letter's budget.
func (p *LetterParser) Parse(text string) ([]Node, error) {
	p.bytes += len(text)
	if p.bytes > 32<<20 {
		return nil, fmt.Errorf("native letter source exceeds 32 MiB")
	}
	p.state.blocks++
	if p.state.blocks > 50000 {
		return nil, fmt.Errorf("native letter block limit exceeded (50000)")
	}
	return p.state.inline(text, 0)
}

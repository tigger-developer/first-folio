// ABOUTME: Owns bounded Pandoc JSON interchange for editable prose adapters.
// ABOUTME: Validates semantic depth/counts and rejects active or unsupported source nodes.
package office

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tigger-developer/first-folio/internal/source"
	"io"
)

type astLimits struct{ bytes, blocks, inlines, depth int }

// Node is Pandoc's tagged JSON node; Content retains its documented array shape.
type Node = source.Node

// Document carries the reader's API version through the office writer.
type Document struct {
	API    []int          `json:"pandoc-api-version"`
	Meta   map[string]any `json:"meta"`
	Blocks []Node         `json:"blocks"`
}

func N(kind string, content any) Node { return Node{Type: kind, Content: content} }

// Decode validates source JSON before allocating the adapter's typed root.
func Decode(raw []byte) (Document, error) {
	var doc Document
	if err := validateAST(raw, astLimits{DataLimit, 50000, 100000, 128}); err != nil {
		return doc, err
	}
	err := json.Unmarshal(raw, &doc)
	return doc, err
}

func validateAST(raw []byte, limits astLimits) error {
	if len(raw) > limits.bytes {
		return fmt.Errorf("Pandoc JSON exceeds byte limit")
	}
	if err := validText(string(raw)); err != nil {
		return err
	}
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&root); err != nil {
		return fmt.Errorf("invalid Pandoc JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing Pandoc JSON content")
	}
	version, ok := root["pandoc-api-version"].([]any)
	if !ok || len(version) < 2 || version[0] != float64(1) || version[1] != float64(23) {
		return fmt.Errorf("incompatible Pandoc JSON API version %v; requires 1.23", root["pandoc-api-version"])
	}
	blocks, ok := root["blocks"].([]any)
	if !ok {
		return fmt.Errorf("Pandoc blocks must be an array")
	}
	if _, ok := root["meta"].(map[string]any); !ok {
		return fmt.Errorf("Pandoc metadata must be an object")
	}
	state := astCounter{limits: limits}
	return state.visit(blocks, 0)
}

type astCounter struct {
	limits          astLimits
	blocks, inlines int
}

var blockKinds = map[string]bool{"Plain": true, "Para": true, "LineBlock": true, "CodeBlock": true, "BlockQuote": true, "OrderedList": true, "BulletList": true, "DefinitionList": true, "Header": true, "HorizontalRule": true, "Table": true, "Div": true, "Null": true}
var inlineKinds = map[string]bool{"Str": true, "Emph": true, "Underline": true, "Strong": true, "Strikeout": true, "Superscript": true, "Subscript": true, "SmallCaps": true, "Quoted": true, "Code": true, "Space": true, "SoftBreak": true, "LineBreak": true, "Link": true, "Note": true, "Span": true}
var valueKinds = map[string]bool{"AlignLeft": true, "AlignRight": true, "AlignCenter": true, "AlignDefault": true, "ColWidth": true, "ColWidthDefault": true, "DefaultStyle": true, "Decimal": true, "LowerRoman": true, "UpperRoman": true, "LowerAlpha": true, "UpperAlpha": true, "DefaultDelim": true, "Period": true, "OneParen": true, "TwoParens": true, "SingleQuote": true, "DoubleQuote": true}

func (s *astCounter) visit(value any, depth int) error {
	switch v := value.(type) {
	case string:
		return validText(v)
	case []any:
		for _, child := range v {
			if err := s.visit(child, depth); err != nil {
				return err
			}
		}
	case map[string]any:
		if kind, ok := v["t"].(string); ok {
			if err := validateAdapterNode(kind, v["c"]); err != nil {
				return err
			}
			switch {
			case blockKinds[kind]:
				s.blocks++
				depth++
			case inlineKinds[kind]:
				s.inlines++
				depth++
			case valueKinds[kind]:
			default:
				return fmt.Errorf("unsupported Pandoc content node %q", kind)
			}
			if s.blocks > s.limits.blocks || s.inlines > s.limits.inlines || depth > s.limits.depth {
				return fmt.Errorf("Pandoc semantic block, inline or nesting limit exceeded")
			}
		}
		for _, child := range v {
			if err := s.visit(child, depth); err != nil {
				return err
			}
		}
	}
	return nil
}

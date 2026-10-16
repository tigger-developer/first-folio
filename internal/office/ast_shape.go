// ABOUTME: Checks Pandoc node shapes consumed by the prose adapters.
// ABOUTME: Rejects malformed subprocess output before typed array access.
package office

import (
	"fmt"
	"math"
)

func validateAdapterNode(kind string, content any) error {
	array, isArray := content.([]any)
	valid := true
	switch kind {
	case "Str":
		_, valid = content.(string)
	case "Header":
		valid = isArray && len(array) == 3
		if valid {
			level, ok := array[0].(float64)
			valid = ok && level >= 1 && level == math.Trunc(level) && pandocAttribute(array[1]) && pandocNodes(array[2])
		}
	case "Code", "CodeBlock":
		valid = isArray && len(array) == 2
		if valid {
			_, valid = array[1].(string)
			valid = valid && pandocAttribute(array[0])
		}
	case "Span", "Div", "Link":
		length := 2
		if kind == "Link" {
			length = 3
		}
		valid = isArray && len(array) == length
		if valid {
			valid = pandocAttribute(array[0]) && pandocNodes(array[1])
			if kind == "Link" {
				valid = valid && pandocStringPair(array[2])
			}
		}
	case "Quoted":
		valid = isArray && len(array) == 2 && pandocNodes(array[1])
	case "Para", "Plain", "Emph", "Strong", "Underline", "Strikeout", "Superscript", "Subscript", "SmallCaps", "Note", "BlockQuote":
		valid = pandocNodes(content)
	}
	if !valid {
		return fmt.Errorf("malformed Pandoc %s content", kind)
	}
	return nil
}

func pandocNodes(value any) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		node, ok := item.(map[string]any)
		if !ok {
			return false
		}
		if kind, ok := node["t"].(string); !ok || kind == "" {
			return false
		}
	}
	return true
}

func pandocAttribute(value any) bool {
	attr, ok := value.([]any)
	if !ok || len(attr) != 3 {
		return false
	}
	if _, ok := attr[0].(string); !ok {
		return false
	}
	classes, ok := attr[1].([]any)
	if !ok {
		return false
	}
	for _, class := range classes {
		if _, ok := class.(string); !ok {
			return false
		}
	}
	pairs, ok := attr[2].([]any)
	if !ok {
		return false
	}
	for _, pair := range pairs {
		if !pandocStringPair(pair) {
			return false
		}
	}
	return true
}

func pandocStringPair(value any) bool {
	pair, ok := value.([]any)
	if !ok || len(pair) != 2 {
		return false
	}
	_, first := pair[0].(string)
	_, second := pair[1].(string)
	return first && second
}

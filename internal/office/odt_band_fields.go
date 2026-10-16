// ABOUTME: Carries changing ODT running text through page-aware native variables.
// ABOUTME: Retains live page fields while avoiding master changes for heading text.
package office

import "fmt"

type odtRunningToken struct{ text, field string }

// Adjacent literal and metadata tokens form a single changing text segment.
// Page and total-page fields remain native fields, not cached strings.
func (p *Prose) odtBandTokens(value string, state Boundary) []odtRunningToken {
	tokens := []odtRunningToken{{}}
	for _, token := range fieldTokens(value) {
		if token.field == "page" || token.field == "total-pages" {
			tokens = append(tokens, odtRunningToken{field: token.field}, odtRunningToken{})
			continue
		}
		text := token.text
		if token.field != "" {
			text = p.runningValue(token.field, state)
		}
		tokens[len(tokens)-1].text += text
	}
	return tokens
}

func odtBandVariable(kind, variant string, index int) string {
	if variant != "-left" {
		variant = "-right"
	}
	return fmt.Sprintf("FolioRunning-%s%s-%d", kind, variant, index)
}

func (p *Prose) odtBandAssignments(state Boundary) []docxNode {
	var fields []docxNode
	for _, item := range []struct {
		kind string
		band Band
	}{{"header", p.Page.Header}, {"footer", p.Page.Footer}} {
		if !item.band.Enabled {
			continue
		}
		for _, variant := range []string{"-left", "-right"} {
			value := item.band.Odd
			if variant == "-left" {
				value = item.band.Even
			}
			if !state.Body {
				value = item.band.FrontOdd
				if variant == "-left" {
					value = item.band.FrontEven
				}
			}
			for index, token := range p.odtBandTokens(value, state) {
				if token.field != "" {
					continue
				}
				fields = append(fields, ox("text:variable-set").attr("text:name", odtBandVariable(item.kind, variant, index)).attr("text:display", "none").attr("office:value-type", "string").attr("office:string-value", token.text))
			}
		}
	}
	return fields
}

func odtVariableDeclarations(flow *docxNode) {
	decls := ox("text:variable-decls")
	seen := map[string]bool{}
	flow.visit(func(node *docxNode) {
		if node.name != "text:variable-set" {
			return
		}
		name := node.get("text:name")
		if !seen[name] {
			seen[name] = true
			decls.children = append(decls.children, ox("text:variable-decl").attr("text:name", name).attr("office:value-type", "string"))
		}
	})
	if len(decls.children) > 0 {
		flow.children = append([]docxNode{decls}, flow.children...)
	}
}

// ABOUTME: Emits typed stage-play documents as Org, Markdown, or Fountain text.
// ABOUTME: Preserves format-specific structure while sharing one semantic model.
package play

import (
	"fmt"
	"sort"
	"strings"
)

func Emit(doc Document, format Format) (Output, error) {
	switch format {
	case FormatOrg:
		return emitOrg(doc), nil
	case FormatMarkdown:
		return emitMarkdown(doc), nil
	case FormatFountain:
		return emitFountain(doc), nil
	default:
		return Output{}, fmt.Errorf("no text emitter for format %q", format)
	}
}

func emitOrg(doc Document) Output {
	lines := []string{"#+SCHEMA: https://github.com/tigger-developer/first-folio/blob/master/schema/script.org"}
	appendOrgMetadata(&lines, doc.Metadata)
	footnotes := map[string]string{}
	for _, event := range doc.Events {
		switch event.Kind {
		case EventFrontMatter:
		case EventActHeader, EventIntroHeader:
			lines = append(lines, "", "* "+event.Text)
		case EventSceneHeader:
			lines = append(lines, "** "+event.Text)
		case EventStageDirection:
			lines = append(lines, "*** "+event.Text)
		case EventCharacter:
			line := "**** " + event.Name
			if event.Direction != "" {
				line += " " + event.Direction
			}
			lines = append(lines, line)
		case EventDialogue, EventIntroText:
			lines = append(lines, event.Text)
		case EventCharacterTableStart:
			lines = append(lines, "", "* "+defaultText(event.Text, "CHARACTERS"))
		case EventCharacterTableRow:
			lines = append(lines, "| "+event.Name+" | "+event.Text+" |")
		case EventPropText:
			lines = append(lines, "*** "+event.Text+strings.Repeat(" ", 40)+":prop:")
		case EventTransition:
			lines = append(lines, "*** "+event.Text+strings.Repeat(" ", 40)+":transition:")
		case EventFootnote:
			footnotes[event.Name] = event.Text
		}
	}
	appendFootnotes(&lines, footnotes, func(name, text string) string { return "[fn:" + name + "] " + text })
	return Output{Text: strings.Join(lines, "\n") + "\n"}
}

func emitMarkdown(doc Document) Output {
	lines := []string{"---", "schema: https://github.com/tigger-developer/first-folio/blob/master/schema/script.md", "---", ""}
	if title := doc.Metadata["title"]; title != "" {
		lines = append(lines, "# "+title)
		if subtitle := doc.Metadata["subtitle"]; subtitle != "" {
			lines = append(lines, "", "**"+subtitle+"**")
		}
		if author := doc.Metadata["author"]; author != "" {
			lines = append(lines, "", "*by "+author+"*")
		}
		var meta []string
		if version := doc.Metadata["version"]; version != "" {
			meta = append(meta, version)
		}
		if date := doc.Metadata["date"]; date != "" {
			meta = append(meta, date)
		}
		if len(meta) > 0 {
			lines = append(lines, "", "--- "+strings.Join(meta, " | ")+" ---")
		}
		lines = append(lines, "")
	}
	footnotes := map[string]string{}
	for _, event := range doc.Events {
		switch event.Kind {
		case EventFrontMatter, EventCharacterTableEnd:
		case EventActHeader, EventIntroHeader:
			appendSeparated(&lines, "## "+event.Text)
		case EventSceneHeader:
			appendSeparated(&lines, "### "+event.Text)
		case EventStageDirection:
			appendSeparated(&lines, "*"+event.Text+"*")
		case EventCharacter:
			line := "**" + event.Name + ":**"
			if event.Direction != "" {
				line += " *(" + event.Direction + ")*"
			}
			appendSeparated(&lines, line)
		case EventDialogue, EventIntroText:
			lines = append(lines, event.Text)
		case EventCharacterTableStart:
			appendSeparated(&lines, "| Character | Description |")
			lines = append(lines, "|-----------|-------------|")
		case EventCharacterTableRow:
			lines = append(lines, "| "+event.Name+" | "+event.Text+" |")
		case EventPropText:
			appendSeparated(&lines, `***"`+event.Text+`"***`)
		case EventTransition:
			appendSeparated(&lines, "> "+event.Text)
		case EventFootnote:
			footnotes[event.Name] = event.Text
		}
	}
	for i := range lines {
		for name := range footnotes {
			lines[i] = strings.ReplaceAll(lines[i], "[fn:"+name+"]", "[^"+name+"]")
		}
	}
	appendFootnotes(&lines, footnotes, func(name, text string) string { return "[^" + name + "]: " + text })
	return Output{Text: strings.Join(lines, "\n") + "\n"}
}

func emitFountain(doc Document) Output {
	var lines []string
	if title := doc.Metadata["title"]; title != "" {
		line := "Title: " + title
		if subtitle := doc.Metadata["subtitle"]; subtitle != "" {
			line += "\n    " + subtitle
		}
		lines = append(lines, line)
	}
	if author := doc.Metadata["author"]; author != "" {
		lines = append(lines, "Author: "+author)
	}
	if version := doc.Metadata["version"]; version != "" {
		lines = append(lines, "Draft date: "+version)
	}
	if date := doc.Metadata["date"]; date != "" {
		lines = append(lines, "Date: "+date)
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	footnotes := map[string]string{}
	var warnings []string
	for _, event := range doc.Events {
		switch event.Kind {
		case EventFrontMatter, EventCharacterTableEnd:
		case EventActHeader:
			appendSeparated(&lines, "===", "", "> **"+strings.ToUpper(event.Text)+"** <")
		case EventIntroHeader:
			appendSeparated(&lines, "> **"+strings.ToUpper(event.Text)+"** <")
		case EventSceneHeader:
			text := strings.ToUpper(event.Text)
			if !isSceneHeading(text) {
				text = "." + text
			}
			appendSeparated(&lines, text)
		case EventStageDirection:
			appendSeparated(&lines, event.Text)
		case EventCharacter:
			appendSeparated(&lines, strings.ToUpper(event.Name))
			if event.Direction != "" {
				lines = append(lines, "("+event.Direction+")")
			}
		case EventDialogue:
			lines = append(lines, event.Text)
		case EventIntroText:
			lines = append(lines, strings.TrimPrefix(event.Text, "- "))
		case EventCharacterTableStart:
			warnings = append(warnings, "warning: character table has no Fountain equivalent, rendering as Action text")
			appendSeparated(&lines, "> **"+strings.ToUpper(defaultText(event.Text, "Characters"))+"** <")
		case EventCharacterTableRow:
			appendSeparated(&lines, event.Name+" - "+event.Text)
		case EventPropText:
			appendSeparated(&lines, ">"+event.Text+"<")
		case EventTransition:
			appendSeparated(&lines, strings.ToUpper(event.Text))
		case EventFootnote:
			footnotes[event.Name] = event.Text
		}
	}
	appendFootnotes(&lines, footnotes, func(_ string, text string) string { return "[[" + text + "]]" })
	return Output{Text: strings.Join(lines, "\n") + "\n", Warnings: warnings}
}

func appendOrgMetadata(lines *[]string, metadata map[string]string) {
	for _, key := range []string{"title", "subtitle", "author", "date", "version"} {
		if value := metadata[key]; value != "" {
			*lines = append(*lines, "#+"+strings.ToUpper(key)+": "+value)
		}
	}
}

func appendFootnotes(lines *[]string, footnotes map[string]string, format func(string, string) string) {
	if len(footnotes) == 0 {
		return
	}
	*lines = append(*lines, "")
	keys := make([]string, 0, len(footnotes))
	for key := range footnotes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		*lines = append(*lines, format(key, footnotes[key]))
	}
}

func appendSeparated(lines *[]string, values ...string) {
	if len(*lines) > 0 && (*lines)[len(*lines)-1] != "" {
		*lines = append(*lines, "")
	}
	*lines = append(*lines, values...)
}

func defaultText(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

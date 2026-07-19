// ABOUTME: Renders typed stage-play events through the file-backed Typst template.
// ABOUTME: Validates layout literals, escapes content contexts, and invokes Typst directly.
package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	folio "github.com/tigger-developer/first-folio"
	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
	typstutil "github.com/tigger-developer/first-folio/internal/typst"
)

var (
	dimensionRE = regexp.MustCompile(`^(?:0|-?(?:\d+(?:\.\d+)?)(?:pt|mm|cm|in|em|%))$`)
	weightRE    = regexp.MustCompile(`^(?:[1-9]00|thin|extralight|light|regular|medium|semibold|bold|extrabold|black)$`)
	alignRE     = regexp.MustCompile(`^(?:left|center|right)$`)
	pageRE      = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

	boldItalicRE = regexp.MustCompile(`\*\*\*([^*\n]+)\*\*\*`)
	boldRE       = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	codeTickRE   = regexp.MustCompile("`([^`\\n]+)`")
	codeOrgRE    = regexp.MustCompile(`=([^=\n]+)=`)
	orgItalicRE  = regexp.MustCompile(`/([^/\n]+?)/`)
	underlineRE  = regexp.MustCompile(`_([^_\n]+)_`)
	mdItalicRE   = regexp.MustCompile(`\*([^*\n]+)\*`)
)

type scriptTemplateData struct {
	Page, Margin, Font, FontSize                     string
	GlobalWeight, GlobalStretch                      string
	DialogueSameLine                                 bool
	SpeechSpace, DialogueIndent, DialogueWrap        string
	SpeakerWeight, SpeakerContent, SpeakerAlign      string
	SpeakerIndent                                    string
	InstructionOpen, InstructionClose                string
	InstructionPrefix, InstructionSuffix             string
	InstructionAlign                                 string
	DirectionSpace, DirectionAlign, DirectionIndent  string
	DirectionOpen, DirectionClose                    string
	TransitionIndent                                 string
	ActSpace, ActAlign, ActFont, ActSize             string
	ActWeight, ActOpen, ActClose                     string
	SceneSpace, SceneAfter, SceneAlign               string
	SceneFont, SceneSize, SceneWeight                string
	SceneOpen, SceneClose                            string
	FrontmatterSpace, FrontmatterAlign               string
	FrontmatterSize, FrontmatterWeight               string
	HasTitle, HasSubtitle, HasAuthor                 bool
	TitlePageNumber                                  bool
	Title, Subtitle, Author, AuthorPrefixInline      string
	AuthorPrefixLines                                []string
	TitleAlign, TitleOffset, TitleFont, TitleSize    string
	TitleWeight, TitleStyle                          string
	SubtitleSpace, SubtitleFont, SubtitleSize        string
	SubtitleStyle                                    string
	AuthorSpace, AuthorFont, AuthorSize, AuthorStyle string
	FooterLeft, FooterRight                          string
	Body                                             string
}

func renderPlayDocument(doc play.Document, cfg config.Config, target string, toStdout bool, force bool, stdout io.Writer) error {
	source, err := renderPlayTypst(doc, cfg)
	if err != nil {
		return err
	}
	if target != "" && strings.HasSuffix(strings.ToLower(target), ".typ") {
		if err := os.WriteFile(target, []byte(source), 0o644); err != nil {
			return fmt.Errorf("cannot write to %s: %w", target, err)
		}
		return nil
	}
	if toStdout && writerIsTerminal(stdout) && !force {
		return fmt.Errorf("PDF is binary output. Refusing to write to terminal\nUse a target file, redirect stdout, or pass --force to override")
	}
	return compilePlayTypst(source, target, toStdout, stdout)
}

func renderPlayTypst(doc play.Document, cfg config.Config) (string, error) {
	data, err := newScriptTemplateData(doc, cfg)
	if err != nil {
		return "", err
	}
	raw, err := folio.Assets.ReadFile("templates/script.typ")
	if err != nil {
		return "", fmt.Errorf("loading script Typst template: %w", err)
	}
	tmpl, err := template.New("script.typ").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing script Typst template: %w", err)
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("executing script Typst template: %w", err)
	}
	return output.String(), nil
}

func newScriptTemplateData(doc play.Document, cfg config.Config) (scriptTemplateData, error) {
	font := cfg.String("folio.font", "Libertinus Serif")
	fontSize := cfg.String("folio.font-size", "12pt")
	data := scriptTemplateData{
		Page: cfg.String("folio.page", "a4"), Margin: cfg.String("folio.margin", "25mm"), Font: escapeTypstString(font), FontSize: fontSize,
		SpeechSpace:       cfg.String("folio.positioning.speech.space-before", "1.6em"),
		DialogueIndent:    typstDimension(cfg.String("folio.positioning.speech.dialogue.indent", "0")),
		DialogueWrap:      cfg.String("folio.positioning.speech.dialogue.wrap-indent", "7em"),
		DialogueSameLine:  cfg.String("folio.positioning.speech.dialogue.placement", "same-line") == "same-line",
		SpeakerWeight:     boolWeight(cfg.Bool("folio.positioning.speech.speaker.bold", true)),
		SpeakerAlign:      validAlign(cfg.String("folio.positioning.speech.speaker.align", "left")),
		SpeakerIndent:     typstDimension(cfg.String("folio.positioning.speech.speaker.indent", "0")),
		InstructionPrefix: instructionDelimiter(cfg.String("folio.positioning.speech.speech-instruction.prefix", "(")),
		InstructionSuffix: instructionDelimiter(cfg.String("folio.positioning.speech.speech-instruction.suffix", ")")),
		InstructionAlign:  validAlign(cfg.String("folio.positioning.speech.speech-instruction.align", "left")),
		DirectionSpace:    cfg.String("folio.positioning.stage-direction.space-before", "1.6em"),
		DirectionAlign:    validAlign(cfg.String("folio.positioning.stage-direction.align", "left")),
		DirectionIndent:   typstDimension(cfg.String("folio.positioning.stage-direction.indent", "0")),
		TransitionIndent:  typstDimension(cfg.String("folio.positioning.transition.indent", "0")),
		ActSpace:          cfg.String("folio.positioning.act-header.space-before", "0em"),
		ActAlign:          validAlign(cfg.String("folio.positioning.act-header.align", "center")),
		ActFont:           escapeTypstString(cfg.InheritedString("folio.positioning.act-header", "font", font)),
		ActSize:           cfg.String("folio.positioning.act-header.font-size", "14pt"),
		ActWeight:         boolWeight(cfg.Bool("folio.positioning.act-header.bold", true)),
		SceneSpace:        cfg.String("folio.positioning.scene-header.space-before", "2em"),
		SceneAfter:        cfg.String("folio.positioning.scene-header.space-after", "0.5em"),
		SceneAlign:        validAlign(cfg.String("folio.positioning.scene-header.align", "left")),
		SceneFont:         escapeTypstString(cfg.InheritedString("folio.positioning.scene-header", "font", font)),
		SceneSize:         cfg.String("folio.positioning.scene-header.font-size", "12pt"),
		SceneWeight:       boolWeight(cfg.Bool("folio.positioning.scene-header.bold", true)),
		FrontmatterSpace:  cfg.String("folio.positioning.frontmatter.header.space-before", "2em"),
		FrontmatterAlign:  validAlign(cfg.String("folio.positioning.frontmatter.header.align", "left")),
		FrontmatterSize:   cfg.String("folio.positioning.frontmatter.header.font-size", "14pt"),
		FrontmatterWeight: boolWeight(cfg.Bool("folio.positioning.frontmatter.header.bold", true)),
		TitlePageNumber:   cfg.Bool("folio.title-page.page-number", false),
		Title:             escapeTypstContent(doc.Metadata["title"]), Subtitle: escapeTypstContent(doc.Metadata["subtitle"]), Author: escapeTypstContent(doc.Metadata["author"]),
		TitleAlign: validAlign(cfg.String("folio.title-page.title.align", "center")), TitleFont: escapeTypstString(cfg.InheritedString("folio.title-page.title", "font", font)),
		TitleSize: cfg.String("folio.title-page.title.font-size", "24pt"), TitleWeight: boolWeight(cfg.Bool("folio.title-page.title.bold", true)), TitleStyle: boolStyle(cfg.Bool("folio.title-page.title.italic", false)),
		SubtitleSpace: cfg.String("folio.title-page.subtitle.space-before", "1em"), SubtitleFont: escapeTypstString(cfg.InheritedString("folio.title-page.subtitle", "font", font)),
		SubtitleSize: cfg.String("folio.title-page.subtitle.font-size", "14pt"), SubtitleStyle: boolStyle(cfg.Bool("folio.title-page.subtitle.italic", true)),
		AuthorSpace: cfg.String("folio.title-page.author.space-before", "2em"), AuthorFont: escapeTypstString(cfg.InheritedString("folio.title-page.author", "font", font)),
		AuthorSize: cfg.String("folio.title-page.author.font-size", "12pt"), AuthorStyle: boolStyle(cfg.Bool("folio.title-page.author.italic", false)),
	}
	data.AuthorPrefixInline, data.AuthorPrefixLines = authorPrefix(cfg.String("folio.title-page.author.prefix", ""))
	data.HasTitle, data.HasSubtitle, data.HasAuthor = data.Title != "", data.Subtitle != "", data.Author != ""
	if cfg.Bool("folio.positioning.stage-direction.italic", true) {
		data.DirectionOpen, data.DirectionClose = "_", " _"
	}
	if cfg.Bool("folio.positioning.speech.speech-instruction.italic", true) {
		data.InstructionOpen, data.InstructionClose = "_", "_"
	}
	data.SpeakerContent = caseExpression(cfg.String("folio.positioning.speech.speaker.case-transform", "upper"), escapeTypstContent(cfg.String("folio.positioning.speech.speaker.prefix", ""))+"#name"+escapeTypstContent(cfg.String("folio.positioning.speech.speaker.suffix", ":")))
	data.ActOpen, data.ActClose = caseDelimiters(cfg.String("folio.positioning.act-header.case-transform", "as-written"))
	data.SceneOpen, data.SceneClose = caseDelimiters(cfg.String("folio.positioning.scene-header.case-transform", "as-written"))
	data.TitleOffset = "30%"
	if cfg.String("folio.title-page.title.position", "third") != "third" {
		data.TitleOffset = "40%"
	}
	data.GlobalWeight = optionalWeight(cfg.String("folio.font-weight", ""))
	data.GlobalStretch = optionalStretch(cfg.String("folio.font-stretch", ""))
	data.FooterLeft, data.FooterRight = titleFooter(doc, cfg)
	data.Body = renderPlayBody(doc, cfg)
	if err := validateScriptData(data); err != nil {
		return scriptTemplateData{}, err
	}
	return data, nil
}

func authorPrefix(value string) (string, []string) {
	if !strings.Contains(value, "\n") {
		return escapeTypstContent(value), nil
	}
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, escapeTypstContent(line))
		}
	}
	return "", lines
}

func typstDimension(value string) string {
	if strings.TrimSpace(value) == "0" {
		return "0pt"
	}
	return value
}

func renderPlayBody(doc play.Document, cfg config.Config) string {
	footnotes := map[string]string{}
	for _, event := range doc.Events {
		if event.Kind == play.EventFootnote {
			footnotes[event.Name] = event.Text
		}
	}
	var lines []string
	for i := 0; i < len(doc.Events); i++ {
		event := doc.Events[i]
		switch event.Kind {
		case play.EventFrontMatter, play.EventFootnote, play.EventCharacterTableEnd:
		case play.EventActHeader:
			if cfg.Bool("folio.positioning.act-header.page-break-before", true) {
				lines = append(lines, "#pagebreak()")
			}
			lines = append(lines, "#act-header["+inlineTypst(event.Text, footnotes)+"]")
		case play.EventSceneHeader:
			lines = append(lines, "#scene-header["+inlineTypst(event.Text, footnotes)+"]")
		case play.EventIntroHeader:
			lines = append(lines, "#frontmatter-header["+inlineTypst(event.Text, footnotes)+"]")
		case play.EventIntroText:
			lines = append(lines, inlineTypst(event.Text, footnotes), "")
		case play.EventStageDirection:
			lines = append(lines, "#stage-direction["+inlineTypst(event.Text, footnotes)+"]")
		case play.EventCharacter:
			name := escapeTypstString(event.Name)
			direction := ""
			if event.Direction != "" {
				direction = `, direction: "` + escapeTypstString(event.Direction) + `"`
			}
			var dialogue []string
			for i+1 < len(doc.Events) && doc.Events[i+1].Kind == play.EventDialogue {
				i++
				dialogue = append(dialogue, inlineTypst(doc.Events[i].Text, footnotes))
			}
			lines = append(lines, "#dialogue(\""+name+"\""+direction+")["+strings.Join(dialogue, " \\\n")+"]")
		case play.EventDialogue:
			lines = append(lines, inlineTypst(event.Text, footnotes), "")
		case play.EventPropText:
			lines = append(lines, "#prop-text["+inlineTypst(event.Text, footnotes)+"]")
		case play.EventTransition:
			align := validAlign(cfg.String("folio.positioning.transition.align", "right"))
			content := caseExpression(cfg.String("folio.positioning.transition.case-transform", "upper"), inlineTypst(event.Text, footnotes))
			indent := typstDimension(cfg.String("folio.positioning.transition.indent", "0"))
			lines = append(lines, "#v("+cfg.String("folio.positioning.transition.space-before", "1.6em")+")", "#pad(left: "+indent+")[#align("+align+")["+content+"]]")
		case play.EventCharacterTableStart:
			var rows []string
			for i+1 < len(doc.Events) && doc.Events[i+1].Kind == play.EventCharacterTableRow {
				i++
				row := doc.Events[i]
				rows = append(rows, "["+inlineTypst(row.Name, footnotes)+"], ["+inlineTypst(row.Text, footnotes)+"],")
			}
			lines = append(lines, "#frontmatter-header["+inlineTypst(defaultString(event.Text, "Characters"), footnotes)+"]", "#table(columns: (30%, 1fr), stroke: none,", strings.Join(rows, "\n"), ")")
		}
	}
	return strings.Join(lines, "\n")
}

func inlineTypst(text string, footnotes map[string]string) string {
	tokens := []string{}
	protect := func(value string) string {
		tokens = append(tokens, value)
		return fmt.Sprintf("\x00%d\x00", len(tokens)-1)
	}
	keys := make([]string, 0, len(footnotes))
	for key := range footnotes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		note := "#footnote[" + inlineTypst(footnotes[key], nil) + "]"
		text = strings.ReplaceAll(text, "[fn:"+key+"]", protect(note))
		text = strings.ReplaceAll(text, "[^"+key+"]", protect(note))
	}
	text = replaceMarkup(text, boldItalicRE, func(s string) string { return "*_" + escapeTypstContent(s) + "_*" }, protect)
	text = replaceMarkup(text, boldRE, func(s string) string { return "*" + escapeTypstContent(s) + "*" }, protect)
	text = replaceMarkup(text, codeTickRE, func(s string) string { return `#raw("` + escapeTypstString(s) + `")` }, protect)
	text = replaceMarkup(text, codeOrgRE, func(s string) string { return `#raw("` + escapeTypstString(s) + `")` }, protect)
	text = replaceMarkup(text, orgItalicRE, func(s string) string { return "_" + escapeTypstContent(s) + "_" }, protect)
	text = replaceMarkup(text, underlineRE, func(s string) string { return "#underline[" + escapeTypstContent(s) + "]" }, protect)
	text = replaceMarkup(text, mdItalicRE, func(s string) string { return "_" + escapeTypstContent(s) + "_" }, protect)
	text = escapeTypstContent(text)
	for i, token := range tokens {
		text = strings.ReplaceAll(text, fmt.Sprintf("\x00%d\x00", i), token)
	}
	return text
}

func replaceMarkup(text string, expression *regexp.Regexp, render func(string) string, protect func(string) string) string {
	return expression.ReplaceAllStringFunc(text, func(match string) string {
		parts := expression.FindStringSubmatch(match)
		return protect(render(parts[1]))
	})
}

func escapeTypstContent(value string) string {
	return typstutil.EscapeContent(value)
}

func escapeTypstString(value string) string {
	return typstutil.EscapeString(value)
}

func validateScriptData(data scriptTemplateData) error {
	for name, value := range map[string]string{
		"font-size": data.FontSize, "margin": data.Margin, "speech spacing": data.SpeechSpace, "dialogue indent": data.DialogueIndent, "dialogue wrap": data.DialogueWrap,
		"speaker indent": data.SpeakerIndent, "direction spacing": data.DirectionSpace, "direction indent": data.DirectionIndent, "transition indent": data.TransitionIndent,
		"act spacing": data.ActSpace, "act size": data.ActSize,
		"scene spacing": data.SceneSpace, "scene after": data.SceneAfter, "scene size": data.SceneSize,
		"frontmatter spacing": data.FrontmatterSpace, "frontmatter size": data.FrontmatterSize,
		"title size": data.TitleSize, "subtitle spacing": data.SubtitleSpace, "subtitle size": data.SubtitleSize,
		"author spacing": data.AuthorSpace, "author size": data.AuthorSize,
	} {
		if !dimensionRE.MatchString(value) {
			return fmt.Errorf("invalid %s value %q", name, value)
		}
	}
	if !pageRE.MatchString(data.Page) {
		return fmt.Errorf("invalid page value %q", data.Page)
	}
	return nil
}

func compilePlayTypst(source string, target string, toStdout bool, stdout io.Writer) error {
	typstFile, err := os.CreateTemp("", "folio-script-*.typ")
	if err != nil {
		return fmt.Errorf("creating temporary Typst source: %w", err)
	}
	typstPath := typstFile.Name()
	defer os.Remove(typstPath)
	if _, err := typstFile.WriteString(source); err != nil {
		typstFile.Close()
		return fmt.Errorf("writing temporary Typst source: %w", err)
	}
	if err := typstFile.Close(); err != nil {
		return fmt.Errorf("closing temporary Typst source: %w", err)
	}
	outputPath := target
	if toStdout {
		pdfFile, err := os.CreateTemp("", "folio-script-*.pdf")
		if err != nil {
			return fmt.Errorf("creating temporary PDF: %w", err)
		}
		outputPath = pdfFile.Name()
		pdfFile.Close()
		defer os.Remove(outputPath)
	}
	cmd := exec.Command("typst", "compile", typstPath, outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("typst compile failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if toStdout {
		raw, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("reading temporary PDF: %w", err)
		}
		_, err = stdout.Write(raw)
		return err
	}
	return nil
}

func writerIsTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func titleFooter(doc play.Document, cfg config.Config) (string, string) {
	slots := map[string][]string{}
	for _, item := range []struct{ value, position, size string }{
		{doc.Metadata["version"], cfg.String("folio.title-page.version.position", "bottom-right"), cfg.String("folio.title-page.version.font-size", "10pt")},
		{doc.Metadata["date"], cfg.String("folio.title-page.date.position", "bottom-left"), cfg.String("folio.title-page.date.font-size", "10pt")},
	} {
		if item.value != "" {
			slots[item.position] = append(slots[item.position], "#text(size: "+item.size+")["+escapeTypstContent(item.value)+"]")
		}
	}
	return strings.Join(slots["bottom-left"], " #linebreak() "), strings.Join(slots["bottom-right"], " #linebreak() ")
}

func optionalWeight(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	if !weightRE.MatchString(value) {
		return ""
	}
	if _, err := strconv.Atoi(value); err == nil {
		return ", weight: " + value
	}
	return `, weight: "` + value + `"`
}

func optionalStretch(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasSuffix(value, "%") {
		value += "%"
	}
	if !dimensionRE.MatchString(value) {
		return ""
	}
	return ", stretch: " + value
}

func caseExpression(transform string, content string) string {
	open, close := caseDelimiters(transform)
	return open + content + close
}

func caseDelimiters(transform string) (string, string) {
	switch transform {
	case "upper":
		return "#upper[", "]"
	case "lower":
		return "#lower[", "]"
	case "small-caps":
		return "#smallcaps[", "]"
	default:
		return "", ""
	}
}

func boolWeight(value bool) string {
	if value {
		return "bold"
	}
	return "regular"
}

func boolStyle(value bool) string {
	if value {
		return "italic"
	}
	return "normal"
}

func validAlign(value string) string {
	if alignRE.MatchString(value) {
		return value
	}
	return "left"
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func instructionDelimiter(value string) string {
	if strings.ContainsAny(value, `\#$@*_`) {
		return escapeTypstContent(value)
	}
	return value
}

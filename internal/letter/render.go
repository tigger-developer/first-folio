// ABOUTME: Renders cover letters through a file-backed Typst template.
// ABOUTME: Converts Org inline markup, formats addresses/dates, and compiles PDFs directly.
package letter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	folio "github.com/tigger-developer/first-folio"
	"github.com/tigger-developer/first-folio/internal/config"
	typstutil "github.com/tigger-developer/first-folio/internal/typst"
)

type templateData struct {
	Page, Font, FontSize, Weight, Stretch                                    string
	MarginTop, MarginBottom, MarginLeft, MarginRight                         string
	SpaceBeforeClosing, SpaceBeforeSignoff                                   string
	SpaceAfterSender, SpaceAfterRecipient, SpaceAfterDate, SpaceAfterSubject string
	Sender, Contact, Recipient, Date, Subject, Body, Closing, Signoff        string
}

var (
	letterCodeRE      = regexp.MustCompile(`=([^=\n]+)=`)
	letterBoldRE      = regexp.MustCompile(`\*([^*\n]+)\*`)
	letterItalicRE    = regexp.MustCompile(`/([^/\n]+?)/`)
	letterUnderlineRE = regexp.MustCompile(`_([^_\n]+)_`)
)

func RenderTypst(letter Letter, cfg config.Config) (string, error) {
	data := templateData{
		Page: cfg.String("folio.letter.page", "a4"), Font: escapeString(cfg.String("folio.letter.font", "Libertinus Serif")), FontSize: cfg.String("folio.letter.font-size", "11pt"),
		MarginTop: cfg.String("folio.letter.margin-top", "25mm"), MarginBottom: cfg.String("folio.letter.margin-bottom", "25mm"), MarginLeft: cfg.String("folio.letter.margin-left", "30mm"), MarginRight: cfg.String("folio.letter.margin-right", "25mm"),
		SpaceBeforeClosing: cfg.String("folio.letter.space-before-closing", "1.2em"), SpaceBeforeSignoff: cfg.String("folio.letter.space-before-signoff", "1.5em"),
		SpaceAfterSender: cfg.String("folio.letter.space-after-sender", "2em"), SpaceAfterRecipient: cfg.String("folio.letter.space-after-recipient", "1em"), SpaceAfterDate: cfg.String("folio.letter.space-after-date", "1em"), SpaceAfterSubject: cfg.String("folio.letter.space-after-subject", "0.5em"),
		Sender: formatAddress(letter.Sender), Recipient: formatAddress(letter.Recipient), Date: escapeContent(formatDate(letter.Date)), Subject: inlineOrg(letter.Subject), Body: bodyMarkup(letter.Body), Closing: inlineOrg(letter.Closing), Signoff: inlineOrg(letter.Signoff),
	}
	if letter.Email != "" {
		data.Contact = escapeContent(letter.Email)
	}
	if letter.Contact != "" {
		if data.Contact != "" {
			data.Contact += " \\\n"
		}
		data.Contact += escapeContent(letter.Contact)
	}
	data.Weight = optionalLetterWeight(cfg.String("folio.letter.font-weight", ""))
	data.Stretch = optionalLetterStretch(cfg.String("folio.letter.font-stretch", ""))
	raw, err := folio.Assets.ReadFile("templates/letter.typ")
	if err != nil {
		return "", fmt.Errorf("loading letter template: %w", err)
	}
	tmpl, err := template.New("letter.typ").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing letter template: %w", err)
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("executing letter template: %w", err)
	}
	return output.String(), nil
}

func CompilePDF(source string, output string) error {
	tmp, err := os.CreateTemp("", "cover-*.typ")
	if err != nil {
		return fmt.Errorf("creating temporary Typst source: %w", err)
	}
	path := tmp.Name()
	keep := os.Getenv("FOLIO_KEEP_TYPST") != ""
	if !keep {
		defer os.Remove(path)
	}
	if _, err := tmp.WriteString(source); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temporary Typst source: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temporary Typst source: %w", err)
	}
	cmd := exec.Command("typst", "compile", path, output)
	if result, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("typst compile failed: %w: %s", err, strings.TrimSpace(string(result)))
	}
	if keep {
		fmt.Fprintf(os.Stderr, "Typst source kept at: %s\n", path)
	}
	return nil
}

func bodyMarkup(body string) string {
	paragraphs := regexp.MustCompile(`\n\n+`).Split(strings.TrimSpace(body), -1)
	for i := range paragraphs {
		paragraphs[i] = inlineOrg(paragraphs[i])
	}
	return strings.Join(paragraphs, "\n\n")
}

func inlineOrg(value string) string {
	tokens := []string{}
	protect := func(markup string) string {
		tokens = append(tokens, markup)
		return fmt.Sprintf("\x00%d\x00", len(tokens)-1)
	}
	for _, item := range []struct {
		re *regexp.Regexp
		fn func(string) string
	}{
		{letterCodeRE, func(s string) string { return `#text(font: "Libertinus Mono")[` + escapeContent(s) + `]` }},
		{letterBoldRE, func(s string) string { return "*" + escapeContent(s) + "*" }},
		{letterItalicRE, func(s string) string { return "_" + escapeContent(s) + "_" }},
		{letterUnderlineRE, func(s string) string { return "#underline[" + escapeContent(s) + "]" }},
	} {
		value = item.re.ReplaceAllStringFunc(value, func(match string) string {
			return protect(item.fn(item.re.FindStringSubmatch(match)[1]))
		})
	}
	value = escapeContent(value)
	for i, token := range tokens {
		value = strings.ReplaceAll(value, fmt.Sprintf("\x00%d\x00", i), token)
	}
	return value
}

func formatDate(value string) string {
	if value == "" {
		return time.Now().Format("2 January 2006")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("2 January 2006")
}

func formatAddress(value string) string {
	parts := strings.Split(value, " / ")
	for i := range parts {
		parts[i] = inlineOrg(parts[i])
	}
	return strings.Join(parts, " \\\n")
}

func escapeContent(value string) string {
	return typstutil.EscapeContent(value)
}

func escapeString(value string) string {
	return typstutil.EscapeString(value)
}

func optionalLetterWeight(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return `, weight: "` + escapeString(strings.ToLower(strings.TrimSpace(value))) + `"`
}

func optionalLetterStretch(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasSuffix(value, "%") {
		value += "%"
	}
	return ", stretch: " + value
}

// ABOUTME: Projects substituted cover letters into editable Pandoc prose paragraphs.
// ABOUTME: Reuses letter syntax and date contracts while resolving letter-only layout.
package letter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/office"
)

// Office builds native prose without first rendering or parsing Typst.
func Office(item Letter, cfg config.Config) (office.Prose, error) {
	raw, _, err := office.Pandoc(nil, "--from=markdown", "--to=json")
	if err != nil {
		return office.Prose{}, err
	}
	doc, err := office.Decode(raw)
	if err != nil {
		return office.Prose{}, err
	}
	font, err := cfg.Font("folio.letter.font")
	if err != nil {
		return office.Prose{}, err
	}
	base, err := office.ResolveStyle(font, 12)
	if err != nil {
		return office.Prose{}, err
	}
	cap := office.CapHeight(base)
	base.Line = cap + base.Size*0.8
	base.After = base.Size * 0.4
	plan := office.Prose{Document: doc, Styles: map[string]office.Style{}, Metadata: map[string]string{"title": item.Subject, "author": item.Author}}
	page := &plan.Page
	page.Width, page.Height, err = office.Paper(cfg.String("folio.letter.page", "a4"))
	if err != nil {
		return plan, err
	}
	for _, margin := range []struct {
		key      string
		target   *float64
		fallback string
	}{{"top", &page.Top, "25mm"}, {"bottom", &page.Bottom, "25mm"}, {"left", &page.Left, "30mm"}, {"right", &page.Right, "25mm"}} {
		*margin.target, err = office.Points(cfg.String("folio.letter.margin-"+margin.key, margin.fallback), base.Size, 0)
		if err != nil {
			return plan, err
		}
	}
	plan.Styles["FolioBody"] = base
	literal := base
	literal.Font.Family = "Libertinus Mono"
	literal.Character = true
	plan.Styles["FolioLiteral"] = literal
	add := func(role, value string, bold bool, gap string, fallback string) error {
		style := base
		style.Bold = bold || base.Bold
		if gap != "" {
			extra, e := office.Points(cfg.String("folio.letter."+gap, fallback), base.Size, 0)
			if e != nil {
				return e
			}
			style.After += extra
		}
		plan.Styles[role] = style
		plan.Document.Blocks = append(plan.Document.Blocks, office.Paragraph(role, officeInlines(value)))
		return nil
	}
	sender := strings.ReplaceAll(item.Sender, " / ", "\n")
	if item.Email != "" {
		sender += "\n" + item.Email
	}
	if item.Contact != "" {
		sender += "\n" + item.Contact
	}
	if err = add("FolioSender", sender, false, "space-after-sender", "2em"); err != nil {
		return plan, err
	}
	senderStyle := plan.Styles["FolioSender"]
	senderStyle.Left = max(0, page.Width-page.Left-page.Right-office.TextWidth(sender, senderStyle))
	plan.Styles["FolioSender"] = senderStyle
	if err = add("FolioRecipient", strings.ReplaceAll(item.Recipient, " / ", "\n"), false, "space-after-recipient", "1em"); err != nil {
		return plan, err
	}
	if err = add("FolioDate", formatDate(item.Date), false, "space-after-date", "1em"); err != nil {
		return plan, err
	}
	if err = add("FolioSubject", item.Subject, true, "space-after-subject", "0.5em"); err != nil {
		return plan, err
	}
	paragraphs := regexp.MustCompile(`\n\n+`).Split(strings.TrimSpace(item.Body), -1)
	for i, paragraph := range paragraphs {
		role := fmt.Sprintf("FolioLetterBody%d", i)
		gap, fallback := "", ""
		if i == len(paragraphs)-1 {
			gap, fallback = "space-before-closing", "1.2em"
		}
		if err = add(role, paragraph, false, gap, fallback); err != nil {
			return plan, err
		}
	}
	if err = add("FolioClosing", item.Closing+",", false, "space-before-signoff", "1.5em"); err != nil {
		return plan, err
	}
	if err = add("FolioSignoff", item.Signoff, false, "", ""); err != nil {
		return plan, err
	}
	return plan, nil
}

func officeInlines(value string) []office.Node {
	var tokens []office.Node
	for _, rule := range []struct {
		re   *regexp.Regexp
		kind string
	}{{letterCodeRE, "Code"}, {letterBoldRE, "Strong"}, {letterItalicRE, "Emph"}, {letterUnderlineRE, "Underline"}} {
		value = rule.re.ReplaceAllStringFunc(value, func(match string) string {
			text := rule.re.FindStringSubmatch(match)[1]
			var n office.Node
			if rule.kind == "Code" {
				n = office.Span("FolioLiteral", office.Text(text))
			} else {
				n = office.N(rule.kind, office.Text(text))
			}
			tokens = append(tokens, n)
			return "\x00" + strconv.Itoa(len(tokens)-1) + "\x00"
		})
	}
	var result []office.Node
	for len(value) > 0 {
		at := strings.IndexByte(value, 0)
		if at < 0 {
			result = append(result, literalOfficeInlines(value)...)
			break
		}
		result = append(result, literalOfficeInlines(value[:at])...)
		value = value[at+1:]
		end := strings.IndexByte(value, 0)
		if end < 0 {
			break
		}
		index, err := strconv.Atoi(value[:end])
		if err == nil && index < len(tokens) {
			result = append(result, tokens[index])
		}
		value = value[end+1:]
	}
	return result
}
func literalOfficeInlines(value string) []office.Node {
	var result []office.Node
	for i, line := range strings.Split(value, "\n") {
		if i > 0 {
			result = append(result, office.N("LineBreak", nil))
		}
		if line != "" {
			result = append(result, office.N("Str", line))
		}
	}
	return result
}

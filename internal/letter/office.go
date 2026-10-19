// ABOUTME: Projects substituted cover letters into editable Pandoc prose paragraphs.
// ABOUTME: Reuses letter syntax and date contracts while resolving letter-only layout.
package letter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/office"
	"github.com/tigger-developer/first-folio/internal/source"
)

// Office builds native prose without first rendering or parsing Typst.
func Office(item Letter, cfg config.Config) (office.Prose, error) {
	doc := office.Document{API: []int{1, 23, 1, 1}, Meta: map[string]any{}, Blocks: []office.Node{}}
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
		inlines, err := officeInlines(value)
		if err != nil {
			return err
		}
		plan.Document.Blocks = append(plan.Document.Blocks, office.Paragraph(role, inlines))
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

func officeInlines(value string) ([]office.Node, error) {
	nodes, err := source.LetterInline(value)
	if err != nil {
		return nil, err
	}
	var mapLiterals func([]office.Node) []office.Node
	mapLiterals = func(nodes []office.Node) []office.Node {
		for i := range nodes {
			if nodes[i].Type == "Code" {
				c := nodes[i].Content.([]any)
				nodes[i] = office.Span("FolioLiteral", office.Text(c[1].(string)))
			} else if children, ok := nodes[i].Content.([]source.Node); ok {
				nodes[i].Content = mapLiterals(children)
			}
		}
		return nodes
	}
	return mapLiterals(nodes), nil
}

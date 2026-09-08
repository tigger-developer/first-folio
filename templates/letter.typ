// ABOUTME: Typst layout template for generated cover letters.
// ABOUTME: Receives validated configuration and escaped letter content from Go.
#set page(paper: "{{.Page}}", margin: (top: {{.MarginTop}}, bottom: {{.MarginBottom}}, left: {{.MarginLeft}}, right: {{.MarginRight}}))
#set text(font: "{{.Font}}", size: {{.FontSize}}, weight: {{.Weight}}, stretch: {{.Stretch}}, style: "{{.Style}}", tracking: {{.LetterSpacing}})
#set par(leading: 0.8em, spacing: 1.2em)

#align(right)[#block(width: auto)[#align(left)[
{{.Sender}}{{if .Contact}}
{{.Contact}}{{end}}
]]]

#v({{.SpaceAfterSender}})

{{.Recipient}}

#v({{.SpaceAfterRecipient}})

{{.Date}}

#v({{.SpaceAfterDate}})

*{{.Subject}}*

#v({{.SpaceAfterSubject}})

{{.Body}}

#v({{.SpaceBeforeClosing}})

{{.Closing}},

#v({{.SpaceBeforeSignoff}})

{{.Signoff}}

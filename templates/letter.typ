// ABOUTME: Typst layout template for generated cover letters.
// ABOUTME: Receives validated configuration and escaped letter content from Go.
#set page(paper: "{{.Page}}", margin: (top: {{.MarginTop}}, bottom: {{.MarginBottom}}, left: {{.MarginLeft}}, right: {{.MarginRight}}))
#set text(font: "{{.Font}}", size: {{.FontSize}}, weight: {{.Weight}}, stretch: 100%, style: "{{.Style}}", tracking: {{.LetterSpacing}})
#set par(leading: 0.8em, spacing: 1.2em)

#let folio-font-role = state("folio-font-role", "role")
#let folio-font-stretch(ratio, body, role: "role") = {
  if ratio == 100% {
    body
  } else {
    folio-font-role.update(role)
    set text(stretch: 100%, spacing: ratio)
    show regex("\\S+"): word => context {
      if folio-font-role.at(here()) != role {
        word
      } else {
        let descent = measure(text(top-edge: "baseline", word)).height
        box(baseline: descent, scale(x: ratio, y: 100%, reflow: true, word))
      }
    }
    body
    folio-font-role.update("role")
  }
}

#let mono-text(body) = text(font: "{{.MonoFont}}", size: {{.MonoFontSize}}, weight: {{.MonoWeight}}, stretch: 100%, style: "{{.MonoStyle}}", tracking: {{.MonoSpacing}})[
  #folio-font-stretch({{.MonoStretch}}, body, role: "mono")
]

#folio-font-stretch({{.Stretch}})[

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
]

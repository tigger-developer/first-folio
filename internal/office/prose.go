// ABOUTME: Defines the resolved prose layout exchanged by manuscript and letter adapters.
// ABOUTME: Keeps mode-specific configuration and source semantics outside native package mechanics.
package office

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
)

// Style contains resolved native point values rather than source-mode YAML.
type Style struct {
	Font                                                    config.Font
	Size, Tracking, Line, Before, After, Left, Right, First float64
	Align                                                   string
	Bold, Italic, KeepNext, Hyphenate, Widow                bool
	Character                                               bool
}

// Page describes physical paper, mirrored margins and running-band geometry.
type Page struct {
	Width, Height, Top, Bottom, Left, Right, Gutter float64
	Header, Footer                                  Band
}
type Band struct {
	Enabled                        bool
	Font                           Style
	Even, Odd, FrontEven, FrontOdd string
	Align                          string
	Edge, Clearance                float64
}

// Boundary applies at a named paragraph style, separating physical and displayed numbering.
type Boundary struct {
	Before, After                              string
	Break, Body, Reset, SkipHeader, SkipFooter bool
	NumberFormat, Vertical                     string
	Title                                      bool
	TitleNumber                                bool
	Blank                                      bool
	Fields                                     map[string]string
}

// Placement keeps independent title/copyright groups editable in anchored text frames.
type Placement struct {
	Vertical   string
	Horizontal string
	Group      string
}

// ContentsEntry binds an editable contents paragraph to its heading bookmark.
type ContentsEntry struct {
	Style, Anchor         string
	Links, Leaders, Pages bool
}

// Prose is a writer plan, not a replacement for either source-mode model.
type Prose struct {
	Document                                       Document
	Styles                                         map[string]Style
	Page                                           Page
	Boundaries                                     map[string]Boundary
	Ends                                           map[string]Boundary
	Placements                                     map[string]Placement
	Contents                                       []ContentsEntry
	Metadata                                       map[string]string
	InitialNumberFormat                            string
	Images                                         map[string][]byte
	Anchors                                        map[string]string
	ListBefore, ListAfter, TableBefore, TableAfter float64
}

// Points resolves the same absolute and font-relative units used by mode configuration.
func Points(raw string, em, extent float64) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "0" || raw == "" {
		return 0, nil
	}
	for _, u := range []struct {
		suffix string
		scale  float64
	}{{"pt", 1}, {"mm", 72 / 25.4}, {"cm", 720 / 25.4}, {"in", 72}, {"em", em}, {"%", extent / 100}} {
		if strings.HasSuffix(raw, u.suffix) {
			n, err := strconv.ParseFloat(strings.TrimSuffix(raw, u.suffix), 64)
			if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
				break
			}
			return n * u.scale, nil
		}
	}
	return 0, fmt.Errorf("invalid office length %q", raw)
}

// ResolveStyle retains all six effective font properties and independent role sizing.
func ResolveStyle(font config.Font, base float64) (Style, error) {
	size, err := Points(font.Size, base, 0)
	if err != nil {
		return Style{}, err
	}
	tracking, err := Points(font.LetterSpacing, size, 0)
	if err != nil {
		return Style{}, err
	}
	bold := font.Weight == "bold" || font.Weight == "semibold" || font.Weight == "extrabold" || font.Weight == "black"
	if n, err := strconv.Atoi(font.Weight); err == nil {
		bold = n >= 600
	}
	return Style{Font: font, Size: size, Tracking: tracking, Align: "left", Bold: bold, Italic: font.Style != "regular" && font.Style != "normal", Widow: true}, nil
}

// CapHeight obtains the font's native cap-height metric without a rendering process.
func CapHeight(style Style) float64 {
	faces, _ := ResolveFonts([]string{style.Font.Family}, fontFaceDirs())
	for _, face := range faces {
		wanted := 0
		if style.Bold {
			wanted++
		}
		if style.Italic {
			wanted += 2
		}
		if face.face != wanted {
			continue
		}
		head, os2 := fontTable(face.Raw, "head"), fontTable(face.Raw, "OS/2")
		if len(head) >= 20 && len(os2) >= 90 {
			units := binary.BigEndian.Uint16(head[18:20])
			cap := int16(binary.BigEndian.Uint16(os2[88:90]))
			if units > 0 && cap > 0 {
				return style.Size * float64(cap) / float64(units)
			}
		}
	}
	// Unavailable fonts retain their family and warn during embedding. Their
	// reader-selected replacement determines final metrics; this is the nominal cap box.
	return style.Size * 0.7
}
func fontTable(raw []byte, tag string) []byte {
	if len(raw) < 12 {
		return nil
	}
	count := int(binary.BigEndian.Uint16(raw[4:6]))
	for i := 0; i < count && 12+(i+1)*16 <= len(raw); i++ {
		item := raw[12+i*16 : 12+(i+1)*16]
		if string(item[:4]) != tag {
			continue
		}
		off, size := uint64(binary.BigEndian.Uint32(item[8:12])), uint64(binary.BigEndian.Uint32(item[12:16]))
		if off <= uint64(len(raw)) && size <= uint64(len(raw))-off {
			return raw[off : off+size]
		}
	}
	return nil
}

func Attr(id string, classes []string, pairs ...[]string) []any {
	if classes == nil {
		classes = []string{}
	}
	if pairs == nil {
		pairs = [][]string{}
	}
	return []any{id, classes, pairs}
}
func Text(value string) []Node {
	if value == "" {
		return []Node{}
	}
	return []Node{N("Str", value)}
}
func Paragraph(style string, inlines []Node) Node {
	if inlines == nil {
		inlines = []Node{}
	}
	return N("Div", []any{Attr("", nil, []string{"custom-style", style}), []Node{N("Para", inlines)}})
}
func Span(style string, inlines []Node) Node {
	return N("Span", []any{Attr("", nil, []string{"custom-style", style}), inlines})
}

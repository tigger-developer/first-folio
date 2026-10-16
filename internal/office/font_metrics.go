// ABOUTME: Measures native font advances for right-positioned letter sender blocks.
// ABOUTME: Bounds cmap and metrics reads and leaves unavailable-font substitution to the reader.
package office

import (
	"encoding/binary"
	"strings"
)

// resolveBaselines aligns cap boxes to the configured content edge and running
// header baselines to their edge distance, independently of the native line box.
func (p *Prose) resolveBaselines() {
	var families []string
	seen := map[string]bool{}
	add := func(name string) {
		if !seen[name] {
			families = append(families, name)
			seen[name] = true
		}
	}
	for _, style := range p.Styles {
		add(style.Font.Family)
	}
	if p.Page.Header.Enabled {
		add(p.Page.Header.Font.Font.Family)
	}
	faces, _ := ResolveFonts(families, fontFaceDirs())
	resolve := func(style Style, baseline bool) Style {
		if style.Character || style.Line == 0 {
			return style
		}
		cap, descent, box := 0.7*style.Size, 0.25*style.Size, 1.25*style.Size
		wanted := 0
		if style.Bold {
			wanted++
		}
		if style.Italic {
			wanted += 2
		}
		for _, face := range faces {
			if face.family != style.Font.Family || face.face != wanted {
				continue
			}
			head, os2 := fontTable(face.Raw, "head"), fontTable(face.Raw, "OS/2")
			if len(head) < 20 || len(os2) < 90 {
				continue
			}
			units := float64(binary.BigEndian.Uint16(head[18:20]))
			if units == 0 {
				continue
			}
			cap = float64(binary.BigEndian.Uint16(os2[88:90])) * style.Size / units
			descent = float64(binary.BigEndian.Uint16(os2[76:78])) * style.Size / units
			box = float64(binary.BigEndian.Uint16(os2[74:76])+binary.BigEndian.Uint16(os2[76:78])) * style.Size / units
			break
		}
		style.BaselineShift = max(0, style.Line-descent-cap)
		if baseline {
			style.BaselineShift = style.Line - descent
		}
		style.FontBox = box
		return style
	}
	for key, style := range p.Styles {
		p.Styles[key] = resolve(style, false)
	}
	if p.Page.Header.Enabled {
		p.Page.Header.Font = resolve(p.Page.Header.Font, true)
	}
}

// TextWidth returns the longest line's advance in points, including configured tracking.
func TextWidth(text string, style Style) float64 {
	faces, _ := ResolveFonts([]string{style.Font.Family}, fontFaceDirs())
	var raw []byte
	wanted := 0
	if style.Bold {
		wanted++
	}
	if style.Italic {
		wanted += 2
	}
	for _, face := range faces {
		if face.face == wanted {
			raw = face.Raw
			break
		}
	}
	head, hhea, hmtx, cmap := fontTable(raw, "head"), fontTable(raw, "hhea"), fontTable(raw, "hmtx"), fontTable(raw, "cmap")
	units, metrics := 0, 0
	if len(head) >= 20 && len(hhea) >= 36 {
		units = int(binary.BigEndian.Uint16(head[18:20]))
		metrics = int(binary.BigEndian.Uint16(hhea[34:36]))
	}
	stretch, err := Points(style.Font.Stretch, 0, 1)
	if err != nil || stretch == 0 {
		stretch = 1
	}
	var longest float64
	for _, line := range strings.Split(text, "\n") {
		var width float64
		count := 0
		for _, r := range line {
			advance := style.Size * 0.6
			glyph := cmapGlyph(cmap, uint32(r))
			if glyph >= metrics {
				glyph = metrics - 1
			}
			if units > 0 && glyph >= 0 && glyph*4+2 <= len(hmtx) {
				advance = float64(binary.BigEndian.Uint16(hmtx[glyph*4:])) * style.Size / float64(units)
			}
			width += advance * stretch
			if count > 0 {
				width += style.Tracking
			}
			count++
		}
		if width > longest {
			longest = width
		}
	}
	return longest
}

func cmapGlyph(cmap []byte, char uint32) int {
	if len(cmap) < 4 {
		return 0
	}
	n := int(binary.BigEndian.Uint16(cmap[2:4]))
	for i := 0; i < n && 4+(i+1)*8 <= len(cmap); i++ {
		record := cmap[4+i*8:]
		platform := binary.BigEndian.Uint16(record)
		if platform != 0 && platform != 3 {
			continue
		}
		offset := uint64(binary.BigEndian.Uint32(record[4:8]))
		if offset+2 > uint64(len(cmap)) {
			continue
		}
		table := cmap[offset:]
		switch binary.BigEndian.Uint16(table) {
		case 12:
			if len(table) < 16 {
				continue
			}
			groups := uint64(binary.BigEndian.Uint32(table[12:16]))
			if groups > uint64((len(table)-16)/12) {
				continue
			}
			for j := uint64(0); j < groups; j++ {
				g := table[16+j*12:]
				first, last := binary.BigEndian.Uint32(g), binary.BigEndian.Uint32(g[4:])
				if char >= first && char <= last {
					return int(binary.BigEndian.Uint32(g[8:]) + char - first)
				}
			}
		case 4:
			if char > 65535 || len(table) < 16 {
				continue
			}
			segments := int(binary.BigEndian.Uint16(table[6:8])) / 2
			if 16+8*segments > len(table) {
				continue
			}
			for j := 0; j < segments; j++ {
				last := uint32(binary.BigEndian.Uint16(table[14+2*j:]))
				first := uint32(binary.BigEndian.Uint16(table[16+2*segments+2*j:]))
				if char < first || char > last {
					continue
				}
				delta := binary.BigEndian.Uint16(table[16+4*segments+2*j:])
				index := 16 + 6*segments + 2*j
				offset := int(binary.BigEndian.Uint16(table[index:]))
				if offset == 0 {
					return int(uint16(char) + delta)
				}
				at := index + offset + int(char-first)*2
				if at+2 > len(table) {
					return 0
				}
				glyph := binary.BigEndian.Uint16(table[at:])
				if glyph != 0 {
					glyph += delta
				}
				return int(glyph)
			}
		}
	}
	return 0
}

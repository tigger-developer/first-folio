// ABOUTME: Resolves complete, editable TrueType/OpenType faces for office exports.
// ABOUTME: Reads bundled and local fonts without installing or downloading anything.
package office

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode/utf16"

	folio "github.com/tigger-developer/first-folio"
)

const fontFaceLimit = 32 << 20
const fontFacesLimit = 64 << 20

type fontFace struct {
	family string
	face   int // regular, bold, italic, bold italic
	ext    string
	Raw    []byte
	rights uint16
}

// SFNT offsets are checked before slicing. Collections and variable fonts are
// deliberately excluded: their faces need extraction/instantiation first.
func ParseFont(raw []byte) (fontFace, error) {
	f, err := readOfficeFontInfo(bytes.NewReader(raw), int64(len(raw)))
	if err == nil {
		f.Raw = raw
	}
	return f, err
}

func readOfficeFontInfo(reader io.ReaderAt, length int64) (fontFace, error) {
	f := fontFace{ext: "ttf"}
	if length < 12 || length > fontFaceLimit {
		return f, fmt.Errorf("invalid font size")
	}
	raw := make([]byte, 12)
	if _, err := reader.ReadAt(raw, 0); err != nil {
		return f, err
	}
	if string(raw[:4]) == "OTTO" {
		f.ext = "otf"
	} else if binary.BigEndian.Uint32(raw[:4]) != 0x00010000 {
		return f, fmt.Errorf("not a standalone TrueType/OpenType font")
	}
	n := int(binary.BigEndian.Uint16(raw[4:6]))
	if n > 256 || int64(n) > (length-12)/16 {
		return f, fmt.Errorf("truncated font directory")
	}
	directory := make([]byte, n*16)
	if _, err := reader.ReadAt(directory, 12); err != nil {
		return f, err
	}
	tables := map[string][]byte{}
	for i := 0; i < n; i++ {
		r := directory[i*16 : (i+1)*16]
		off, size := uint64(binary.BigEndian.Uint32(r[8:12])), uint64(binary.BigEndian.Uint32(r[12:16]))
		if off > uint64(length) || size > uint64(length)-off {
			return f, fmt.Errorf("font table outside file")
		}
		tag := string(r[:4])
		if tag == "fvar" {
			return f, fmt.Errorf("variable fonts are not supported")
		}
		if tag != "name" && tag != "OS/2" {
			continue
		}
		if size > 1<<20 {
			return f, fmt.Errorf("font metadata exceeds 1 MiB")
		}
		data := make([]byte, int(size))
		if _, err := reader.ReadAt(data, int64(off)); err != nil {
			return f, err
		}
		tables[tag] = data
	}
	os2, names := tables["OS/2"], tables["name"]
	if len(os2) < 64 || len(names) < 6 {
		return f, fmt.Errorf("missing font metadata")
	}
	f.rights = binary.BigEndian.Uint16(os2[8:10])
	weight := binary.BigEndian.Uint16(os2[4:6])
	if weight != 400 && weight != 700 {
		return f, fmt.Errorf("only regular/bold faces are supported")
	}
	if weight == 700 {
		f.face++
	}
	if binary.BigEndian.Uint16(os2[62:64])&1 != 0 {
		f.face += 2
	}
	count, start := int(binary.BigEndian.Uint16(names[2:4])), int(binary.BigEndian.Uint16(names[4:6]))
	if count > (len(names)-6)/12 || start > len(names) {
		return f, fmt.Errorf("truncated font names")
	}
	for i := 0; i < count; i++ {
		r := names[6+i*12 : 18+i*12]
		platform, id := binary.BigEndian.Uint16(r[:2]), binary.BigEndian.Uint16(r[6:8])
		if id != 1 || (platform != 0 && platform != 3) {
			continue
		}
		size, off := int(binary.BigEndian.Uint16(r[8:10])), start+int(binary.BigEndian.Uint16(r[10:12]))
		if off > len(names) || size > len(names)-off || size%2 != 0 {
			return f, fmt.Errorf("invalid font name")
		}
		units := make([]uint16, size/2)
		for j := range units {
			units[j] = binary.BigEndian.Uint16(names[off+j*2 : off+j*2+2])
		}
		family := string(utf16.Decode(units))
		if f.family == "" || binary.BigEndian.Uint16(r[4:6]) == 0x409 {
			f.family = family
		}
	}
	if f.family == "" {
		return f, fmt.Errorf("font has no Unicode family name")
	}
	return f, nil
}

func (f fontFace) editable() bool {
	// OS/2 fsType: installable (0) or editable (8), never restricted,
	// preview/print-only or bitmap-only. Full faces also honour no-subsetting.
	return f.rights&0x202 == 0 && (f.rights&0xC == 0 || f.rights&8 != 0)
}

func fontFaceDirs() []string {
	dirs := filepath.SplitList(os.Getenv("TYPST_FONT_PATHS"))
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, filepath.Join(home, "Library/Fonts"), "/Library/Fonts", "/System/Library/Fonts")
	case "windows":
		dirs = append(dirs, filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft/Windows/Fonts"), filepath.Join(os.Getenv("WINDIR"), "Fonts"))
	default:
		data := os.Getenv("XDG_DATA_HOME")
		if data == "" {
			data = filepath.Join(home, ".local/share")
		}
		dirs = append(dirs, filepath.Join(data, "fonts"), filepath.Join(home, ".fonts"), "/usr/local/share/fonts", "/usr/share/fonts")
	}
	return dirs
}

func ResolveFonts(families []string, dirs []string) ([]fontFace, []string) {
	wanted := map[string]bool{}
	for _, name := range families {
		wanted[name] = true
	}
	selected := map[string]map[int]fontFace{}
	issues := map[string]string{}
	total := 0
	accept := func(raw []byte) {
		f, err := ParseFont(raw)
		if err != nil || !wanted[f.family] {
			return
		}
		if _, exists := selected[f.family][f.face]; exists {
			return
		}
		if !f.editable() {
			issues[f.family] = "a face prohibits editable embedding"
			return
		}
		if selected[f.family] == nil {
			selected[f.family] = map[int]fontFace{}
		}
		if len(raw) > fontFacesLimit-total {
			issues[f.family] = "64 MiB font budget exceeded"
			return
		}
		selected[f.family][f.face] = f
		total += len(raw)
	}
	entries, _ := fs.ReadDir(folio.Assets, "fonts/iosevka-publish")
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ttf") {
			raw, err := folio.Assets.ReadFile("fonts/iosevka-publish/" + e.Name())
			if err == nil {
				accept(raw)
			}
		}
	}
	// Avoid scanning host fonts when the bundled faces satisfy every family.
	allFound := func() bool {
		for _, f := range families {
			if len(selected[f]) != 4 {
				return false
			}
		}
		return true
	}
	visited := 0
	for _, dir := range dirs {
		if dir == "" || allFound() {
			continue
		}
		_ = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			visited++
			if visited > 10000 {
				return fs.SkipAll
			}
			if err != nil || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".otf" {
				return nil
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || info.Size() > fontFaceLimit {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			metadata, metadataErr := readOfficeFontInfo(file, info.Size())
			_, exists := selected[metadata.family][metadata.face]
			var raw []byte
			var readErr error
			if metadataErr == nil && wanted[metadata.family] && !exists {
				if metadata.editable() {
					raw, readErr = io.ReadAll(io.LimitReader(file, fontFaceLimit+1))
				} else {
					issues[metadata.family] = "a face prohibits editable embedding"
				}
			}
			closeErr := file.Close()
			if len(raw) > 0 && readErr == nil && closeErr == nil {
				accept(raw)
			}
			return nil
		})
	}
	var fonts []fontFace
	var warnings []string
	sort.Strings(families)
	for _, family := range families {
		if len(selected[family]) == 0 {
			warnings = append(warnings, fmt.Sprintf("font %q not embedded: no supported editable local face found", family))
		}
		if issue := issues[family]; issue != "" {
			warnings = append(warnings, fmt.Sprintf("font %q: some faces not embedded: %s", family, issue))
		}
		for face := 0; face < 4; face++ {
			if f, ok := selected[family][face]; ok {
				fonts = append(fonts, f)
			}
		}
	}
	return fonts, warnings
}

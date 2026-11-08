// ABOUTME: Covers explicit and title-page-driven suppression of source-defined script covers.
// ABOUTME: Proves disabled covers are not validated or loaded for PDF, DOCX or ODT conversion.
package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRT081_1DisabledScriptCoverIsIgnoredAcrossRenderedFormats(t *testing.T) {
	binary := officeBuild(t)
	const source = `#+TITLE: Disabled cover
#+AUTHOR: Example Author
#+FOLIO.COVER-IMAGE.PATH: missing-cover.png
#+FOLIO.COVER-IMAGE.MODE: background
* ACT ONE
**** CÁIT
Hello.
`
	for _, control := range []struct {
		name string
		yaml string
	}{
		{"cover disabled", "folio:\n  cover-image:\n    enabled: false\n"},
		{"title page disabled", "folio:\n  title-page:\n    enabled: false\n    author:\n      position: bottom-center\n"},
	} {
		for _, format := range []string{"pdf", "docx", "odt"} {
			t.Run(control.name+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "play.org")
				target := filepath.Join(dir, "play."+format)
				writeAppFile(t, src, source)
				writeAppFile(t, filepath.Join(dir, "script.yaml"), control.yaml)
				_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, target)
				if err != nil {
					t.Fatalf("disabled cover %s conversion: %v: %s", format, err, stderr)
				}
				if info, statErr := os.Stat(target); statErr != nil || info.Size() == 0 {
					t.Fatalf("disabled cover %s target: %v size=%d", format, statErr, fileSize(info))
				}
			})
		}
	}
}

func fileSize(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}

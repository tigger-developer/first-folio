package app

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT077_1OrgPercentageControlsScriptTitlePosition(t *testing.T) {
	requireW073Tools(t)
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	writeW073PNG(t, filepath.Join(dir, "cover.png"), 976, 976, color.RGBA{R: 30, G: 40, B: 50, A: 255})

	type titleGroup struct {
		title, subtitle wordBoundsW073
	}
	render := func(name, position string, coverLines []string) titleGroup {
		t.Helper()
		source := filepath.Join(dir, name+".org")
		target := filepath.Join(dir, name+".pdf")
		lines := []string{
			"#+TITLE: POSITIONMARKER",
			"#+SUBTITLE: SUBTITLEMARKER",
			"#+FOLIO.TITLE-PAGE.TITLE.POSITION: " + position,
			"#+FOLIO.COVER-IMAGE.PATH: cover.png",
		}
		lines = append(lines, coverLines...)
		lines = append(lines, "* ACT ONE", "**** CÁIT", "Body.")
		writeAppFile(t, source, strings.Join(lines, "\n")+"\n")
		status, _, stderr := runApp(t, "convert", source, target)
		if status != 0 {
			t.Fatalf("%s at %s: status=%d stderr=%s", name, position, status, stderr)
		}
		return titleGroup{
			title:    w073WordBounds(t, target, "POSITIONMARKER"),
			subtitle: w073WordBounds(t, target, "SUBTITLEMARKER"),
		}
	}

	for _, tc := range []struct {
		name       string
		coverLines []string
	}{
		{
			name: "background-overlay",
			coverLines: []string{
				"#+FOLIO.COVER-IMAGE.MODE: background",
			},
		},
		{
			name: "inset-dodge",
			coverLines: []string{
				"#+FOLIO.COVER-IMAGE.MODE: inset",
				"#+FOLIO.COVER-IMAGE.PLACEMENT: dodge",
				"#+FOLIO.COVER-IMAGE.SIZE: 80mm",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			at20 := render(tc.name+"-20", "20%", tc.coverLines)
			at40 := render(tc.name+"-40", "40%", tc.coverLines)
			if at20.title.yMin >= at40.title.yMin || math.Abs(at40.title.yMin-at20.title.yMin) < 20 {
				t.Fatalf("percentage did not control title position: 20%%=%#v 40%%=%#v", at20, at40)
			}
			titleDelta := at40.title.yMin - at20.title.yMin
			subtitleDelta := at40.subtitle.yMin - at20.subtitle.yMin
			if subtitleDelta < 20 || math.Abs(titleDelta-subtitleDelta) > 0.2 {
				t.Fatalf("subtitle did not move with title group: title delta=%g subtitle delta=%g; 20%%=%#v 40%%=%#v", titleDelta, subtitleDelta, at20, at40)
			}
		})
	}
}

func TestRT077_1UnsupportedScriptTitlePositionFailsAtomically(t *testing.T) {
	for _, position := range []string{"floating", "10mm", "-1%", "101%"} {
		t.Run(position, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			writeAppFile(t, source, "#+TITLE: Position\n#+FOLIO.TITLE-PAGE.TITLE.POSITION: "+position+"\n* ACT ONE\n**** CÁIT\nBody.\n")
			status, _, stderr := runApp(t, "convert", source, target)
			if status == 0 || !strings.Contains(stderr, "folio.title-page.title.position") {
				t.Fatalf("position=%q status=%d stderr=%s", position, status, stderr)
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatalf("position=%q published target: %v", position, err)
			}
		})
	}
}

func TestRT077_1PercentageBoundariesAreAccepted(t *testing.T) {
	for _, position := range []string{"0%", "100%"} {
		t.Run(position, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			writeAppFile(t, source, "#+TITLE: Position\n#+FOLIO.TITLE-PAGE.TITLE.POSITION: "+position+"\n* ACT ONE\n**** CÁIT\nBody.\n")
			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("position=%q status=%d stderr=%s", position, status, stderr)
			}
			if _, err := os.Stat(target); err != nil {
				t.Fatalf("position=%q target: %v", position, err)
			}
		})
	}
}

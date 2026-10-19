// ABOUTME: Checks the independent novel recipe against its signed-off scale and text diversity.
// ABOUTME: Prevents homogeneous ASCII prose from standing in for the required Unicode fixture.
package testfixture

import (
	"strings"
	"testing"
	"unicode"
)

func TestNovelInventory(t *testing.T) {
	for _, format := range []string{"md", "org"} {
		text, paragraphs := Novel(format)
		if len(paragraphs) != 10000 || len(strings.Fields(text)) < 120000 {
			t.Fatalf("%s novel inventory below its specified scale", format)
		}
		lengths := map[int]bool{}
		for _, paragraph := range paragraphs {
			lengths[len(strings.Fields(paragraph))] = true
		}
		if len(lengths) < 3 {
			t.Errorf("%s novel needs varied paragraph lengths", format)
		}
		if !strings.ContainsFunc(text, func(r rune) bool { return r > unicode.MaxASCII }) {
			t.Errorf("%s novel needs Unicode text", format)
		}
	}
}

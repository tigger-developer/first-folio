// ABOUTME: Tests manuscript-first sidecar publication and truthful partial-success errors.
// ABOUTME: Injects publication failures without filesystem races or permission assumptions.
package office

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestRT063_9_SidecarPublicationOrder(t *testing.T) {
	for _, fail := range []string{"", "book.docx", "book.barcode.svg"} {
		t.Run("fail-"+fail, func(t *testing.T) {
			storage := map[string]string{"book.docx": "prior-document", "book.barcode.svg": "prior-sidecar"}
			var calls []string
			fault := errors.New("injected publication failure")
			publish := func(path string, raw []byte) error {
				calls = append(calls, path)
				if path == fail {
					return fault
				}
				storage[path] = string(raw)
				return nil
			}
			err := publishManuscript("book.docx", []byte("new-document"), "book.barcode.svg", []byte("new-sidecar"), publish)
			wantCalls := []string{"book.docx", "book.barcode.svg"}
			wantDocument, wantSidecar := "new-document", "new-sidecar"
			if fail == "book.docx" {
				wantCalls = wantCalls[:1]
				wantDocument, wantSidecar = "prior-document", "prior-sidecar"
			} else if fail != "" {
				wantSidecar = "prior-sidecar"
			}
			if !slices.Equal(calls, wantCalls) {
				t.Errorf("publication order %v; want %v", calls, wantCalls)
			}
			if storage["book.docx"] != wantDocument || storage["book.barcode.svg"] != wantSidecar {
				t.Errorf("partial outcome %#v; want %s / %s", storage, wantDocument, wantSidecar)
			}
			if fail == "" && err != nil {
				t.Fatal(err)
			}
			if fail != "" && (!errors.Is(err, fault) || !strings.Contains(fmt.Sprint(err), fail)) {
				t.Errorf("failure must retain cause and identify target %s: %v", fail, err)
			}
			if fail == "book.barcode.svg" && (!strings.Contains(fmt.Sprint(err), "book.docx") || !strings.Contains(strings.ToLower(fmt.Sprint(err)), "published")) {
				t.Errorf("sidecar failure conceals published manuscript: %v", err)
			}
		})
	}
}

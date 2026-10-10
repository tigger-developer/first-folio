// ABOUTME: Provides executable-boundary and XML package helpers for office export tests.
// ABOUTME: Confines synthetic Pandoc failures to a controlled subprocess boundary.
package app

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Only a child invoked as the synthetic Pandoc enters this path.
func init() {
	if mode := os.Getenv("FOLIO_TEST_PANDOC_FAILURE"); mode != "" {
		if strings.Contains(strings.Join(os.Args[1:], " "), "-t json") {
			fmt.Print(`{"pandoc-api-version":[1,23,1],"meta":{},"blocks":[]}`)
			os.Exit(0)
		}
		switch mode {
		case "exit":
			fmt.Fprintln(os.Stderr, "synthetic conversion failure")
			os.Exit(7)
		case "malformed":
			fmt.Print("not a zip")
		case "timeout":
			time.Sleep(125 * time.Second)
		}
		os.Exit(0)
	}
}

type officeXML struct {
	Name     xml.Name
	Attr     []xml.Attr
	Text     string
	Children []*officeXML
}

func (n *officeXML) attr(name string) string {
	for _, a := range n.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
func (n *officeXML) all(name string) []*officeXML {
	var out []*officeXML
	if n.Name.Local == name {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, c.all(name)...)
	}
	return out
}
func (n *officeXML) text() string {
	s := n.Text
	for _, c := range n.Children {
		s += c.text()
	}
	return s
}
func officeParseXML(t *testing.T, raw []byte) *officeXML {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(raw))
	root := &officeXML{}
	stack := []*officeXML{root}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		parent := stack[len(stack)-1]
		switch v := tok.(type) {
		case xml.StartElement:
			n := &officeXML{Name: v.Name, Attr: v.Attr}
			parent.Children = append(parent.Children, n)
			stack = append(stack, n)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			parent.Children = append(parent.Children, &officeXML{Text: string(v)})
		}
	}
	return root
}
func officePackage(t *testing.T, raw []byte) map[string]*officeXML {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("office package: %v", err)
	}
	out := map[string]*officeXML{}
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, ".xml") {
			r, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(r)
			closeErr := r.Close()
			if err != nil {
				t.Fatal(err)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			out[f.Name] = officeParseXML(t, data)
		}
	}
	if len(out) == 0 {
		t.Fatal("no XML package members")
	}
	return out
}
func officeBuild(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "folio")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/folio")
	cmd.Dir = root
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, raw)
	}
	return binary
}
func officeEnv(t *testing.T, home string, extra map[string]string) []string {
	t.Helper()
	values := map[string]string{"HOME": home, "TMPDIR": t.TempDir()}
	for k, v := range extra {
		values[k] = v
	}
	var env []string
	for _, e := range os.Environ() {
		k, _, _ := strings.Cut(e, "=")
		if _, replace := values[k]; !replace {
			env = append(env, e)
		}
	}
	for k, v := range values {
		env = append(env, k+"="+v)
	}
	return env
}
func officeRun(t *testing.T, binary, dir string, env []string, args ...string) ([]byte, string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.String(), err
}
func officeExport(t *testing.T, binary, source, format, style string) map[string]*officeXML {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "play"+source)
	target := filepath.Join(dir, "play."+format)
	writeAppFile(t, src, officeSources[source])
	env := officeEnv(t, t.TempDir(), nil)
	_, stderr, err := officeRun(t, binary, dir, env, "convert", src, target, "--style", style)
	if err != nil {
		t.Fatalf("export %s/%s/%s: %v: %s", source, format, style, err, stderr)
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	return officePackage(t, raw)
}
func officeBody(t *testing.T, pkg map[string]*officeXML, format string) *officeXML {
	t.Helper()
	name := "word/document.xml"
	if format == "odt" {
		name = "content.xml"
	}
	if pkg[name] == nil {
		t.Fatalf("missing %s", name)
	}
	return pkg[name]
}
func officeAssertAttr(t *testing.T, n *officeXML, element, attr, value string) {
	t.Helper()
	for _, v := range n.all(element) {
		if v.attr(attr) == value {
			return
		}
	}
	t.Errorf("missing native property %s.%s=%s", element, attr, value)
}
func officeStyle(t *testing.T, pkg map[string]*officeXML, format, name string) *officeXML {
	t.Helper()
	file := "word/styles.xml"
	if format == "odt" {
		file = "styles.xml"
	}
	if pkg[file] == nil {
		t.Fatalf("missing %s", file)
	}
	for _, s := range pkg[file].all("style") {
		if format == "odt" && (s.attr("name") == name || s.attr("display-name") == name) {
			return s
		}
		if format == "docx" {
			for _, n := range s.all("name") {
				if n.attr("val") == name {
					return s
				}
			}
		}
	}
	t.Fatalf("missing %s style %q", format, name)
	return nil
}

var officeSources = map[string]string{
	".org":      "#+TITLE: Matrix Play\n#+AUTHOR: Example Author\n* ACT ONE\n** SCENE ONE\n*** Night.\n**** CÁIT (softly)\nHello.\nSecond line.\n",
	".md":       "# Matrix Play\n\n*by Example Author*\n\n## ACT ONE\n\n### SCENE ONE\n\n*Night.*\n\n**CÁIT:** *(softly)*\nHello.\nSecond line.\n",
	".fountain": "Title: Matrix Play\nAuthor: Example Author\n\n> **ACT ONE** <\n\n.SCENE ONE\n\nNight.\n\nCÁIT\n(softly)\nHello.\nSecond line.\n",
}

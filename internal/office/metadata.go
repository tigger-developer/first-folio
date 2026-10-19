// ABOUTME: Writes resolved prose metadata without asking Pandoc to synthesize visible title blocks.
// ABOUTME: Keeps document properties separate from configured editable generated matter.
package office

func (p *Prose) nativeMetadata(files map[string][]byte, format string) error {
	path := "docProps/core.xml"
	if format == "odt" {
		path = "meta.xml"
	}
	root, err := parseXML(files[path])
	if err != nil {
		return err
	}
	root.set("xmlns:dc", "http://purl.org/dc/elements/1.1/")
	parent := root
	if format == "odt" {
		parent = root.ensure("office:meta")
	}
	for key, value := range p.Metadata {
		native := "dc:" + key
		if key == "author" {
			native = "dc:creator"
		}
		if key != "author" && key != "title" {
			continue
		}
		node := parent.ensure(native)
		node.text = value
		node.children = nil
	}
	files[path], err = docxEncode(*root, DataLimit)
	return err
}

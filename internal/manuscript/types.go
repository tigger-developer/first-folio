// ABOUTME: Shared data structures for manuscript parsing and rendering.
// ABOUTME: Keeps prose manuscript semantics separate from the stage-play event stream.
package manuscript

type Metadata struct {
	Title             string
	Subtitle          string
	Author            string
	AuthorAttribution string
	Date              string
	Version           string
	WordCount         string
	ContactName       string
	Address           string
	Phone             string
	Email             string
	Website           string
}

type Block struct {
	Kind           string
	Level          int
	Text           string
	Lang           string
	ChapterOpening bool
	// AC18.1 / AC18.2: semantic-authoring fields for part and chapter blocks.
	// Name is the semantic heading text after the parser has stripped any
	// `Part N: ` / `Chapter N: ` prefix. Number is the source-order-derived
	// number (part counter starts at 1 and increments per H1; chapter counter
	// resets to 1 at each new part and increments per H2). SourceNumber and
	// SourceSeparator capture the number literal and separator glyph as they
	// appeared in the source, in case the manuscript opts in to
	// `explicit-numbering: source`.
	Name            string
	Number          int
	SourceNumber    string
	SourceSeparator string
}

type Document struct {
	Metadata Metadata
	Blocks   []Block
}

type InputSet struct {
	Format string
	Paths  []string
}

type Options struct {
	Style             string
	Output            string
	DryRun            bool
	ShowHelp          bool
	ShowVersion       bool
	Title             string
	Subtitle          string
	Author            string
	AuthorAttribution string
	Date              string
	Version           string
	WordCount         string
	ContactName       string
}

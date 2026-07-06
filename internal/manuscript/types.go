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
	Address           string
	Email             string
	Website           string
}

type Block struct {
	Kind  string
	Level int
	Text  string
	Lang  string
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
}

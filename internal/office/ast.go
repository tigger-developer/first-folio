// ABOUTME: Declares the bounded Pandoc JSON validation boundary for prose export tests.
// ABOUTME: W063 supplies placeholder declarations only until its test-code gate passes.
package office

type astLimits struct {
	bytes, blocks, inlines, depth int
}

func validateAST(raw []byte, limits astLimits) error {
	return nil
}

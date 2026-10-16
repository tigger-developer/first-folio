// ABOUTME: Declares the prose package validation boundary for preimplementation regression tests.
// ABOUTME: Placeholder bodies implement no export behaviour; W063 test-code review precedes implementation.
package office

type packageLimits struct {
	compressed, expanded, member uint64
	entries                      int
}

func validatePackage(raw []byte, format string, limits packageLimits) error {
	return nil
}

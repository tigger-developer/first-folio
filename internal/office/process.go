// ABOUTME: Declares the bounded subprocess boundary needed by prose export regressions.
// ABOUTME: This preimplementation placeholder has no callers outside the W063 test package.
package office

import "context"

type processLimits struct {
	output, diagnostic int
}

func runProcess(ctx context.Context, executable string, args []string, input []byte, limits processLimits) ([]byte, string, error) {
	return nil, "", nil
}

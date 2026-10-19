// ABOUTME: Runs bounded native document tools without a shell or caller-supplied filters.
// ABOUTME: Cancels on output overflow and preserves deadlines and subprocess diagnostics.
package office

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type processLimits struct{ output, diagnostic int }

type boundedOutput struct {
	buffer   bytes.Buffer
	limit    int
	cancel   context.CancelFunc
	overflow bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	left := b.limit - b.buffer.Len()
	if len(p) > left {
		b.overflow = true
		n, err := b.buffer.Write(p[:left])
		b.cancel()
		return n, errors.Join(fmt.Errorf("generated data exceeds %d bytes", b.limit), err)
	}
	return b.buffer.Write(p)
}

func runProcess(ctx context.Context, executable string, args []string, input []byte, limits processLimits) ([]byte, string, error) {
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	output := boundedOutput{limit: limits.output, cancel: cancel}
	diagnostic := boundedOutput{limit: limits.diagnostic, cancel: cancel}
	cmd := exec.CommandContext(child, executable, args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = &diagnostic
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	if output.overflow || diagnostic.overflow {
		err = errors.Join(err, fmt.Errorf("process output or diagnostic limit exceeded"))
	}
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err())
	}
	if err != nil {
		return nil, diagnostic.buffer.String(), fmt.Errorf("%s failed: %w", executable, err)
	}
	return output.buffer.Bytes(), diagnostic.buffer.String(), nil
}

// Pandoc applies the same limits to every canonical reader and writer invocation.
func Pandoc(input []byte, args ...string) ([]byte, string, error) {
	if len(input) > DataLimit {
		return nil, "", fmt.Errorf("Pandoc input exceeds 128 MiB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	raw, diagnostic, err := runProcess(ctx, "pandoc", args, input, processLimits{DataLimit, 64 << 10})
	if err != nil {
		return nil, diagnostic, fmt.Errorf("Pandoc conversion (requires compatible Pandoc 3.x): %w: %s", err, diagnostic)
	}
	return raw, diagnostic, nil
}

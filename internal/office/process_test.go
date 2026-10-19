// ABOUTME: Forces deterministic subprocess faults at the office process boundary.
// ABOUTME: Uses the Go test executable as a local child, with no shell or external service.
package office

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProseProcessFixture(t *testing.T) {
	index := -1
	for i, arg := range os.Args {
		if arg == "--office-process-fixture" {
			index = i
			break
		}
	}
	if index < 0 {
		return
	}
	if len(os.Args) != index+3 {
		t.Fatal("invalid subprocess fixture arguments")
	}
	mode, value := os.Args[index+1], os.Args[index+2]
	switch mode {
	case "echo":
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			t.Fatal(err)
		}
	case "stdout", "stderr":
		n, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		writer := os.Stdout
		if mode == "stderr" {
			writer = os.Stderr
		}
		if _, err := io.WriteString(writer, strings.Repeat("x", n)); err != nil {
			t.Fatal(err)
		}
	case "fail":
		if _, err := io.WriteString(os.Stderr, "fixture failure"); err != nil {
			t.Fatal(err)
		}
		os.Exit(23)
	case "stall":
		// Deliberately unresponsive external process; the parent must cancel it.
		timer := time.NewTimer(time.Hour)
		defer timer.Stop()
		<-timer.C
	default:
		t.Fatalf("unknown subprocess fixture %q", mode)
	}
	os.Exit(0) // Avoid the test runner appending PASS to the protocol bytes.
}

func processFixture(t *testing.T, ctx context.Context, mode string, n int, input []byte, limits processLimits) ([]byte, string, error) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-test.run=^TestProseProcessFixture$", "--", "--office-process-fixture", mode, strconv.Itoa(n)}
	return runProcess(ctx, executable, args, input, limits)
}

func TestRT063_9_ProcessPreservesInputAndOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	input := []byte("<source>literal & UTF-8: café</source>\n")
	got, diagnostic, err := processFixture(t, ctx, "echo", 0, input, processLimits{1024, 1024})
	if err != nil || !bytes.Equal(got, input) || diagnostic != "" {
		t.Fatalf("input/output mismatch: output=%q diagnostic=%q error=%v", got, diagnostic, err)
	}
}

func TestRT063_9_ProcessOutputThresholds(t *testing.T) {
	for _, stream := range []string{"stdout", "stderr"} {
		for _, n := range []int{65536, 65537} {
			t.Run(stream+"/"+strconv.Itoa(n), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				output, diagnostic, err := processFixture(t, ctx, stream, n, nil, processLimits{65536, 65536})
				if len(output) > 65536 || len(diagnostic) > 65536 {
					t.Fatal("subprocess exceeded captured-data budget")
				}
				if n == 65536 {
					got := len(output)
					if stream == "stderr" {
						got = len(diagnostic)
					}
					if err != nil || got != n {
						t.Errorf("exact limit must succeed without truncation: bytes=%d error=%v", got, err)
					}
				} else if err == nil {
					t.Error("overflow must fail, not truncate successfully")
				}
			})
		}
	}
}

func TestRT063_9_ProcessDeadlineAndExit(t *testing.T) {
	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, _, err := processFixture(t, ctx, "stall", 0, nil, processLimits{1024, 1024})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected propagated deadline, got %v", err)
		}
		if time.Since(start) > 3*time.Second {
			t.Error("process was not reaped promptly after cancellation")
		}
	})
	t.Run("nonzero", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, diagnostic, err := processFixture(t, ctx, "fail", 0, nil, processLimits{1024, 1024})
		if err == nil || !strings.Contains(diagnostic, "fixture failure") {
			t.Errorf("child failure lost: diagnostic=%q error=%v", diagnostic, err)
		}
	})
	t.Run("missing-executable", func(t *testing.T) {
		_, _, err := runProcess(context.Background(), t.TempDir()+"/absent-pandoc", nil, nil, processLimits{1024, 1024})
		if err == nil {
			t.Error("missing executable accepted")
		}
	})
}

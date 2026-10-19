// ABOUTME: Reads bounded, regular UTF-8 source files for editable prose exports.
// ABOUTME: Rejects text that native XML cannot preserve before conversion begins.
package office

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

const SourceLimit = 32 << 20
const DataLimit = 128 << 20

func validText(text string) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("office source requires UTF-8 text")
	}
	for _, r := range text {
		if r == 9 || r == 10 || r == 13 || r >= 0x20 && r <= 0xD7FF || r >= 0xE000 && r <= 0xFFFD || r >= 0x10000 && r <= 0x10FFFF {
			continue
		}
		return fmt.Errorf("office XML cannot represent character U+%04X", r)
	}
	return nil
}

// ReadSources retains per-file bytes and applies one aggregate source budget.
func ReadSources(paths []string) ([][]byte, error) { return readSources(paths, SourceLimit) }

func readSources(paths []string, limit int64) ([][]byte, error) {
	var result [][]byte
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("checking office source %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("office source %s must be a regular file", path)
		}
		if info.Size() > limit {
			return nil, fmt.Errorf("office input exceeds aggregate 32 MiB limit: %s", path)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("reading office source %s: %w", path, err)
		}
		opened, statErr := f.Stat()
		if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
			closeErr := f.Close()
			return nil, errors.Join(fmt.Errorf("office source %s changed while opening", path), statErr, closeErr)
		}
		raw, err := io.ReadAll(io.LimitReader(f, limit+1))
		err = errors.Join(err, f.Close())
		if err != nil {
			return nil, fmt.Errorf("reading office source %s: %w", path, err)
		}
		if int64(len(raw)) > limit {
			return nil, fmt.Errorf("office input exceeds aggregate 32 MiB limit: %s", path)
		}
		if err := validText(string(raw)); err != nil {
			return nil, fmt.Errorf("office source %s: %w", path, err)
		}
		limit -= int64(len(raw))
		result = append(result, raw)
	}
	return result, nil
}

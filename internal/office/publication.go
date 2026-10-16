// ABOUTME: Declares the atomic publication filesystem boundary for preimplementation tests.
// ABOUTME: W063 tests inject only staging and rename operations; placeholder bodies publish nothing.
package office

import "os"

type stagedFile interface {
	Write([]byte) (int, error)
	Close() error
	Chmod(os.FileMode) error
	Name() string
}

type publicationFS struct {
	createTemp func(string, string) (stagedFile, error)
	rename     func(string, string) error
	remove     func(string) error
}

func publishFile(path string, raw []byte, fs publicationFS) error {
	return nil
}

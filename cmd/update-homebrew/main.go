// ABOUTME: Updates and optionally publishes the First Folio Homebrew formula.
// ABOUTME: Replaces shell release logic with checked HTTP, hashing, file, and git operations.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const formulaPath = "homebrew/Formula/first-folio.rb"

func main() {
	publishTap := flag.String("publish-tap", "", "path to the Homebrew tap checkout")
	flag.Parse()
	if flag.NArg() != 1 {
		fatal("usage: update-homebrew [--publish-tap PATH] VERSION")
	}
	version := strings.TrimPrefix(flag.Arg(0), "v")
	url := "https://github.com/tigger-developer/first-folio/archive/refs/tags/v" + version + ".tar.gz"
	digest, err := downloadDigest(url)
	if err != nil {
		fatal(err.Error())
	}
	if err := updateFormula(formulaPath, url, digest); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("Updated %s to %s (%s)\n", formulaPath, version, digest)
	if *publishTap != "" {
		if err := publishFormula(*publishTap, version); err != nil {
			fatal(err.Error())
		}
	}
}

func downloadDigest(url string) (string, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	response, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading release tarball: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading release tarball: %s", response.Status)
	}
	hash := sha256.New()
	count, err := io.Copy(hash, response.Body)
	if err != nil {
		return "", fmt.Errorf("hashing release tarball: %w", err)
	}
	if count == 0 {
		return "", fmt.Errorf("release tarball was empty: %s", url)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func updateFormula(path string, url string, digest string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading formula: %w", err)
	}
	content := string(raw)
	content = regexp.MustCompile(`(?m)^  url ".*"$`).ReplaceAllString(content, `  url "`+url+`"`)
	content = regexp.MustCompile(`(?m)^  sha256 ".*"$`).ReplaceAllString(content, `  sha256 "`+digest+`"`)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing formula: %w", err)
	}
	return nil
}

func publishFormula(tap string, version string) error {
	target := filepath.Join(tap, "Formula", "first-folio.rb")
	raw, err := os.ReadFile(formulaPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return fmt.Errorf("publishing formula file: %w", err)
	}
	for _, args := range [][]string{{"add", "Formula/first-folio.rb"}, {"commit", "-m", "first-folio " + version}, {"push"}} {
		command := exec.Command("git", args...)
		command.Dir = tap
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}

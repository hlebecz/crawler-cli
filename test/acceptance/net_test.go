//go:build acceptance_network

package acceptance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCLI_RealSite(t *testing.T) {
	dir := t.TempDir()

	res := runCLI(t, 30*time.Second, dir,
		"--urls", "https://example.com",
		"--depth", "0",
		"--timeout", "20s",
		"--request-timeout", "10s",
		"--output", "result.json",
		"--log", "crawler.log",
	)

	if res.exitCode != 0 {
		t.Fatalf("expected exit code 0 against a real site, got %d (stderr: %s)", res.exitCode, res.stderr)
	}

	f, err := os.Open(filepath.Join(dir, "result.json"))
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	defer f.Close()

	nodes, maxDepth, err := CountNodes(f)
	if err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if nodes != 1 {
		t.Errorf("got %d nodes, expected 1 (just the seed page, depth=0)", nodes)
	}
	if maxDepth != 1 {
		t.Errorf("got max depth %d, expected 1", maxDepth)
	}

	if _, err := os.Stat(filepath.Join(dir, "crawler.log")); err != nil {
		t.Errorf("expected log file to exist: %v", err)
	}
}

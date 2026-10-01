package acceptance

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

var binPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "crawler-cli-bin-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	binPath = filepath.Join(tmpDir, "crawler")

	build := exec.Command("go", "build", "-o", binPath, "github.com/hlebecz/crawler-cli/cmd/app")
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("failed to build crawler binary: " + err.Error())
	}

	os.Exit(m.Run())
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runCLI(t *testing.T, timeout time.Duration, dir string, args ...string) cliResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run crawler binary: %v (stderr: %s)", err, stderr.String())
		}
	}

	return cliResult{stdout: stdout.String(), stderr: stderr.String(), exitCode: exitCode}
}

func TestCLI_MissingRequiredURLsFlag(t *testing.T) {
	dir := t.TempDir()

	res := runCLI(t, 5*time.Second, dir, "--depth", "1")

	if res.exitCode == 0 {
		t.Fatalf("expected a non-zero exit code when --urls is missing, got 0 (stderr: %s)", res.stderr)
	}
	if !strings.Contains(res.stderr, "urls") {
		t.Errorf("expected the error output to mention the missing 'urls' flag, got: %s", res.stderr)
	}
}

func TestCLI_CrawlsSeedAndWritesOutputAndLog(t *testing.T) {
	srv := serverFromMux(t, testMux())
	dir := t.TempDir()

	res := runCLI(t, 10*time.Second, dir,
		"--urls", srv.URL,
		"--depth", "1",
		"--timeout", "5s",
		"--request-timeout", "2s",
		"--output", "result.json",
		"--log", "crawler.log",
	)

	if res.exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", res.exitCode, res.stderr)
	}

	f, err := os.Open(filepath.Join(dir, "result.json"))
	if err != nil {
		t.Fatalf("expected --output file to exist: %v", err)
	}
	defer f.Close()

	outBytes, err := os.ReadFile(filepath.Join(dir, "result.json"))
	if err != nil {
		t.Fatalf("expected --output file to exist: %v", err)
	}
	if len(outBytes) == 0 {
		t.Error("got an empty output file")
	}

	logBytes, err := os.ReadFile(filepath.Join(dir, "crawler.log"))
	if err != nil {
		t.Fatalf("expected --log file to exist: %v", err)
	}
	if len(logBytes) == 0 {
		t.Error("got an empty log file")
	}
}

func TestCLI_ExitsNonZeroWhenOutputFileCannotBeCreated(t *testing.T) {
	srv := serverFromMux(t, testMux())
	dir := t.TempDir()

	res := runCLI(t, 10*time.Second, dir,
		"--urls", srv.URL,
		"--depth", "0",
		"--output", filepath.Join("no-such-directory", "result.json"),
	)

	if res.exitCode == 0 {
		t.Fatalf("expected a non-zero exit code when the output path is invalid, got 0 (stderr: %s)", res.stderr)
	}
}

func TestCLI_GracefulShutdownOnSIGINT(t *testing.T) {
	var once sync.Once
	started := make(chan struct{})

	mux := slowMux(func() {
		once.Do(func() { close(started) })
	}, 30*time.Second)
	srv := serverFromMux(t, mux)

	dir := t.TempDir()

	cmd := exec.Command(binPath,
		"--urls", srv.URL,
		"--depth", "0",
		"--timeout", "30s",
		"--request-timeout", "25s",
		"--output", "result.json",
		"--log", "crawler.log",
	)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start crawler binary: %v", err)
	}

	select {
	case <-started:

	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("test server never received the crawler's request")
	}

	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	waitErr := make(chan error, 1)
	go func() { waitErr <- cmd.Wait() }()

	select {
	case err := <-waitErr:
		if err != nil {
			t.Fatalf("expected a clean exit after SIGINT, got: %v (stderr: %s)", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5s of SIGINT despite --timeout=30s — graceful shutdown is broken")
	}
}

func TestCrawlDepth_NodeCountAndTreeDepth(t *testing.T) {
	cases := []struct {
		name      string
		goFlag    string
		depth     string
		wantNodes int
		wantDepth int
	}{
		{"concurrent crawler, depth=0 crawls only the seed page", "4", "0", 1, 1},
		{"concurrent crawler, depth=1 also reaches the seed's direct links", "4", "1", 2, 2},
		{"concurrent crawler, depth=2 also reaches links-of-links", "4", "2", 4, 3},
		{"recursive crawler, depth=0 crawls only the seed page", "1", "0", 1, 1},
		{"recursive crawler, depth=1 also reaches the seed's direct links", "1", "1", 2, 2},
		{"recursive crawler, depth=2 also reaches links-of-links", "1", "2", 4, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := serverFromMux(t, testMux())
			dir := t.TempDir()

			res := runCLI(t, 10*time.Second, dir,
				"--urls", srv.URL,
				"--depth", tc.depth,
				"--timeout", "5s",
				"--request-timeout", "2s",
				"--output", "result.json",
				"--go", tc.goFlag,
			)
			if res.exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d (stderr: %s)", res.exitCode, res.stderr)
			}

			f, err := os.Open(filepath.Join(dir, "result.json"))
			if err != nil {
				t.Fatalf("reopen output file: %v", err)
			}
			defer f.Close()

			nodes, maxDepth, err := CountNodes(f)
			if err != nil {
				t.Fatalf("count nodes: %v", err)
			}

			if nodes != tc.wantNodes {
				t.Errorf("got %d nodes, expected %d", nodes, tc.wantNodes)
			}
			if maxDepth != tc.wantDepth {
				t.Errorf("got max depth %d, expected %d", maxDepth, tc.wantDepth)
			}
		})
	}
}

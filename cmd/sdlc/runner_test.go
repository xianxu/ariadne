package main

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGitCmd keeps stderr out of a successful command's output — a parsed
// stream (a -z status, a SHA) must not carry warnings — and returns it after
// stdout on failure so error messages keep the diagnostics (#259).
func TestRunGitCmdSeparatesStderr(t *testing.T) {
	out, err := runGitCmd(exec.Command("sh", "-c", "printf 'data'; printf 'warning: noise' >&2"))
	if err != nil || string(out) != "data" {
		t.Fatalf("success = (%q, %v), want stdout alone", out, err)
	}
	out, err = runGitCmd(exec.Command("sh", "-c", "printf 'partial'; printf 'fatal: broke' >&2; exit 3"))
	if err == nil || string(out) != "partialfatal: broke" {
		t.Fatalf("failure = (%q, %v), want stdout then stderr with the error", out, err)
	}
}

func TestNetworkProductionFetch(t *testing.T) {
	dir := t.TempDir()
	script := `#!/bin/sh
printf 'partial'
printf '%s\n' "fatal: unable to access 'https://user:SECRET@git.example/repo': Could not resolve host: git.example" >&2
exit 128
`
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, runner := range []struct {
		name string
		run  func() ([]byte, error)
	}{
		{"git", func() ([]byte, error) { return (execGitRunner{}).Git("fetch", "origin") }},
		{"git-in-dir", func() ([]byte, error) { return (execGitRunner{}).GitInDir(dir, "fetch", "origin") }},
		{"weekly", func() ([]byte, error) { return weeklyGit(context.Background(), dir, "fetch", "origin") }},
	} {
		t.Run(runner.name, func(t *testing.T) {
			_, err := runner.run()
			var network *gitx.GitNetworkFailure
			if !errors.As(err, &network) || network.Host != "git.example" {
				t.Fatalf("fetch must expose infrastructure cause: %v", err)
			}
			if strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("error leaks credential: %v", err)
			}
		})
	}
}

func TestNetworkProductionLocalFetchSuccess(t *testing.T) {
	source, target := t.TempDir(), t.TempDir()
	run := func(dir string, args ...string) []byte {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return out
	}
	run(source, "init", "-b", "main")
	run(source, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "initial")
	run(target, "init")
	out, err := (execGitRunner{}).GitInDir(target, "fetch", source, "main")
	if err != nil || len(out) != 0 {
		t.Fatalf("local fetch=(%q,%v)", out, err)
	}
	want := strings.TrimSpace(string(run(source, "rev-parse", "HEAD")))
	got, err := (execGitRunner{}).GitInDir(target, "rev-parse", "FETCH_HEAD")
	if err != nil || strings.TrimSpace(string(got)) != want {
		t.Fatalf("parsed output=(%q,%v), want %s", got, err, want)
	}
}

func TestNetworkProductionNativeGitLoopbackRefusal(t *testing.T) {
	// A closed ephemeral listener gives Git a real refused loopback connection,
	// without production DNS or a dependency on a particular well-known port.
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	remote := "git://" + listener.Addr().String() + "/no-repo"
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := boundedLandingGit(t.Context(), execGitRunner{}).GitInDir(t.TempDir(), "ls-remote", remote)
	t.Logf("real Git diagnostic: %s", out)
	var network *gitx.GitNetworkFailure
	var exit *exec.ExitError
	if !errors.As(err, &network) || network.Host != "127.0.0.1" || !errors.As(err, &exit) {
		t.Fatalf("real native transport must retain network/exit cause: %v", err)
	}
}

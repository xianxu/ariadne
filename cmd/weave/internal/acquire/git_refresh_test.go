package acquire

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func refreshGitScript(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte("#!/bin/sh\n"+script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func TestExecGitRawPreservesOutputAndRedactsErrors(t *testing.T) {
	dir := refreshGitScript(t, "printf ' path\\n\\000tail \\n'; printf 'secret-stderr' >&2\n")
	out, err := (ExecGit{Raw: true}).Run(context.Background(), dir, "probe")
	if err != nil || out != " path\n\x00tail \n" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	refreshGitScript(t, "printf 'password-in-stderr' >&2; exit 7\n")
	_, err = (ExecGit{Raw: true}).Run(context.Background(), dir, "fetch", "https://password-in-argv@example.com/repo")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("exit error lost: %v", err)
	}
	if strings.Contains(err.Error(), "password") || !strings.Contains(err.Error(), "fetch") || !strings.Contains(err.Error(), dir) {
		t.Fatalf("unsafe or unhelpful error: %v", err)
	}
}

func TestExecGitBoundedStreamsAndCancellation(t *testing.T) {
	for _, stream := range []string{"", " >&2"} {
		t.Run("overflow"+stream, func(t *testing.T) {
			dir := refreshGitScript(t, "while :; do printf '0123456789'"+stream+"; done\n")
			start := time.Now()
			_, err := (ExecGit{Raw: true, MaxOutputBytes: 64, Timeout: 5 * time.Second}).Run(context.Background(), dir, "probe")
			if err == nil || !strings.Contains(err.Error(), "output limit") {
				t.Fatalf("got %v", err)
			}
			if time.Since(start) > 3*time.Second {
				t.Fatal("overflow shutdown was not bounded")
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		dir := refreshGitScript(t, "sleep 30 &\nwait\n")
		start := time.Now()
		_, err := (ExecGit{Raw: true, Timeout: 30 * time.Millisecond}).Run(context.Background(), dir, "probe")
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("got %v", err)
		}
		if time.Since(start) > 3*time.Second {
			t.Fatal("cancellation shutdown was not bounded")
		}
	})
}

func TestExecGitDefaultCompatibility(t *testing.T) {
	dir := refreshGitScript(t, "printf '  first '; printf 'second  \\n' >&2\n")
	got, err := (ExecGit{}).Run(context.Background(), dir, "probe")
	if err != nil || got != "first second" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestExecGitRawInheritsExtraFiles(t *testing.T) {
	dir := refreshGitScript(t, "cat <&3\n")
	file, err := os.CreateTemp(t.TempDir(), "lease")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.WriteString("inherited")
	file.Seek(0, 0)
	got, err := (ExecGit{Raw: true, ExtraFiles: []*os.File{file}}).Run(context.Background(), dir, "probe")
	if err != nil || got != "inherited" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestExecGitRawOwnedRequiresStage(t *testing.T) {
	dir := refreshGitScript(t, "exit 0\n")
	_, err := (ExecGit{Raw: true}).RunOwned(context.Background(), dir, "", "probe")
	if err == nil {
		t.Fatal("owned operation accepted empty stage")
	}
}

func TestRestoreWithRawGitPaths(t *testing.T) {
	base := t.TempDir()
	source := origin(t, base, "layer", "", true)
	dest := filepath.Join(base, "clone with spaces ")
	client := Client{Git: ExecGit{Raw: true, Timeout: time.Second, MaxOutputBytes: 4096}}
	if err := client.Ensure(context.Background(), dest, source, true); err != nil {
		t.Fatal(err)
	}
	if err := client.Ensure(context.Background(), dest, source, true); err != nil {
		t.Fatal(err)
	}
}

func TestExecGitOutputLimitIncludesBothStreams(t *testing.T) {
	dir := refreshGitScript(t, "printf '12345678'; printf '12345678' >&2\n")
	_, err := (ExecGit{Raw: true, MaxOutputBytes: 12}).Run(context.Background(), dir, "status")
	if err == nil || !strings.Contains(err.Error(), "output limit") {
		t.Fatalf("got %v", err)
	}
}

func TestExecGitResourceFailureDoesNotExposePredicateExit(t *testing.T) {
	for _, raw := range []bool{false, true} {
		t.Run(fmt.Sprintf("raw=%t", raw), func(t *testing.T) {
			dir := refreshGitScript(t, "printf 'output-too-large'; exit 1\n")
			_, err := (ExecGit{Raw: raw, MaxOutputBytes: 4}).Run(context.Background(), dir, "config", "--get", "remote.origin.url")
			if err == nil || !strings.Contains(err.Error(), "output limit") {
				t.Fatalf("got %v", err)
			}
			var code interface{ ExitCode() int }
			if errors.As(err, &code) {
				t.Fatalf("resource error exposes predicate exit %d: %v", code.ExitCode(), err)
			}
		})
	}
}

func TestOriginRefusesOutputLimitOnMissingValue(t *testing.T) {
	dir := refreshGitScript(t, "case \"$*\" in *--list*) exit 0;; esac\nprintf 'output-too-large'; exit 1\n")
	origin, err := (Client{Git: ExecGit{Raw: true, MaxOutputBytes: 4}}).Origin(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "output limit") {
		t.Fatalf("origin=%q err=%v", origin, err)
	}
}

func TestExecGitCancellationDoesNotExposePredicateExit(t *testing.T) {
	dir := refreshGitScript(t, "sleep 30 &\nwait\n")
	_, err := (ExecGit{Raw: true, Timeout: 30 * time.Millisecond}).Run(context.Background(), dir, "config")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	var code interface{ ExitCode() int }
	if errors.As(err, &code) {
		t.Fatalf("cancellation exposes process exit %d: %v", code.ExitCode(), err)
	}
}

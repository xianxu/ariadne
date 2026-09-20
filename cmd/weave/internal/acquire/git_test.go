package acquire

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type exitFailure int

func (e exitFailure) Error() string { return "injected Git failure" }
func (e exitFailure) ExitCode() int { return int(e) }

type stateGit struct {
	calls [][]string
	run   func(int, string, []string) (string, error)
}

func (s *stateGit) Run(_ context.Context, dir string, args ...string) (string, error) {
	s.calls = append(s.calls, append([]string{}, args...))
	return s.run(len(s.calls), dir, args)
}
func (s *stateGit) RunOwned(ctx context.Context, dir, stage string, args ...string) (string, error) {
	return s.Run(ctx, dir, args...)
}

func TestOriginMissingVersusFailedInspection(t *testing.T) {
	for _, tc := range []struct {
		name      string
		failure   error
		wantError bool
	}{{"absent", exitFailure(1), false}, {"bad config", exitFailure(128), true}, {"missing git", errors.New("executable not found"), true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := &stateGit{run: func(n int, _ string, args []string) (string, error) {
				if n == 1 {
					return "", nil
				}
				return "", tc.failure
			}}
			_, err := (Client{Git: g}).Origin(context.Background(), t.TempDir())
			if (err != nil) != tc.wantError {
				t.Fatal(err)
			}
			if len(g.calls) != 2 || !reflect.DeepEqual(g.calls[0], []string{"config", "--local", "--list"}) {
				t.Fatal(g.calls)
			}
		})
	}
	g := &stateGit{run: func(_ int, _ string, _ []string) (string, error) { return "", exitFailure(128) }}
	if _, err := (Client{Git: g}).Origin(context.Background(), t.TempDir()); err == nil || len(g.calls) != 1 {
		t.Fatalf("%v %v", g.calls, err)
	}
}
func TestEnsurePublicationConflictAfterClone(t *testing.T) {
	base := t.TempDir()
	dest := filepath.Join(base, "peer")
	g := &stateGit{run: func(_ int, _ string, args []string) (string, error) {
		put(t, filepath.Join(args[len(args)-1], "construct/base.manifest"), "")
		put(t, filepath.Join(dest, "authored"), "keep")
		return "", nil
	}}
	err := (Client{Git: g}).Ensure(context.Background(), dest, "https://github.com/org/peer.git", true)
	if err == nil || !strings.Contains(err.Error(), "appeared") {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dest, "authored"))
	if string(b) != "keep" {
		t.Fatal(string(b))
	}
	matches, _ := filepath.Glob(filepath.Join(base, ".peer-weave-*"))
	if len(matches) != 0 {
		t.Fatal(matches)
	}
}
func TestRestoreInjectedCloneFailureStopsLaterSources(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "construct/deps"), "substrate ../base https://github.com/org/base.git\nsubstrate ../later https://github.com/org/later.git\n")
	g := &stateGit{run: func(_ int, _ string, args []string) (string, error) {
		put(t, filepath.Join(args[len(args)-1], "partial"), "partial")
		return "", errors.New("disconnected")
	}}
	if _, err := (Client{Git: g}).Restore(context.Background(), root, false); err == nil || len(g.calls) != 1 {
		t.Fatalf("calls=%v err=%v", g.calls, err)
	}
}

func TestOriginRealGitAbsentAndMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	gitFixture(t, dir, "init")
	if got, err := Origin(context.Background(), dir); err != nil || got != "" {
		t.Fatalf("%q %v", got, err)
	}
	put(t, filepath.Join(dir, ".git/config"), "[unterminated\n")
	if _, err := Origin(context.Background(), dir); err == nil {
		t.Fatal("malformed configuration treated as absent origin")
	}
}

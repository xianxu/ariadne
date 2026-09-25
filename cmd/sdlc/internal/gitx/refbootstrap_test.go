package gitx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestBootstrapCreatesOnlyOrphanSnapshot(t *testing.T) {
	repo, origin := trunkFixture(t, "unshipped source")
	before := testfix.Capture(t, repo, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(repo, "note.md"), []byte("staged code"), 0600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, repo, "add", "note.md")
	if err := os.WriteFile(filepath.Join(repo, "note.md"), []byte("unstaged code"), 0600); err != nil {
		t.Fatal(err)
	}
	indexBefore := testfix.Capture(t, repo, "write-tree")
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "issue-tracker")
	called := false
	result, err := tf.Bootstrap(map[string][]byte{"cards/1.md": []byte("card")}, "bootstrap token-one", func(r BootstrapResult) error {
		called = true
		if r.Candidate == "" || r.Outcome != BootstrapNotPublished {
			t.Fatalf("unprepared receipt: %+v", r)
		}
		if got := testfix.Capture(t, origin, "for-each-ref", "refs/heads/issue-tracker"); got != "" {
			t.Fatalf("published before receipt: %s", got)
		}
		return nil
	})
	if err != nil || result.Outcome != BootstrapCreated || !called {
		t.Fatalf("result %+v, err %v", result, err)
	}
	if got := testfix.Capture(t, origin, "show", "issue-tracker:cards/1.md"); got != "card" {
		t.Fatal(got)
	}
	if got := strings.TrimSpace(testfix.Capture(t, origin, "ls-tree", "-r", "--name-only", "issue-tracker")); got != "cards/1.md" {
		t.Fatal(got)
	}
	if got := testfix.Capture(t, origin, "rev-list", "--parents", "-1", "issue-tracker"); strings.Fields(got)[0] != result.Candidate || len(strings.Fields(got)) != 1 {
		t.Fatalf("not orphan: %s", got)
	}
	if got := testfix.Capture(t, repo, "rev-parse", "HEAD"); got != before {
		t.Fatal("changed caller HEAD")
	}
	if got := testfix.Capture(t, repo, "write-tree"); got != indexBefore {
		t.Fatal("changed caller index")
	}
	if got, err := os.ReadFile(filepath.Join(repo, "note.md")); err != nil || string(got) != "unstaged code" {
		t.Fatalf("changed caller worktree: %q %v", got, err)
	}
	_, err = tf.Bootstrap(map[string][]byte{"other": []byte("x")}, "second", func(BootstrapResult) error { t.Fatal("existing ref reached candidate callback"); return nil })
	if !errors.Is(err, ErrBootstrapExists) {
		t.Fatalf("existing branch: %v", err)
	}
}

func TestBootstrapExpectedAbsenceRace(t *testing.T) {
	repo, origin := trunkFixture(t, "source")
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "my token", func(BootstrapResult) error {
		winner, winErr := tf.Bootstrap(map[string][]byte{"card": []byte("peer")}, "peer token", func(BootstrapResult) error { return nil })
		if winErr != nil || winner.Outcome != BootstrapCreated {
			t.Fatalf("peer %+v %v", winner, winErr)
		}
		return nil
	})
	if !errors.Is(err, ErrBootstrapRejected) || result.Outcome != BootstrapRejected {
		t.Fatalf("lost race: %+v %v", result, err)
	}
	if got := testfix.Capture(t, origin, "show", "issue-tracker:card"); got != "peer" {
		t.Fatalf("overwrote peer: %s", got)
	}
}

func TestBootstrapLostAcknowledgmentNeverClaimsOwnership(t *testing.T) {
	repo, origin := trunkFixture(t, "source")
	real := runGitInContext
	t.Cleanup(func() { runGitInContext = real })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		out, diag, err := real(ctx, dir, env, args...)
		if len(args) > 0 && args[0] == "push" && err == nil {
			return nil, nil, errors.New("lost acknowledgment")
		}
		return out, diag, err
	}
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "unique token", func(BootstrapResult) error { return nil })
	if !errors.Is(err, ErrPublicationUncertain) || result.Outcome != BootstrapUncertain {
		t.Fatalf("unknown: %+v %v", result, err)
	}
	if got := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "issue-tracker")); got != result.Candidate {
		t.Fatalf("push did not happen: %s", got)
	}
}

func TestBootstrapIdenticalPeerCandidateDoesNotWin(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "token", func(r BootstrapResult) error {
		testfix.Git(t, repo, "push", "origin", r.Candidate+":refs/heads/issue-tracker")
		return nil
	})
	if !errors.Is(err, ErrBootstrapRejected) || result.Outcome != BootstrapRejected {
		t.Fatalf("identical candidate falsely won: %+v %v", result, err)
	}
}

func TestBootstrapCancellationDuringPushIsUncertain(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	real := runGitInContext
	t.Cleanup(func() { runGitInContext = real })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		out, diag, err := real(ctx, dir, env, args...)
		if len(args) > 0 && args[0] == "push" {
			cancel()
		}
		return out, diag, err
	}
	tf, _ := NewTrunkFileContext(ctx, repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "unique token", func(BootstrapResult) error { return nil })
	if !errors.Is(err, ErrPublicationUncertain) || result.Outcome != BootstrapUncertain {
		t.Fatalf("cancelled publication: %+v %v", result, err)
	}
}

func TestBootstrapPathValidation(t *testing.T) {
	for _, p := range []string{"", ".", "/absolute", "../escape", "a/../b", "a//b", "a/", ".git/config", "a/.GIT/config", "a\\b", "a\x00b", ":(glob)*", "*", "a\nb"} {
		if _, err := bootstrapPaths(map[string][]byte{p: nil}); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
	if _, err := bootstrapPaths(map[string][]byte{"a": nil, "a-b": nil, "a/b": nil}); err == nil {
		t.Fatal("accepted overlapping paths")
	}
	if _, err := bootstrapPaths(map[string][]byte{"cards/one two.md": nil}); err != nil {
		t.Fatal(err)
	}
}

func TestBootstrapReadFailureIsNotAbsence(t *testing.T) {
	repo, origin := trunkFixture(t, "source")
	real := runGitInContext
	t.Cleanup(func() { runGitInContext = real })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "ls-remote" {
			return nil, nil, errors.New("remote inspection failed")
		}
		return real(ctx, dir, env, args...)
	}
	tf, _ := NewTrunkFileContext(context.Background(), repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "unique token", func(BootstrapResult) error { t.Fatal("read failure reached publish preparation"); return nil })
	if err == nil || result.Outcome != BootstrapNotPublished {
		t.Fatalf("read failure: %+v %v", result, err)
	}
	if got := testfix.Capture(t, origin, "for-each-ref", "refs/heads/issue-tracker"); got != "" {
		t.Fatalf("published: %s", got)
	}
}

func TestBootstrapCancelledPushDoesNotTrustPartialRejection(t *testing.T) {
	repo, _ := trunkFixture(t, "source")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	real := runGitInContext
	t.Cleanup(func() { runGitInContext = real })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == "push" {
			cancel()
			return []byte("!\tlocal:refs/heads/issue-tracker\t[remote rejected] (ref lock)\n"), nil, ctx.Err()
		}
		return real(ctx, dir, env, args...)
	}
	tf, _ := NewTrunkFileContext(ctx, repo, "origin", "issue-tracker")
	result, err := tf.Bootstrap(map[string][]byte{"card": []byte("mine")}, "unique token", func(BootstrapResult) error { return nil })
	if !errors.Is(err, ErrPublicationUncertain) || result.Outcome != BootstrapUncertain {
		t.Fatalf("trusted cancelled query: %+v %v", result, err)
	}
}

func TestBootstrapRefusesBeforePublication(t *testing.T) {
	for _, scenario := range []string{"cancel", "receipt", "path", "empty", "nil callback"} {
		t.Run(scenario, func(t *testing.T) {
			repo, origin := trunkFixture(t, "source")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tf, _ := NewTrunkFileContext(ctx, repo, "origin", "issue-tracker")
			files := map[string][]byte{"card": []byte("mine")}
			callback := func(BootstrapResult) error {
				if scenario == "cancel" {
					cancel()
				}
				if scenario == "receipt" {
					return errors.New("cannot save receipt")
				}
				return nil
			}
			if scenario == "path" {
				files = map[string][]byte{"../escape": []byte("bad")}
			}
			if scenario == "empty" {
				files = nil
			}
			if scenario == "nil callback" {
				callback = nil
			}
			result, err := tf.Bootstrap(files, "unique token", callback)
			if err == nil || result.Outcome != BootstrapNotPublished {
				t.Fatalf("expected safe refusal: %+v %v", result, err)
			}
			if got := testfix.Capture(t, origin, "for-each-ref", "refs/heads/issue-tracker"); got != "" {
				t.Fatalf("published: %s", got)
			}
		})
	}
}

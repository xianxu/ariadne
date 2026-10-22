package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestQueueHeadOrdinaryMainIntegrationRealGit(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "checkout", "-qb", "000010-work")
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#10: own change")
	first := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	main := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", base, "-m", "#99: unrelated main"))
	testfix.Git(t, root, "update-ref", "refs/remotes/origin/main", main)
	testfix.Git(t, root, "merge", "--no-edit", "origin/main")
	merge := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	subject := strings.TrimSpace(testfix.Capture(t, root, "show", "-s", "--format=%s", merge))
	if !strings.HasPrefix(subject, "Merge remote-tracking branch 'origin/main'") {
		t.Fatalf("fixture did not retain Git's default subject: %q", subject)
	}
	got, err := readQueueHeadMemberships(t.Context(), []string{merge}, "ariadne", gitQueueHeadReader(root))
	if err != nil || !got[merge].Known || fmt.Sprint(got[merge].Issues) != "[10]" || got[merge].Train != 0 {
		t.Fatalf("default merge lost first-parent issue %s or borrowed main's issue: %+v err=%v", first, got, err)
	}
}

func TestQueueHeadAncestryProbeIsolationAndCache(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	commit := func(message string, parents ...string) string {
		args := []string{"commit-tree", tree, "-m", message}
		for _, p := range parents {
			args = append(args, "-p", p)
		}
		return strings.TrimSpace(testfix.Capture(t, root, args...))
	}
	own := commit("#10: own", base)
	main := commit("#99: main", base)
	first := commit("first integration", own, main)
	second := commit("second integration", own, main)
	older := commit("older main integration", own, base)
	testfix.Git(t, root, "update-ref", "refs/remotes/upstream/main", main)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	counter := filepath.Join(bin, "counter")
	wrapper := `#!/bin/sh
for arg in "$@"; do
  case "$arg" in
    cat-file) [ "$GIT_NO_LAZY_FETCH" = 1 ] || exit 73 ;;
    rev-parse) printf 'pin\n' >> "$QUEUE_PROBE_COUNTER" ;;
    merge-base)
      [ "$GIT_NO_LAZY_FETCH" = 1 ] && [ "$GIT_GRAFT_FILE" = /dev/null ] && [ "$GIT_SHALLOW_FILE" = /dev/null ] || exit 74
      printf 'probe\n' >> "$QUEUE_PROBE_COUNTER" ;;
  esac
done
exec "$QUEUE_REAL_GIT" "$@"
`
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(wrapper), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("QUEUE_REAL_GIT", realGit)
	t.Setenv("QUEUE_PROBE_COUNTER", counter)
	t.Setenv("GIT_NO_LAZY_FETCH", "")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	reader := gitQueueHeadReader(root, "refs/remotes/upstream/main")
	for i := 0; i < 2; i++ {
		got, err := readQueueHeadMemberships(t.Context(), []string{first, second}, "ariadne", reader)
		if err != nil || !got[first].Known || !got[second].Known {
			t.Fatalf("shared graph: %+v err=%v", got, err)
		}
	}
	raw, err := os.ReadFile(counter)
	if err != nil || string(raw) != "pin\nprobe\n" {
		t.Fatalf("proof must pin once and probe each shared side parent once: %q err=%v", raw, err)
	}
	old := defaultQueueHeadLimits
	defaultQueueHeadLimits.Nodes = 1
	t.Cleanup(func() { defaultQueueHeadLimits = old })
	limited := gitQueueHeadReader(root, "refs/remotes/upstream/main")
	if _, err := limited(t.Context(), []string{first}); err != nil {
		t.Fatal(err)
	}
	if got, err := limited(t.Context(), []string{older}); err == nil || !strings.Contains(err.Error(), "node safety budget") || got != nil {
		t.Fatalf("ancestry cache overflow returned partial evidence: %+v err=%v", got, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if got, err := reader(ctx, []string{first}); err == nil || got != nil {
		t.Fatalf("cancelled pass returned evidence: %+v err=%v", got, err)
	}
}

func TestQueueMainAncestorSeparatesNegativeAndFailure(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	foreign := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", base, "-m", "foreign"))
	if yes, err := queueMainAncestor(t.Context(), root, base, foreign); err != nil || !yes {
		t.Fatalf("real ancestor: %v %v", yes, err)
	}
	if yes, err := queueMainAncestor(t.Context(), root, foreign, base); err != nil || yes {
		t.Fatalf("non-ancestor must be confirmed false: %v %v", yes, err)
	}
	if yes, err := queueMainAncestor(t.Context(), root, strings.Repeat("d", 40), base); err == nil || yes {
		t.Fatalf("missing object became a negative proof: %v %v", yes, err)
	}
}

func TestQueueHeadStructuralHistoryOverrides(t *testing.T) {
	for _, kind := range []string{"replace", "graft", "shallow"} {
		t.Run(kind, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
			tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
			commit := func(message string, parents ...string) string {
				args := []string{"commit-tree", tree, "-m", message}
				for _, p := range parents {
					args = append(args, "-p", p)
				}
				return strings.TrimSpace(testfix.Capture(t, root, args...))
			}
			own := commit("#10: own", base)
			main := commit("#99: main", base)
			foreign := commit("#20: foreign", base)
			head := commit("arbitrary merge", own, foreign)
			testfix.Git(t, root, "update-ref", "refs/remotes/origin/main", main)
			switch kind {
			case "replace":
				fakeMain := commit("forged history", foreign)
				testfix.Git(t, root, "replace", main, fakeMain)
				testfix.Git(t, root, "merge-base", "--is-ancestor", foreign, main)
			case "graft":
				if err := os.WriteFile(filepath.Join(root, ".git", "info", "grafts"), []byte(main+" "+foreign+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
				testfix.Git(t, root, "merge-base", "--is-ancestor", foreign, main)
			case "shallow":
				head = commit("integrate an older main", own, base)
				if err := os.WriteFile(filepath.Join(root, ".git", "shallow"), []byte(main+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
				var exit *exec.ExitError
				if err := exec.CommandContext(t.Context(), "git", "-C", root, "merge-base", "--is-ancestor", base, main).Run(); !errors.As(err, &exit) || exit.ExitCode() != 1 {
					t.Fatalf("shallow control did not hide the real ancestor: %v", err)
				}
			}
			got, err := readQueueHeadMemberships(t.Context(), []string{head}, "ariadne", gitQueueHeadReader(root))
			if err != nil {
				t.Fatal(err)
			}
			if kind == "shallow" {
				if !got[head].Known || fmt.Sprint(got[head].Issues) != "[10]" {
					t.Fatalf("raw history lost to shallow override: %+v", got)
				}
			} else if got[head].Known {
				t.Fatalf("forged ancestry accepted: %+v", got)
			}
		})
	}
}

func TestQueueHeadMissingMainIsNotAnAncestryProof(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	own := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", base, "-m", "#10: own"))
	main := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", base, "-m", "#99: main"))
	head := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", own, "-p", main, "-m", "unattributed integration"))
	reader := gitQueueHeadReader(root, "refs/remotes/missing/main")
	if got, err := readQueueHeadMemberships(t.Context(), []string{head}, "ariadne", reader); err == nil || got != nil || !strings.Contains(err.Error(), "cannot pin publication main") {
		t.Fatalf("missing pin returned evidence: %+v err=%v", got, err)
	}
	got, err := readQueueHeadMemberships(t.Context(), []string{own}, "ariadne", gitQueueHeadReader(root, "refs/remotes/missing/main"))
	if err != nil || !got[own].Known {
		t.Fatalf("unneeded main pin poisoned direct membership: %+v err=%v", got, err)
	}
}

func TestQueueHeadStructuralIntegrationControls(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	tree := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD^{tree}"))
	commit := func(message string, parents ...string) string {
		args := []string{"commit-tree", tree, "-m", message}
		for _, parent := range parents {
			args = append(args, "-p", parent)
		}
		return strings.TrimSpace(testfix.Capture(t, root, args...))
	}
	own := commit("#10: own", base)
	main := commit("#99: main", base)
	foreign := commit("#20: foreign", base)
	unknown := commit("unattributed", base)
	testfix.Git(t, root, "update-ref", "refs/remotes/origin/main", main)
	nested := commit("anything at all", own, main)
	for _, tc := range []struct {
		name, head, issues string
	}{
		{"arbitrary-message", nested, "[10]"},
		{"nested", commit("another arbitrary message", nested, main), "[10]"},
		{"octopus", commit("three parents", own, base, main), "[10]"},
		{"older-main", commit("old main", own, base), "[10]"},
		{"foreign-parent", commit("Merge remote-tracking branch 'origin/main'", own, foreign), ""},
		{"mixed-octopus", commit("multiple sources", own, main, foreign), ""},
		{"unknown-first", commit("main does not supply ownership", unknown, main), ""},
		{"ordinary-single-parent", commit("unattributed descendant", own), ""},
		{"malformed-train", commit("Landing train 123: #20 (PR #12)", own, main), ""},
		{"train-car-precedence", commit("Landing train 123: #20 (PR #12)\n\nLanding-Train: 123\nLanding-PR: 12", own, main), "[20]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readQueueHeadMemberships(t.Context(), []string{tc.head}, "ariadne", gitQueueHeadReader(root))
			if err != nil {
				t.Fatal(err)
			}
			m := got[tc.head]
			if tc.issues == "" {
				if m.Known {
					t.Fatalf("unproved integration accepted: %+v", m)
				}
			} else if !m.Known || fmt.Sprint(m.Issues) != tc.issues {
				t.Fatalf("membership=%+v want %s", m, tc.issues)
			}
		})
	}
	t.Run("one-main-generation", func(t *testing.T) {
		reader := gitQueueHeadReader(root)
		if _, err := reader(t.Context(), []string{nested}); err != nil {
			t.Fatal(err)
		}
		testfix.Git(t, root, "update-ref", "refs/remotes/origin/main", foreign)
		head := commit("late foreign integration", own, foreign)
		got, err := readQueueHeadMemberships(t.Context(), []string{head}, "ariadne", reader)
		if err != nil || got[head].Known {
			t.Fatalf("moving main changed an existing pass: %+v err=%v", got, err)
		}
		fresh, err := readQueueHeadMemberships(t.Context(), []string{head}, "ariadne", gitQueueHeadReader(root))
		if err != nil || !fresh[head].Known || fmt.Sprint(fresh[head].Issues) != "[10]" {
			t.Fatalf("next pass did not observe its own pin: %+v err=%v", fresh, err)
		}
	})
}

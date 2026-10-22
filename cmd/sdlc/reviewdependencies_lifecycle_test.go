package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func reviewDependencySourceFixture(t *testing.T) (string, string) {
	t.Helper()
	src := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, src, "construct/base.manifest", "# fixture substrate\n")
	writeRepoFile(t, src, "fixture", "published\n")
	testfix.Git(t, src, "add", ".")
	testfix.Git(t, src, "commit", "-qm", "dependency fixture")
	head := strings.TrimSpace(testfix.Capture(t, src, "rev-parse", "HEAD"))
	testfix.Git(t, src, "update-ref", "refs/remotes/origin/main", head)
	return src, head
}

func reviewDependencyConsumerFixture(t *testing.T, declaration string) judge.ReviewWindowManifest {
	t.Helper()
	repo := testfix.Repo(t, testfix.InitialCommit())
	if declaration != "" {
		writeRepoFile(t, repo, "construct/deps", declaration)
		testfix.Git(t, repo, "add", "construct/deps")
		testfix.Git(t, repo, "commit", "-qm", "declare fixture dependencies")
	}
	head := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
	return judge.ReviewWindowManifest{RepoRoot: repo, HeadSHA: head}
}

func TestReviewDependenciesSetupFailureCleansOwnedClones(t *testing.T) {
	source, _ := reviewDependencySourceFixture(t)
	other, _ := reviewDependencySourceFixture(t)
	noRef, _ := reviewDependencySourceFixture(t)
	testfix.Git(t, noRef, "update-ref", "-d", "refs/remotes/origin/main")
	for _, tc := range []struct{ name, declaration, want string }{
		{"malformed", "substrate\n", "construct/deps"},
		{"conflicting", "substrate " + source + "\nsubstrate " + other + "\n", "conflicting"},
		{"missing-ref", "substrate " + noRef + "\n", "origin/main"},
		{"missing-source", "substrate " + filepath.Join(t.TempDir(), "absent") + "\n", "dependency"},
		// The substrate has already been cloned when the absent data mount fails.
		{"failure-after-first-child", "substrate " + source + "\ndata https://invalid.example/data.git missing-data\n", "dependency"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
			manifest := reviewDependencyConsumerFixture(t, tc.declaration)
			owner, _, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), manifest)
			if owner != nil {
				defer owner.Close()
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("setup error=%v; want %q", err, tc.want)
			}
			entries, readErr := os.ReadDir(root)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("failed setup leaked clones or locks: %v %v", entries, readErr)
			}
		})
	}
}

func TestReviewDependenciesMissingDeclarationIsEmpty(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
	owner, iso, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), reviewDependencyConsumerFixture(t, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if len(iso.Dependencies) != 0 {
		t.Fatalf("invented dependencies: %+v", iso.Dependencies)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("no-dependency layout=%v %v", entries, err)
	}
}

func TestReviewDependenciesResolveSourceLessSubstrateAndDirectDataMount(t *testing.T) {
	source, sourceHead := reviewDependencySourceFixture(t)
	data, dataHead := reviewDependencySourceFixture(t)
	manifest := reviewDependencyConsumerFixture(t, "substrate "+source+"\ndata https://never-contact.invalid/data.git data/input\n")
	mounted := filepath.Join(manifest.RepoRoot, "data/input")
	if err := os.MkdirAll(filepath.Dir(mounted), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(data, mounted); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
	owner, iso, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), manifest)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if len(iso.Dependencies) != 2 {
		t.Fatalf("bindings=%+v", iso.Dependencies)
	}
	expected := map[string]string{"substrate": sourceHead, "data": dataHead}
	for _, dep := range iso.Dependencies {
		if dep.Head != expected[dep.Kind] || dep.Dir == source || dep.Dir == mounted {
			t.Fatalf("wrong disposable binding: %+v", dep)
		}
		if got := strings.TrimSpace(testfix.Capture(t, dep.Dir, "rev-parse", "HEAD")); got != dep.Head {
			t.Fatalf("wrong checkout pin: %s != %s", got, dep.Head)
		}
	}
}

func TestReviewDependenciesDedupeAndRetainPinAcrossReset(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			source, pin := reviewDependencySourceFixture(t)
			manifest := reviewDependencyConsumerFixture(t, strings.Repeat("substrate "+source+"\n", n))
			root := t.TempDir()
			t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
			owner, iso, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), manifest)
			if err != nil {
				t.Fatal(err)
			}
			defer owner.Close()
			if len(iso.Dependencies) != 1 {
				t.Fatalf("%d rows became %d bindings", n, len(iso.Dependencies))
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 4 {
				t.Fatalf("want primary+one child+two locks: %v %v", entries, err)
			}
			child := iso.Dependencies[0]
			writeRepoFile(t, source, "fixture", "later fetched main\n")
			testfix.Git(t, source, "commit", "-qam", "advance source main")
			testfix.Git(t, source, "update-ref", "refs/remotes/origin/main", "HEAD")
			writeRepoFile(t, child.Dir, "fixture", "reviewer mutation\n")
			if err := owner.ResetContext(t.Context()); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(testfix.Capture(t, child.Dir, "rev-parse", "HEAD")); got != pin {
				t.Fatalf("reset repinned dependency: %s != %s", got, pin)
			}
			got, err := os.ReadFile(filepath.Join(child.Dir, "fixture"))
			if err != nil || string(got) != "published\n" {
				t.Fatalf("reset bytes=%q %v", got, err)
			}
			if iso.Dependencies[0].Dir != child.Dir || iso.Dependencies[0].Head != pin {
				t.Fatal("reset changed descriptor")
			}
			if err := owner.Close(); err != nil {
				t.Fatal(err)
			}
			entries, err = os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("cleanup=%v %v", entries, err)
			}
		})
	}
}

func TestReviewDependenciesProtectLinkedSourceCommonAndClone(t *testing.T) {
	primary, pin := reviewDependencySourceFixture(t)
	linked := filepath.Join(t.TempDir(), "linked")
	testfix.Git(t, primary, "worktree", "add", "--quiet", "--detach", linked, pin)
	manifest := reviewDependencyConsumerFixture(t, "substrate "+linked+"\n")
	root := t.TempDir()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
	owner, iso, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), manifest)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	for _, path := range []string{linked, primary, filepath.Join(primary, ".git"), iso.Dependencies[0].Dir} {
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, p := range iso.Protect {
			if p == canonical {
				found = true
			}
		}
		if !found {
			t.Errorf("fence omits %s: %v", canonical, iso.Protect)
		}
	}
	// Inspect the actual fence argv with the resolved paths rather than invoking
	// an OS fence that may be unavailable inside an agent's enclosing sandbox.
	iso.Fence = reviewbox.FenceBwrap
	_, args := iso.Wrap("reviewer", nil)
	for _, path := range iso.Protect {
		found := false
		for i := 0; i+2 < len(args); i++ {
			if args[i] == "--ro-bind" && args[i+1] == path && args[i+2] == path {
				found = true
			}
		}
		if !found {
			t.Errorf("fence argv omits read-only binding %s", path)
		}
	}
}

func TestReviewDependenciesPrepareCancellationCleansWithOwner(t *testing.T) {
	source, _ := reviewDependencySourceFixture(t)
	manifest := reviewDependencyConsumerFixture(t, "substrate "+source+"\n")
	root := t.TempDir()
	owner, err := reviewbox.CreateContext(t.Context(), root, manifest.RepoRoot, manifest.HeadSHA, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	tools := t.TempDir()
	ready := filepath.Join(tools, "git-child.pid")
	writeRepoFile(t, tools, "git", "#!/bin/sh\nsleep 60 &\nprintf '%s' \"$!\" > \"$REVIEW_PREPARE_READY\"\nwait\n")
	if err := os.Chmod(filepath.Join(tools, "git"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REVIEW_PREPARE_READY", ready)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, _, err := prepareReviewDependencies(ctx, owner, manifest.RepoRoot); done <- err }()
	pid := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
	if pid == 0 {
		t.Fatal("preparation never entered Git")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("canceled setup leaked: %v %v", entries, err)
	}
}

func TestReviewFetchAgeMatchesOnlySingleOriginMainObservation(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	head := strings.Repeat("a", 40)
	other := strings.Repeat("b", 40)
	origin := "https://example.invalid/base.git"
	line := head + "\t\tbranch 'main' of " + origin + "\n"
	for _, tc := range []struct {
		name, raw, source string
		modified          time.Time
		known             bool
	}{
		{"matching", line, origin, now.Add(-90 * time.Second), true},
		{"matching-without-dotgit", strings.ReplaceAll(line, origin, strings.TrimSuffix(origin, ".git")), origin, now.Add(-90 * time.Second), true},
		{"unrelated-head", strings.ReplaceAll(line, head, other), origin, now.Add(-90 * time.Second), false},
		{"unrelated-origin", line, "https://elsewhere.invalid/base.git", now.Add(-90 * time.Second), false},
		{"multi-entry", line + line, origin, now.Add(-90 * time.Second), false},
		{"future", line, origin, now.Add(time.Second), false},
		{"missing-origin", line, "", now.Add(-time.Second), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := reviewFetchAge(tc.raw, head, tc.source, tc.modified, now)
			if tc.known {
				if !strings.Contains(got, "1m30s") || !strings.Contains(got, "remote freshness unverified") {
					t.Fatalf("known observation=%q", got)
				}
			} else if !strings.HasPrefix(got, "unknown") {
				t.Fatalf("unproven age=%q", got)
			}
		})
	}
}

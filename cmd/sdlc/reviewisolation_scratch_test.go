//go:build unix

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// A RAM disk that refuses the write (an agent sandbox) costs the review its
// RAM, never the review: the checkout falls back to the cache root (#468).
func TestReviewIsolationFallsBackFromScratch(t *testing.T) {
	home := t.TempDir() // os.UserCacheDir follows HOME, so no real cache is touched
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", "")
	ram := t.TempDir()
	if err := os.Chmod(ram, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ram, 0o700) })
	t.Setenv("WF_TEST_SCRATCH", ram)

	repo := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "x"}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	head, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	checkout, _, err := openReviewIsolation(&stderr, judge.ReviewWindowManifest{RepoRoot: repo, HeadSHA: strings.TrimSpace(string(head))})
	if err != nil {
		t.Fatalf("review isolation: %v\n%s", err, stderr.String())
	}
	defer checkout.Close()
	cache, _ := os.UserCacheDir()
	want, _ := filepath.EvalSymlinks(filepath.Join(cache, "sdlc", "review-checkouts"))
	if filepath.Dir(checkout.Dir) != want || !strings.Contains(stderr.String(), "using") {
		t.Fatalf("checkout at %s, want under %s; stderr:\n%s", checkout.Dir, want, stderr.String())
	}
}

// Only a refused write falls back: a review that fails for its own reasons (a
// head the source doesn't have) fails, on the RAM disk, with no fallback.
func TestReviewIsolationKeepsOtherErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", "")
	t.Setenv("WF_TEST_SCRATCH", t.TempDir())
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	var stderr bytes.Buffer
	_, _, err := openReviewIsolation(&stderr, judge.ReviewWindowManifest{RepoRoot: repo, HeadSHA: strings.Repeat("0", 40)})
	if err == nil {
		t.Fatal("a missing head opened a review checkout")
	}
	if strings.Contains(stderr.String(), "refused the review checkout") {
		t.Fatalf("fell back on a non-write error:\n%s", stderr.String())
	}
}

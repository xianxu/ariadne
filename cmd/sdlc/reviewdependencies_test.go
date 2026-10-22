package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Echo #9's real BR-1: a reviewer could not inspect the pinned upstream
// harness because the disposable consumer had no sibling dependency.
func TestBoundaryReviewProvidesPinnedDependency(t *testing.T) {
	for _, milestone := range []string{"M1", ""} {
		t.Run("boundary-"+milestone, func(t *testing.T) { testBoundaryPinnedDependency(t, milestone) })
	}
}

func testBoundaryPinnedDependency(t *testing.T, milestone string) {
	capturedFinding, err := os.ReadFile("testdata/review-dependency-echo9.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	t.Setenv("SDLC_DEP_SUBSTRATE", "/stale-dependency")
	t.Setenv("SDLC_DEP_OLD", "must be cleared")
	p := boundaryWindow(t)
	p.Milestone = milestone
	if milestone == "" {
		p.Label = "#69"
	}
	source := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, source, "construct/base.manifest", "# substrate\n")
	writeRepoFile(t, source, "fixture", "published fixture\n")
	testfix.Git(t, source, "add", ".")
	testfix.Git(t, source, "commit", "-qm", "published dependency")
	pin := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "HEAD"))
	testfix.Git(t, source, "update-ref", "refs/remotes/origin/main", pin)
	writeRepoFile(t, source, "fixture", "unpublished local branch\n")
	testfix.Git(t, source, "commit", "-qam", "private dependency")
	writeRepoFile(t, source, "fixture", "dirty local work\n")
	writeRepoFile(t, ".", "construct/deps", "substrate "+source+"\n")
	git(t, "", "add", "construct/deps")
	git(t, "", "commit", "-qm", "#69 M1: declare substrate")
	p.Head = strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	// The declaration must come from the reviewed commit, not caller edits.
	writeRepoFile(t, ".", "construct/deps", "invalid caller edit\n")

	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	var dependency string
	judge.Run = func(ctx context.Context, onStart func(int), name string, args ...string) (judge.ProcessOutput, error) {
		_, agentArgs, _ := unwrapReviewArgv(t, name, args)
		prompt := agentArgs[len(agentArgs)-1]
		match := regexp.MustCompile(`(?m)^SDLC_DEP_SUBSTRATE=(.+)$`).FindStringSubmatch(prompt)
		if len(match) != 2 {
			t.Errorf("missing disposable SDLC_DEP_SUBSTRATE binding; original live finding:\n%s", capturedFinding)
			return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
		}
		var err error
		dependency, err = strconv.Unquote(match[1])
		if err != nil || dependency == source {
			t.Fatalf("bad disposable dependency path %q: %v", dependency, err)
		}
		if got, err := os.ReadFile(filepath.Join(dependency, "fixture")); err != nil || string(got) != "published fixture\n" {
			t.Errorf("reviewer dependency bytes %q, %v", got, err)
		}
		if !strings.Contains(prompt, pin) || !strings.Contains(prompt, "fetch age") || !strings.Contains(prompt, "disposable") {
			t.Error("dependency context omitted pin, freshness or mode")
		}
		// Exercise the actual process environment, not only a prompt projection.
		command, err := orig(ctx, onStart, "sh", "-c", `test -z "${SDLC_DEP_OLD+x}" && cat "$SDLC_DEP_SUBSTRATE/fixture"`)
		if err != nil || string(command.Stdout) != "published fixture\n" {
			t.Errorf("reviewer command cannot read the exported pinned fixture: %+v, %v", command, err)
		}
		return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
	}
	var out, errout bytes.Buffer
	r := dispatchBoundaryReview(&out, &errout, p)
	if r.Verdict != judge.VerdictShip {
		t.Fatalf("review failed: %+v\n%s", r, errout.String())
	}
	if dependency != "" && (!strings.Contains(errout.String(), "SDLC_DEP_SUBSTRATE="+strconv.Quote(dependency)) || !strings.Contains(errout.String(), pin)) {
		t.Errorf("close output omitted the actual disposable dependency manifest: %s", errout.String())
	}
	if dependency != "" {
		if _, err := os.Stat(dependency); !os.IsNotExist(err) {
			t.Errorf("dependency outlived the review: %v", err)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(source, "fixture")); string(got) != "dirty local work\n" {
		t.Error("review reached live dependency bytes")
	}
	if entries := reviewRootEntries(t); len(entries) != 0 {
		t.Errorf("review left %d scratch entries", len(entries))
	}
}

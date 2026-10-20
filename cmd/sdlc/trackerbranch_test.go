package main

import (
	"context"
	"strings"
	"testing"
)

func TestRequireIssueBranch(t *testing.T) {
	const path = "workshop/issues/000009-s09.md"
	for _, tc := range []struct {
		name, current, named string
		wantError            bool
	}{
		{"correct", "000009-s09", "", false},
		{"other issue", "000010-other", "000010-other", true},
		{"rest", "main-slot1", "main-slot1", true},
		{"detached", "", "a detached HEAD", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := requireIssueBranch("000009", path, tc.current)
			if !tc.wantError {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "000009-s09") || !strings.Contains(err.Error(), tc.named) {
				t.Fatalf("wrong branch error = %v", err)
			}
		})
	}
}

func TestRequireCloseIssueBranchRenamedAtTheIssueTip(t *testing.T) {
	r, paths, _ := startedHere(t)
	want := r.git("rev-parse", "HEAD")
	r.git("switch", "-q", "-c", "000009-s09-close")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := requireCloseIssueBranch(env, "000009", paths["000009"]); err != nil {
		t.Fatalf("same issue tip under a recovery name: %v", err)
	}
	r.git("commit", "-qm", "#10: unrelated work", "--allow-empty")
	env, err = openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := requireCloseIssueBranch(env, "000009", paths["000009"]); err == nil || !strings.Contains(err.Error(), "000009-s09-close") {
		t.Fatalf("different head on renamed branch must refuse: %v", err)
	}
	if got := r.git("rev-parse", "refs/heads/"+s09Branch); got != want {
		t.Fatalf("recovery probe moved the issue branch to %s", got)
	}
}

func TestRequireCloseIssueBranchUsesRemoteOnlyWithoutLocalRef(t *testing.T) {
	r, paths, _ := startedHere(t)
	r.git("push", "-q", "origin", "HEAD:refs/heads/"+s09Branch)
	r.git("switch", "-q", "-c", "000009-s09-close")
	r.git("branch", "-q", "-D", s09Branch)
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := requireCloseIssueBranch(env, "000009", paths["000009"]); err != nil {
		t.Fatalf("renamed branch at the fresh remote issue tip: %v", err)
	}
}

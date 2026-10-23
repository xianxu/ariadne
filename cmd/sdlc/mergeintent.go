package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// Merge intents (#565): the running sdlc merge holds one per branch, so a
// full shutdown that kills it leaves an orphan that status surfaces print
// with the one command that re-enters the landing. Re-entry needs nothing
// more: sdlc merge --branch resumes from Git and GitHub evidence, and the
// #332 gate reuses a passing recorded suite on the unchanged head.

var issueBranchNumber = regexp.MustCompile(`^0*([0-9]+)-`)

// holdMergeIntent records branch's in-flight merge before any landing work.
// release takes the merge's result: an interruption keeps the intent as an
// orphan, any other outcome removes it. A branch with no local head has
// nothing to re-enter, so it records nothing.
func holdMergeIntent(ctx context.Context, r gitRunner, t landingTarget, requested string, stderr io.Writer) (func(error), error) {
	none := func(error) {}
	branch, _, err := landingIssueBranch(r, t, requested)
	if err != nil {
		return none, nil // the attempt refuses with the same error
	}
	lineDir, err := landingLineDir(r, t.Root)
	if err != nil {
		return none, err
	}
	store := landingline.OpenIntents(lineDir)
	sweepMergeIntents(r, t.Root, store, stderr)
	head, err := landingOptional(r, t.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil || head == "" {
		return none, err
	}
	in := landingline.Intent{Branch: branch, Head: head, Root: canonicalRoot(t.Root), PID: os.Getpid(), StartedAt: time.Now().UTC()}
	if m := issueBranchNumber.FindStringSubmatch(branch); m != nil {
		in.Issue, _ = strconv.Atoi(m[1])
	}
	if identity, err := workspace.Resolve(queueGitReader{ctx}, t.Root, ""); err == nil && identity.Address != nil {
		in.Workspace = *identity.Address
	}
	hold, err := store.Hold(in)
	if err != nil {
		return none, err
	}
	return func(result error) {
		if err := hold.Release(queueInterruptionReason(result) != ""); err != nil {
			fmt.Fprintf(stderr, "merge intent: %v\n", err)
		}
	}, nil
}

// sweepMergeIntents removes unlocked intents whose branch is gone or that
// outlived landingline.IntentMaxAge; a failure only warns, since it leaves
// nothing worse than a stale line in status.
func sweepMergeIntents(r gitRunner, root string, store landingline.Intents, stderr io.Writer) {
	exists := func(branch string) (bool, error) {
		head, err := landingOptional(r, root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
		return head != "", err
	}
	if _, err := store.Sweep(exists, time.Now(), landingline.IntentMaxAge); err != nil {
		fmt.Fprintf(stderr, "merge intent sweep: %v\n", err)
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
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
	head, err := localBranchHead(r, t.Root, branch)
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

// localBranchHead is the branch's local head, "" when it has none.
func localBranchHead(r gitRunner, root, branch string) (string, error) {
	return landingOptional(r, root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
}

// sweepMergeIntents removes unlocked intents whose branch is gone or that
// outlived landingline.IntentMaxAge; a failure only warns, since it leaves
// nothing worse than a stale line in status.
func sweepMergeIntents(r gitRunner, root string, store landingline.Intents, stderr io.Writer) {
	exists := func(branch string) (bool, error) {
		head, err := localBranchHead(r, root, branch)
		return head != "", err
	}
	if _, err := store.Sweep(exists, time.Now(), landingline.IntentMaxAge); err != nil {
		fmt.Fprintf(stderr, "merge intent sweep: %v\n", err)
	}
}

// mergeIntentReport is one intent as the status surfaces show it; it is also
// sdlc state --json's merge_intents[] member, which Couch's restart plan reads.
type mergeIntentReport = fleet.MergeIntentReport

// reportMergeIntent is the pure reading of one observed intent. The command
// is built from the validated branch, never copied from the file.
func reportMergeIntent(o landingline.ObservedIntent, branchExists bool, paused []queuegate.PauseOwner) mergeIntentReport {
	if o.Err != nil {
		state := "unreadable"
		if o.Locked {
			state = string(landingline.IntentLive) // its merge may still be writing it
		}
		return mergeIntentReport{State: state, Note: o.Err.Error(), Path: o.Path, Identity: o.Path}
	}
	in := o.Intent
	r := mergeIntentReport{State: string(landingline.ClassifyIntent(o.Locked, branchExists)), Branch: in.Branch, Issue: in.Issue,
		Head: in.Head, Root: in.Root, Workspace: in.Workspace, PID: in.PID, StartedAt: in.StartedAt, Path: o.Path, Identity: o.Path}
	if r.State == string(landingline.IntentOrphaned) {
		r.Reentry = "sdlc merge --branch " + in.Branch + " --yes"
		r.Note = mergePauseOwnersNote(paused)
	}
	return r
}

// formatMergeIntent is the status line for a report that needs attention.
func formatMergeIntent(r mergeIntentReport) string {
	switch r.State {
	case "unreadable":
		return fmt.Sprintf("unreadable merge intent %q: %q (inspect it, then remove it)", r.Path, r.Note)
	case string(landingline.IntentOrphaned):
		issue := ""
		if r.Issue > 0 {
			issue = fmt.Sprintf("#%d ", r.Issue)
		}
		line := fmt.Sprintf("orphaned merge: %sbranch %q at %s, started %s by pid %d, which is gone; re-enter in %q: %s",
			issue, r.Branch, abbrevSHA(r.Head), r.StartedAt.Format(time.RFC3339), r.PID, r.Root, r.Reentry)
		if r.Note != "" {
			line += "\n  " + r.Note
		}
		return line
	}
	return ""
}

// observeMergeIntents reads the repository's intents for the status
// surfaces; keep selects which to report (all, or one slot's own).
func observeMergeIntents(r gitRunner, root string, keep func(mergeIntentReport) bool) ([]mergeIntentReport, error) {
	return observeMergeIntentsWithPause(r, root, keep, queuePauseNote)
}

func observeMergeIntentsWithPause(r gitRunner, root string, keep func(mergeIntentReport) bool, pauseNote func() string) ([]mergeIntentReport, error) {
	lineDir, err := landingLineDir(r, root)
	if err != nil {
		return nil, err
	}
	observed, err := landingline.OpenIntents(lineDir).List()
	if len(observed) == 0 {
		return nil, err
	}
	readErr := err
	note := pauseNote()
	var out []mergeIntentReport
	for _, o := range observed {
		if physical, pathErr := workspace.CanonicalPath(o.Path); pathErr == nil {
			o.Path = physical
		} else {
			o.Err = errors.Join(o.Err, fmt.Errorf("merge intent path identity unavailable: %w", pathErr))
		}
		exists := false
		var branchErr error
		if o.Err == nil {
			head, err := localBranchHead(r, root, o.Intent.Branch)
			branchErr = err
			exists = head != ""
		}
		rep := reportMergeIntent(o, exists, nil)
		if branchErr != nil {
			rep.State, rep.Reentry, rep.Note = "unreadable", "", branchErr.Error()
			readErr = errors.Join(readErr, branchErr)
		} else if rep.State == string(landingline.IntentOrphaned) {
			rep.Note = note
		}
		if keep(rep) {
			out = append(out, rep)
		}
	}
	slices.SortFunc(out, func(a, b mergeIntentReport) int { return strings.Compare(a.Path, b.Path) })
	return out, readErr
}

func queuePauseNote() string {
	s, err := defaultQueueCommandDeps().store()
	if err != nil {
		return "queue pause information unavailable: " + err.Error()
	}
	owners, err := s.PauseOwners()
	if err != nil {
		return "queue pause information unavailable: " + err.Error()
	}
	return mergePauseOwnersNote(owners)
}

func mergePauseOwnersNote(owners []queuegate.PauseOwner) string {
	if len(owners) == 0 {
		return ""
	}
	var by []string
	for _, p := range owners {
		by = append(by, fmt.Sprintf("%q (%q)", p.Owner, p.Reason))
	}
	return "the machine queue is paused by " + strings.Join(by, ", ") + "; re-enter after it resumes"
}

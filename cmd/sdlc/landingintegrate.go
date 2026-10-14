// landingintegrate.go — integrate a moved main before testing and landing (#332).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const integrationTrailer = "Integrated-Main"
const integrationLimit = 3

// Mutable durations let tests exercise the real polling boundary without waiting
// minutes. CI's independent budget is owned by waitCIGate.
var integrationPollInterval = 5 * time.Second
var integrationSyncTimeout = 2 * time.Minute

type integrationStep string

const (
	integrationMerge  integrationStep = "merge"
	integrationAbort  integrationStep = "abort"
	integrationReset  integrationStep = "reset"
	integrationPush   integrationStep = "push"
	integrationGates  integrationStep = "gates"
	integrationLand   integrationStep = "land"
	integrationRefuse integrationStep = "refuse"
)

type integrationObs struct {
	MainIncluded, GatesPassed, Resume bool
	Attempts                          int
	// Result is an observed effect, not a command to execute.
	Result string
}

// nextIntegrationStep is the ordering contract. Suite runs and pending CI waits
// are substeps of gates, owned by the CI gate rather than duplicated here.
func nextIntegrationStep(o integrationObs) integrationStep {
	switch o.Result {
	case "conflict":
		return integrationAbort
	case "replay-refused":
		return integrationReset
	case "replay-passed":
		return integrationPush
	case "lease-refused":
		return integrationRefuse
	}
	if o.Resume {
		return integrationPush
	}
	if !o.MainIncluded {
		if o.Attempts >= integrationLimit {
			return integrationRefuse
		}
		return integrationMerge
	}
	if !o.GatesPassed {
		return integrationGates
	}
	return integrationLand
}

func integrationAncestor(r gitRunner, root, ancestor, head string) (bool, error) {
	_, err := landingGit(r, root, "merge-base", "--is-ancestor", ancestor, head)
	if err == nil {
		return true, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// integrationResume validates the commit, not just its message. Only an exact
// clean two-parent merge on top of the observed remote branch may be published
// automatically. Ordinary unpublished edits keep the existing push refusal.
func integrationResume(r gitRunner, t landingTarget, head, remote, main string) (bool, error) {
	if head == "" || remote == "" || head == remote {
		return false, nil
	}
	parents, err := landingGit(r, t.Root, "show", "-s", "--format=%P", head)
	if err != nil {
		return false, err
	}
	p := strings.Fields(parents)
	if len(p) != 2 || p[0] != remote {
		return false, nil
	}
	trailer, err := landingGit(r, t.Root, "show", "-s", "--format=%(trailers:key="+integrationTrailer+",valueonly)", head)
	if err != nil {
		return false, err
	}
	integrated := strings.TrimSpace(trailer)
	if !landingOIDValid(integrated) || integrated != p[1] {
		return false, nil
	}
	included, err := integrationAncestor(r, t.Root, integrated, main)
	if err != nil || !included {
		return false, err
	}
	tree, err := landingGit(r, t.Root, "merge-tree", "--write-tree", p[0], p[1])
	if err != nil {
		return false, fmt.Errorf("cannot verify interrupted integration's clean merge: %w", err)
	}
	actual, err := landingGit(r, t.Root, "rev-parse", head+"^{tree}")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(tree) == actual, nil
}

// selectIntegrationPR preserves strict ordinary selection. The sole exception
// is a validated integration whose push (or GitHub registration) was interrupted.
func selectIntegrationPR(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch, head, current string, prs []landingPR, stderr io.Writer) (landingPR, bool, error) {
	pr, selectionErr := selectLandingPR(prs, t, branch, head)
	if selectionErr == nil {
		return pr, false, nil
	}
	candidate, err := selectLandingPR(prs, t, branch, "")
	if err != nil || candidate.State != "OPEN" || current != branch {
		return landingPR{}, false, selectionErr
	}
	main, err := t.fetchMain(r)
	if err != nil {
		return landingPR{}, false, err
	}
	remote, err := remoteRefTip(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, "refs/heads/"+branch)
	if err != nil {
		return landingPR{}, false, err
	}
	if remote != candidate.HeadOID && remote != head {
		return landingPR{}, false, selectionErr
	}
	valid, err := integrationResume(r, t, head, candidate.HeadOID, main)
	if err != nil {
		return landingPR{}, false, err
	}
	if !valid {
		return landingPR{}, false, selectionErr
	}
	if remote == head {
		pr, err = syncIntegrationPR(ctx, gh, t, branch, head, candidate, stderr)
		return pr, false, err
	}
	return candidate, true, nil
}

func syncIntegrationPR(ctx context.Context, gh landingGH, t landingTarget, branch, head string, previous landingPR, stderr io.Writer) (landingPR, error) {
	ctx, cancel := context.WithTimeout(ctx, integrationSyncTimeout)
	defer cancel()
	for {
		prs, err := gh.LandingPRs(ctx, t.Repo, branch)
		if err != nil {
			return landingPR{}, err
		}
		pr, err := selectLandingPR(prs, t, branch, "")
		if err != nil {
			return landingPR{}, err
		}
		if pr.Number != previous.Number || pr.State != "OPEN" {
			return landingPR{}, errors.New("PR identity or state changed while registering integration; preserve checkout and retry")
		}
		if pr.HeadOID == head {
			return selectLandingPR(prs, t, branch, head)
		}
		if pr.HeadOID != previous.HeadOID {
			return landingPR{}, errors.New("PR head changed concurrently while registering integration; preserve checkout and retry")
		}
		fmt.Fprintf(stderr, "waiting for GitHub to register pushed head %s on PR #%d (limit %v)\n", shortOID(head), pr.Number, integrationSyncTimeout)
		timer := time.NewTimer(integrationPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return landingPR{}, fmt.Errorf("GitHub has not registered the pushed head; re-run sdlc merge: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// integrateLandingMain makes the merge, or resumes the already validated one;
// replay must pass before a leased push can make either commit public.
func integrateLandingMain(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch, head, main string, pr landingPR, resume bool, noJudge bool, issuesDir string, stderr io.Writer) (string, landingPR, error) {
	oldHead := head
	if resume {
		oldHead = pr.HeadOID
	}
	if err := revalidateLanding(r, t, branch, head, branch); err != nil {
		return head, pr, err
	}
	if !resume {
		msg := fmt.Sprintf("Merge %s into %s (sdlc merge, #332)\n\n%s: %s", main, branch, integrationTrailer, main)
		_, mergeErr := landingGit(r, t.Root, "-c", "merge.autoStash=false", "-c", "rerere.enabled=false", "merge", "--no-ff", "--no-edit", "-m", msg, main)
		if mergeErr != nil {
			paths, pathErr := landingGit(r, t.Root, "diff", "--name-only", "--diff-filter=U", "-z")
			// A failed merge may not have started (for example an untracked collision).
			mergeHead, probeErr := landingOptional(r, t.Root, "rev-parse", "--verify", "--quiet", "MERGE_HEAD")
			if probeErr != nil {
				return head, pr, probeErr
			}
			if mergeHead != "" && nextIntegrationStep(integrationObs{Result: "conflict"}) == integrationAbort {
				if _, err := landingGit(r, t.Root, "merge", "--abort"); err != nil {
					return head, pr, fmt.Errorf("integration failed and abort failed: %w; original: %v", err, mergeErr)
				}
			}
			if pathErr != nil {
				return head, pr, pathErr
			}
			return head, pr, fmt.Errorf("main integration conflict or merge failure in %s (branch preserved): %w\n  → merge %s/main manually, resolve the paths, re-close with --fixed-then-ship --full-suite, then retry sdlc merge", strings.Join(splitNUL(paths), ", "), mergeErr, t.Remote)
		}
		var err error
		head, err = landingGit(r, t.Root, "rev-parse", "HEAD")
		if err != nil {
			return oldHead, pr, err
		}
	}
	integrated := pr
	integrated.HeadOID, integrated.BaseOID = head, main
	var replayErr error
	if !noJudge {
		replayErr = runLandingPublishGate(ctx, integrated, issuesDir, stderr)
	}
	outcome := "replay-passed"
	if replayErr != nil {
		outcome = "replay-refused"
	}
	switch nextIntegrationStep(integrationObs{Result: outcome}) {
	case integrationReset:
		if err := revalidateLanding(r, t, branch, head, branch); err != nil {
			return head, pr, fmt.Errorf("replay refused (%v); preserving changed checkout: %w", replayErr, err)
		}
		if _, err := landingGit(r, t.Root, "reset", "--keep", oldHead); err != nil {
			return head, pr, fmt.Errorf("replay refused (%v); rollback failed: %w", replayErr, err)
		}
		return oldHead, pr, fmt.Errorf("integration replay refused; restored branch before pushing: %w", replayErr)
	case integrationPush:
		if err := revalidateLanding(r, t, branch, head, branch); err != nil {
			return head, pr, err
		}
		if _, err := landingGit(r, t.Root, "push", "--force-with-lease=refs/heads/"+branch+":"+oldHead, t.Remote, head+":refs/heads/"+branch); err != nil {
			return head, pr, fmt.Errorf("integration lease push refused; concurrent remote work is preserved (local integration retained): %w", err)
		}
	}
	synced, err := syncIntegrationPR(ctx, gh, t, branch, head, pr, stderr)
	return head, synced, err
}

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// This seam also lets legacy tests explicitly declare ordinary-worktree identity.
var resolveLandingWorkspace = workspace.Resolve

type landingTarget struct {
	Root, RepoIdentity, Rest, RestOID, Remote, Repo, URL, PushURL string
}

func landingGit(r gitRunner, root string, args ...string) (string, error) {
	out, err := r.GitInDir(root, args...)
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
func landingOptional(r gitRunner, root string, args ...string) (string, error) {
	return optionalGitOutput(func(args ...string) (string, error) { return landingGit(r, root, args...) }, args...)
}
func landingRepoURL(raw string) (string, error) {
	path := ""
	if strings.HasPrefix(raw, "git@github.com:") {
		path = strings.TrimPrefix(raw, "git@github.com:")
	} else {
		u, err := url.Parse(raw)
		if err != nil || u.Host != "github.com" || (u.Scheme != "https" && u.Scheme != "ssh") || u.RawQuery != "" || u.Fragment != "" {
			return "", fmt.Errorf("unsupported effective remote URL; expected an HTTPS or SSH github.com repository")
		}
		if u.User != nil && (u.Scheme != "ssh" || u.User.String() != "git") {
			return "", fmt.Errorf("unsupported credentials in GitHub remote URL")
		}
		path = strings.TrimPrefix(u.Path, "/")
	}
	path = strings.TrimSuffix(path, ".git")
	if !landingRepoValid(path) {
		return "", fmt.Errorf("invalid GitHub repository in remote URL")
	}
	return path, nil
}
func resolveLandingTarget(r gitRunner) (*landingTarget, error) {
	id, err := resolveLandingWorkspace(r, ".", "")
	if err != nil {
		return nil, err
	}
	if id.Kind != "primary" && id.Kind != "slot" {
		return nil, nil
	}
	if id.RestingBranch == nil {
		return nil, errors.New("addressable workspace has no resting branch")
	}
	t := &landingTarget{Root: id.WorktreeRoot, RepoIdentity: id.RepoIdentity, Rest: *id.RestingBranch}
	t.RestOID, err = landingGit(r, t.Root, "rev-parse", "--verify", "refs/heads/"+t.Rest+"^{commit}")
	if err != nil {
		return nil, err
	}
	t.Remote, err = landingGit(r, t.Root, "config", "--get-all", "branch."+t.Rest+".remote")
	if err != nil {
		return nil, fmt.Errorf("configure %s to track a named remote/main: %w", t.Rest, err)
	}
	if t.Remote == "." || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`).MatchString(t.Remote) {
		return nil, errors.New("resting upstream must use one named remote")
	}
	if _, err = landingGit(r, t.Root, "check-ref-format", gitx.RemoteTrackingRef(t.Remote, "main")); err != nil {
		return nil, err
	}
	base, err := landingGit(r, t.Root, "config", "--get-all", "branch."+t.Rest+".merge")
	if err != nil || base != "refs/heads/main" {
		return nil, errors.New("resting upstream must be exactly remote/main")
	}
	// Use effective destinations: Git URL rewrites and push overrides are part
	// of the target, not incidental transport configuration.
	t.URL, err = landingGit(r, t.Root, "remote", "get-url", "--all", t.Remote)
	if err != nil {
		return nil, err
	}
	push, err := landingGit(r, t.Root, "remote", "get-url", "--push", "--all", t.Remote)
	if err != nil {
		return nil, err
	}
	t.PushURL = push
	fetchRepo, err := landingRepoURL(t.URL)
	if err != nil {
		return nil, err
	}
	pushRepo, err := landingRepoURL(push)
	if err != nil || !strings.EqualFold(pushRepo, fetchRepo) {
		return nil, errors.New("fetch and push must target the same GitHub repository")
	}
	t.Repo, err = landingRepoURL(t.URL)
	if err != nil {
		return nil, err
	}
	return t, nil
}
func (t landingTarget) mainRef() string { return gitx.RemoteTrackingRef(t.Remote, "main") }
func (t landingTarget) fetchMain(r gitRunner) (string, error) {
	if _, err := landingGit(r, t.Root, "-c", "submodule.recurse=false", "fetch", "--no-tags", "--no-recurse-submodules", t.Remote, "+refs/heads/main:"+t.mainRef()); err != nil {
		return "", err
	}
	return landingGit(r, t.Root, "rev-parse", "--verify", t.mainRef()+"^{commit}")
}
func landingNoOperation(r gitRunner, root string) error {
	dir, err := landingGit(r, root, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	operation, err := workspace.ActiveOperation(strings.TrimSpace(dir), workspace.Lstat)
	if err != nil {
		return err
	}
	if operation != "" {
		return fmt.Errorf("active Git operation %s; finish it before landing", operation)
	}
	return nil
}

func landingClean(r gitRunner, root string) error {
	if err := landingNoOperation(r, root); err != nil {
		return err
	}
	out, err := r.GitInDir(root, "status", "--porcelain=v1", "-z", "--untracked-files=no", "--ignore-submodules=none")
	if err != nil {
		return fmt.Errorf("read worktree status: %w", err)
	}
	if len(out) != 0 {
		return errors.New("tracked changes present; commit or preserve them before landing")
	}
	return nil
}

// Once already on rest, cleanup touches only the proven issue ref/config.
// New staged or unstaged resting work is preserved without refreshing its index.
func landingCheckoutReady(r gitRunner, t landingTarget, current string) error {
	if current == t.Rest {
		return landingNoOperation(r, t.Root)
	}
	return landingClean(r, t.Root)
}

func landingIssueBranch(r gitRunner, t landingTarget, requested string) (string, string, error) {
	current, err := landingGit(r, t.Root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", "", err
	}
	branch := requested
	if branch == "" {
		branch = current
	}
	if branch == "main" || regexp.MustCompile(`^main-slot[0-9]+$`).MatchString(branch) {
		return "", "", errors.New("a resting branch cannot be a landing target")
	}
	if _, err := landingGit(r, t.Root, "check-ref-format", "refs/heads/"+branch); err != nil {
		return "", "", err
	}
	if current != branch && current != t.Rest {
		return "", "", errors.New("recovery must run from the issue branch or this workspace's resting branch")
	}
	return branch, current, nil
}

// liveLandingPRs validates every identity before filtering completed history.
// OPEN PRs remain candidates even at stale heads, so ambiguity cannot be hidden.
func liveLandingPRs(prs []landingPR, t landingTarget, branch, head string) ([]landingPR, error) {
	var live []landingPR
	for _, pr := range prs {
		if pr.Repo != t.Repo || pr.HeadRef != branch || pr.BaseRef != "main" {
			return nil, errors.New("GitHub PR identity does not match landing target")
		}
		if pr.State == "OPEN" || pr.State == "MERGED" && (head == "" || pr.HeadOID == head) {
			live = append(live, pr)
		}
	}
	return live, nil
}

// selectLandingCandidate is shared by strict selection and integration recovery
// and polling, where the sole OPEN PR may still show its previous head.
func selectLandingCandidate(prs []landingPR, t landingTarget, branch, head string) (landingPR, error) {
	live, err := liveLandingPRs(prs, t, branch, head)
	if err != nil {
		return landingPR{}, err
	}
	if len(live) != 1 {
		var names []string
		for _, pr := range live {
			names = append(names, fmt.Sprintf("#%d", pr.Number))
		}
		return landingPR{}, fmt.Errorf("need one exact matching PR for %s; found %d, candidates [%s] (preserving local work)", branch, len(live), strings.Join(names, ", "))
	}
	return live[0], nil
}

func selectLandingPR(prs []landingPR, t landingTarget, branch, head string) (landingPR, error) {
	pr, err := selectLandingCandidate(prs, t, branch, head)
	if err != nil {
		return landingPR{}, err
	}
	// #267: the branch's sole stale OPEN PR needs a push, not a new PR.
	if head != "" && pr.HeadOID != head {
		return landingPR{}, fmt.Errorf("PR #%d for %s is at %s but the local branch is at %s; push it with `sdlc pr`, then retry (preserving local work)", pr.Number, branch, shortOID(pr.HeadOID), shortOID(head))
	}
	return pr, nil
}

type landingAction string

const (
	landingMerge    landingAction = "merge"
	landingArchive  landingAction = "archive"
	landingReturn   landingAction = "return"
	landingDelete   landingAction = "delete"
	landingComplete landingAction = "complete"
)

type landingObservation struct{ Merged, Integrated, Archived, AtRest, RefExists bool }

func nextLandingAction(o landingObservation) (landingAction, error) {
	if !o.Merged {
		if o.AtRest || !o.RefExists {
			return "", errors.New("cannot initiate merge from rest or an absent issue branch")
		}
		return landingMerge, nil
	}
	if !o.Integrated {
		return "", errors.New("merged PR integration is not reachable from configured remote/main")
	}
	if !o.Archived {
		if !o.RefExists {
			return "", errors.New("missing issue branch requires completed archive proof")
		}
		return landingArchive, nil
	}
	if !o.AtRest {
		return landingReturn, nil
	}
	if o.RefExists {
		return landingDelete, nil
	}
	return landingComplete, nil
}
func revalidateLanding(r gitRunner, t landingTarget, branch, head, current string) error {
	now, err := resolveLandingTarget(r)
	if err != nil {
		return err
	}
	if now == nil || *now != t {
		return errors.New("workspace, resting ref or upstream changed; refusing cleanup")
	}
	actual, err := landingGit(r, t.Root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return err
	}
	if actual != current {
		return errors.New("checkout branch changed; refusing cleanup")
	}
	actual, err = landingOptional(r, t.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil {
		return err
	}
	if actual != head {
		return errors.New("issue branch changed; preserving new work")
	}
	return landingCheckoutReady(r, t, current)
}
func landingUnoccupied(r gitRunner, t landingTarget, branch string) error {
	out, err := r.GitInDir(t.Root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return err
	}
	trees, err := workspace.ParseWorktrees(out)
	if err != nil {
		return err
	}
	for _, w := range trees {
		if w.Branch == branch {
			return fmt.Errorf("branch %s is checked out at %s", branch, w.Path)
		}
	}
	return nil
}
func returnLandingToRest(ctx context.Context, r gitRunner, t landingTarget, branch, head, issuesDir string) error {
	if err := revalidateLanding(r, t, branch, head, branch); err != nil {
		return err
	}
	if err := landingUnoccupied(r, t, t.Rest); err != nil {
		return err
	}
	reader, err := gitx.NewTrunkFileContext(ctx, t.Root, t.Remote, "main")
	if err != nil {
		return err
	}
	view := reader.ViewOf(t.RestOID)
	if err := verifyAssetMaterialization(ctx, t.Root, view, issuesDir); err != nil {
		return err
	}
	if _, err := landingGit(r, t.Root, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", t.Rest); err != nil {
		return err
	}
	if err := verifyMaterializedAssets(t.Root, view, issuesDir); err != nil {
		return fmt.Errorf("returned to %s, but asset verification failed: %w; the issue branch is preserved", t.Rest, err)
	}
	return revalidateLanding(r, t, branch, head, t.Rest)
}
func deleteLandingBranch(r gitRunner, t landingTarget, branch, head string, stderr io.Writer) error {
	if err := revalidateLanding(r, t, branch, head, t.Rest); err != nil {
		return err
	}
	if err := landingUnoccupied(r, t, branch); err != nil {
		return err
	}
	if head != "" {
		if _, err := landingGit(r, t.Root, "update-ref", "-d", "refs/heads/"+branch, head); err != nil {
			return err
		}
	}
	// Re-entry after ref deletion completes this second effect. A recreated ref
	// or changed topology refuses rather than stripping its upstream.
	if err := revalidateLanding(r, t, branch, "", t.Rest); err != nil {
		return err
	}
	if err := landingUnoccupied(r, t, branch); err != nil {
		return err
	}
	cleanupBranchConfig(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, branch, stderr)
	return nil
}
func runDurableMerge(stdout, stderr io.Writer, f *mergeFlags, t landingTarget) (err error) {
	ctx := commandContext(f.Context)
	r := boundedLandingGit(ctx, mergeRunner)
	branch, current, err := landingIssueBranch(r, t, f.Branch)
	if err != nil {
		return err
	}
	if err = landingCheckoutReady(r, t, current); err != nil {
		return err
	}
	if mergeNeedsTTY(f.Yes, f.DryRun, isTTY(os.Stdin)) {
		return errors.New("sdlc merge needs confirmation; rerun with --yes")
	}
	if f.DryRun {
		fmt.Fprintf(stdout, "Would land %s into %s/main, archive remotely, and return to unchanged %s; retry: sdlc merge --branch %s --yes\n", branch, t.Remote, t.Rest, branch)
		return nil
	}
	gh, ok := ghClient.(landingGH)
	if !ok {
		return errors.New("GitHub adapter does not support structured landing evidence")
	}
	// Confirm before queueing, so an operator's prompt never holds the line.
	if !f.Yes {
		if ans := mergePrompter.Ask("Land PR, archive remotely and return to unchanged baseline? [y/N] ", stderr); ans != "y" && ans != "Y" {
			return errors.New("aborted by operator")
		}
	}
	// Enrollment covers every merge path, including already-merged cleanup.
	// Only actual line admission (or bypass admission below) activates it.
	queueHead, err := landingOptional(r, t.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil {
		return err
	}
	ctx, q, err := startQueueOperation(ctx, t.Root, queueHead, []string{"merge", "landing"})
	if err != nil {
		return err
	}
	localFlags := *f
	f = &localFlags
	f.Context = ctx
	defer func() {
		reason := queueInterruptionReason(err)
		if finishErr := q.finish(ctx, reason); finishErr != nil {
			if err == nil {
				err = finishErr
			} else {
				fmt.Fprintf(stderr, "queue retirement: %v\n", finishErr)
			}
		}
	}()
	// The landing line (#393). Rule: a merge without a turn acts only on an
	// already-merged PR. So it skips the line only when the first read shows
	// a merged PR and no open one (a resume), and the selected PR is checked
	// against the rule below, because a later read can differ (a lagging
	// listing). Everything else is observed after admission (PQ-1).
	pre, err := gh.LandingPRs(ctx, t.Repo, branch)
	if err != nil {
		return err
	}
	var turn *landingTurn
	var sched func() landingline.Schedule
	gates := productionCarGates(r, ghClient, t, f)
	readiness := func(ctx context.Context) (landingline.Readiness, bool) {
		return carReadiness(ctx, r, gh, t, branch, gates, landingNow())
	}
	landedByTrain := false
	var trainOutcome *landingline.Outcome
	if !resumeOnly(pre) {
		sched = landingScheduleSource(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, stderr)
		if turn, err = enterLandingLine(ctx, r, t, branch, sched, readiness, stderr); err != nil {
			// A train landed this car (#449): its PR is merged by the train,
			// and this merge finishes as a resume, without a turn.
			var settled *landingline.Settled
			switch {
			case errors.Is(err, errMergedWhileWaiting):
				fmt.Fprintln(stderr, "this PR was merged while it waited (a train with no word for it); finishing cleanup")
			case errors.As(err, &settled) && settled.Outcome.Kind == landingline.OutcomeLanded:
				trainOutcome = &settled.Outcome
				fmt.Fprintf(stderr, "landing train %d landed this PR (train PR #%d); finishing cleanup\n", settled.Outcome.Train, settled.Outcome.PR)
			default:
				return err
			}
			landedByTrain = true
		} else {
			defer func() { turn.leave(err) }()
		}
	}
	if turn == nil {
		// Resume-only and train-landed cleanup bypass the repository line, but
		// still honor the machine pause before doing active cleanup work.
		if err = q.waitAdmission(ctx, stderr); err != nil {
			return err
		}
	}
	ctx = withQueueAdmitted(ctx)
	f.Context = ctx
	head, err := landingOptional(r, t.Root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil {
		return err
	}
	if head == "" && current != t.Rest {
		return errors.New("issue branch absent outside resting checkout")
	}
	prs, err := gh.LandingPRs(ctx, t.Repo, branch)
	if err != nil {
		return err
	}
	if landedByTrain {
		// GitHub marks a car merged by reachability shortly after the train's
		// merge: wait for that, bounded, rather than misread a lag as open.
		if prs, err = awaitMergedByTrain(ctx, gh, t, branch, prs, stderr); err != nil {
			return err
		}
	}
	pr, resumeIntegration, err := selectIntegrationPR(ctx, r, gh, t, branch, head, current, prs, f.withLock, stderr)
	if err != nil {
		return err
	}
	if turn == nil && pr.State != "MERGED" {
		return fmt.Errorf("PR #%d appeared open after this merge chose not to queue (GitHub's listing lagged); nothing was changed, re-run sdlc merge --branch %s --yes", pr.Number, branch)
	}
	if turn != nil {
		turn.PR, turn.Resumed = pr.Number, pr.State == "MERGED"
	}
	fmt.Fprintf(stderr, "Recovery: sdlc merge --branch %s --yes\n", branch)
	if turn != nil && pr.State != "MERGED" && sched != nil {
		// Trains on (#449): this head may conduct a train that lands its PR
		// with others'; one car is today's single landing, below.
		merged, conducted, terr := maybeConductTrain(ctx, r, gh, ghClient, t, f, turn, branch, head, pr, sched(), readiness, stderr)
		if terr != nil {
			return terr
		}
		if conducted {
			pr = merged
			turn.Resumed = false
		}
	}
	if pr.State != "MERGED" {
		if _, err = nextLandingAction(landingObservation{AtRest: current == t.Rest, RefExists: head != ""}); err != nil {
			return err
		}

		landingGitCall := func(args ...string) (string, error) {
			readRunner := r
			if len(args) > 1 && args[0] == "worktree" && (args[1] == "remove" || args[1] == "prune") {
				readRunner = recoveryLandingGit(r)
			}
			return landingGit(readRunner, t.Root, args...)
		}
		if f.NoValidate {
			cwarn(stderr, "--no-validate: skipping instance and duplicate-ID gates")
		}
		if f.NoJudge {
			cwarn(stderr, "--no-judge: skipping reviewed-HEAD publish gate")
		}
		sourceHead := head
		landingChecks := landingCheckOwner{}
		state := newIntegrationState(head, pr, resumeIntegration)
		for {
			var targetMain string
			if err = f.withLock(func() error {
				if err := revalidateLanding(r, t, branch, state.head, current); err != nil {
					return err
				}
				var fetchErr error
				targetMain, fetchErr = t.fetchMain(r)
				return fetchErr
			}); err != nil {
				return err
			}
			remoteHead, err := remoteRefTip(landingGitCall, t.Remote, "refs/heads/"+branch)
			if err != nil {
				return err
			}
			if remoteHead != state.head && (!state.resume || remoteHead != state.pr.HeadOID) {
				return errors.New("local, remote and PR heads must match; push selected branch before landing")
			}
			observation, err := observeIntegrationMain(r, t, state, targetMain, stderr)
			if err != nil {
				return err
			}
			var step integrationStep
			state, step, err = transitionIntegration(state, observation)
			if err != nil {
				return err
			}
			if step == integrationRefuse {
				return fmt.Errorf("main moved too often: stopped after %d integrations; re-run sdlc merge", integrationLimit)
			}
			if step == integrationMerge || step == integrationReplay {
				if state.chargeIntegration {
					fmt.Fprintf(stderr, "Integrating %s/main before landing (attempt %d of %d)\n", t.Remote, state.attempts+1, integrationLimit)
				} else {
					fmt.Fprintf(stderr, "Integrating issue records from %s/main; code integration budget unchanged\n", t.Remote)
				}
				state, err = integrateLandingMain(ctx, r, gh, t, branch, state, f.withLock, f.NoJudge, f.IssuesDir, f.PlansDir, stderr)
				if err != nil {
					return err
				}
				// PR-head registration may have waited while another slot moved main.
				// Observe it again before spending suite time on this integration.
				if err = f.withLock(func() error {
					if err := revalidateLanding(r, t, branch, state.head, current); err != nil {
						return err
					}
					latest, err := t.fetchMain(r)
					if err != nil {
						return err
					}
					observation, err = observeIntegrationMain(r, t, state, latest, stderr)
					if err != nil {
						return err
					}
					state, step, err = transitionIntegration(state, observation)
					return err
				}); err != nil {
					return err
				}
				if step == integrationRefuse {
					return fmt.Errorf("main moved too often: stopped after %d integrations; re-run sdlc merge", integrationLimit)
				}
				if step != integrationGates {
					continue
				}
				targetMain = state.main
			}
			if state.phase != integrationGates {
				return fmt.Errorf("integration cannot run gates from phase %s", state.phase)
			}
			head, pr = state.head, state.pr
			// Gate this exact integrated tree. A resumed integration of an older
			// main is brought fully current before spending time on its suite.
			observation, err = observeIntegrationMain(r, t, state, targetMain, stderr)
			if err != nil {
				return err
			}
			if !observation.included && !observation.issueOnly {
				continue
			}
			if !f.NoValidate {
				if err = runLandingInstanceGate("", targetMain, head, nounGates(f.IssuesDir), stdout, stderr, r); err != nil {
					return err
				}
				if err = runLandingDuplicateGate(targetMain, head, f.IssuesDir, f.HistoryDir, r); err != nil {
					return err
				}
			}
			if !f.NoJudge {
				if err = runLandingPublishGate(ctx, pr, f.IssuesDir, f.PlansDir, stderr); err != nil {
					return err
				}
			}
			localDeps := localRunDeps{GH: ghClient, Repo: t.Repo, Root: t.Root, Git: landingGitCall, CleanupGit: localCheckCleanupGit(ctx, r, t.Root), PR: pr.Number, Stderr: stderr, FlakeRegistry: gateFlakeRegistry(t.Root)}
			localRun := newLocalCheckRunner(localDeps)
			if err = waitCIGate(ctx, ghClient, t.Repo, pr.Number, targetMain, head, landingGitCall, localRun, f.NoCI, stderr, func(ioCtx context.Context) func(...string) (string, error) {
				readRunner := withLandingGitContext(ioCtx, r)
				return func(args ...string) (string, error) { return landingGit(readRunner, t.Root, args...) }
			}); err != nil {
				state, _, _ = transitionIntegration(state, integrationEvent{kind: integrationGatesRejected})
				return diagnoseSuiteFailure(ctx, err, targetMain, localDeps)
			}
			if err := landingChecks.run(ctx, landingGitCall, targetMain, head, []string{sourceHead}, !f.NoCI, localRun, func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
				return runLocalCheck(ctx, localDeps, c, sha)
			}); err != nil {
				return err
			}
			// Read effective readiness outside the transaction. Its live base
			// target must still match the final fetched main; no status is copied.
			readiness, readinessErr := landingReadiness{}, fmt.Errorf("adapter has no readiness evidence")
			if source, ok := gh.(landingReadinessGH); ok {
				readiness, readinessErr = source.LandingReadiness(ctx, t.Repo, pr.Number)
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Final revalidation and expected-head merge remain one transaction.

			if err = f.withLock(func() error {
				if err := revalidateLanding(r, t, branch, head, current); err != nil {
					return err
				}
				latest, err := t.fetchMain(r)
				if err != nil {
					return err
				}
				observation, err = observeIntegrationMain(r, t, state, latest, stderr)
				if err != nil {
					return err
				}
				observation.kind = integrationGatesChecked
				if !observation.included && observation.issueOnly {
					if readinessErr == nil {
						observation.ready, err = readinessAllowsUnintegrated(pr, readiness, latest)
						if err != nil {
							return err
						}
					} else {
						fmt.Fprintf(stderr, "issue-record readiness unavailable (%v); integrating normally without charging the code budget\n", readinessErr)
					}
					if observation.ready && !landingChecks.allowsMainMove(landingGitCall, targetMain, latest, []string{sourceHead}) {
						observation.ready = false
						fmt.Fprintln(stderr, "main movement changes landing-check evidence; integrating before checking again")
					}
					if observation.ready {
						// Main may have gained issue IDs or published details
						// during CI. Reuse the gates against that exact destination.
						if !f.NoValidate {
							if err = runLandingInstanceGate("", latest, head, nounGates(f.IssuesDir), stdout, stderr, r); err != nil {
								return err
							}
							if err = runLandingDuplicateGate(latest, head, f.IssuesDir, f.HistoryDir, r); err != nil {
								return err
							}
						}
						if !f.NoJudge {
							fresh := pr
							fresh.BaseOID = latest
							if err = runLandingPublishGate(ctx, fresh, f.IssuesDir, f.PlansDir, stderr); err != nil {
								return err
							}
						}
					}
				}
				state, step, err = transitionIntegration(state, observation)
				if err != nil || step != integrationLand {
					return err
				}
				if err := gh.LandingMerge(ctx, t.Repo, pr.Number, head); err != nil {
					return fmt.Errorf("merge outcome uncertain; use recovery command: %w", err)
				}
				return nil
			}); err != nil {
				return err
			}
			if step == integrationRefuse {
				return fmt.Errorf("main moved too often: stopped after %d integrations; re-run sdlc merge", integrationLimit)
			}
			if step == integrationLand {
				prs, err = gh.LandingPRs(ctx, t.Repo, branch)
				if err != nil {
					return err
				}
				observed, err := selectLandingPR(prs, t, branch, head)
				if err != nil {
					return err
				}
				if observed.Number != pr.Number {
					return errors.New("PR identity changed after merge request")
				}
				pr = observed
				break
			}
		}
	}
	if err := f.withLock(func() error {
		if err := revalidateLanding(r, t, branch, head, current); err != nil {
			return err
		}
		if pr.State != "MERGED" {
			return errors.New("PR is not yet merged (possibly queued); preserve checkout and retry after integration")
		}
		mainOID, err := t.fetchMain(r)
		if err != nil {
			return err
		}
		landed, ok := landedCommit(pr, func(oid string) bool {
			_, err := landingGit(r, t.Root, "merge-base", "--is-ancestor", oid, mainOID)
			return err == nil
		})
		if !ok {
			return fmt.Errorf("cannot confirm integration on %s/main: neither PR #%d's merge commit nor its head %s is on main", t.Remote, pr.Number, shortSHA(pr.HeadOID))
		}
		pr.LandedOID = landed
		if err = revalidateLanding(r, t, branch, head, current); err != nil {
			return err
		}
		// #252: the landing is confirmed, so the closes the PR owns go done on their
		// cards, landed at its integrated merge (squash and rebase included).
		if err = completeLandingPR(commandContext(f.Context), t.Root, f.IssuesDir, pr); err != nil {
			return err
		}
		complete, err := landingArchiveComplete(ctx, t.Root, mainOID, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir)
		if complete && err == nil && trainOutcome != nil {
			fmt.Fprintf(stderr, "landing train %d: PR #%d's issue archived by the train (train PR #%d); finishing cleanup\n", trainOutcome.Train, pr.Number, trainOutcome.PR)
		}
		if err != nil {
			return err
		}
		action, err := nextLandingAction(landingObservation{Merged: true, Integrated: true, Archived: complete, AtRest: current == t.Rest, RefExists: head != ""})
		if err != nil {
			return err
		}
		if action == landingArchive {
			if err = archiveLandingPR(ctx, t.Root, t.Remote, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir); err != nil {
				return err
			}
			mainOID, err = t.fetchMain(r)
			if err != nil {
				return err
			}
			complete, err = landingArchiveComplete(ctx, t.Root, mainOID, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir)
			if err != nil {
				return err
			}
			if !complete {
				return errors.New("archive publication not confirmed; preserve checkout and retry")
			}
		}
		if current != t.Rest {
			if err = returnLandingToRest(ctx, r, t, branch, head, f.IssuesDir); err != nil {
				return err
			}
		}
		if err = deleteLandingBranch(r, t, branch, head, stderr); err != nil {
			return err
		}
		// #286: the landed branch leaves origin too, leased on the landed head so
		// a push made after the PR merged is never thrown away.
		if derr := deleteRemoteBranch(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, branch, pr.HeadOID); derr != nil {
			cwarn(stderr, fmt.Sprintf("landed, but %s was not deleted on %s: %v — `sdlc merge --branch %s --yes` retries it", branch, t.Remote, derr, branch))
		}
		// #287: the next merge also finishes work merged outside sdlc — its card
		// went done above (settled by its evidence); archive it and drop its branch.
		if tracked, terr := repositoryTracked(ctx, t.Root); terr == nil && tracked {
			if env, oerr := openTrackerAt(ctx, t.Root); oerr != nil {
				cwarn(stderr, fmt.Sprintf("landed; work merged outside sdlc not checked: %v", oerr))
			} else if ferr := finishLandedLeftovers(env, stderr, archiveDirs{Issues: f.IssuesDir, Plans: f.PlansDir, History: f.HistoryDir}); ferr != nil {
				cwarn(stderr, fmt.Sprintf("landed; earlier work merged outside sdlc not archived: %v — `sdlc issue recovery reconcile --issue N` retries", ferr))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// #373: tick the landed issue in every live project citing it — after the
	// lock: it is warn-only IO on other repositories' origins.
	tickOnLanding(ctx, stdout, stderr, t.Root, branch, strconv.Itoa(pr.Number))
	fmt.Fprintf(stdout, "Landed PR #%d; workspace retained on unchanged %s. Refresh is separate.\n", pr.Number, t.Rest)
	return nil
}

func runDurablePR(stdout, stderr io.Writer, f *prFlags, t landingTarget) error {
	ctx := commandContext(f.Context)
	branch, _, err := landingIssueBranch(prRunner, t, "")
	if err != nil {
		return err
	}
	head, err := landingGit(prRunner, t.Root, "rev-parse", "--verify", "refs/heads/"+branch)
	if err != nil {
		return err
	}
	base := t.mainRef()
	if !f.DryRun {
		base, err = t.fetchMain(prRunner)
		if err != nil {
			return err
		}
	}
	changedPaths, err := landingGit(prRunner, t.Root, "diff", "--name-only", base+".."+head, "--", f.IssuesDir+"/*.md")
	if err != nil {
		return err
	}
	commits, err := landingGit(prRunner, t.Root, "log", base+".."+head, "--pretty=format:- %s")
	if err != nil {
		return err
	}
	ghNums, lerr := collectGitHubIssueNumbers(ctx, splitNonEmptyLines(changedPaths))
	if lerr != nil {
		cwarn(stderr, fmt.Sprintf("PR body lacks Fixes lines: %v", lerr))
	}
	body := combineBody(commits, formatFixes(ghNums))
	if f.DryRun {
		// #267: the dry run stays offline, so it names both outcomes of the PR query.
		fmt.Fprintf(stdout, "Would: git push %s refs/heads/%s:refs/heads/%s\nWould: update the branch's open PR if one exists, else gh pr create --repo %s --base main --head %s\n%s\n", t.Remote, branch, branch, t.Repo, branch, body)
		return nil
	}
	now, err := resolveLandingTarget(prRunner)
	if err != nil {
		return err
	}
	if now == nil || *now != t {
		return errors.New("workspace target changed before PR creation")
	}
	if err = revalidateLanding(prRunner, t, branch, head, branch); err != nil {
		return err
	}
	gh, ok := ghClient.(landingGH)
	if !ok {
		return errors.New("GitHub adapter does not support structured landing evidence")
	}
	prs, err := gh.LandingPRs(ctx, t.Repo, branch)
	if err != nil {
		return err
	}
	live, err := liveLandingPRs(prs, t, branch, head)
	if err != nil {
		return err
	}
	var open []landingPR
	for _, pr := range live {
		if pr.State == "OPEN" {
			open = append(open, pr)
		}
	}
	if len(open) > 1 {
		return fmt.Errorf("found %d open PRs for %s; close all but one, then retry", len(open), branch)
	}
	// #286: the same leased push as the boundaries, so a branch rebased since
	// its last push publishes without a blind force.
	if err = leasedBranchPush(func(args ...string) (string, error) { return landingGit(prRunner, t.Root, args...) }, t.Remote, branch); err != nil {
		return err
	}
	// #267: the branch's open PR follows its head; the push was the update.
	if len(open) == 1 {
		fmt.Fprintf(stdout, "updated PR #%d to %s\n", open[0].Number, shortOID(head))
		return nil
	}
	link, err := ghClient.PRCreate(t.Repo, "main", branch, body)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, link)
	return nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
	out, err := r.GitInDir(root, args...)
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return "", nil
		}
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSuffix(string(out), "\n"), nil
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

// liveLandingPRs refuses any PR whose identity (repository, head ref, base)
// differs from the landing target and drops closed ones.
func liveLandingPRs(prs []landingPR, t landingTarget, branch string) ([]landingPR, error) {
	var live []landingPR
	for _, pr := range prs {
		if pr.Repo != t.Repo || pr.HeadRef != branch || pr.BaseRef != "main" {
			return nil, errors.New("GitHub PR identity does not match landing target")
		}
		if pr.State != "CLOSED" {
			live = append(live, pr)
		}
	}
	return live, nil
}

func selectLandingPR(prs []landingPR, t landingTarget, branch, head string) (landingPR, error) {
	live, err := liveLandingPRs(prs, t, branch)
	if err != nil {
		return landingPR{}, err
	}
	var matches, stale []landingPR
	for _, pr := range live {
		if head == "" || pr.HeadOID == head {
			matches = append(matches, pr)
		} else if pr.State == "OPEN" {
			stale = append(stale, pr)
		}
	}
	// #267: the branch's open PR behind local work is still its PR; name the push.
	if len(matches) == 0 && len(stale) == 1 {
		return landingPR{}, fmt.Errorf("PR #%d for %s is at %s but the local branch is at %s; push it with `sdlc pr`, then retry (preserving local work)", stale[0].Number, branch, shortOID(stale[0].HeadOID), shortOID(head))
	}
	if len(matches) != 1 {
		return landingPR{}, fmt.Errorf("need one exact matching PR for %s; found %d (preserving local work)", branch, len(matches))
	}
	return matches[0], nil
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
func returnLandingToRest(r gitRunner, t landingTarget, branch, head string) error {
	if err := revalidateLanding(r, t, branch, head, branch); err != nil {
		return err
	}
	if err := landingUnoccupied(r, t, t.Rest); err != nil {
		return err
	}
	if _, err := landingGit(r, t.Root, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", t.Rest); err != nil {
		return err
	}
	return revalidateLanding(r, t, branch, head, t.Rest)
}
func deleteLandingBranch(r gitRunner, t landingTarget, branch, head string) error {
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
	config, err := landingOptional(r, t.Root, "config", "--local", "--get-regexp", `^branch\.`+regexp.QuoteMeta(branch)+`\.`)
	if err != nil {
		return err
	}
	if config != "" {
		_, err = landingGit(r, t.Root, "config", "--local", "--remove-section", "branch."+branch)
	}
	return err
}
func runDurableMerge(stdout, stderr io.Writer, f *mergeFlags, t landingTarget) error {
	r := mergeRunner
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
	ctx := f.Context
	if ctx == nil {
		ctx = context.Background()
	}
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
	pr, err := selectLandingPR(prs, t, branch, head)
	if err != nil {
		return err
	}
	if !f.Yes {
		if ans := mergePrompter.Ask("Land PR, archive remotely and return to unchanged baseline? [y/N] ", stderr); ans != "y" && ans != "Y" {
			return errors.New("aborted by operator")
		}
	}
	fmt.Fprintf(stderr, "Recovery: sdlc merge --branch %s --yes\n", branch)
	if pr.State != "MERGED" {
		if _, err = nextLandingAction(landingObservation{AtRest: current == t.Rest, RefExists: head != ""}); err != nil {
			return err
		}
		targetMain, fetchErr := t.fetchMain(r)
		if fetchErr != nil {
			return fetchErr
		}
		remoteHead, err := remoteRefTip(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, "refs/heads/"+branch)
		if err != nil {
			return err
		}
		if remoteHead != head {
			return errors.New("local, remote and PR heads must match; push selected branch before landing")
		}
		if f.NoValidate {
			cwarn(stderr, "--no-validate: skipping instance and duplicate-ID gates")
		}
		if f.NoJudge {
			cwarn(stderr, "--no-judge: skipping reviewed-HEAD publish gate")
		}
		if !f.NoValidate {
			if err = validateChangedInstancesFn(pr.BaseOID, "", nounGates(f.IssuesDir), stdout, stderr); err != nil {
				return err
			}
			if err = runLandingDuplicateGate(targetMain, f.IssuesDir, f.HistoryDir, r); err != nil {
				return err
			}
		}
		if !f.NoJudge {
			if err = runLandingPublishGate(commandContext(f.Context), pr, f.IssuesDir, stderr); err != nil {
				return err
			}
		}
		if err = revalidateLanding(r, t, branch, head, current); err != nil {
			return err
		}
		if err = gh.LandingMerge(ctx, t.Repo, pr.Number, head); err != nil {
			return fmt.Errorf("merge outcome uncertain; use recovery command: %w", err)
		}
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
	}
	if pr.State != "MERGED" {
		return errors.New("PR is not yet merged (possibly queued); preserve checkout and retry after integration")
	}
	mainOID, err := t.fetchMain(r)
	if err != nil {
		return err
	}
	if _, err = landingGit(r, t.Root, "merge-base", "--is-ancestor", pr.MergeOID, mainOID); err != nil {
		return fmt.Errorf("cannot confirm integration on %s/main: %w", t.Remote, err)
	}
	if err = revalidateLanding(r, t, branch, head, current); err != nil {
		return err
	}
	// #252: the landing is confirmed, so the closes the PR owns go done on their
	// cards, landed at its integrated merge (squash and rebase included).
	if err = completeLandingPR(commandContext(f.Context), t.Root, f.IssuesDir, pr); err != nil {
		return err
	}
	complete, err := landingArchiveComplete(ctx, t.Root, mainOID, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir)
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
		if err = returnLandingToRest(r, t, branch, head); err != nil {
			return err
		}
	}
	if err = deleteLandingBranch(r, t, branch, head); err != nil {
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
		fmt.Fprintf(stdout, "Would: git push -u %s %s\nWould: update the branch's open PR if one exists, else gh pr create --repo %s --base main --head %s\n%s\n", t.Remote, branch, t.Repo, branch, body)
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
	live, err := liveLandingPRs(prs, t, branch)
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
	// #267: git exits 0 when it pushes but cannot write the upstream config,
	// and its stderr is not returned on success; verify the result instead.
	if upstream, cerr := landingOptional(prRunner, t.Root, "config", "--get", "branch."+branch+".merge"); cerr != nil {
		cwarn(stderr, fmt.Sprintf("pushed, but could not read %s's upstream: %v", branch, cerr))
	} else if upstream == "" {
		cwarn(stderr, fmt.Sprintf("pushed, but git did not record %s's upstream (a write-protected .git/config, as in an agent sandbox, does this); a bare `git push` will fail — re-run `sdlc pr` to publish later commits", branch))
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

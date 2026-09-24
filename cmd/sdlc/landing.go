package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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
	if _, err = landingGit(r, t.Root, "check-ref-format", "refs/remotes/"+t.Remote+"/main"); err != nil {
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
func (t landingTarget) mainRef() string { return "refs/remotes/" + t.Remote + "/main" }
func (t landingTarget) fetchMain(r gitRunner) (string, error) {
	if _, err := landingGit(r, t.Root, "-c", "submodule.recurse=false", "fetch", "--no-tags", "--no-recurse-submodules", t.Remote, "+refs/heads/main:"+t.mainRef()); err != nil {
		return "", err
	}
	return landingGit(r, t.Root, "rev-parse", "--verify", t.mainRef()+"^{commit}")
}
func landingClean(r gitRunner, root string) error {
	for _, operation := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply", "sequencer", "BISECT_START"} {
		p, err := landingGit(r, root, "rev-parse", "--git-path", operation)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		if _, err = os.Stat(p); err == nil {
			return fmt.Errorf("active Git operation %s; finish it before landing", operation)
		} else if !os.IsNotExist(err) {
			return err
		}
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
func selectLandingPR(prs []landingPR, t landingTarget, branch, head string) (landingPR, error) {
	var matches []landingPR
	for _, pr := range prs {
		if pr.Repo != t.Repo || pr.HeadRef != branch || pr.BaseRef != "main" {
			return landingPR{}, errors.New("GitHub PR identity does not match landing target")
		}
		if pr.State == "CLOSED" {
			continue
		}
		if head == "" || pr.HeadOID == head {
			matches = append(matches, pr)
		}
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
	return landingClean(r, t.Root)
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
	if err = landingClean(r, t.Root); err != nil {
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
		remoteHead, err := landingGit(r, t.Root, "ls-remote", "--heads", t.Remote, "refs/heads/"+branch)
		if err != nil {
			return err
		}
		fields := strings.Fields(remoteHead)
		if len(fields) != 2 || fields[0] != head || fields[1] != "refs/heads/"+branch {
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
			if err = runLandingPublishGate(pr, f.IssuesDir, stderr); err != nil {
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
	complete, err := landingArchiveComplete(t.Root, mainOID, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir)
	if err != nil {
		return err
	}
	action, err := nextLandingAction(landingObservation{Merged: true, Integrated: true, Archived: complete, AtRest: current == t.Rest, RefExists: head != ""})
	if err != nil {
		return err
	}
	if action == landingArchive {
		if err = archiveLandingPR(t.Root, t.Remote, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir); err != nil {
			return err
		}
		mainOID, err = t.fetchMain(r)
		if err != nil {
			return err
		}
		complete, err = landingArchiveComplete(t.Root, mainOID, t.Repo, pr, f.IssuesDir, f.PlansDir, f.HistoryDir)
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
	fmt.Fprintf(stdout, "Landed PR #%d; workspace retained on unchanged %s. Refresh is separate.\n", pr.Number, t.Rest)
	return nil
}

func runDurablePR(stdout, stderr io.Writer, f *prFlags, t landingTarget) error {
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
	body := combineBody(commits, formatFixes(collectGitHubIssueNumbers(splitNonEmptyLines(changedPaths))))
	if f.DryRun {
		fmt.Fprintf(stdout, "Would: git push -u %s %s\nWould: gh pr create --repo %s --base main --head %s\n%s\n", t.Remote, branch, t.Repo, branch, body)
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
	if _, err = landingGit(prRunner, t.Root, "push", "-u", t.Remote, "refs/heads/"+branch+":refs/heads/"+branch); err != nil {
		return err
	}
	link, err := ghClient.PRCreate(t.Repo, "main", branch, body)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, link)
	return nil
}

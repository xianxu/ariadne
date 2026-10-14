// ghclient.go — shared `gh` CLI shim. Extracted in M5 so fetch / pr /
// merge can all stub `gh issue view` / `gh pr list` / `gh pr create` /
// `gh pr merge` through one seam.
//
// The interface intentionally enumerates the exact calls our verbs make
// rather than exposing a generic Run(args...). That lets each method's
// argv shape be tested in isolation; the production realGH wires the
// concrete `gh` flags. If a future verb needs a new gh call, add a
// method here rather than reaching for a generic Run.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

// ghCaller is the shared interface fetch/pr/merge use to talk to the
// gh CLI. realGH below is the production implementation; tests
// substitute stubGH (in fetch_test.go) or richer fakes per-verb.
type ghCaller interface {
	// RepoVisibility requires a full HTTPS GitHub repository URL; it never
	// infers the authorization host from GH_HOST or the current checkout.
	RepoVisibility(context.Context, string) (string, error)
	// TitleAndBody returns the title + body of GitHub issue issueNum
	// in repo (owner/repo slug). Used by `sdlc issue new --from-github`
	// (and its `sdlc fetch` alias).
	TitleAndBody(repo, issueNum string) (title, body string, err error)

	// IssueClose closes GitHub issue issueNum in repo with the given
	// comment. Used by `sdlc push` after archiving done issues.
	IssueClose(repo, issueNum, comment string) error

	// PRCreate opens a pull request and returns the URL of the new
	// PR. baseRef and headRef are git refs; body is the body text
	// (empty allowed). If body == "" we pass --fill (commit-derived
	// body); otherwise --fill-first plus --body. Mirrors the shell
	// pull-request: target's branching.
	PRCreate(repo, baseRef, headRef, body string) (url string, err error)

	// PRListForBranch returns the GitHub PR number for the open PR
	// targeting `headRef` (the branch name without origin/) in repo,
	// or "" if there is none. Used by `sdlc merge` to find an existing PR.
	PRListForBranch(repo, headRef string) (number string, err error)

	// PRMerge merges PR for branch on repo via the GitHub API
	// (--merge --delete-branch), only while the PR head is expectedHead
	// (--match-head-commit): the head whose checks the CI gate read (#309).
	// Used by `sdlc merge`.
	PRMerge(repo, branch, expectedHead string) error

	// PRChecks returns the checks GitHub reports on PR number's head (#309).
	// Used by `sdlc merge`'s CI gate (cigate.go).
	PRChecks(ctx context.Context, repo string, number int) ([]prCheck, error)

	// PRHeadOID returns the head commit GitHub currently records for PR
	// number. `gh pr checks` reads checks of that head, not of a SHA we name,
	// so the CI gate brackets its read with this to bind the checks to the
	// head it pins (#309 BR-1).
	PRHeadOID(ctx context.Context, repo string, number int) (string, error)

	// PRMergedForBranch reports whether a MERGED PR exists for headRef in
	// repo. `sdlc merge` uses it to detect + resume an interrupted run — PR
	// merged server-side (irreversible) but the local cleanup never finished —
	// instead of erroring on "no open PR" (#62 M3).
	PRMergedForBranch(repo, headRef string) (merged bool, err error)

	// JobRun reads one Actions job, so the CI gate can tell a job that never got
	// a runner from one that ran (#345, localchecks.go jobRun.runnerless).
	JobRun(ctx context.Context, repo string, id int64) (jobRun, error)

	// PostStatus records a commit status on sha: how a local run is recorded
	// as evidence on the exact head (#345, context localStatusContext).
	PostStatus(ctx context.Context, repo, sha, statusContext, state, description string) error

	// CommitStatuses returns the latest state of each commit status on sha
	// (#332: a suite recorded on an ancestor of the landing head).
	CommitStatuses(ctx context.Context, repo, sha string) (map[string]string, error)

	// PRComment posts a comment on PR number (#345: a failed local run's tail).
	PRComment(ctx context.Context, repo string, number int, body string) error
}

// ghClient is the package-level instance every verb resolves through.
// Tests swap it; production keeps the realGH default.
var ghClient ghCaller = realGH{}

// realGH dispatches to the gh CLI on PATH.
type realGH struct{}

func (realGH) TitleAndBody(repo, issueNum string) (string, string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return "", "", fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	titleOut, err := exec.Command("gh", "issue", "view", issueNum,
		"--repo", repo, "--json", "title", "--jq", ".title").Output()
	if err != nil {
		return "", "", fmt.Errorf("gh issue view --jq .title: %w", err)
	}
	bodyOut, err := exec.Command("gh", "issue", "view", issueNum,
		"--repo", repo, "--json", "body", "--jq", ".body // \"\"").Output()
	if err != nil {
		return "", "", fmt.Errorf("gh issue view --jq .body: %w", err)
	}
	title := strings.TrimRight(string(titleOut), "\n")
	body := strings.TrimRight(string(bodyOut), "\n")
	return title, body, nil
}

func (realGH) IssueClose(repo, issueNum, comment string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	args := []string{"issue", "close", issueNum, "--repo", repo}
	if comment != "" {
		args = append(args, "--comment", comment)
	}
	out, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh issue close %s: %w\n%s", issueNum, err, string(out))
	}
	return nil
}

func (realGH) PRCreate(repo, baseRef, headRef, body string) (string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	args := []string{"pr", "create", "--repo", repo, "--base", baseRef, "--head", headRef}
	if body != "" {
		// --fill-first prefills title from the first commit subject, then
		// our --body overrides the body. Matches the shell pull-request
		// target's `gh pr create ... --fill-first --body "$body"` path.
		args = append(args, "--fill-first", "--body", body)
	} else {
		// No collected body (no touched issues, no commits) → let gh
		// derive title + body from the commit history.
		args = append(args, "--fill")
	}
	out, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (realGH) PRListForBranch(repo, headRef string) (string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	out, err := exec.Command("gh", "pr", "list",
		"--repo", repo, "--head", headRef,
		"--json", "number", "--jq", ".[0].number // \"\"",
	).Output()
	if err != nil {
		// `gh pr list` errors when not authed or repo missing; the shell
		// target treats this as "no PR" (`|| true`). Mirror it.
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func (realGH) PRMergedForBranch(repo, headRef string) (bool, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return false, fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	out, err := exec.Command("gh", "pr", "list",
		"--repo", repo, "--head", headRef, "--state", "merged",
		"--json", "number", "--jq", ".[0].number // \"\"",
	).Output()
	if err != nil {
		// Mirror PRListForBranch: a query failure (not authed, repo missing)
		// is treated as "no merged PR" rather than a hard error.
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (realGH) PRMerge(repo, branch, expectedHead string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not on PATH: %w", err)
	}
	out, err := exec.Command("gh", prMergeArgs(repo, branch, expectedHead)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh pr merge: %w\n%s", err, string(out))
	}
	return nil
}

// prMergeArgs is PRMerge's argv: it merges only while the PR head is
// expectedHead, so a push after the CI gate read the checks cannot land
// unchecked (#309).
func prMergeArgs(repo, branch, expectedHead string) []string {
	return []string{"pr", "merge", "--repo", repo, "--merge", "--delete-branch", "--match-head-commit", expectedHead, branch}
}

var visibilityRepositoryURL = regexp.MustCompile(`^https://github[.]com/[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// RepoVisibility queries only the visibility field; raw host errors may contain
// credentials and are deliberately not returned to policy help or refusals.
func (realGH) RepoVisibility(ctx context.Context, repo string) (string, error) {
	if !visibilityRepositoryURL.MatchString(repo) {
		return "", fmt.Errorf("visibility repository must include its publication host: use the full https://github.com/owner/repository URL")
	}
	cmd := exec.CommandContext(ctx, "gh", "repo", "view", repo, "--json", "visibility")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("visibility unavailable")
	}
	if err = cmd.Start(); err != nil {
		return "", fmt.Errorf("visibility unavailable")
	}
	raw, readErr := io.ReadAll(io.LimitReader(out, 4097))
	if len(raw) > 4096 || readErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", fmt.Errorf("invalid visibility response")
	}
	if err = cmd.Wait(); err != nil {
		return "", fmt.Errorf("visibility unavailable")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if token, err := dec.Token(); err != nil || token != json.Delim('{') {
		return "", fmt.Errorf("invalid visibility response")
	}
	if key, err := dec.Token(); err != nil || key != "visibility" {
		return "", fmt.Errorf("invalid visibility response")
	}
	var visibility string
	if err := dec.Decode(&visibility); err != nil || dec.More() {
		return "", fmt.Errorf("invalid visibility response")
	}
	if token, err := dec.Token(); err != nil || token != json.Delim('}') {
		return "", fmt.Errorf("invalid visibility response")
	}
	if _, err := dec.Token(); err != io.EOF {
		return "", fmt.Errorf("invalid visibility response")
	}
	switch visibility {
	case "PUBLIC", "PRIVATE", "INTERNAL":
		return visibility, nil
	}
	return "", fmt.Errorf("unknown repository visibility")
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// landingPR binds integration evidence to an exact repository, PR and head.
// A successful merge request is not evidence: callers must observe MERGED and
// resolve LandedOID (landedCommit) against the freshly fetched configured
// main. MergeOID is GitHub's raw mergeCommit, read only here and by
// landedCommit (#449: a PR merged by reachability, as a train's car, may
// report none, or its own head).
type landingPR struct {
	Number                                                    int
	State, HeadRef, HeadOID, BaseRef, BaseOID, MergeOID, Repo string
	// LandedOID is the commit proven on main that carries this PR: set once,
	// by landedCommit, before any completion or archive reads it.
	LandedOID string
	// LocalGatesVerified: the PR carries the operator's local-gates label, so
	// it rides a train alone (#449). Parsed here, at the boundary.
	LocalGatesVerified bool
}

type landingGH interface {
	LandingPRs(context.Context, string, string) ([]landingPR, error)
	LandingMerge(context.Context, string, int, string) error
}

const landingPRFields = "number,state,headRefName,headRefOid,baseRefName,baseRefOid,mergeCommit,url,isCrossRepository,headRepository,headRepositoryOwner,labels"
const landingPRLimit = 100

func landingOIDValid(oid string) bool {
	if len(oid) != 40 && len(oid) != 64 {
		return false
	}
	if strings.Trim(oid, "0") == "" {
		return false
	}
	for _, c := range oid {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func landingRepoValid(repo string) bool {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasPrefix(part, "-") {
			return false
		}
		for _, c := range part {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
				return false
			}
		}
	}
	return true
}

func parseLandingPRs(data []byte, repo, branch string) ([]landingPR, error) {
	if !landingRepoValid(repo) || strings.TrimSpace(branch) == "" {
		return nil, fmt.Errorf("landing: invalid repository or branch")
	}
	// A null payload is missing evidence, distinct from a successful empty list.
	if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		return nil, fmt.Errorf("landing PR response must be an array")
	}
	var raw []struct {
		Number                                                       int
		State, HeadRefName, HeadRefOID, BaseRefName, BaseRefOID, URL string
		MergeCommit                                                  *struct{ OID string }
		IsCrossRepository                                            *bool
		HeadRepository                                               *struct{ Name, NameWithOwner string }
		HeadRepositoryOwner                                          *struct{ Login string }
		Labels                                                       []struct{ Name string }
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode landing PR evidence: %w", err)
	}
	if len(raw) > landingPRLimit {
		return nil, fmt.Errorf("landing PR evidence exceeds %d records; narrow the branch selection", landingPRLimit)
	}
	out := make([]landingPR, 0, len(raw))
	seen := map[int]bool{}
	for _, r := range raw {
		fail := func(why string) ([]landingPR, error) { return nil, fmt.Errorf("landing PR #%d: %s", r.Number, why) }
		if r.Number <= 0 || seen[r.Number] {
			return fail("invalid or duplicate PR number")
		}
		seen[r.Number] = true
		u, err := url.Parse(r.URL)
		wantPath := "/" + repo + "/pull/" + strconv.Itoa(r.Number)
		if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.EqualFold(u.EscapedPath(), wantPath) {
			return fail("base repository URL does not match the selected GitHub repository")
		}
		if r.IsCrossRepository == nil || *r.IsCrossRepository || r.HeadRepository == nil || r.HeadRepositoryOwner == nil {
			return fail("missing or cross-repository head identity")
		}
		if !strings.EqualFold(r.HeadRepositoryOwner.Login+"/"+r.HeadRepository.Name, repo) || !strings.EqualFold(r.HeadRepository.NameWithOwner, repo) {
			return fail("head repository does not match the selected repository")
		}
		if r.HeadRefName != branch || r.BaseRefName != "main" {
			return fail("head branch or main base does not match")
		}
		if !landingOIDValid(r.HeadRefOID) || !landingOIDValid(r.BaseRefOID) {
			return fail("missing or malformed full head/base commit OID")
		}
		if r.State != "OPEN" && r.State != "MERGED" && r.State != "CLOSED" {
			return fail("unknown PR state")
		}
		merge := ""
		if r.MergeCommit != nil {
			merge = r.MergeCommit.OID
			if !landingOIDValid(merge) {
				return fail("malformed integration OID")
			}
		}
		// A MERGED PR without mergeCommit is legal (#449): a car merged by
		// reachability. landedCommit proves its head is on main instead.
		verified := false
		for _, l := range r.Labels {
			verified = verified || l.Name == localGatesLabel
		}
		out = append(out, landingPR{Number: r.Number, State: r.State, HeadRef: r.HeadRefName, HeadOID: r.HeadRefOID, BaseRef: r.BaseRefName, BaseOID: r.BaseRefOID, MergeOID: merge, Repo: repo, LocalGatesVerified: verified})
	}
	return out, nil
}

// runLandingGH bounds process lifetime and pipe draining, inheriting cancellation
// from the command. stdout is kept separate from diagnostic stderr for JSON.
func runLandingGH(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("gh landing interrupted: %w", ctx.Err())
		}
		return nil, fmt.Errorf("gh landing: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("gh landing interrupted: %w", err)
	}
	return stdout.Bytes(), nil
}
func (realGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	if !landingRepoValid(repo) || strings.TrimSpace(branch) == "" {
		return nil, fmt.Errorf("landing: invalid repository or branch")
	}
	data, err := runLandingGH(ctx, "pr", "list", "--repo", repo, "--head", branch, "--state", "all", "--limit", strconv.Itoa(landingPRLimit+1), "--json", landingPRFields)
	if err != nil {
		return nil, err
	}
	return parseLandingPRs(data, repo, branch)
}
func (realGH) LandingMerge(ctx context.Context, repo string, number int, expectedHead string) error {
	if !landingRepoValid(repo) || number <= 0 || !landingOIDValid(expectedHead) {
		return fmt.Errorf("landing: invalid repository, PR number or expected head")
	}
	_, err := runLandingGH(ctx, "pr", "merge", strconv.Itoa(number), "--repo", repo, "--merge", "--match-head-commit", expectedHead)
	return err
}

var _ landingGH = realGH{}

// landedCommit resolves the commit that proves a MERGED PR is on main: its
// mergeCommit when that is a valid commit on main, else its own head when
// that is on main (a car GitHub marked merged by reachability, #449, may
// report no mergeCommit or one that isn't the train's). Either is the PR's
// exact reviewed head reaching main; neither on main refuses.
func landedCommit(pr landingPR, onMain func(oid string) bool) (string, bool) {
	if pr.State != "MERGED" {
		return "", false
	}
	if landingOIDValid(pr.MergeOID) && onMain(pr.MergeOID) {
		return pr.MergeOID, true
	}
	if landingOIDValid(pr.HeadOID) && onMain(pr.HeadOID) {
		return pr.HeadOID, true
	}
	return "", false
}

// OpenTrainPRs lists open landing-train PRs (#449): the open PRs' head
// names, then each train branch through the validated landing listing.
func (g realGH) OpenTrainPRs(ctx context.Context, repo string) ([]landingPR, error) {
	if !landingRepoValid(repo) {
		return nil, fmt.Errorf("landing: invalid repository")
	}
	data, err := runLandingGH(ctx, "pr", "list", "--repo", repo, "--state", "open", "--limit", strconv.Itoa(landingPRLimit+1), "--json", "headRefName")
	if err != nil {
		return nil, err
	}
	var heads []struct{ HeadRefName string }
	if err := json.Unmarshal(data, &heads); err != nil {
		return nil, fmt.Errorf("decode open PR heads: %w", err)
	}
	var out []landingPR
	for _, h := range heads {
		if !strings.HasPrefix(h.HeadRefName, trainBranchPrefix) {
			continue
		}
		prs, err := g.LandingPRs(ctx, repo, h.HeadRefName)
		if err != nil {
			return nil, err
		}
		for _, p := range prs {
			if p.State == "OPEN" {
				out = append(out, p)
			}
		}
	}
	return out, nil
}

func (realGH) CreateTrainPR(ctx context.Context, repo, branch, title, body string) error {
	if !landingRepoValid(repo) || !strings.HasPrefix(branch, trainBranchPrefix) {
		return fmt.Errorf("landing: invalid repository or train branch")
	}
	_, err := runLandingGH(ctx, "pr", "create", "--repo", repo, "--base", "main", "--head", branch, "--title", title, "--body", body)
	return err
}

func (realGH) CloseTrainPR(ctx context.Context, repo string, number int, branch string) error {
	if !landingRepoValid(repo) || number <= 0 || !strings.HasPrefix(branch, trainBranchPrefix) {
		return fmt.Errorf("landing: invalid repository, PR number or train branch")
	}
	// The branch is deleted by the caller, leased (deleteTrainRef): gh's
	// --delete-branch would delete it unleased, outside the repo lock.
	_, err := runLandingGH(ctx, "pr", "close", strconv.Itoa(number), "--repo", repo)
	return err
}

var _ trainGH = realGH{}

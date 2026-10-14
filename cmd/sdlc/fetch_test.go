package main

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// stubGH stubs ghClient for tests so we don't shell out to `gh`.
// Implements the full ghCaller interface; methods other than the one
// under test are no-ops returning zero values.
type stubGH struct {
	title, body string
	err         error
}

func (s stubGH) TitleAndBody(repo, issueNum string) (string, string, error) {
	return s.title, s.body, s.err
}

func (s stubGH) IssueClose(repo, issueNum, comment string) error        { return nil }
func (s stubGH) PRCreate(repo, base, head, body string) (string, error) { return "", nil }
func (s stubGH) PRListForBranch(repo, headRef string) (string, error)   { return "", nil }
func (s stubGH) PRMerge(repo, branch, expectedHead string) error        { return nil }

// PRChecks defaults to one green check, so a merge test that is not about CI
// passes the gate; the CI gate's own tests override it (cigate_test.go).
func (s stubGH) PRChecks(context.Context, string, int) ([]prCheck, error) {
	return []prCheck{{Name: "merge-check", Bucket: "pass"}}, nil
}

// PRHeadOID: the stub knows no PR head; a fake that reaches the CI gate
// answers with the head its PR carries (e2eGH, landingFakeGH).
func (s stubGH) PRHeadOID(context.Context, string, int) (string, error) {
	return "", errors.New("stubGH has no PR head")
}
func (s stubGH) PRMergedForBranch(repo, headRef string) (bool, error) { return false, nil }

// The #345 methods: a test that reaches them uses ciFakeGH (cigate_fake_test.go).
func (s stubGH) JobRun(context.Context, string, int64) (jobRun, error) {
	return jobRun{}, errors.New("stubGH: JobRun not implemented")
}
func (s stubGH) PostStatus(context.Context, string, string, string, string, string) error {
	return errors.New("stubGH: PostStatus not implemented")
}
func (s stubGH) PRComment(context.Context, string, int, string) error {
	return errors.New("stubGH: PRComment not implemented")
}

// TestFetchAlias_ThroughTree: the hidden `fetch --github-issue N` alias still
// delegates to `issue new --from-github`, now reserving a tracker card (#252).
func TestFetchAlias_ThroughTree(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	prevGH, prevRepo := ghClient, detectRepo
	ghClient = stubGH{title: "Folded Fetch", body: "GH body text."}
	detectRepo = func() (string, error) { return "xianxu/ariadne", nil }
	t.Cleanup(func() { ghClient, detectRepo = prevGH, prevRepo })

	root := buildRoot()
	root.SetArgs([]string{"fetch", "--github-issue", "7"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("execute fetch alias: %v", err)
	}
	body := r.mainFile("workshop/issues/000008-folded-fetch.md") // #308: published at creation
	if body == "" {
		t.Fatal("expected created details on main")
	}
	probIdx, specIdx, ghIdx := strings.Index(body, "## Problem"), strings.Index(body, "## Spec"), strings.Index(body, "GH body text.")
	if probIdx < 0 || ghIdx < probIdx || ghIdx > specIdx {
		t.Errorf("GH body should sit under ## Problem:\n%s", body)
	}
	if card := r.card("workshop/issue-cards/000008-folded-fetch.md"); !strings.Contains(card, "github_issue: 7") {
		t.Errorf("card lost github_issue:\n%s", card)
	}
}

func TestDetectRepo_OriginShapes(t *testing.T) {
	cases := []struct {
		url, want string
	}{
		{"git@github.com:xianxu/ariadne.git\n", "xianxu/ariadne"},
		{"https://github.com/xianxu/ariadne.git\n", "xianxu/ariadne"},
		{"https://github.com/xianxu/ariadne\n", "xianxu/ariadne"},
	}
	for _, c := range cases {
		t.Run(c.url, func(t *testing.T) {
			m := originRE.FindStringSubmatch(c.url)
			if m == nil {
				t.Fatalf("originRE did not match %q", c.url)
			}
			if m[1] != c.want {
				t.Errorf("got %q want %q", m[1], c.want)
			}
		})
	}
}

// gitInit creates a minimal git repo at dir with origin set to remoteURL.
// Returns silently — test fails on git errors via t.Fatal.
func gitInit(t *testing.T, dir, remoteURL string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", remoteURL},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
}

func (s stubGH) RepoVisibility(context.Context, string) (string, error) {
	return "", errors.New("visibility unavailable")
}

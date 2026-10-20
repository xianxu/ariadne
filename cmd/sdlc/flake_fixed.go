package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// fixedFlakeIssue is a terminal card's structured test ownership and landing.
type fixedFlakeIssue struct {
	ID, Merge string
	Tests     []string
}

type fixedFlakeState uint8

const (
	flakeUnknown fixedFlakeState = iota
	flakeFixed
	flakeRecurred
)

var errFixedGitUnknown = errors.New("Git ancestry is unknown")

type flakeListView struct {
	Registry flakes.Registry
	Cards    []flakes.Card
	Fixed    []fixedFlakeIssue
	Main     string
	FixedErr error
}

// loadFlakeListView reads the tracker only once. Fixed-issue evidence is
// advisory to this read-only view, so its failure never masks a failure row.
func loadFlakeListView(ctx context.Context, issuesDir string) (flakeListView, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil {
		return flakeListView{}, err
	}
	cards, err := flakeCards(rs)
	if err != nil {
		return flakeListView{}, err
	}
	registry, err := flakes.BuildRegistry(cards, flakeIssueOpen)
	if err != nil {
		return flakeListView{}, err
	}
	view := flakeListView{Registry: registry, Cards: cards}
	view.Fixed, view.Main, view.FixedErr = publishedFixedIssues(ctx, issuesDir, rs)
	return view, nil
}

func flakeGitContains(ctx context.Context, root, ancestor, head string) (bool, error) {
	_, err := gitx.GitOutputContext(ctx, "-C", root, "merge-base", "--is-ancestor", ancestor, head)
	if err == nil {
		return true, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

func publishedFixedIssues(ctx context.Context, issuesDir string, rs tracker.Records) ([]fixedFlakeIssue, string, error) {
	root := repoRootOfContext(ctx, issuesDir)
	if root == "" {
		return nil, "", errors.New("issue cards are outside a repository")
	}
	trunk, err := peerTrunk(ctx, root)
	if err != nil {
		return nil, "", err
	}
	main, err := trunk.Snapshot()
	if err != nil {
		return nil, "", err
	}
	var fixed []fixedFlakeIssue
	for _, rec := range rs.All() {
		if rec.Duplicate || rec.Card == nil || !vocab.Issue().IsTerminal(rec.Status()) {
			continue
		}
		tests, _, err := issue.FieldList(rec.Card.Card.Frontmatter, flakeTestsField)
		if err != nil {
			return nil, "", fmt.Errorf("issue %s flake_tests: %w", rec.ID, err)
		}
		completion, _, err := issue.CardCompletion(rec.Card.Raw)
		if err != nil {
			return nil, "", fmt.Errorf("issue %s completion: %w", rec.ID, err)
		}
		fixed = append(fixed, fixedFlakeIssue{ID: rec.ID, Merge: completion.LandedCommit, Tests: tests})
	}
	return fixed, main.Ref(), nil
}

// fixedCandidates are the terminal issues whose structured flake_tests cover
// failing leaf name by the gate's rule (flakes.Covers, #479): the nearest
// covering name wins, as in Registry.Cover, and issues tied at it all return.
func fixedCandidates(issues []fixedFlakeIssue, name string) []fixedFlakeIssue {
	var candidates []fixedFlakeIssue
	best := ""
	for _, candidate := range issues {
		for _, test := range candidate.Tests {
			if !flakes.Covers(test, name) || len(test) < len(best) {
				continue
			}
			if len(test) > len(best) {
				best, candidates = test, nil
			}
			if !slices.ContainsFunc(candidates, func(c fixedFlakeIssue) bool { return c.ID == candidate.ID }) {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

// classifyFixedFlake coordinates ownership with injected Git ancestry. It is
// integration logic: fixedCandidates is the pure card matcher.
func classifyFixedFlake(name string, issues []fixedFlakeIssue, contains func(string, string) (bool, error), main, head string) (fixedFlakeState, fixedFlakeIssue, string) {
	candidates := fixedCandidates(issues, name)
	if len(candidates) > 1 {
		return flakeUnknown, fixedFlakeIssue{}, "multiple terminal issues cover this test"
	}
	if len(candidates) == 0 {
		return flakeUnknown, fixedFlakeIssue{}, ""
	}
	if candidates[0].Merge == "" {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s has no landing merge", candidates[0].ID)
	}
	onMain, err := contains(candidates[0].Merge, main)
	if err != nil {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("Git ancestry unavailable for issue #%s: %v", candidates[0].ID, err)
	}
	if !onMain {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("issue #%s landing merge is not on main", candidates[0].ID)
	}
	inHead, err := contains(candidates[0].Merge, head)
	if err != nil {
		return flakeUnknown, fixedFlakeIssue{}, fmt.Sprintf("Git ancestry unavailable for issue #%s: %v", candidates[0].ID, err)
	}
	if inHead {
		return flakeRecurred, candidates[0], ""
	}
	return flakeFixed, candidates[0], ""
}

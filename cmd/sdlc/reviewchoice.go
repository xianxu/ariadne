// reviewchoice.go — the IO side of #409's reviewer choice: the window's author
// families, the reviewer policy and what can run here, resolved once for every
// review site (boundary, fix, plan-quality, estimate-quality, planning review,
// sdlc judge).
package main

import (
	"context"
	"os/exec"
	"runtime"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// reviewAuthorFamilies is every vendor family that wrote base..head: the
// window's own commits' Agent: trailers (first parent, no merges, as #367's
// report reads them) and always the session running the review. The trailers
// are untrusted, so a forged one can add a family but never remove the
// closer's own. When the window cannot be read, unread says why: the caller
// records it as degraded rather than quietly narrowing the exclusion.
func reviewAuthorFamilies(ctx context.Context, base, head string) (families []judge.Family, unread string) {
	session := judge.FamilyOf(string(judge.MainAgent(judge.CurrentAgentDefaultEnv())))
	if base == "" || head == "" {
		return []judge.Family{session}, "authors unread: no window base (origin/main unknown)"
	}
	out, err := gitx.GitOutputContext(ctx, "log", "--first-parent", "--no-merges", "--format="+agentid.LogFormat, base+".."+head)
	if err != nil {
		return []judge.Family{session}, "authors unread: " + strings.TrimSpace(firstLineOf(err.Error()))
	}
	return authorFamiliesFromLog(string(out), session), ""
}

// authorFamiliesFromLog is the families an agentid.LogFormat log names in its
// Agent: trailers (`<agent>/<model>/<effort>`), plus the session's: merges and
// agents sdlc does not know contribute nothing.
func authorFamiliesFromLog(log string, session judge.Family) []judge.Family {
	families := []judge.Family{session}
	for _, c := range agentid.ParseLog(log) {
		if c.Merge {
			continue
		}
		for _, v := range c.Agents {
			agent, _, _ := strings.Cut(strings.TrimSpace(v), "/")
			if f := judge.FamilyOf(agent); f != "" {
				families = append(families, f)
			}
		}
	}
	return families
}

func firstLineOf(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// resolveReviewChoice is a review site's reviewers: the repo policy read from
// root, the families that authored base..head, the change kind and what can
// launch under containment here.
func resolveReviewChoice(ctx context.Context, root, base, head, explicit string, explicitSet bool, kind judge.ChangeKind, containment judge.Containment) (judge.ReviewChoice, error) {
	policy, err := loadReviewerPolicy(root)
	if err != nil {
		return judge.ReviewChoice{}, err
	}
	return resolveReviewChoiceWith(ctx, policy, base, head, explicit, explicitSet, kind, containment), nil
}

func resolveReviewChoiceWith(ctx context.Context, policy judge.ReviewerPolicy, base, head, explicit string, explicitSet bool, kind judge.ChangeKind, containment judge.Containment) judge.ReviewChoice {
	authors, unread := reviewAuthorFamilies(ctx, base, head)
	choice := judge.ResolveReview(explicit, explicitSet, judge.CurrentAgentDefaultEnv(), policy, authors, kind, reviewAvailability(containment))
	if unread != "" {
		choice.Degraded = append(choice.Degraded, unread)
	}
	return choice
}

// reviewAvailability is what can launch here: the host's PATH and platform.
// Tests replace it so no result depends on which CLIs the machine has.
var reviewAvailability = func(containment judge.Containment) judge.Availability {
	return judge.LocalAvailability(containment, runtime.GOOS, exec.LookPath)
}

// branchWindowBase is where the current branch left main, for a review that
// has no boundary window of its own (plan-quality, planning review): "" when
// it cannot be read, so only the session's own family counts as an author.
func branchWindowBase(ctx context.Context) string {
	out, err := gitx.GitOutputContext(ctx, "merge-base", "HEAD", "origin/main")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// boundaryChangeKind is a boundary review's change kind: the full review judges
// a design-heavy change, the small-diff and fix reviews a bounded one.
func boundaryChangeKind(cat judge.Category) judge.ChangeKind {
	if cat == judge.MilestoneReview {
		return judge.KindDesignHeavy
	}
	return judge.KindBounded
}

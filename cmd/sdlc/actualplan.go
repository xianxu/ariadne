package main

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// pinnedActualPlan resolves the canonical issue body at the measurement's code
// generation, independently of the caller's checkout and later history.
func pinnedActualPlan(ctx context.Context, root, pin, id string) ([]string, error) {
	n, err := strconv.Atoi(id)
	if err != nil || n <= 0 {
		return nil, fmt.Errorf("invalid measurement issue %q", id)
	}
	active := envOr("WF_ISSUES_DIR", "workshop/issues")
	if filepath.IsAbs(active) {
		active, err = filepath.Rel(root, active)
		if err != nil {
			return nil, err
		}
	}
	active = filepath.ToSlash(filepath.Clean(active))
	if active == ".." || strings.HasPrefix(active, "../") {
		return nil, fmt.Errorf("measurement issue root must be inside repository")
	}
	listing, err := ownedActualGit(ctx, root, "--literal-pathspecs", "ls-tree", "-r", "--name-only", pin, "--", active, "workshop/history/issues")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, path := range strings.Split(listing, "\n") {
		found, _, ok := issue.ParseFilename(filepath.Base(path))
		if ok && found == fmt.Sprintf("%06d", n) && (filepath.ToSlash(filepath.Dir(path)) == active || filepath.ToSlash(filepath.Dir(path)) == "workshop/history/issues") {
			paths = append(paths, path)
		}
	}
	if len(paths) != 1 {
		return nil, fmt.Errorf("pinned issue #%d needs one canonical Plan file at %s; found %d", n, pin, len(paths))
	}
	raw, err := ownedActualGit(ctx, root, "show", pin+":"+paths[0])
	if err != nil {
		return nil, err
	}
	fm, body, err := issue.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", paths[0], err)
	}
	found, err := issue.FieldString(fm, "id")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", paths[0], err)
	}
	if found != fmt.Sprintf("%06d", n) {
		return nil, fmt.Errorf("%s: pinned issue identity differs from #%d", paths[0], n)
	}
	lines := strings.Split(body, "\n")
	if !slices.Equal(issue.FenceSpans(lines, issue.UnterminatedIsFenced), issue.FenceSpans(lines, issue.UnterminatedIsProse)) {
		return nil, fmt.Errorf("%s: unclosed code fence makes Plan unverified", paths[0])
	}
	count := 0
	for _, line := range strings.Split(issue.StripFenced(body), "\n") {
		if strings.HasPrefix(line, "## ") && strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Plan" {
			count++
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("%s: expected one canonical Plan section, found %d", paths[0], count)
	}
	plan, _ := issue.PlanItemsBody(body)
	return issue.MilestonesInPlanOrder(plan), nil
}

func freshActualRounds(ctx context.Context, root, pin, id, milestone string, rounds []activetime.BoundaryRound) ([]activetime.BoundaryRound, []string, error) {
	finalized := map[string]bool{}
	for _, r := range rounds {
		if r.Finalizing {
			finalized[r.Boundary] = true
		}
	}
	needsPlan := false
	for _, r := range rounds {
		needsPlan = needsPlan || r.Boundary != "" && !finalized[r.Boundary]
	}
	if !needsPlan {
		return rounds, nil, nil
	}
	milestones, err := pinnedActualPlan(ctx, root, pin, id)
	if err != nil {
		return nil, nil, fmt.Errorf("measurement Plan unavailable: %w; restore the canonical issue Plan, commit it, then retry sdlc actual --issue %s", err, issue.CLIRef(id))
	}
	if milestone != "" && !finalized[milestone] && !slices.Contains(milestones, milestone) {
		return nil, nil, fmt.Errorf("milestone %s is absent from the pinned executable Plan at %s; restore and commit its executable row or select its replacement with sdlc actual --issue %s --milestone <replacement>", milestone, pin, issue.CLIRef(id))
	}
	selected, retired := selectFreshSplitRounds(rounds, milestones)
	var notes []string
	if len(retired) > 0 {
		notes = append(notes, fmt.Sprintf("retired unfinalized boundaries %s absent from executable Plan at code %s; replacement build includes work since the last finalized boundary (or issue start)", strings.Join(retired, ", "), pin))
	}
	if milestone != "" {
		seen := map[string]bool{}
		for _, r := range selected {
			if r.Boundary != "" && r.Boundary != milestone && !finalized[r.Boundary] && !seen[r.Boundary] {
				notes = append(notes, fmt.Sprintf("retained executable unfinalized boundary %s; its open convergence may leave %s with no selected phases", r.Boundary, milestone))
				seen[r.Boundary] = true
			}
		}
	}
	return selected, notes, nil
}

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pair#283's names (#235): the plan was saved under a shorter slug than the
// issue file's, and exact-name lookup never saw it.
const (
	pair283Stem    = "000283-default-cursor-style-overrides-terminal"
	pair283OffSlug = "000283-default-cursor-style-plan.md"
)

func writePlanFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("# "+n+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestPlanArtifactPathFindsThePlanByID drives the lookup over sets of the
// id's family. Whatever the set, the result is the canonical path or one
// present -plan.md file, and two plans never resolve to either.
func TestPlanArtifactPathFindsThePlanByID(t *testing.T) {
	sidecars := []string{
		"000283-default-cursor-style-overrides-terminal-plan-gate.md",
		"000283-default-cursor-style-overrides-terminal-close-gate.md",
		"000283-default-cursor-style-overrides-terminal-m1-review.md",
		"000283-default-cursor-style-overrides-terminal-close-review.md",
		"0002830-longer-id-plan.md",
		"000284-neighbour-plan.md",
	}
	canonical := pair283Stem + "-plan.md"
	tests := map[string]struct {
		stem    string
		files   []string
		want    string // base name; "" with wantErr
		wantErr []string
	}{
		"canonical name":          {stem: pair283Stem, files: []string{canonical}, want: canonical},
		"pair#283's off-slug":     {stem: pair283Stem, files: []string{pair283OffSlug}, want: pair283OffSlug},
		"no plan":                 {stem: pair283Stem, want: canonical},
		"sidecars only":           {stem: pair283Stem, files: sidecars, want: canonical},
		"off-slug among sidecars": {stem: pair283Stem, files: append([]string{pair283OffSlug}, sidecars...), want: pair283OffSlug},
		"two plans refuse": {stem: pair283Stem, files: append([]string{canonical, pair283OffSlug}, sidecars...),
			wantErr: []string{canonical, pair283OffSlug, "keep one"}},
		"stem without an id": {stem: "freeform", files: []string{"freeform-plan.md", pair283OffSlug}, want: "freeform-plan.md"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writePlanFiles(t, dir, tc.files...)
			got, err := planArtifactPath(dir, tc.stem)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("picked %s from an ambiguous set", got)
				}
				for _, w := range tc.wantErr {
					if !strings.Contains(err.Error(), w) {
						t.Errorf("refusal %q does not name %q", err, w)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Dir(got) != dir || filepath.Base(got) != tc.want {
				t.Fatalf("plan = %s, want %s", got, filepath.Join(dir, tc.want))
			}
			if !strings.HasSuffix(got, "-plan.md") {
				t.Fatalf("took %s for the plan", got)
			}
		})
	}
}

func TestReadOptionalPlanFileRefusesAmbiguity(t *testing.T) {
	dir := t.TempDir()
	writePlanFiles(t, dir, pair283OffSlug)
	got, err := readOptionalPlanFile(dir, pair283Stem)
	if err != nil || got != "# "+pair283OffSlug+"\n" {
		t.Fatalf("off-slug read = %q, %v", got, err)
	}
	writePlanFiles(t, dir, pair283Stem+"-plan.md")
	if got, err := readOptionalPlanFile(dir, pair283Stem); err == nil {
		t.Fatalf("two plans read as %q, want a refusal", got)
	}
}

func TestPlanNameNote(t *testing.T) {
	if n := planNameNote("p", pair283Stem, filepath.Join("p", pair283Stem+"-plan.md")); n != "" {
		t.Fatalf("canonical plan noted: %q", n)
	}
	n := planNameNote("p", pair283Stem, filepath.Join("p", pair283OffSlug))
	if !strings.Contains(n, pair283OffSlug) || !strings.Contains(n, pair283Stem+"-plan.md") {
		t.Fatalf("note %q names neither the found nor the canonical plan", n)
	}
}

// TestResolveBoundaryReviewManifest_FindsOffSlugPlan: the close review sees a
// plan saved under another slug, and refuses two.
func TestResolveBoundaryReviewManifest_FindsOffSlugPlan(t *testing.T) {
	root := t.TempDir()
	issues, plans := "workshop/issues", "workshop/plans"
	writeRepoFile(t, root, filepath.Join(issues, pair283Stem+".md"), "# issue\n")
	writePlanFiles(t, filepath.Join(root, plans), pair283OffSlug, pair283Stem+"-plan-gate.md")
	req := boundaryReviewManifestRequest{
		BaseRef: "base", HeadRef: "head", IssuesDir: issues, HistoryDir: "workshop/history", IssueNum: 283, PlansDir: plans,
	}
	fake := &fakeBoundaryGit{root: root, refs: map[string]string{"base": testReviewBaseSHA, "head": testReviewHeadSHA}}
	m, err := resolveBoundaryReviewManifest(fake, req)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(plans, pair283OffSlug); m.PlanFile != want {
		t.Fatalf("review plan = %q, want %q", m.PlanFile, want)
	}

	writePlanFiles(t, filepath.Join(root, plans), pair283Stem+"-plan.md")
	fake = &fakeBoundaryGit{root: root, refs: map[string]string{"base": testReviewBaseSHA, "head": testReviewHeadSHA}}
	if _, err := resolveBoundaryReviewManifest(fake, req); err == nil || !strings.Contains(err.Error(), pair283OffSlug) {
		t.Fatalf("two plans in the review manifest: %v", err)
	}
}

// TestChangeCodeFindsOffSlugPlan reproduces pair#283 end to end: a durable plan
// past the design limit, saved under another slug, makes change-code infer
// full and reaches the plan-quality prompt.
func TestChangeCodeFindsOffSlugPlan(t *testing.T) {
	const slug = "default-cursor-style-overrides-terminal"
	cardPath, card, detailPath, detail := seededIssue(t, "000283", slug)
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(283)); err != nil {
		t.Fatal(err)
	}
	if err := startPlanBranch(context.Background(), &out, 283); err != nil {
		t.Fatal(err)
	}
	const marker = "OFF-SLUG PLAN MARKER"
	writeRepoFile(t, r.root, "workshop/plans/"+pair283OffSlug, marker+"\n"+strings.Repeat("- design line\n", 600))

	f := &changeCodeFlags{Issue: 283, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", Worktree: "no",
		NoStructural: true, NoEstimate: true, NoEstimateRecon: true, DryRun: true}
	out.Reset()
	errs.Reset()
	if err := runChangeCode(strings.NewReader(""), &out, &errs, f); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if !strings.Contains(errs.String(), "flow: full (inferred") {
		t.Errorf("flow not inferred full from the off-slug plan:\n%s", errs.String())
	}
	if !strings.Contains(errs.String(), "its canonical name is "+pair283Stem+"-plan.md") {
		t.Errorf("no note naming the canonical plan name:\n%s", errs.String())
	}
	if !strings.Contains(out.String(), marker) {
		t.Errorf("plan-quality prompt lacks the off-slug plan:\n%s", out.String())
	}

	// The manual render reads the plan through the same lookup.
	errs.Reset()
	prompt := planQualityJudgePrompt(&errs, &judgeFlags{Issue: 283, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", DryRun: true})
	if !strings.Contains(prompt, marker) {
		t.Errorf("judge plan-quality --dry-run lacks the off-slug plan")
	}

	// The design checkpoint commits the plan where it lies.
	f.DryRun, f.NoJudge = false, true
	out.Reset()
	errs.Reset()
	if err := runChangeCode(strings.NewReader(""), &out, &errs, f); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if files := r.git("show", "--name-only", "--format=", "HEAD"); !strings.Contains(files, "workshop/plans/"+pair283OffSlug) {
		t.Errorf("design checkpoint carried %q, not the off-slug plan", files)
	}
}

// TestPlanLookupToleratesGlobMetacharacters: a repository path containing
// glob metacharacters still finds the off-slug plan (#235 close review).
func TestPlanLookupToleratesGlobMetacharacters(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "repo[1]*?")
	writePlanFiles(t, dir, pair283OffSlug)
	got, err := planArtifactPath(dir, pair283Stem)
	if err != nil || filepath.Base(got) != pair283OffSlug {
		t.Fatalf("plan = %s, %v; want %s", got, err, pair283OffSlug)
	}
}

// TestReviewPlanPathsRefusesTwoPlans: close's review snapshot resolves the
// plan through reviewPlanPaths, which passes the ambiguity refusal up rather
// than reading it as no plan.
func TestReviewPlanPathsRefusesTwoPlans(t *testing.T) {
	root := t.TempDir()
	writePlanFiles(t, filepath.Join(root, "workshop/plans"), pair283OffSlug, pair283Stem+"-plan.md")
	display, stat, err := reviewPlanPaths(root, "workshop/plans", "workshop/issues/"+pair283Stem+".md")
	if err == nil || !strings.Contains(err.Error(), pair283OffSlug) {
		t.Fatalf("two plans resolved to %q (%q), %v", display, stat, err)
	}
}

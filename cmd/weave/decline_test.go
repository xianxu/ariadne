package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// declineFixture extends buildFixture's base with two seeded workflows and a
// seeded entrypoint, the shape of ariadne's own manifest (#467). The derived
// manifest gets extra rows appended.
func declineFixture(t *testing.T, derivedRows string) (base, derived string) {
	t.Helper()
	_, base, derived = buildFixture(t)
	mkfile(t, filepath.Join(base, "construct", "base.manifest"),
		"prose AGENTS.local.md\nscaffold .claude/skills\nsymlink shared.md\n"+
			"seed .github/workflows/merge-check.yml\nseed .github/workflows/local-gates-guard.yml\nseed bootstrap.sh\n")
	mkfile(t, filepath.Join(base, ".github/workflows/merge-check.yml"), "MERGE CHECK")
	mkfile(t, filepath.Join(base, ".github/workflows/local-gates-guard.yml"), "GUARD")
	mkfile(t, filepath.Join(base, "bootstrap.sh"), "BOOTSTRAP")
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"),
		"prose AGENTS.local.md\nsymlink selfdoc.md\n"+derivedRows)
	return base, derived
}

func compileOK(t *testing.T, root string, dryRun bool) string {
	t.Helper()
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, root, plan.TargetAll, dryRun, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	return out.String()
}

func TestCompileSeedsWorkflowsWithoutDecline(t *testing.T) {
	_, derived := declineFixture(t, "")
	out := compileOK(t, derived, false)
	if b, err := os.ReadFile(filepath.Join(derived, ".github/workflows/merge-check.yml")); err != nil || string(b) != "MERGE CHECK" {
		t.Fatalf("normal repo did not get its workflow: %q %v", b, err)
	}
	if strings.Contains(out, "declined") {
		t.Fatalf("report mentions a decline the repo never made:\n%s", out)
	}
}

func TestCompileDeclinedWorkflowsStayAbsent(t *testing.T) {
	_, derived := declineFixture(t, "decline .github/workflows\n")
	for _, dryRun := range []bool{true, false} {
		out := compileOK(t, derived, dryRun)
		want := "weave: declined .github/workflows: opted out of 2 inherited row(s): seed .github/workflows/merge-check.yml, seed .github/workflows/local-gates-guard.yml"
		if !strings.Contains(out, want) {
			t.Fatalf("dryRun=%v: report missing %q:\n%s", dryRun, want, out)
		}
		if dryRun && strings.Contains(out, ".github/workflows/merge-check.yml\n") {
			t.Fatalf("dry-run previews a declined seed:\n%s", out)
		}
	}
	if _, err := os.Stat(filepath.Join(derived, ".github")); !os.IsNotExist(err) {
		t.Fatalf("declined workflows were seeded: %v", err)
	}
	// The repo's other seeded files still refresh.
	if b, err := os.ReadFile(filepath.Join(derived, "bootstrap.sh")); err != nil || string(b) != "BOOTSTRAP" {
		t.Fatalf("other seed did not land: %q %v", b, err)
	}
}

func TestCompileDeclineLeavesExistingWorkflowsUntouched(t *testing.T) {
	base, derived := declineFixture(t, "")
	compileOK(t, derived, false) // seeded once, before the repo opted out
	kept := filepath.Join(derived, ".github/workflows/merge-check.yml")
	if err := os.WriteFile(kept, []byte("REPO'S OWN"), 0644); err != nil {
		t.Fatal(err)
	}
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"),
		"prose AGENTS.local.md\nsymlink selfdoc.md\ndecline .github/workflows\n")
	mkfile(t, filepath.Join(base, ".github/workflows/merge-check.yml"), "MERGE CHECK v2")
	compileOK(t, derived, false)
	if b, err := os.ReadFile(kept); err != nil || string(b) != "REPO'S OWN" {
		t.Fatalf("an opted-out repo's workflow was modified: %q %v", b, err)
	}
	// The unchanged one weave seeded earlier is neither rewritten nor retired.
	if b, err := os.ReadFile(filepath.Join(derived, ".github/workflows/local-gates-guard.yml")); err != nil || string(b) != "GUARD" {
		t.Fatalf("an earlier seed was retired or rewritten: %q %v", b, err)
	}
}

func TestCompileReportsStaleAndDependentDeclines(t *testing.T) {
	base, derived := declineFixture(t, "decline nowhere\ndecline .claude\n")
	mkfile(t, filepath.Join(base, "construct", "base.manifest"),
		"prose AGENTS.local.md\nmerge .claude/settings.base.json .claude/settings.json\n")
	mkfile(t, filepath.Join(base, ".claude/settings.base.json"), "{}")
	out := compileOK(t, derived, false)
	if !strings.Contains(out, "weave: warning: decline nowhere matches no inherited row") {
		t.Fatalf("a stale decline was silent:\n%s", out)
	}
	// A merge into a declined path is named, never dropped silently.
	if !strings.Contains(out, "weave: declined .claude: opted out of 1 inherited row(s): merge .claude/settings.json") {
		t.Fatalf("a declined merge was not reported:\n%s", out)
	}
}

func TestCompileDeclinedSymlinkRetires(t *testing.T) {
	_, derived := declineFixture(t, "")
	compileOK(t, derived, false)
	if _, err := os.Lstat(filepath.Join(derived, "shared.md")); err != nil {
		t.Fatalf("symlink not lowered before the decline: %v", err)
	}
	mkfile(t, filepath.Join(derived, "construct", "base.manifest"),
		"prose AGENTS.local.md\nsymlink selfdoc.md\ndecline shared.md\n")
	compileOK(t, derived, false)
	// A declined link is no longer produced, so its unchanged recorded link retires.
	if _, err := os.Lstat(filepath.Join(derived, "shared.md")); !os.IsNotExist(err) {
		t.Fatalf("declined symlink survived: %v", err)
	}
}

func TestCompileTwoPathDeclineWarnsAndDeclinesNothing(t *testing.T) {
	// First compile, so bootstrap.sh exists only if the ignored row did not decline it.
	_, derived := declineFixture(t, "decline bootstrap.sh extra\n")
	out := compileOK(t, derived, false)
	if !strings.Contains(out, "weave: warning: decline takes one path; ignored the row: decline bootstrap.sh extra") {
		t.Fatalf("a two-path decline was silent:\n%s", out)
	}
	if b, err := os.ReadFile(filepath.Join(derived, "bootstrap.sh")); err != nil || string(b) != "BOOTSTRAP" {
		t.Fatalf("the ignored row declined its path: %q %v", b, err)
	}
}

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectCloseRequiresExecutingAndPointsPausedAtResume(t *testing.T) {
	tests := []struct {
		status, want string
	}{
		{"committed", "requires status executing"},
		{"paused", "sdlc project set-status --to executing"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			f, _, _ := projectCloseFixture(t, tt.status, true, true)
			if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("runProjectClose error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestProjectCloseRequiresRetroUnlessBypassed(t *testing.T) {
	f, _, _ := projectCloseFixture(t, "executing", false, true)
	if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), "--no-retro") {
		t.Fatalf("runProjectClose error = %v, want --no-retro pointer", err)
	}
	f.NoRetro = true
	var stderr bytes.Buffer
	if err := runProjectClose(&bytes.Buffer{}, &stderr, f); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "--no-retro (or --force)") {
		t.Fatalf("stderr missing bypass acknowledgement: %s", stderr.String())
	}
}

func TestProjectCloseRecordsFogAndArchives(t *testing.T) {
	f, projectPath, ledgerPath := projectCloseFixture(t, "executing", true, true)
	originalLookup := projectIssueLookupFn
	projectIssueLookupFn = func(ref, _ string) (issueMeta, error) {
		return map[string]issueMeta{
			"ariadne#1": {ActualHours: 10},
			"ariadne#2": {ActualHours: 30},
		}[ref], nil
	}
	t.Cleanup(func() { projectIssueLookupFn = originalLookup })

	var stdout, stderr bytes.Buffer
	if err := runProjectClose(&stdout, &stderr, f); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Fatalf("live project still exists: %v", err)
	}
	archived := filepath.Join(f.HistoryDir, "projects", "alpha.md")
	b, err := os.ReadFile(archived)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, want := range []string{"status: done", "updated: 2026-07-16", "### 2026-07-16 — close", "phase-a: 40h", "actuals: 40h", "fog: 1.00"} {
		if !strings.Contains(text, want) {
			t.Errorf("archived project missing %q:\n%s", want, text)
		}
	}
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ledger), "| alpha | 40h | 40h | 1.00 | 2026-07-16 |") {
		t.Fatalf("ledger row missing:\n%s", ledger)
	}
	if !strings.Contains(stdout.String(), archived) || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestProjectCloseMissingPhaseAWarnsAndLogsNAWithoutLedger(t *testing.T) {
	f, _, ledgerPath := projectCloseFixture(t, "executing", true, false)
	before, _ := os.ReadFile(ledgerPath)
	var stderr bytes.Buffer
	if err := runProjectClose(&bytes.Buffer{}, &stderr, f); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "phase-a") {
		t.Fatalf("missing phase-a warning: %s", stderr.String())
	}
	after, _ := os.ReadFile(ledgerPath)
	if string(after) != string(before) {
		t.Fatalf("ledger changed without phase-a:\nbefore=%s\nafter=%s", before, after)
	}
	b, _ := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
	if !strings.Contains(string(b), "fog: n/a") {
		t.Fatalf("missing fog n/a close log:\n%s", b)
	}
}

func TestProjectCloseDropFromPausedArchivesWithoutLedger(t *testing.T) {
	f, _, ledgerPath := projectCloseFixture(t, "paused", true, true)
	f.Drop = true
	before, _ := os.ReadFile(ledgerPath)
	if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "status: dropped") {
		t.Fatalf("drop did not set terminal status:\n%s", b)
	}
	after, _ := os.ReadFile(ledgerPath)
	if string(after) != string(before) {
		t.Fatal("drop unexpectedly wrote fog ledger")
	}
}

func TestProjectCloseDropRejectsPreExecutionFunnelStatus(t *testing.T) {
	f, _, _ := projectCloseFixture(t, "committed", true, true)
	f.Drop = true
	if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), "requires status executing or paused") {
		t.Fatalf("runProjectClose error = %v, want executing-or-paused refusal", err)
	}
}

func projectCloseFixture(t *testing.T, status string, retro, phaseA bool) (*projectCloseFlags, string, string) {
	t.Helper()
	root := t.TempDir()
	projects := filepath.Join(root, "workshop", "projects")
	history := filepath.Join(root, "workshop", "history")
	brain := filepath.Join(root, "brain")
	if err := os.MkdirAll(projects, 0o755); err != nil {
		t.Fatal(err)
	}
	log := ""
	if retro {
		log = "\n### 2026-07-15 — retro\n\nLearned.\n"
	}
	estimate := ""
	if phaseA {
		estimate = "\n**phase-a:** 40h\n"
	}
	project := "---\ntype: project\nname: alpha\ngoal: ship\ndone_when: shipped\nstatus: " + status + "\ncreated: 2026-07-01\nupdated: 2026-07-15\ndeadline: 2026-08-01\nplanned_finish: 2026-07-30\nmvp_scope: [ariadne#1, ariadne#2]\n---\n\n# alpha\n\n## Estimate\n" + estimate + "\n## Breakdown\n\n- [x] one [ariadne#1]\n- [x] two [ariadne#2]\n\n## Log\n" + log
	projectPath := filepath.Join(projects, "alpha.md")
	if err := os.WriteFile(projectPath, []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(brain, "data", "life", "42shots", "velocity", "estimate-logic-project-v1.md")
	if err := os.MkdirAll(filepath.Dir(ledgerPath), 0o755); err != nil {
		t.Fatal(err)
	}
	ledger := "# Phase-A\n\n## Fog ledger\n\n| project | phase-a | actuals | fog | closed |\n|---|---:|---:|---:|---|\n"
	if err := os.WriteFile(ledgerPath, []byte(ledger), 0o644); err != nil {
		t.Fatal(err)
	}
	return &projectCloseFlags{Slug: "alpha", ProjectsDir: projects, HistoryDir: history, BrainDir: brain}, projectPath, ledgerPath
}

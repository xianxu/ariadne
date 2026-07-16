package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
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
	f.NoLedger = true
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
			"ariadne#1": {ActualHours: 10, ActualAvailable: true},
			"ariadne#2": {ActualHours: 30, ActualAvailable: true},
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

func TestAppendProjectLedgerRowHandlesEOFAndStructuralHeadings(t *testing.T) {
	row := "| alpha | 40h | 20h | 0.50 | 2026-07-16 |"
	for _, text := range []string{
		"# Ledger\n\n## Fog ledger\n\n| project | phase-a | actuals | fog | closed |\n|---|---:|---:|---:|---|",
		"# Ledger\n\n## Fog ledger\n\n| project | phase-a | actuals | fog | closed |\n|---|---:|---:|---:|---|\n| old | 1h | 1h | 1.00 | 2026-01-01 |",
	} {
		got, err := appendProjectLedgerRow(text, row)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(got, row) {
			t.Fatalf("row not appended safely at EOF:\n%s", got)
		}
	}

	text := "# Ledger\n\n`## Fog ledger` is the required section.\n\n```md\n## Fog ledger\n| fake | table |\n|---|---|\n```\n\n## Fog ledger\n\n| project | phase-a | actuals | fog | closed |\n|---|---:|---:|---:|---|\n\n## Notes\nkeep\n"
	got, err := appendProjectLedgerRow(text, row)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, row) != 1 || !strings.Contains(got, "|---|---:|---:|---:|---|\n"+row+"\n\n## Notes") {
		t.Fatalf("row was not confined to the real Fog ledger section:\n%s", got)
	}
	if !strings.Contains(got, "| fake | table |\n|---|---|\n```") {
		t.Fatalf("fenced example changed:\n%s", got)
	}
}

func TestProjectCloseReadsQuotedStatusAndBlockMVPScope(t *testing.T) {
	f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
	b, _ := os.ReadFile(projectPath)
	text := strings.Replace(string(b), "status: executing", "status: 'executing'", 1)
	text = strings.Replace(text, "mvp_scope: [ariadne#1, ariadne#2]", "mvp_scope:\n  - ariadne#1\n  - ariadne#2", 1)
	if err := os.WriteFile(projectPath, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	original := projectIssueLookupFn
	projectIssueLookupFn = func(string, string) (issueMeta, error) {
		return issueMeta{ActualHours: 20, ActualAvailable: true}, nil
	}
	t.Cleanup(func() { projectIssueLookupFn = original })
	if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	archived, err := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
	if err != nil || !strings.Contains(string(archived), "fog: 1.00") {
		t.Fatalf("block-list close failed: err=%v\n%s", err, archived)
	}
}

func TestProjectCloseFailsClosedOnUnknownModeledGuard(t *testing.T) {
	f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
	original := projectCloseTransitionFn
	originalLookup := projectIssueLookupFn
	projectIssueLookupFn = func(string, string) (issueMeta, error) {
		return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
	}
	projectCloseTransitionFn = func(from, event string) *vocab.Transition {
		tr := *vocab.Project().TransitionForEvent(from, event)
		tr.Guards = append(append([]string(nil), tr.Guards...), "future-close-guard")
		return &tr
	}
	t.Cleanup(func() {
		projectCloseTransitionFn = original
		projectIssueLookupFn = originalLookup
	})

	err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), `unknown project close guard "future-close-guard"`) {
		t.Fatalf("runProjectClose error = %v, want unknown modeled guard refusal", err)
	}
	if _, statErr := os.Stat(projectPath); statErr != nil {
		t.Fatalf("unknown guard mutated live project: %v", statErr)
	}
}

func TestProjectCloseRefusesIncompleteActualsUnlessLedgerBypassed(t *testing.T) {
	tests := []struct {
		name   string
		lookup func(string, string) (issueMeta, error)
	}{
		{"lookup error", func(ref, _ string) (issueMeta, error) {
			if ref == "ariadne#2" {
				return issueMeta{}, errors.New("peer unavailable")
			}
			return issueMeta{ActualHours: 10, ActualAvailable: true}, nil
		}},
		{"unset actual", func(ref, _ string) (issueMeta, error) {
			if ref == "ariadne#2" {
				return issueMeta{}, nil
			}
			return issueMeta{ActualHours: 10, ActualAvailable: true}, nil
		}},
		{"explicit N/A", func(ref, _ string) (issueMeta, error) {
			if ref == "ariadne#2" {
				return issueMeta{ActualNA: true}, nil
			}
			return issueMeta{ActualHours: 10, ActualAvailable: true}, nil
		}},
		{"all unavailable", func(string, string) (issueMeta, error) { return issueMeta{}, nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
			original := projectIssueLookupFn
			projectIssueLookupFn = tt.lookup
			t.Cleanup(func() { projectIssueLookupFn = original })
			err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
			if err == nil || !strings.Contains(err.Error(), "incomplete MVP actuals") || !strings.Contains(err.Error(), "--no-ledger") {
				t.Fatalf("runProjectClose error = %v, want incomplete-actual refusal", err)
			}
			if _, statErr := os.Stat(projectPath); statErr != nil {
				t.Fatalf("refusal mutated live project: %v", statErr)
			}
		})
	}
}

func TestProjectCloseNoLedgerAllowsIncompleteActualsButLogsNA(t *testing.T) {
	f, _, _ := projectCloseFixture(t, "executing", true, true)
	f.NoLedger = true
	original := projectIssueLookupFn
	projectIssueLookupFn = func(string, string) (issueMeta, error) { return issueMeta{}, nil }
	t.Cleanup(func() { projectIssueLookupFn = original })
	var stderr bytes.Buffer
	if err := runProjectClose(&bytes.Buffer{}, &stderr, f); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "actuals: incomplete") || !strings.Contains(string(b), "fog: n/a") {
		t.Fatalf("incomplete bypass wrote a calibrated-looking log:\n%s", b)
	}
}

func TestProjectCloseRejectsNonFiniteActuals(t *testing.T) {
	for _, actual := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
		original := projectIssueLookupFn
		projectIssueLookupFn = func(string, string) (issueMeta, error) {
			return issueMeta{ActualHours: actual, ActualAvailable: true}, nil
		}
		err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
		projectIssueLookupFn = original
		if err == nil || !strings.Contains(err.Error(), "incomplete MVP actuals") {
			t.Errorf("actual %v error = %v, want refusal", actual, err)
		}
		if _, statErr := os.Stat(projectPath); statErr != nil {
			t.Errorf("actual %v mutated project: %v", actual, statErr)
		}
	}
}

func TestProjectCloseRejectsDuplicateLogicalMVPScopeRefs(t *testing.T) {
	for _, scope := range []string{
		"mvp_scope: [ariadne#1, ariadne#1]",
		"mvp_scope: [ariadne#1, '#1']",
	} {
		f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
		b, _ := os.ReadFile(projectPath)
		text := strings.Replace(string(b), "mvp_scope: [ariadne#1, ariadne#2]", scope, 1)
		if err := os.WriteFile(projectPath, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		original := projectIssueLookupFn
		lookups := 0
		projectIssueLookupFn = func(string, string) (issueMeta, error) {
			lookups++
			return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
		}
		err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
		projectIssueLookupFn = original
		if err == nil || !strings.Contains(err.Error(), "duplicate logical MVP issue") {
			t.Errorf("scope %q error = %v, want duplicate refusal", scope, err)
		}
		if _, statErr := os.Stat(projectPath); statErr != nil {
			t.Errorf("scope %q mutated project: %v", scope, statErr)
		}
		if lookups != 0 {
			t.Errorf("scope %q performed %d issue lookups before duplicate refusal", scope, lookups)
		}
	}
}

func TestProjectCloseTreatsUnavailablePeerAsIncompleteActual(t *testing.T) {
	for _, noLedger := range []bool{false, true} {
		f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
		f.NoLedger = noLedger
		b, _ := os.ReadFile(projectPath)
		text := strings.Replace(string(b), "ariadne#2", "missing-peer#2", 1)
		if err := os.WriteFile(projectPath, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		original := projectIssueLookupFn
		projectIssueLookupFn = func(ref string, _ string) (issueMeta, error) {
			if ref == "missing-peer#2" {
				return issueMeta{}, errors.New("peer unavailable")
			}
			return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
		}
		var stderr bytes.Buffer
		err := runProjectClose(&bytes.Buffer{}, &stderr, f)
		projectIssueLookupFn = original
		if !noLedger {
			if err == nil || !strings.Contains(err.Error(), "incomplete MVP actuals") {
				t.Fatalf("ledger-backed close error = %v, want incomplete actuals", err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("--no-ledger close: %v", err)
		}
		archived, readErr := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
		if readErr != nil || !strings.Contains(string(archived), "actuals: incomplete") || !strings.Contains(string(archived), "fog: n/a") {
			t.Fatalf("degraded close did not archive incomplete calibration: err=%v\n%s", readErr, archived)
		}
	}
}

func TestProjectCloseCapturesTodayOnce(t *testing.T) {
	f, _, _ := projectCloseFixture(t, "executing", true, true)
	f.NoLedger = true
	originalToday := projectTodayFn
	originalLookup := projectIssueLookupFn
	calls := 0
	projectTodayFn = func() string {
		calls++
		if calls == 1 {
			return "2026-07-16"
		}
		return "2026-07-17"
	}
	projectIssueLookupFn = func(string, string) (issueMeta, error) {
		return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
	}
	t.Cleanup(func() {
		projectTodayFn = originalToday
		projectIssueLookupFn = originalLookup
	})
	if err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("projectTodayFn called %d times, want one transaction date", calls)
	}
}

func TestProjectCloseLedgerStageFailureLeavesBothRecordsUnchanged(t *testing.T) {
	f, projectPath, ledgerPath := projectCloseFixture(t, "executing", true, true)
	projectBefore, _ := os.ReadFile(projectPath)
	ledgerBefore, _ := os.ReadFile(ledgerPath)
	originalLookup := projectIssueLookupFn
	projectIssueLookupFn = func(string, string) (issueMeta, error) {
		return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
	}
	originalStage := projectCloseStageFileFn
	projectCloseStageFileFn = func(dir, pattern string, data []byte) (string, error) {
		if strings.Contains(pattern, "ledger") {
			return "", errors.New("forced ledger stage failure")
		}
		return originalStage(dir, pattern, data)
	}
	t.Cleanup(func() {
		projectIssueLookupFn = originalLookup
		projectCloseStageFileFn = originalStage
	})

	err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), "forced ledger stage failure") {
		t.Fatalf("runProjectClose error = %v, want forced stage failure", err)
	}
	projectAfter, _ := os.ReadFile(projectPath)
	ledgerAfter, _ := os.ReadFile(ledgerPath)
	if string(projectAfter) != string(projectBefore) || string(ledgerAfter) != string(ledgerBefore) {
		t.Fatal("stage failure changed a durable record")
	}
	if _, statErr := os.Stat(filepath.Join(f.HistoryDir, "projects", "alpha.md")); !os.IsNotExist(statErr) {
		t.Fatalf("stage failure left archived project: %v", statErr)
	}
}

func TestProjectCloseArchiveRenameFailureRestoresProjectAndLedger(t *testing.T) {
	f, projectPath, ledgerPath := projectCloseFixture(t, "executing", true, true)
	projectBefore, _ := os.ReadFile(projectPath)
	ledgerBefore, _ := os.ReadFile(ledgerPath)
	originalLookup := projectIssueLookupFn
	projectIssueLookupFn = func(string, string) (issueMeta, error) {
		return issueMeta{ActualHours: 2, ActualAvailable: true}, nil
	}
	originalRename := projectCloseRenameFn
	projectCloseRenameFn = func(oldPath, newPath string) error {
		if strings.Contains(filepath.Base(oldPath), ".sdlc-project-close-project-") {
			return errors.New("forced archive rename failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() {
		projectIssueLookupFn = originalLookup
		projectCloseRenameFn = originalRename
	})
	err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), "forced archive rename failure") {
		t.Fatalf("runProjectClose error = %v", err)
	}
	projectAfter, _ := os.ReadFile(projectPath)
	ledgerAfter, _ := os.ReadFile(ledgerPath)
	if string(projectAfter) != string(projectBefore) || string(ledgerAfter) != string(ledgerBefore) {
		t.Fatal("post-ledger archive failure did not restore both originals")
	}
}

func TestProjectCloseMissingPhaseAWarnsAndLogsNAWithoutLedger(t *testing.T) {
	f, _, ledgerPath := projectCloseFixture(t, "executing", true, false)
	before, _ := os.ReadFile(ledgerPath)
	var stderr bytes.Buffer
	if err := runProjectClose(&bytes.Buffer{}, &stderr, f); err == nil || !strings.Contains(err.Error(), "--no-ledger") {
		t.Fatalf("missing phase-a error = %v, want explicit bypass requirement", err)
	}
	f.NoLedger = true
	stderr.Reset()
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

func TestProjectCloseRejectsMalformedOrNonPositivePhaseA(t *testing.T) {
	for _, value := range []string{"bogus", "0h"} {
		t.Run(value, func(t *testing.T) {
			f, projectPath, _ := projectCloseFixture(t, "executing", true, true)
			b, _ := os.ReadFile(projectPath)
			text := strings.Replace(string(b), "**phase-a:** 40h", "**phase-a:** "+value, 1)
			if err := os.WriteFile(projectPath, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			f.NoLedger = true
			err := runProjectClose(&bytes.Buffer{}, &bytes.Buffer{}, f)
			if err == nil || !strings.Contains(err.Error(), "invalid phase-a") {
				t.Fatalf("phase-a %q error = %v, want invalid refusal", value, err)
			}
		})
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

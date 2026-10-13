package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestProjectCloseRequiresClosingAndPointsExecutingAtWrap(t *testing.T) {
	tests := []struct {
		status, want string
	}{
		{"defined", "requires status closing"},
		{"executing", "sdlc project set-status --to closing"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			f, _ := projectCloseFixture(t, tt.status, true)
			if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("runProjectClose error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestProjectCloseRequiresRetroUnlessBypassed(t *testing.T) {
	f, _ := projectCloseFixture(t, "closing", false)
	if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), "--no-retro") {
		t.Fatalf("runProjectClose error = %v, want --no-retro pointer", err)
	}
	f.NoRetro = true
	var stderr bytes.Buffer
	if err := runProjectClose(context.Background(), &bytes.Buffer{}, &stderr, f); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "--no-retro (or --force)") {
		t.Fatalf("stderr missing bypass acknowledgement: %s", stderr.String())
	}
}

func TestProjectCloseArchives(t *testing.T) {
	f, projectPath := projectCloseFixture(t, "'closing'", true) // a quoted status reads the same
	var stdout bytes.Buffer
	if err := runProjectClose(context.Background(), &stdout, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(f.HistoryDir, "projects", "alpha.md")
	if strings.TrimSpace(stdout.String()) != dest {
		t.Fatalf("stdout = %q, want %q", stdout.String(), dest)
	}
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		t.Fatalf("live record still present: %v", err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"status: closed", "updated: 2026-07-16", "### 2026-07-16 — close\n\n- status: closing → closed"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("archived project missing %q:\n%s", want, b)
		}
	}
}

func TestProjectCloseFailsClosedOnUnknownModeledGuard(t *testing.T) {
	f, projectPath := projectCloseFixture(t, "closing", true)
	original := projectCloseTransitionFn
	projectCloseTransitionFn = func(from, event string) *vocab.Transition {
		tr := *vocab.Project().TransitionForEvent(from, event)
		tr.Guards = append(append([]string(nil), tr.Guards...), "future-close-guard")
		return &tr
	}
	t.Cleanup(func() { projectCloseTransitionFn = original })
	err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), `unknown project close guard "future-close-guard"`) {
		t.Fatalf("runProjectClose error = %v, want unknown modeled guard refusal", err)
	}
	if _, statErr := os.Stat(projectPath); statErr != nil {
		t.Fatalf("unknown guard mutated live project: %v", statErr)
	}
}

func TestProjectCloseCapturesTodayOnce(t *testing.T) {
	f, _ := projectCloseFixture(t, "closing", true)
	calls := 0
	projectTodayFn = func() string {
		calls++
		if calls == 1 {
			return "2026-07-16"
		}
		return "2026-07-17"
	}
	if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("projectTodayFn called %d times, want one transaction date", calls)
	}
}

func TestProjectCloseStageFailureLeavesProjectUnchanged(t *testing.T) {
	f, projectPath := projectCloseFixture(t, "closing", true)
	before, _ := os.ReadFile(projectPath)
	original := projectCloseStageFileFn
	projectCloseStageFileFn = func(string, string, []byte) (string, error) { return "", errors.New("forced stage failure") }
	t.Cleanup(func() { projectCloseStageFileFn = original })
	if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), "forced stage failure") {
		t.Fatalf("runProjectClose error = %v", err)
	}
	after, _ := os.ReadFile(projectPath)
	if string(after) != string(before) {
		t.Fatal("stage failure changed the live project")
	}
}

func TestProjectCloseArchiveRenameFailureRestoresProject(t *testing.T) {
	f, projectPath := projectCloseFixture(t, "closing", true)
	before, _ := os.ReadFile(projectPath)
	originalRename := projectCloseRenameFn
	projectCloseRenameFn = func(oldPath, newPath string) error {
		if strings.Contains(filepath.Base(oldPath), ".sdlc-project-close-project-") {
			return errors.New("forced archive rename failure")
		}
		return originalRename(oldPath, newPath)
	}
	t.Cleanup(func() { projectCloseRenameFn = originalRename })
	err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), "forced archive rename failure") {
		t.Fatalf("runProjectClose error = %v", err)
	}
	after, _ := os.ReadFile(projectPath)
	if string(after) != string(before) {
		t.Fatal("archive failure did not restore the live project")
	}
}

func TestProjectCloseDropFromLiveStatuses(t *testing.T) {
	for _, status := range []string{"executing", "closing"} {
		t.Run(status, func(t *testing.T) {
			f, _ := projectCloseFixture(t, status, true)
			f.Drop = true
			if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(f.HistoryDir, "projects", "alpha.md"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(b), "status: dropped") || !strings.Contains(string(b), "— drop\n\n- status: "+status+" → dropped") {
				t.Fatalf("drop record wrong:\n%s", b)
			}
		})
	}
}

func TestProjectCloseDropRejectsDefined(t *testing.T) {
	f, _ := projectCloseFixture(t, "defined", true)
	f.Drop = true
	if err := runProjectClose(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, f); err == nil || !strings.Contains(err.Error(), "set-status --to dropped") {
		t.Fatalf("runProjectClose error = %v, want the set-status pointer", err)
	}
}

func projectCloseFixture(t *testing.T, status string, retro bool) (*projectCloseFlags, string) {
	t.Helper()
	root := t.TempDir()
	projects := filepath.Join(root, "workshop", "projects")
	history := filepath.Join(root, "workshop", "history")
	if err := os.MkdirAll(projects, 0o755); err != nil {
		t.Fatal(err)
	}
	log := ""
	if retro {
		log = "\n### 2026-07-15 — retro\n\nLearned.\n"
	}
	project := "---\ntype: project\nname: alpha\ngoal: ship\ndone_when: shipped\nstatus: " + status + "\ncreated: 2026-07-01\nupdated: 2026-07-15\nmvp_scope: [ariadne#1, ariadne#2]\n---\n\n# alpha\n\n## Charter\nWhy.\n\n## Trains\n\n## Breakdown\n\n- [x] one [ariadne#1]\n- [x] two [ariadne#2]\n\n## Log\n" + log + "\n## Decisions\n"
	projectPath := filepath.Join(projects, "alpha.md")
	if err := os.WriteFile(projectPath, []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}
	original := projectTodayFn
	projectTodayFn = func() string { return "2026-07-16" }
	t.Cleanup(func() { projectTodayFn = original })
	return &projectCloseFlags{Slug: "alpha", ProjectsDir: projects, HistoryDir: history}, projectPath
}

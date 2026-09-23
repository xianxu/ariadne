package main

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func calibrationWorkspace(t *testing.T) (string, string) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	fleet := t.TempDir()
	primary := filepath.Join(fleet, "sample")
	if err := os.Mkdir(primary, 0755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", primary}, args...)...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git: %s %v", out, err)
		}
	}
	git("init", "-b", "main")
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	checkout := filepath.Join(fleet, "worktree", "sample", "topic")
	git("worktree", "add", "-b", "topic", checkout)
	nested := filepath.Join(checkout, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	fleet, _ = filepath.EvalSymlinks(fleet)
	return fleet, nested
}

func TestBrainDefaultsLinkedNested(t *testing.T) {
	fleet, nested := calibrationWorkspace(t)
	t.Chdir(nested)
	for _, explicit := range []bool{false, true} {
		var brain string
		var got string
		root := &cobra.Command{Use: "test"}
		cmd := &cobra.Command{Use: "calibrate", Run: func(*cobra.Command, []string) { got = brain }}
		cmd.Flags().StringVar(&brain, "brain-dir", "../brain", "")
		root.AddCommand(cmd)
		wrapBrainDefaults(root)
		args := []string{"calibrate"}
		if explicit {
			args = append(args, "--brain-dir", "../brain")
		}
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(fleet, "brain")
		if explicit {
			want = "../brain"
		}
		if got != want {
			t.Errorf("explicit=%v got %q want %q", explicit, got, want)
		}
	}
	c := gatherBaseContention(nested, 0)
	if c.Repo != "sample" {
		t.Errorf("contention repo %q want sample", c.Repo)
	}
	var out bytes.Buffer
	runStartPlan(&out, 0)
	if !strings.Contains(out.String(), filepath.Join(fleet, "brain")) {
		t.Errorf("planning source not fleet brain: %s", out.String())
	}
}

func TestBrainDefaultsIdentityFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	root := &cobra.Command{Use: "test"}
	called := false
	cmd := &cobra.Command{Use: "calibrate", Run: func(*cobra.Command, []string) { called = true }}
	cmd.Flags().String("brain-dir", "../brain", "")
	root.AddCommand(cmd)
	wrapBrainDefaults(root)
	root.SetArgs([]string{"calibrate"})
	if err := root.Execute(); err == nil || called {
		t.Fatalf("missing identity: called=%v err=%v", called, err)
	}
	var out bytes.Buffer
	runStartPlan(&out, 0)
	if !strings.Contains(out.String(), "cannot resolve estimator brain") {
		t.Errorf("optional planning should warn: %s", out.String())
	}
}

func TestBrainDefaultsExplicitEstimatorOverride(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("WF_ESTIMATOR_SRC", "calibration.md")
	if err := os.WriteFile("calibration.md", []byte("calibration"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := NewEstimateSourceCmd()
	wrapBrainDefaults(cmd)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "calibration.md") {
		t.Fatal(out.String())
	}
}

func TestActualPeersCanonicalFromLinkedCheckout(t *testing.T) {
	_, nested := calibrationWorkspace(t)
	t.Chdir(nested)
	cmd := exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "#1: implement sample#2 with foreign#3")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git: %s %v", out, err)
	}
	result := computeActual(filepath.Dir(nested), "", "1")
	if len(result.Peers) != 2 || result.Peers[0] != "1" || result.Peers[1] != "2" {
		t.Fatalf("canonical peers = %v, status=%v detail=%s", result.Peers, result.Status, result.Detail)
	}
}

func TestPlanningEstimateUsesCurrentCheckout(t *testing.T) {
	_, nested := calibrationWorkspace(t)
	checkout := filepath.Dir(nested)
	tracker := filepath.Join(checkout, "workshop", "issues")
	if err := os.MkdirAll(tracker, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tracker, "000001-example.md"), []byte("---\ntype: issue\nstatus: working\nestimate_hours: 3\n---\n# Example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	t.Setenv("WF_ISSUES_DIR", "")
	if got, err := issueEstimate(1); err != nil || got != "3" {
		t.Fatalf("issueEstimate nested checkout = %q, %v; want 3", got, err)
	}
}

func TestPlanningContentionReportsUnavailableIdentity(t *testing.T) {
	root := t.TempDir()
	got := gatherBaseContention(root, 0)
	if got.Clean() {
		t.Fatal("unavailable Git identity reported clean")
	}
	line := baseContentionSummary(got)
	if !strings.Contains(line, "unavailable") || !strings.Contains(line, root) {
		t.Fatalf("missing failure evidence: %s", line)
	}
}

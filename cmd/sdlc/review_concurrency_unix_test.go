//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// A real process waits on a FIFO; the test decides exactly when its review may
// return. No wall-clock sleep is used to produce the concurrency schedule.
func blockedBoundaryReviewer(t *testing.T, verdict string) (<-chan struct{}, func()) {
	t.Helper()
	dir := t.TempDir()
	fifo := filepath.Join(dir, "release")
	response := filepath.Join(dir, "response")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	gate, err := os.OpenFile(fifo, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	body := "```verdict\nverdict: " + verdict + "\nconfidence: high\n```\n\n```findings\n```\n"
	if err := os.WriteFile(response, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	original := judge.Run
	judge.Run = func(ctx context.Context, onStart func(int), _ string, _ ...string) (judge.ProcessOutput, error) {
		return original(ctx, func(pid int) {
			if onStart != nil {
				onStart(pid)
			}
			close(started)
		}, "sh", "-c", `IFS= read -r release < "$1"; cat "$2"`, "review-fixture", fifo, response)
	}
	t.Cleanup(func() { _, _ = gate.WriteString("release\n"); _ = gate.Close(); judge.Run = original })
	return started, func() {
		if _, err := gate.WriteString("release\n"); err != nil {
			t.Fatal(err)
		}
	}
}

func reviewPlanFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			files[entry.Name()] = "<directory>"
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name()] = string(raw)
	}
	return files
}

func TestReviewConcurrencySchedules(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		for _, change := range []string{"branch", "boundary-ledger", "plan-seed", "directory", "cancel", "relock", "unrelated"} {
			t.Run(verb+"/"+change, func(t *testing.T) {
				issues := closeRepo(t, 69)
				plans := filepath.Join("workshop", "plans")
				if err := os.MkdirAll(plans, 0755); err != nil {
					t.Fatal(err)
				}
				if change == "unrelated" {
					writeIssue(t, issues, "000070-other.md", "---\nid: 000070\nstatus: open\n---\n# Other\n")
					git(t, "", "add", "--", filepath.Join(issues, "000070-other.md"))
					git(t, "", "commit", "-qm", "seed unrelated issue")
				}
				started, release := blockedBoundaryReviewer(t, "SHIP")
				if change == "relock" {
					original := repoLockAcquireForCommand
					calls := 0
					repoLockAcquireForCommand = func(cmd *cobra.Command) (func() error, error) {
						calls++
						if calls == 2 {
							return nil, errors.New("injected reacquire failure")
						}
						return original(cmd)
					}
					t.Cleanup(func() { repoLockAcquireForCommand = original })
				}
				args := []string{verb, "--issue", "69", "--actual", "1", "--verified", "test evidence", "--no-atlas", "--issues-dir", issues, "--plans-dir", plans, "--brain-dir", "../nonexistent-brain"}
				if verb == "milestone-close" {
					args = append(args, "--milestone", "M1")
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan error, 1)
				go func() {
					cmd := buildRoot()
					cmd.SetArgs(args)
					cmd.SetContext(ctx)
					cmd.SetOut(&bytes.Buffer{})
					var diagnostics bytes.Buffer
					cmd.SetErr(&diagnostics)
					err := cmd.Execute()
					if err != nil {
						err = fmt.Errorf("%w; diagnostics: %s", err, diagnostics.String())
					}
					done <- err
				}()
				select {
				case <-started:
				case err := <-done:
					t.Fatalf("review returned before process start: %v", err)
				case <-time.After(20 * time.Second):
					cancel()
					release()
					select {
					case <-done:
					case <-time.After(7 * time.Second):
						t.Fatal("review did not stop after start timeout")
					}
					t.Fatal("reviewer process did not start")
				}
				switch change {
				case "branch":
					git(t, "", "switch", "-c", "parallel-branch-same-head")
				case "boundary-ledger":
					if err := os.WriteFile(boundaryGatePath(plans, "000069-x.md"), []byte("new concurrent boundary generation"), 0644); err != nil {
						t.Fatal(err)
					}
				case "plan-seed":
					if err := os.WriteFile(planGatePath(plans, "000069-x.md"), []byte("new concurrent seed findings"), 0644); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Mkdir(filepath.Join(plans, "000069-x-plan.md"), 0755); err != nil {
						t.Fatal(err)
					}
				case "cancel":
					cancel()
				case "unrelated":
					p := filepath.Join(issues, "000070-other.md")
					f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0600)
					if err != nil {
						t.Fatal(err)
					}
					_, err = f.WriteString("\nindependent work\n")
					_ = f.Close()
					if err != nil {
						t.Fatal(err)
					}
					independent := make(chan error, 1)
					go func() { _, _, err := executeSDLCTestCommand("issue", "sync", "--issue", "70"); independent <- err }()
					select {
					case err := <-independent:
						if err != nil {
							t.Fatal(err)
						}
					case <-time.After(3 * time.Second):
						t.Fatal("unrelated operation blocked behind external review")
					}
					select {
					case err := <-done:
						t.Fatalf("review returned before release: %v", err)
					default:
					}
				}
				beforePlans := reviewPlanFiles(t, plans)
				beforeIssue := readIssue(t, issues)
				if change != "cancel" {
					release()
				}
				var err error
				select {
				case err = <-done:
				case <-time.After(7 * time.Second):
					t.Fatal("review command did not terminate")
				}
				if change == "unrelated" {
					if err != nil {
						t.Fatalf("unrelated documentation change should remain acceptable: %v", err)
					}
					return
				}
				if err == nil {
					t.Fatal("invalid review finalized")
				}
				switch change {
				case "cancel":
					if !errors.Is(err, context.Canceled) {
						t.Errorf("cancellation lost: %v", err)
					}
				case "relock":
					if !strings.Contains(err.Error(), "reacquire") {
						t.Errorf("relock error lost: %v", err)
					}
				default:
					if !strings.Contains(err.Error(), "stale") {
						t.Errorf("missing stale refusal: %v", err)
					}
				}
				if after := reviewPlanFiles(t, plans); !reflect.DeepEqual(beforePlans, after) {
					t.Errorf("invalid review wrote sidecar/ledger:\nbefore=%v\nafter=%v", beforePlans, after)
				}
				if after := readIssue(t, issues); after != beforeIssue {
					t.Errorf("invalid review changed issue: %s", after)
				}
			})
		}
	}
}

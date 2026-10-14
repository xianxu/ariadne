package judge

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
)

func preparationRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "--allow-empty", "-qm", "base"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	return dir
}

func TestPreparationRetryLifecycle(t *testing.T) {
	dir := preparationRepo(t)
	evidence := &DispatchEvidence{}
	runs := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: REWORK")}, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
	saved := Run
	Run = func(ctx context.Context, onStart func(int), name string, args ...string) (ProcessOutput, error) {
		if _, err := os.Stat(filepath.Join(dir, "generated")); err != nil {
			t.Fatal("review preceded preparation", err)
		}
		return saved(ctx, onStart, name, args...)
	}
	resets := 0
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir,
		Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", "test ! -e generated && echo asset > generated"}, TimeoutMinutes: 1}, Evidence: evidence,
		BeforeRetry: func() error { resets++; return os.Remove(filepath.Join(dir, "generated")) },
		ValidateResponse: func(s string) error {
			if strings.Contains(s, "REWORK") {
				return errors.New("REWORK needs a finding")
			}
			return nil
		},
	})
	if err != nil || *runs != 2 || resets != 1 || out != "VERDICT: SHIP" || len(evidence.Preparation) != 2 || len(evidence.Protocol) != 1 {
		t.Fatalf("out=%q err=%v runs=%d resets=%d evidence=%+v", out, err, *runs, resets, evidence)
	}
}

func TestPreparationFailureStopsReview(t *testing.T) {
	for _, script := range []string{"echo failure; exit 7", "git -c user.name=T -c user.email=t@example.com commit --allow-empty -qm changed"} {
		t.Run(script, func(t *testing.T) {
			dir := preparationRepo(t)
			calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
			evidence := &DispatchEvidence{}
			out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", script}, TimeoutMinutes: 1}, Evidence: evidence})
			if err == nil || out != "" || *calls != 0 || len(evidence.Preparation) != 1 || evidence.Preparation[0].Error == "" {
				t.Fatalf("out=%q err=%v calls=%d evidence=%+v", out, err, *calls, evidence)
			}
		})
	}
}

func TestPreparationOutputAndCancellation(t *testing.T) {
	dir := preparationRepo(t)
	calls := fakeRuns(t, ProcessOutput{})
	evidence := &DispatchEvidence{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", "yes diagnostic & wait"}, TimeoutMinutes: 1}, Evidence: evidence})
	if err == nil || *calls != 0 || len(evidence.Preparation) != 1 {
		t.Fatalf("err=%v calls=%d evidence=%+v", err, *calls, evidence)
	}
	if r := evidence.Preparation[0]; len(r.Output) > reviewprepare.OutputLimit || !r.Truncated {
		t.Fatalf("output bound: len=%d truncated=%v", len(r.Output), r.Truncated)
	}
}

func TestInvalidResponsesNeverBecomeSemanticOutput(t *testing.T) {
	dir := preparationRepo(t)
	calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: REWORK")})
	evidence := &DispatchEvidence{}
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Evidence: evidence, ValidateResponse: func(string) error { return errors.New("no finding") }})
	if err == nil || out != "" || *calls != 2 || len(evidence.Protocol) != 2 {
		t.Fatalf("out=%q err=%v calls=%d evidence=%+v", out, err, *calls, evidence)
	}
}

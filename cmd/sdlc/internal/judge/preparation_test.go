package judge

import (
	"context"
	"errors"
	"fmt"
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
	ready := filepath.Join(dir, "output-ready")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		_, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{
			Command:        []string{"sh", "-c", fmt.Sprintf("head -c %d /dev/zero && touch \"$1\" && sleep 60", 64*reviewprepare.OutputLimit), "sh", ready},
			TimeoutMinutes: 1,
		}, Evidence: evidence})
		finished <- err
	}()
	// The producer signals only after writing far beyond the output cap. The
	// timeout protects a broken setup; it does not trigger the cancellation.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			cancel()
			<-finished
			t.Fatal(err)
		}
		select {
		case err := <-finished:
			t.Fatalf("preparation ended before output readiness: %v", err)
		case <-ticker.C:
		case <-deadline.C:
			cancel()
			err := <-finished
			t.Fatalf("preparation never signaled output readiness: %v", err)
		}
	}
	cancel()
	err := <-finished
	if !errors.Is(err, context.Canceled) || *calls != 0 || len(evidence.Preparation) != 1 {
		t.Fatalf("err=%v calls=%d evidence=%+v", err, *calls, evidence)
	}
	if r := evidence.Preparation[0]; len(r.Output) > reviewprepare.OutputLimit || !r.Truncated {
		t.Fatalf("output bound: len=%d truncated=%v", len(r.Output), r.Truncated)
	}
}

func TestPreparationCanceledBeforeOutput(t *testing.T) {
	dir := preparationRepo(t)
	calls := fakeRuns(t, ProcessOutput{})
	evidence := &DispatchEvidence{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{
		Command: []string{"sh", "-c", "echo unexpected"}, TimeoutMinutes: 1,
	}, Evidence: evidence})
	if !errors.Is(err, context.Canceled) || *calls != 0 || len(evidence.Preparation) != 1 {
		t.Fatalf("err=%v calls=%d evidence=%+v", err, *calls, evidence)
	}
	if r := evidence.Preparation[0]; r.Output != "" || r.Truncated {
		t.Fatalf("early cancellation gained output: %+v", r)
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

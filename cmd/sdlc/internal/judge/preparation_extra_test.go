//go:build unix

package judge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestPreparationTrackedContentGuard(t *testing.T) {
	for _, tc := range []struct {
		name, script string
		blocked      bool
	}{{"tracked", "printf changed > tracked", true}, {"generated", "printf asset > generated", false}} {
		t.Run(tc.name, func(t *testing.T) {
			dir := preparationRepo(t)
			if err := os.WriteFile(filepath.Join(dir, "tracked"), []byte("original"), 0644); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"add", "tracked"}, {"-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "tracked"}} {
				if b, e := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); e != nil {
					t.Fatalf("%v %s", e, b)
				}
			}
			calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
			evidence := &DispatchEvidence{}
			out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", tc.script}, TimeoutMinutes: 1}, Evidence: evidence})
			if tc.blocked {
				if err == nil || !strings.Contains(err.Error(), "tracked checkout state") || out != "" || *calls != 0 {
					t.Fatalf("mutation admitted: out=%q calls=%d err=%v", out, *calls, err)
				}
			} else if err != nil || *calls != 1 {
				t.Fatalf("generated refused: calls=%d err=%v", *calls, err)
			}
		})
	}
}

func TestPreparationTrackedGuardIgnoresTextConversion(t *testing.T) {
	dir := preparationRepo(t)
	for path, body := range map[string]string{"tracked": "original\n", ".gitattributes": "tracked diff=constant\n"} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "."}, {"-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "tracked"}, {"config", "diff.constant.textconv", "sh -c 'echo constant' --"}} {
		if b, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, b)
		}
	}
	calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
	_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", "printf changed > tracked"}, TimeoutMinutes: 1}})
	if err == nil || !strings.Contains(err.Error(), "tracked checkout state") || *calls != 0 {
		t.Fatalf("text conversion hid tracked mutation: err=%v reviewer calls=%d", err, *calls)
	}
}

func TestPreparationUsesReviewerWrapperAndEnvironment(t *testing.T) {
	dir := preparationRepo(t)
	calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
	originalRun := Run
	var wrapped []string
	Run = func(ctx context.Context, onStart func(int), name string, args ...string) (ProcessOutput, error) {
		conf := ctx.Value(confinementKey{}).(*reviewbox.Confinement)
		got, err := os.ReadFile(filepath.Join(dir, "environment"))
		if err != nil {
			t.Fatal(err)
		}
		want := conf.Tmp() + "\n" + conf.BinDir() + "\n"
		if string(got) != want {
			t.Fatalf("preparation environment=%q want=%q", got, want)
		}
		return originalRun(ctx, onStart, name, args...)
	}
	evidence := &DispatchEvidence{}
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Evidence: evidence,
		Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", `printf '%s\n%s\n' "$TMPDIR" "${PATH%%:*}" > environment; printf 'VERDICT: REWORK\n'`}, TimeoutMinutes: 1},
		Wrap: func(name string, args []string) (string, []string) {
			wrapped = append(wrapped, name)
			return "env", append([]string{name}, args...)
		},
	})
	if err != nil || *calls != 1 || out != "VERDICT: SHIP" || len(wrapped) != 2 || wrapped[0] != "sh" || wrapped[1] != "codex" {
		t.Fatalf("out=%q err=%v calls=%d wrapped=%v", out, err, *calls, wrapped)
	}
	if len(evidence.Preparation) != 1 || !strings.Contains(evidence.Preparation[0].Output, "VERDICT: REWORK") {
		t.Fatalf("prep diagnostic missing: %+v", evidence)
	}
}

func TestPreparationDeadlineKillsDescendant(t *testing.T) {
	dir := preparationRepo(t)
	calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
	old := reviewShutdownGrace
	reviewShutdownGrace = 50 * time.Millisecond
	t.Cleanup(func() { reviewShutdownGrace = old })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", "sleep 60 & echo $! > child.pid; wait"}, TimeoutMinutes: 1}})
	if err == nil || out != "" || *calls != 0 {
		t.Fatalf("out=%q err=%v calls=%d", out, err, *calls)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "child.pid"))
	if err != nil {
		t.Fatal("child never started", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatal("preparation descendant survived deadline", err)
	}
}

//go:build unix

package judge

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
	//lint:allow test-shrunk-deadline — the grace starts at the cued deadline, after the child reported ready
	reviewShutdownGrace = 50 * time.Millisecond
	t.Cleanup(func() { reviewShutdownGrace = old })
	// The preparation's deadline passes on cue once its child reports in (#434).
	ctx, expire := testfix.DeadlineOnCue(context.Background())
	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", "sleep 60 & echo $! > child.pid; wait"}, TimeoutMinutes: 1}})
		done <- result{out, err}
	}()
	pid := testfix.AwaitPIDFile(filepath.Join(dir, "child.pid"), testfix.HangCeiling, nil)
	if pid == 0 {
		t.Fatal("child never started")
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	expire()
	var r result
	select {
	case r = <-done:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("preparation did not end at its deadline")
	}
	if r.err == nil || r.out != "" || *calls != 0 {
		t.Fatalf("out=%q err=%v calls=%d", r.out, r.err, *calls)
	}
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatal("preparation descendant survived deadline", err)
	}
}

// Git's normalized view cannot be the authority for bytes the reviewer executes.
func TestPreparationTrackedBytesIgnoreGitSuppression(t *testing.T) {
	for _, tc := range []struct{ name, configure, script string }{
		{"assume-unchanged", "", "git update-index --assume-unchanged tracked && printf changed > tracked"},
		{"skip-worktree", "", "git update-index --skip-worktree tracked && printf changed > tracked"},
		{"clean-filter", "git config filter.constant.clean 'printf canonical'; printf 'tracked filter=constant\n' > .gitattributes", "printf changed > tracked"},
		{"permission-bits", "", "chmod 600 tracked"},
		{"missing", "", "rm tracked"},
		{"fifo", "", "rm tracked && mkfifo tracked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := preparationRepo(t)
			if err := os.WriteFile(filepath.Join(dir, "tracked"), []byte("original"), 0644); err != nil {
				t.Fatal(err)
			}
			if tc.configure != "" {
				cmd := exec.Command("sh", "-c", tc.configure)
				cmd.Dir = dir
				if b, e := cmd.CombinedOutput(); e != nil {
					t.Fatalf("configure: %v %s", e, b)
				}
			}
			preparationGit(t, dir, "add", ".")
			preparationGit(t, dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "tracked")
			calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
			out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", tc.script}, TimeoutMinutes: 1}})
			if err == nil || out != "" || *calls != 0 {
				t.Fatalf("raw tracked mutation admitted: output=%q err=%v calls=%d", out, err, *calls)
			}
		})
	}
}

func TestPreparationTrackedPathsAndDirtyBaseline(t *testing.T) {
	for _, tc := range []struct {
		name, script string
		blocked      bool
	}{
		{"ancestor-symlink", "cp -R directory copy && rm -r directory && ln -s copy directory", true},
		{"leaf-symlink", "rm link && ln -s alternate link", true},
		{"dirty-overlay-and-generated", "mkdir -p generated; printf build > generated/output; printf ignored > ignored", false},
		{"newline-path", "printf changed > 'line\nbreak'", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := preparationRepo(t)
			if err := os.Mkdir(filepath.Join(dir, "directory"), 0755); err != nil {
				t.Fatal(err)
			}
			for name, content := range map[string]string{"directory/file": "original", "line\nbreak": "original", ".gitignore": "ignored\n"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink("directory/file", filepath.Join(dir, "link")); err != nil {
				t.Fatal(err)
			}
			preparationGit(t, dir, "add", ".")
			preparationGit(t, dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "tracked")
			if err := os.WriteFile(filepath.Join(dir, "directory/file"), []byte("existing dirty overlay"), 0644); err != nil {
				t.Fatal(err)
			}
			calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
			out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", tc.script}, TimeoutMinutes: 1}})
			if tc.blocked {
				if err == nil || out != "" || *calls != 0 {
					t.Fatalf("tracked path mutation admitted: %q %v calls=%d", out, err, *calls)
				}
			} else if err != nil || *calls != 1 {
				t.Fatalf("unchanged dirty baseline refused: %v calls=%d", err, *calls)
			}
		})
	}
}

func preparationGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestPreparationGitlinkContents(t *testing.T) {
	for _, tc := range []struct {
		name        string
		initialized bool
		script      string
		blocked     bool
	}{
		{"initialized-unchanged", true, "printf asset > module/generated", false},
		{"initialized-tracked-change", true, "git -C module update-index --assume-unchanged tracked && printf changed > module/tracked", true},
		{"uninitialized-unchanged", false, ":", false},
		{"uninitialized-populated", false, "printf unowned > module/file", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := preparationRepo(t)
			module := filepath.Join(dir, "module")
			if err := os.Mkdir(module, 0755); err != nil {
				t.Fatal(err)
			}
			oid := preparationGit(t, dir, "rev-parse", "HEAD")
			if tc.initialized {
				preparationGit(t, module, "init", "-q")
				if err := os.WriteFile(filepath.Join(module, "tracked"), []byte("original"), 0644); err != nil {
					t.Fatal(err)
				}
				preparationGit(t, module, "add", "tracked")
				preparationGit(t, module, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "module")
				oid = preparationGit(t, module, "rev-parse", "HEAD")
			}
			preparationGit(t, dir, "update-index", "--add", "--cacheinfo", "160000,"+oid+",module")
			preparationGit(t, dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "gitlink")
			calls := fakeRuns(t, ProcessOutput{Stdout: []byte("VERDICT: SHIP")})
			out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Checkout: dir, Preparation: &reviewprepare.Config{Command: []string{"sh", "-c", tc.script}, TimeoutMinutes: 1}})
			if tc.blocked {
				if err == nil || out != "" || *calls != 0 {
					t.Fatalf("gitlink mutation admitted: %q %v calls=%d", out, err, *calls)
				}
			} else if err != nil || *calls != 1 {
				t.Fatalf("unchanged gitlink refused: %v calls=%d", err, *calls)
			}
		})
	}
}

func TestPreparationFingerprintRejectsAncestorSymlinkBaseline(t *testing.T) {
	dir := preparationRepo(t)
	if err := os.Mkdir(filepath.Join(dir, "directory"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "directory/file"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	preparationGit(t, dir, "add", ".")
	preparationGit(t, dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "commit", "-qm", "tracked")
	if err := os.Rename(filepath.Join(dir, "directory"), filepath.Join(dir, "copy")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("copy", filepath.Join(dir, "directory")); err != nil {
		t.Fatal(err)
	}
	if _, err := trackedPreparationState(context.Background(), dir); err == nil {
		t.Fatal("ancestor symlink baseline would hide mutable tracked descendants")
	}
}

// Cancel from the stream itself: the check must run during hashing, not merely
// before the Git subprocesses, and an unbounded stream must not be retained.
type cancelPreparationReader struct {
	cancel context.CancelFunc
	reads  int
}

func (r *cancelPreparationReader) Read(p []byte) (int, error) {
	r.reads++
	clear(p)
	r.cancel()
	return len(p), nil
}
func TestPreparationContentsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &cancelPreparationReader{cancel: cancel}
	var retained bytes.Buffer
	err := copyPreparationContents(ctx, &retained, source)
	if !errors.Is(err, context.Canceled) || source.reads != 1 || retained.Len() > 32<<10 {
		t.Fatalf("cancel not bounded: err=%v reads=%d bytes=%d", err, source.reads, retained.Len())
	}
}

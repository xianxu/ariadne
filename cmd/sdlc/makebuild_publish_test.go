package main

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeGoScript stands in for `go build -o OUT ...` the way the real one
// publishes: it writes OUT in place, partially first. FAKE_GO_MODE=pause then
// waits for $FAKE_GO_DIR/release before finishing; fail exits 1 with OUT half
// written.
const fakeGoScript = `#!/bin/sh
[ "$1" = build ] && [ "$2" = -o ] || exit 2
out="$3"
printf partial > "$out"
case "$FAKE_GO_MODE" in
pause)
  : > "$FAKE_GO_DIR/started"
  while [ ! -e "$FAKE_GO_DIR/release" ]; do sleep 0.05; done ;;
fail)
  exit 1 ;;
esac
printf new-complete > "$out"
`

// makeBuildRepo is a repository with one cmd/x, an old complete bin/x, and
// the real Makefile.workflow; it returns the repo and a PATH with the fake go.
func makeBuildRepo(t *testing.T) (repo, fakeDir string, env []string) {
	t.Helper()
	root := realRepoRoot()
	if root == "" {
		t.Fatal("could not resolve repository root")
	}
	repo = t.TempDir()
	fakeDir = t.TempDir()
	for p, body := range map[string]string{
		"go.mod":        "module x\n",
		"cmd/x/main.go": "package main\n",
		"bin/x":         "old-complete",
		"Makefile":      "include Makefile.workflow\n",
		"fakebin/go":    fakeGoScript,
	} {
		full := filepath.Join(repo, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(root, "Makefile.workflow"), filepath.Join(repo, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	env = append(os.Environ(), "PATH="+filepath.Join(repo, "fakebin")+":"+os.Getenv("PATH"), "FAKE_GO_DIR="+fakeDir)
	return repo, fakeDir, env
}

func readBin(t *testing.T, repo string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repo, "bin", "x"))
	if err != nil {
		t.Fatalf("bin/x: %v", err)
	}
	return string(b)
}

func noStaging(t *testing.T, repo string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repo, "bin"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".go-build.") {
			t.Errorf("staging dir left behind: bin/%s", e.Name())
		}
	}
}

// #419: while the compiler runs, bin/x stays the old complete binary; the new
// one appears only by the rename, and no staging dir remains.
func TestMakeBuildPublishesByRename(t *testing.T) {
	repo, fakeDir, env := makeBuildRepo(t)
	cmd := exec.Command("make", "build")
	cmd.Dir = repo
	cmd.Env = append(env, "FAKE_GO_MODE=pause")
	out := &strings.Builder{}
	cmd.Stdout, cmd.Stderr = out, out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	testfix.AwaitCondition(t, "the fake compiler started", func() bool {
		_, err := os.Stat(filepath.Join(fakeDir, "started"))
		return err == nil
	}, []chan error{done})
	if got := readBin(t, repo); got != "old-complete" {
		t.Errorf("bin/x mid-build = %q, want the old complete binary", got)
	}
	if err := os.WriteFile(filepath.Join(fakeDir, "release"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("make build: %v\n%s", err, out)
	}
	if got := readBin(t, repo); got != "new-complete" {
		t.Errorf("bin/x after the build = %q, want the new binary", got)
	}
	noStaging(t, repo)
}

// #419: a failed build leaves the old binary and no staging dir.
func TestMakeBuildFailureKeepsOldBinary(t *testing.T) {
	repo, _, env := makeBuildRepo(t)
	cmd := exec.Command("make", "build")
	cmd.Dir = repo
	cmd.Env = append(env, "FAKE_GO_MODE=fail")
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("make build succeeded with a failing compiler:\n%s", out)
	}
	if got := readBin(t, repo); got != "old-complete" {
		t.Errorf("bin/x after a failed build = %q, want the old complete binary", got)
	}
	noStaging(t, repo)
}

// #419: a build interrupted mid-compile (SIGTERM to its process group, as a
// terminal's Ctrl-C reaches it) keeps the old binary and removes its staging dir.
func TestMakeBuildInterruptKeepsOldBinary(t *testing.T) {
	repo, fakeDir, env := makeBuildRepo(t)
	cmd := exec.Command("make", "build")
	cmd.Dir = repo
	cmd.Env = append(env, "FAKE_GO_MODE=pause")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	testfix.AwaitCondition(t, "the fake compiler started", func() bool {
		_, err := os.Stat(filepath.Join(fakeDir, "started"))
		return err == nil
	}, []chan error{done})
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err == nil {
		t.Fatal("make build succeeded after SIGTERM")
	}
	if got := readBin(t, repo); got != "old-complete" {
		t.Errorf("bin/x after an interrupted build = %q, want the old complete binary", got)
	}
	// make exits on the signal without waiting for its recipe shell, so the
	// staging dir goes when that shell's EXIT trap runs, just after.
	for end := time.Now().Add(testfix.HangCeiling); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
		if ents, err := os.ReadDir(filepath.Join(repo, "bin")); err == nil && len(ents) == 1 {
			break
		}
	}
	noStaging(t, repo)
}

// #419: a staging dir a SIGKILL left behind is swept by the next build once it
// is an hour old; a recent one (a build still running) is kept.
func TestMakeBuildSweepsStaleStaging(t *testing.T) {
	repo, _, env := makeBuildRepo(t)
	stale := filepath.Join(repo, "bin", ".go-build.stale1")
	recent := filepath.Join(repo, "bin", ".go-build.recent")
	for _, d := range []string{stale, recent} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("make", "build")
	cmd.Dir = repo
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make build: %v\n%s", err, out)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("an hour-old staging dir survived the build: %v", err)
	}
	if _, err := os.Stat(recent); err != nil {
		t.Errorf("a recent staging dir (a live build's) was swept: %v", err)
	}
}

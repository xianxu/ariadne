package weavefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeRunner records each Run call (dir, argv) and returns a canned error. The
// generate stage's unit tests use it so no real binary is spawned; this file
// exercises the SEAM itself (the recording fake + the production execRunner).
type fakeRunner struct {
	calls []runCall
	err   error
}

type runCall struct {
	dir  string
	argv []string
}

func (f *fakeRunner) Run(dir string, argv []string) error {
	f.calls = append(f.calls, runCall{dir: dir, argv: argv})
	return f.err
}

var _ Runner = (*fakeRunner)(nil)

// TestFakeRunner_RecordsCalls locks the fake's contract: it records (dir, argv)
// in order and surfaces the canned error verbatim.
func TestFakeRunner_RecordsCalls(t *testing.T) {
	f := &fakeRunner{}
	if err := f.Run("/a/dir", []string{"./.dynamic-skill"}); err != nil {
		t.Fatalf("nil-err fake should return nil, got %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("got %d calls, want 1", len(f.calls))
	}
	if f.calls[0].dir != "/a/dir" || len(f.calls[0].argv) != 1 || f.calls[0].argv[0] != "./.dynamic-skill" {
		t.Errorf("recorded call = %+v, want dir=/a/dir argv=[./.dynamic-skill]", f.calls[0])
	}
}

// TestExecRunner_NonZeroExitErrors is the production-seam integration test: a
// real /bin/sh `exit 3` in a tempdir → a non-nil error; `exit 0` → nil. This is
// the one place a real process is spawned (the generate-stage tests use the
// fake), matching weave's "faithful seam over mocks" stance.
func TestExecRunner_NonZeroExitErrors(t *testing.T) {
	dir := t.TempDir()
	r := ExecRunner{}

	if err := r.Run(dir, []string{"/bin/sh", "-c", "exit 3"}); err == nil {
		t.Error("exit 3 should surface a non-nil error")
	}
	if err := r.Run(dir, []string{"/bin/sh", "-c", "exit 0"}); err != nil {
		t.Errorf("exit 0 should be nil, got %v", err)
	}
}

// TestExecRunner_SetsCwd asserts the command runs with cwd = dir (the package
// dir is where a `.dynamic-skill` resolves its relative paths from). The shell
// writes its cwd to a marker file; we read it back and compare against the
// resolved tempdir.
func TestExecRunner_SetsCwd(t *testing.T) {
	dir := t.TempDir()
	r := ExecRunner{}
	if err := r.Run(dir, []string{"/bin/sh", "-c", "pwd -P > cwd.txt"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cwd.txt"))
	if err != nil {
		t.Fatalf("read cwd marker: %v", err)
	}
	got := strings.TrimSpace(string(data))
	// macOS resolves /tmp → /private/tmp; compare against the physical form.
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("cwd = %q, want %q (cmd.Dir must be the package dir)", got, want)
	}
}

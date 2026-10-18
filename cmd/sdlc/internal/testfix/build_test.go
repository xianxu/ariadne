package testfix

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeBuilds counts builds and writes a stand-in binary at the -o path.
func fakeBuilds(t *testing.T, fail bool) *atomic.Int32 {
	t.Helper()
	var n atomic.Int32
	prev := buildCommand
	buildCommand = func(dir string, args ...string) ([]byte, error) {
		n.Add(1)
		if fail {
			return []byte("compile error"), errors.New("exit status 1")
		}
		for i, a := range args {
			if a == "-o" {
				return nil, os.WriteFile(args[i+1], []byte("bin"), 0o755)
			}
		}
		return nil, errors.New("no -o")
	}
	t.Cleanup(func() { buildCommand = prev; RemoveBuiltBinaries() })
	RemoveBuiltBinaries()
	return &n
}

func TestBuiltBinaryBuildsOncePerProcess(t *testing.T) {
	n := fakeBuilds(t, false)
	dir := t.TempDir()
	var wg sync.WaitGroup
	paths := make([]string, 16)
	for i := range paths {
		wg.Add(1)
		go func() { defer wg.Done(); paths[i] = BuiltBinary(t, dir, ".") }()
	}
	wg.Wait()
	if got := n.Load(); got != 1 {
		t.Fatalf("16 parallel callers ran %d builds, want 1", got)
	}
	for _, p := range paths {
		if p != paths[0] {
			t.Fatalf("callers got different binaries: %v", paths)
		}
	}
	if filepath.Base(paths[0]) != filepath.Base(dir) {
		t.Errorf("binary %s is not named after its package", paths[0])
	}
	// Another package or other flags are another build.
	BuiltBinary(t, dir, "./cmd/x")
	BuiltBinary(t, dir, ".", "-trimpath")
	BuiltBinary(t, dir, ".")
	if got := n.Load(); got != 3 {
		t.Errorf("builds = %d, want 3 (one per package and flag set)", got)
	}
	// The process's directory goes with RemoveBuiltBinaries.
	RemoveBuiltBinaries()
	if _, err := os.Stat(paths[0]); !os.IsNotExist(err) {
		t.Errorf("binary survived RemoveBuiltBinaries: %v", err)
	}
}

func TestBuiltBinaryFailureIsSharedNotRetried(t *testing.T) {
	n := fakeBuilds(t, true)
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		ft := &fatalRecorder{TB: t}
		func() {
			defer func() { _ = recover() }()
			BuiltBinary(ft, dir, ".")
		}()
		if !strings.Contains(ft.msg, "compile error") {
			t.Fatalf("caller %d: failure not reported: %q", i, ft.msg)
		}
	}
	if got := n.Load(); got != 1 {
		t.Errorf("a failed build was retried: %d builds", got)
	}
}

// fatalRecorder captures a Fatal without ending the real test.
type fatalRecorder struct {
	testing.TB
	msg string
}

func (f *fatalRecorder) Helper() {}
func (f *fatalRecorder) Fatal(args ...any) {
	for _, a := range args {
		if err, ok := a.(error); ok {
			f.msg += err.Error()
		}
	}
	panic("fatal")
}

func TestSweepBuiltDirsRemovesOnlyStaleOnes(t *testing.T) {
	tmp := t.TempDir()
	stale, fresh, other := filepath.Join(tmp, builtBinPrefix+"old"), filepath.Join(tmp, builtBinPrefix+"new"), filepath.Join(tmp, "unrelated-old")
	for _, d := range []string{stale, fresh, other} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-48 * time.Hour)
	for _, d := range []string{stale, other} {
		if err := os.Chtimes(d, old, old); err != nil {
			t.Fatal(err)
		}
	}
	sweepBuiltDirs(tmp, time.Now())
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("a stale binary directory survived the sweep")
	}
	for _, d := range []string{fresh, other} {
		if _, err := os.Stat(d); err != nil {
			t.Errorf("%s was swept: %v", d, err)
		}
	}
}

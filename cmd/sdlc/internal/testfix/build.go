package testfix

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// builtBinPrefix names a test process's binary directory under TMPDIR.
const builtBinPrefix = "sdlc-testbin-"

// builtStale is when a directory a killed test process left behind is swept.
const builtStale = 24 * time.Hour

type builtEntry struct {
	once sync.Once
	path string
	err  error
}

var built struct {
	sync.Mutex
	dir     string
	entries map[string]*builtEntry
}

// buildTimeout bounds one build, so a hung compiler fails its callers
// instead of running to the test binary's own -timeout.
const buildTimeout = 5 * time.Minute

// buildCommand runs the build; tests replace it to count or fail builds.
var buildCommand = func(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	return runBounded(cmd, buildTimeout)
}

// runBounded runs cmd with its combined output captured, for at most limit:
// past it the whole process group is killed (a compiler or linker child, not
// only go itself), and WaitDelay stops a descendant still holding the output
// pipe from stretching the wait.
func runBounded(cmd *exec.Cmd, limit time.Duration) ([]byte, error) {
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	cmd.WaitDelay = time.Second
	ownGroup(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var expired atomic.Bool
	timer := time.AfterFunc(limit, func() {
		expired.Store(true)
		killGroup(cmd)
	})
	err := cmd.Wait()
	timer.Stop()
	killGroup(cmd) // reclaim descendants a finished parent left behind
	if expired.Load() && err != nil { // a timely finish that raced the timer stays a success
		err = fmt.Errorf("exceeded %s: %w", limit, err)
	}
	return out.Bytes(), err
}

// BuiltBinary returns the binary `go build <flags> <pkg>` produces in dir,
// building it at most once per test process (#438): every caller with the same
// dir, package and flags shares one build, so a suite builds a binary once per
// process instead of once per test. A failed build fails every caller with the
// same error. The binary is named after the package's directory. The build
// runs once, under the first caller's environment (HOME, GOFLAGS, PATH); a
// test that needs a differently built binary passes different flags.
func BuiltBinary(t testing.TB, dir, pkg string, flags ...string) string {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	key := abs + "\x00" + pkg + "\x00" + strings.Join(flags, "\x00")
	built.Lock()
	if built.entries == nil {
		built.entries = map[string]*builtEntry{}
	}
	e := built.entries[key]
	if e == nil {
		e = &builtEntry{}
		built.entries[key] = e
	}
	built.Unlock()
	e.once.Do(func() { e.path, e.err = buildOnce(abs, pkg, key, flags) })
	if e.err != nil {
		t.Fatal(e.err)
	}
	return e.path
}

func buildOnce(dir, pkg, key string, flags []string) (string, error) {
	root, err := builtDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(key))
	out := filepath.Join(root, hex.EncodeToString(sum[:6]))
	if err := os.Mkdir(out, 0o755); err != nil {
		return "", err
	}
	name := filepath.Base(filepath.Join(dir, pkg))
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out = filepath.Join(out, name)
	args := append(append([]string{"build"}, flags...), "-o", out, pkg)
	if output, err := buildCommand(dir, args...); err != nil {
		return "", fmt.Errorf("go %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, output)
	}
	return out, nil
}

// builtDir is this process's binary directory, made on first use. Making it
// also sweeps directories that killed processes left behind.
func builtDir() (string, error) {
	built.Lock()
	defer built.Unlock()
	if built.dir != "" {
		return built.dir, nil
	}
	sweepBuiltDirs(os.TempDir(), time.Now())
	dir, err := os.MkdirTemp("", builtBinPrefix+"*")
	if err != nil {
		return "", err
	}
	built.dir = dir
	return dir, nil
}

// sweepBuiltDirs removes binary directories older than builtStale: a live
// process's directory is younger than any test run can be.
func sweepBuiltDirs(tmp string, now time.Time) {
	matches, _ := filepath.Glob(filepath.Join(tmp, builtBinPrefix+"*"))
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && info.IsDir() && now.Sub(info.ModTime()) > builtStale {
			_ = os.RemoveAll(m)
		}
	}
}

// RemoveBuiltBinaries deletes this process's binaries; a package's TestMain
// calls it after m.Run.
func RemoveBuiltBinaries() {
	built.Lock()
	defer built.Unlock()
	if built.dir != "" {
		_ = os.RemoveAll(built.dir)
	}
	built.dir, built.entries = "", nil
}

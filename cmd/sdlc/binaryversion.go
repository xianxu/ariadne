package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// The running process's identity, never the current PATH's replacement binary.
type binaryBuild struct{ Path, Revision, Modified, Source string }

var runningSDLC = sync.OnceValue(func() binaryBuild { return readBinaryBuild(os.Executable, debug.ReadBuildInfo) })

func readBinaryBuild(executable func() (string, error), metadata func() (*debug.BuildInfo, bool)) binaryBuild {
	var b binaryBuild
	if path, err := executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			b.Path = resolved
		} else {
			b.Path = path
		}
		if filepath.Base(b.Path) == "sdlc" && filepath.Base(filepath.Dir(b.Path)) == "bin" {
			candidate := filepath.Dir(filepath.Dir(b.Path))
			if v, err := workspace.NormalizeVantage(execGitRunner{}, candidate); err == nil && v.WorktreeRoot == candidate {
				b.Source = candidate
			}
		}
	}
	if info, ok := metadata(); ok && info != nil {
		b.Revision, b.Modified = buildRevision(info)
	}
	return b
}

func (b binaryBuild) version() string {
	return fmt.Sprintf("path=%s revision=%s modified=%s", orNone(b.Path), orNone(b.Revision), orNone(b.Modified))
}

var expectedFleetSDLC = func() (string, error) {
	v, err := workspace.NormalizeVantage(execGitRunner{}, ".")
	if err != nil {
		return "", err
	}
	return filepath.Join(v.FleetRoot, "ariadne", "bin", "sdlc"), nil
}

func warnFleetBinary(out io.Writer) {
	if os.Getenv("COUCH_TREE") == "" {
		return
	}
	b := runningSDLC()
	expected, err := expectedFleetSDLC()
	if err != nil {
		fmt.Fprintf(out, "WARNING: sdlc fleet binary could not be verified (%v); %s\n", err, b.version())
		return
	}
	if resolved, err := filepath.EvalSymlinks(expected); err == nil {
		expected = resolved
	}
	if b.Path != expected {
		fmt.Fprintf(out, "WARNING: sdlc is not the fleet :0 binary: %s; expected=%s. Refresh aliases or relaunch.\n", b.version(), expected)
	}
}

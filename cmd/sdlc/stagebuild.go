package main

import (
	"context"
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
)

// readBuildInfo reads a binary's embedded build info; a variable so tests fake
// it without compiling a real binary.
var readBuildInfo = buildinfo.ReadFile

// stageBuild runs the target's `make build` (#344) and reports each bin/* Go
// binary's vcs.revision against the staged commit, so one look verifies what
// the operator will run. A failed build leaves the stage in place.
func stageBuild(ctx context.Context, root, commit string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "make", "build")
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, stderr, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("staged %s, but `make build` in %s failed: %w", shortOID(commit), root, err)
	}
	bins, _ := filepath.Glob(filepath.Join(root, "bin", "*"))
	sort.Strings(bins)
	built := 0
	for _, bin := range bins {
		if st, err := os.Stat(bin); err != nil || !st.Mode().IsRegular() {
			continue
		}
		info, err := readBuildInfo(bin)
		if err != nil {
			continue // not a Go binary
		}
		built++
		rev, modified := buildRevision(info)
		rel, _ := filepath.Rel(root, bin)
		fmt.Fprintf(stdout, "  %s vcs.revision=%s%s\n", rel, orNone(rev), map[bool]string{true: " (modified)"}[modified])
		if rev != commit || modified {
			cwarn(stderr, fmt.Sprintf("%s was not built from the staged %s exactly", rel, shortOID(commit)))
		}
	}
	if built == 0 {
		cinfo(stderr, fmt.Sprintf("`make build` left no Go binaries in %s to verify", filepath.Join(root, "bin")))
	}
	return nil
}

// buildRevision is a binary's VCS stamp: the revision and whether the tree
// was modified (pure).
func buildRevision(info *debug.BuildInfo) (rev string, modified bool) {
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	return rev, modified
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

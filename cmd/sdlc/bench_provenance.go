package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// The native backend cannot use Metis's Python read sensor. Describe its own
// first-party source closure through the existing read-set channel, and retain
// the actual executable digest/build pin as a regular provenance artifact.
func writeBenchStepProvenance(ctx context.Context, f benchStepFolder, source, expectedDigest, expectedRevision string) error {
	physical, err := filepath.EvalSymlinks(source)
	if err != nil || !filepath.IsAbs(source) || physical != source {
		return errors.New("tool_source must name the physical candidate source checkout")
	}
	head, err := runBenchPrepareCommand(ctx, source, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	status, err := runBenchPrepareCommand(ctx, source, nil, "git", "status", "--porcelain")
	if err != nil {
		return err
	}
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("backend build provenance unavailable")
	}
	revision, modified := "", ""
	for _, setting := range build.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}
	if revision == "" || revision != expectedRevision || modified != "false" || strings.TrimSpace(string(head)) != revision || len(status) != 0 {
		return errors.New("candidate backend must be built from this exact committed clean tool_source; rebuild after checkpointing")
	}
	listed, err := runBenchPrepareCommand(ctx, source, nil, "git", "ls-files", "-z")
	if err != nil {
		return err
	}
	paths, err := benchNativeSourcePaths(string(listed))
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	digest, err := hashBenchStepFile(exe, 128<<20)
	if err != nil {
		return err
	}
	if digest != expectedDigest {
		return errors.New("backend executable differs from the immutable Metis tool_sha256 input")
	}
	if err := writeBenchJSONExclusive(filepath.Join(f.Dir, "reads.json"), map[string]any{"roots": map[string][]string{source: paths}, "used_site_packages": false}); err != nil {
		return fmt.Errorf("write native Metis read set: %w", err)
	}
	return writeBenchJSONExclusive(filepath.Join(f.Dir, "tool-provenance.json"), map[string]any{"version": 1, "executable": exe, "sha256": digest, "source": source, "revision": revision, "modified": false})
}

// A Go build reads go:embed assets as well as source/module files. Capturing
// every tracked input is intentionally conservative and avoids an incomplete
// second implementation of the Go package dependency/embed resolver.
func benchNativeSourcePaths(listed string) ([]string, error) {
	paths := []string{}
	for _, path := range strings.Split(listed, "\x00") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	if len(paths) == 0 || len(paths) > 10000 {
		return nil, errors.New("native source dependency closure is empty or exceeds 10000 paths")
	}
	return paths, nil
}

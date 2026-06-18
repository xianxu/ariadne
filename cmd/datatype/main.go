package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// main runs the DAG-aware datatype subsystem. With no subcommand it renders the
// per-repo SKILL.md (the #115 dynamic-skill render):
//
//	datatype --output <dir> [--datatype-dir <ignored>]
//
// finds the repo root from cwd, walks its construct/deps layer graph, DAG-merges
// the prototypes (local/leaf shadows shared, keyed by filename), and writes
// <dir>/SKILL.md. Invoked by the datatype package's `.dynamic-skill` at `weave
// compile` time (cwd = the package dir; `--output .` writes the package dir).
//
// --datatype-dir is ACCEPTED-BUT-IGNORED (deprecated): the DAG merge across
// construct/deps supersedes the single-dir enumeration, but the flag is still
// parsed so the legacy marker's `--datatype-dir …` arg doesn't error.
func main() {
	output := flag.String("output", "", "directory to write SKILL.md into (required for the render command)")
	_ = flag.String("datatype-dir", "construct/datatype", "DEPRECATED + IGNORED: the DAG merge across construct/deps supersedes this; accepted only so the legacy .dynamic-skill marker's arg doesn't error")
	flag.Parse()

	if *output == "" {
		fmt.Fprintln(os.Stderr, "datatype: --output <dir> is required")
		os.Exit(2)
	}
	if err := renderToOutput(*output); err != nil {
		fmt.Fprintln(os.Stderr, "datatype:", err)
		os.Exit(1)
	}
}

// renderToOutput resolves the repo root from cwd, DAG-merges its prototypes,
// renders the SKILL.md, and writes it to <output>/SKILL.md. The thin IO shell
// over findRepoRoot + layergraph.Walk + mergeTypes + renderSkill.
func renderToOutput(output string) error {
	protos, err := resolveTypes()
	if err != nil {
		return err
	}
	names := make([]string, len(protos))
	for i, p := range protos {
		names[i] = p.Name
	}
	out := renderSkill(names)
	dst := filepath.Join(output, "SKILL.md")
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}

// resolveTypes is the shared resolution path for every command: from cwd find
// the repo root, walk its layer graph foundation-first, and DAG-merge the
// prototypes (leaf's project-local datatype/ overlaid last). The single
// DAG-aware access point — render, list, and show all consume it, so eager and
// apply-time can never disagree about a repo's type set.
func resolveTypes() ([]TypeProto, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}
	root, err := findRepoRoot(cwd)
	if err != nil {
		return nil, err
	}
	fs := layergraph.OSFS{}
	roots, err := layergraph.Walk(fs, root)
	if err != nil {
		return nil, fmt.Errorf("walk layer graph from %s: %w", root, err)
	}
	leafLocal := filepath.Join(root, "datatype")
	return mergeTypes(fs, roots, leafLocal)
}

// findRepoRoot walks up from start to the nearest ancestor directory that
// contains a construct/ subdir (the layer-root signal), falling back to the
// nearest ancestor with a .git dir. So apply-time `datatype list`/`show` anchor
// the repo root even when the agent's cwd is a deep subdir, and the legacy
// marker's cwd = construct/local/datatype resolves the repo root (its nearest
// construct/-bearing ancestor). Pure path-walk (the only IO is a Stat per
// ancestor at the edge). Returns an error if no construct/ or .git is found up
// to the filesystem root.
func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs %s: %w", start, err)
	}
	// First pass: nearest ancestor with construct/ (the layer-root signal).
	for d := dir; ; {
		if isDir(filepath.Join(d, "construct")) {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	// Fallback: nearest ancestor with .git.
	for d := dir; ; {
		if isDir(filepath.Join(d, ".git")) {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	return "", fmt.Errorf("no repo root found above %s (no ancestor has construct/ or .git)", start)
}

// isDir reports whether path exists and is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

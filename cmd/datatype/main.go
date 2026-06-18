package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// main runs the DAG-aware datatype subsystem. It dispatches on the first
// positional argument:
//
//	datatype list                 — print each DAG-resolved type's name + description
//	datatype show <name>          — print the leaf-winning prototype body
//	datatype --output <dir> [...] — render the per-repo SKILL.md (the render command)
//
// The render command (no subcommand, --output given) finds the repo root from
// cwd, walks its construct/deps layer graph, DAG-merges the prototypes
// (local/leaf shadows shared, keyed by filename), and writes <dir>/SKILL.md. It
// is invoked by the datatype package's `.dynamic-skill` at `weave compile` time
// (cwd = the package dir; `--output .` writes the package dir).
//
// --datatype-dir is ACCEPTED-BUT-IGNORED (deprecated): the DAG merge across
// construct/deps supersedes the single-dir enumeration, but the flag is still
// parsed so the legacy marker's `--datatype-dir …` arg doesn't error.
func main() {
	// Subcommand dispatch keys on the first positional arg BEFORE flag parsing,
	// so `datatype list` / `datatype show <name>` need no flags while the render
	// command keeps its `--output` flag.
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "list":
			if err := runList(os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "datatype:", err)
				os.Exit(1)
			}
			return
		case "show":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "datatype: show <name> requires a type name")
				os.Exit(2)
			}
			if err := runShow(args[1], os.Stdout, os.Stderr); err != nil {
				if err != errUnknownType {
					// errUnknownType already printed the available-names guidance.
					fmt.Fprintln(os.Stderr, "datatype:", err)
				}
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "datatype: unknown subcommand %q (want list, show, or --output)\n", args[0])
			os.Exit(2)
		}
	}

	output := flag.String("output", "", "directory to write SKILL.md into (required for the render command)")
	_ = flag.String("datatype-dir", "construct/datatype", "DEPRECATED + IGNORED: the DAG merge across construct/deps supersedes this; accepted only so the legacy .dynamic-skill marker's arg doesn't error")
	flag.Parse()

	if *output == "" {
		fmt.Fprintln(os.Stderr, "datatype: --output <dir> is required (or use the `list` / `show <name>` subcommand)")
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

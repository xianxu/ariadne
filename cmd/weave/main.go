// Command weave compiles a repo's agentic context from its layer DAG: it walks
// the layers (construct/deps), composes each layer's intents into an ordered
// []Action (the pure planner), and applies them to the filesystem.
//
//	weave              compile the current working directory's repo
//	weave --dry-run    print the planned []Action; mutate nothing
//	weave skills       print the skill menu (name — description)
//	weave skill <name> print a skill's SKILL.md body (served directly)
//
// The pure core (intent/, layer/, plan/, skill/) never touches disk; weave's
// only IO is the walk (reading manifests/deps/prose/skills) and plan.Apply (the
// mutations), behind weavefs.FS (ARCH-PURE). M3 part 1 adds the skill server:
// weave serves skills DIRECTLY (no .claude/skills/ discovery) — the menu is
// compiled into the composed AGENTS.md (always-on), bodies served on demand via
// `weave skill <name>`. M3 part 2 adds the `tool` lowering + `depend-on`.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/weave/internal/golden"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func main() {
	if err := buildRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the cobra command. Extracted from main so the wiring is
// testable. The root command IS the compile action (no subcommand needed);
// --dry-run flips it to print-only.
func buildRoot() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:           "weave",
		Short:         "Compile a repo's agentic context from its layer DAG",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return run(weavefs.OSFS{}, root, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the planned actions; mutate nothing")
	cmd.AddCommand(buildGolden())
	cmd.AddCommand(buildSkills())
	cmd.AddCommand(buildSkill())
	return cmd
}

// buildSkills assembles `weave skills` — print the agent-agnostic skill menu
// (the same menu compiled into AGENTS.md): one `name — description` line per
// skill, foundation-first with the downstream cascade. Read-only.
func buildSkills() *cobra.Command {
	return &cobra.Command{
		Use:           "skills",
		Short:         "Print the skill menu (name — description) served for this repo",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runSkills(weavefs.OSFS{}, root, cmd.OutOrStdout())
		},
	}
}

// buildSkill assembles `weave skill <name>` — serve the named skill's SKILL.md
// body on stdout (the agent-agnostic on-demand face). Errors non-zero with a
// helpful message on an unknown name. Read-only.
func buildSkill() *cobra.Command {
	return &cobra.Command{
		Use:           "skill <name>",
		Short:         "Print a skill's SKILL.md body (serve it directly, no .claude/skills)",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runSkill(weavefs.OSFS{}, root, args[0], cmd.OutOrStdout())
		},
	}
}

// buildGolden assembles `weave golden [repoPath...]` — the golden-diff harness.
// It verifies weave's INTENDED file-ops (a dry-run Plan, never applied) match
// what setup.sh already produced on the live repos (the live on-disk state IS
// setup.sh's output). For each given repo (or, with none given, each present
// sibling of the cwd's workspace root), it walks → Plans → observes the live FS
// → classifies every divergence as MATCH/EXPECTED/UNEXPECTED, prints a ledger,
// and exits non-zero if ANY divergence is UNEXPECTED. STRICTLY read-only.
func buildGolden() *cobra.Command {
	return &cobra.Command{
		Use:   "golden [repoPath...]",
		Short: "Verify weave's intended file-ops match setup.sh's live output (read-only)",
		Long: "Compares weave's planned actions (dry-run, never applied) against the\n" +
			"live repos' current filesystem — which IS setup.sh's output — and\n" +
			"classifies divergences. Exits non-zero on any UNEXPECTED divergence.\n" +
			"With no args, auto-discovers present sibling repos of the cwd's parent.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runGolden(weavefs.OSFS{}, cwd, args, cmd.OutOrStdout())
		},
	}
}

// runGolden is the harness pipeline over a set of repos: resolve the target
// repos (explicit args, or auto-discovered present siblings), then for each
// run walk → Plan (NO apply) → Gather (observe live) → Classify → Render. It
// prints each per-repo ledger and returns an error iff any repo had an
// UNEXPECTED divergence (the non-zero exit). Skips an absent repo with a note
// (skip-if-absent). Injecting fs + out keeps it testable.
func runGolden(fs weavefs.FS, cwd string, args []string, out io.Writer) error {
	repos := goldenTargets(cwd, args)
	anyUnexpected := false
	for _, repo := range repos {
		if !dirPresent(repo) {
			fmt.Fprintf(out, "== golden-diff: %s ==\n  SKIP — repo not present\n\n", repo)
			continue
		}
		// Canonicalize to the same physical namespace the walk uses (pwd -P),
		// so the relative-symlink targets weave computes match the live links.
		root := repo
		if resolved, err := filepath.EvalSymlinks(root); err == nil {
			root = resolved
		}
		layers, err := walk.Walk(fs, root)
		if err != nil {
			return fmt.Errorf("golden: walk %s: %w", root, err)
		}
		// Golden classifies weave's FILE-OPS against setup.sh's live output;
		// the skill menu is served via AGENTS.md prose, not file-ops, and weave's
		// skill mechanism intentionally DIVERGES from setup.sh's .claude/skills
		// symlinks (an expected M5 divergence — the classifier doesn't model it
		// yet), so pass a nil menu here: the AGENTS.md prose body is already
		// classified; the skill section isn't a setup.sh-comparable file-op.
		actions, err := plan.Plan(layers, nil)
		if err != nil {
			return fmt.Errorf("golden: plan %s: %w", root, err)
		}
		deferred := golden.DeferredIntents(layers)
		in := golden.Gather(fs, root, actions, deferred)
		divs := golden.Classify(in)
		fmt.Fprint(out, golden.Render(root, divs))
		fmt.Fprintln(out)
		if golden.HasUnexpected(divs) {
			anyUnexpected = true
		}
	}
	if anyUnexpected {
		return fmt.Errorf("golden-diff: UNEXPECTED divergence(s) found — see ledger above")
	}
	return nil
}

// goldenTargets resolves which repos to check. Explicit args win (each made
// absolute against cwd). With no args, it auto-discovers the present sibling
// repos of cwd's parent (the workspace root): the canonical ariadne layers
// (ariadne, nous, brain, metis). Pure (string in/out — presence is filtered by
// the IO caller via dirPresent), so it's unit-testable.
func goldenTargets(cwd string, args []string) []string {
	if len(args) > 0 {
		out := make([]string, 0, len(args))
		for _, a := range args {
			if !filepath.IsAbs(a) {
				a = filepath.Join(cwd, a)
			}
			out = append(out, a)
		}
		return out
	}
	// No args: the canonical layer repos as siblings of the workspace root.
	// The worktree lives at …/workspace/worktree/ariadne/<branch>; the LIVE
	// repos are at …/workspace/<name>. Walk up to the dir that holds them.
	ws := workspaceRoot(cwd)
	var out []string
	for _, name := range []string{"ariadne", "nous", "brain", "metis"} {
		out = append(out, filepath.Join(ws, name))
	}
	return out
}

// workspaceRoot finds the dir that holds the live sibling repos, given the cwd.
// A normal repo's parent is the workspace; a worktree at
// …/workspace/worktree/<repo>/<branch> must climb past worktree/<repo>. We
// detect the worktree shape by the literal "worktree" path segment. Pure.
func workspaceRoot(cwd string) string {
	parent := filepath.Dir(cwd)
	if filepath.Base(filepath.Dir(parent)) == "worktree" {
		// cwd = …/workspace/worktree/<repo>/<branch>
		//   parent             = …/workspace/worktree/<repo>
		//   dir(parent)        = …/workspace/worktree   (base == "worktree")
		//   dir(dir(parent))   = …/workspace            ← the workspace root
		return filepath.Dir(filepath.Dir(parent))
	}
	return parent
}

// dirPresent reports whether path is an existing directory (skip-if-absent).
func dirPresent(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// run is the compile pipeline: walk → Plan → (Apply | print). Injecting fs +
// out keeps it testable against a t.TempDir-rooted OSFS and a buffer.
//
// root is canonicalized to its physical form up front (filepath.EvalSymlinks ≈
// pwd -P) so it lives in the SAME namespace as the layer Paths the walk
// canonicalizes — without this, on macOS (/tmp → /private/tmp) Apply would
// compute a relative symlink target between a logical dst-dir and a physical
// upstream src that resolves wrong when the OS follows the link (the exact bug
// setup.sh's pwd -P guards against, lines 39-45).
func run(fs weavefs.FS, root string, dryRun bool, out io.Writer) error {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	layers, err := walk.Walk(fs, root)
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	idx, err := buildSkillIndex(fs, layers)
	if err != nil {
		return fmt.Errorf("gather skills: %w", err)
	}
	actions, err := plan.Plan(layers, idx.Menu())
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if dryRun {
		fmt.Fprint(out, formatActions(actions))
		return nil
	}
	if err := plan.Apply(fs, root, actions); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	fmt.Fprintf(out, "weave: applied %d action(s) to %s\n", len(actions), root)
	return nil
}

// buildSkillIndex is the skill-server pipeline up to the pure index: walk's
// discovery seam (walk.GatherSkills, ports sync-local-skills.sh) → skill.Build
// (the pure menu + body lookup). Shared by the compile path (for the AGENTS.md
// menu) and the skills/skill subcommands. layers must already be walked.
func buildSkillIndex(fs weavefs.FS, layers []layer.Layer) (skill.SkillIndex, error) {
	entries, err := walk.GatherSkills(fs, layers)
	if err != nil {
		return skill.SkillIndex{}, err
	}
	return skill.Build(entries), nil
}

// resolveSkillIndex canonicalizes root, walks the layers, and builds the index
// — the read-only front half the skills/skill subcommands share. No mutation.
func resolveSkillIndex(fs weavefs.FS, root string) (skill.SkillIndex, error) {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	layers, err := walk.Walk(fs, root)
	if err != nil {
		return skill.SkillIndex{}, fmt.Errorf("walk %s: %w", root, err)
	}
	return buildSkillIndex(fs, layers)
}

// runSkills prints the skill menu (one `name — description` line per skill) for
// the repo at root. Read-only; injecting fs + out keeps it testable.
func runSkills(fs weavefs.FS, root string, out io.Writer) error {
	idx, err := resolveSkillIndex(fs, root)
	if err != nil {
		return err
	}
	menu := idx.Menu()
	if len(menu) == 0 {
		fmt.Fprintln(out, "weave: no skills")
		return nil
	}
	for _, m := range menu {
		if m.Description != "" {
			fmt.Fprintf(out, "%s — %s\n", m.Name, m.Description)
		} else {
			fmt.Fprintln(out, m.Name)
		}
	}
	return nil
}

// runSkill serves the named skill's SKILL.md body on out. Unknown name → a
// non-nil error (non-zero exit) listing the available names. Read-only.
func runSkill(fs weavefs.FS, root, name string, out io.Writer) error {
	idx, err := resolveSkillIndex(fs, root)
	if err != nil {
		return err
	}
	bodyPath, ok := idx.BodyPath(name)
	if !ok {
		return fmt.Errorf("unknown skill %q; run `weave skills` to list available skills", name)
	}
	body, err := fs.ReadFile(bodyPath)
	if err != nil {
		return fmt.Errorf("read skill %q (%s): %w", name, bodyPath, err)
	}
	fmt.Fprint(out, string(body))
	if len(body) > 0 && body[len(body)-1] != '\n' {
		fmt.Fprintln(out)
	}
	return nil
}

// formatActions renders a []Action as one line per action for --dry-run. Pure
// (string in/out), so it's unit-tested directly.
func formatActions(actions []plan.Action) string {
	if len(actions) == 0 {
		return "weave: no actions\n"
	}
	var b []byte
	for _, a := range actions {
		switch act := a.(type) {
		case plan.Symlink:
			b = append(b, fmt.Sprintf("symlink   %s -> %s\n", act.Dst, act.Src)...)
		case plan.WriteFile:
			b = append(b, fmt.Sprintf("writefile %s (%d bytes)\n", act.Path, len(act.Content))...)
		case plan.Mkdir:
			b = append(b, fmt.Sprintf("mkdir     %s\n", act.Path)...)
		case plan.Touch:
			b = append(b, fmt.Sprintf("touch     %s\n", act.Path)...)
		case plan.MergeSettings:
			b = append(b, fmt.Sprintf("merge     %s -> %s\n", act.Source, act.Target)...)
		case plan.ToolDep:
			b = append(b, fmt.Sprintf("tool      %s (%s)\n", act.Path, act.Owner)...)
		default:
			b = append(b, fmt.Sprintf("unknown   %T\n", a)...)
		}
	}
	return string(b)
}

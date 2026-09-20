// Command weave compiles a repo's agentic context from its layer DAG: it walks
// the layers (construct/deps), composes each layer's intents into an ordered
// []Action (the pure planner), and applies them to the filesystem.
//
//	weave                          (root) print help — the bare command does NOT compile
//	weave compile [--target T]     compile the cwd repo for backend T (default claude)
//	weave compile --dry-run        print the planned []Action; mutate nothing
//	weave golden [--target T]      verify weave's plan matches setup.sh's live output
//	weave verify-complete [--target T]  assert the plan covers every managed path
//	weave skills                   print the skill listing (name — description)
//	weave skill <name>             print a skill's SKILL.md body (served directly)
//	weave link <path>              record a `substrate <path>` dep in construct/deps
//
// Per-harness faces (Option B, #107): the Union (default) lowers every harness's
// FACE — a prose entry file (CLAUDE.md/AGENTS.md/GEMINI.md) + a skill dir
// (.claude/skills for Claude, .agents/skills for Codex/Gemini); `--target T`
// lowers only T's face and prunes the others. There is no `## Skills` menu — each
// harness discovers its own skill dir. Every non-skill, non-prose file-op is
// target-independent. See plan.Target / plan.Faces.
//
// The pure core (intent/, layer/, plan/, skill/) never touches disk; weave's
// only IO is the walk (reading manifests/deps/prose/skills) and plan.Apply (the
// mutations), behind weavefs.FS (ARCH-PURE). M3 part 1 adds the skill server:
// weave serves skill bodies on demand via `weave skill <name>`; harnesses
// discover the lowered skill dirs directly (no menu). M3 part 2 adds `link`
// (substrate deps). M5 makes the compile an explicit subcommand with `--target`
// face selection and retires the `tool` verb (Go-tool ownership is location-based
// via dev-aliases.sh, not a go.mod edit).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/golden"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
	"github.com/xianxu/ariadne/cmd/weave/internal/startup"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := buildRoot().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the cobra command. Extracted from main so the wiring is
// testable. The root command no longer compiles (M5): with RunE nil, a bare
// `weave` prints help/usage and mutates nothing. Compiling is now the explicit
// `weave compile` subcommand, which carries --dry-run and --target.
func buildRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weave",
		Short: "Compile a repo's agentic context from its layer DAG",
		Long: "weave compiles a repo's agentic context from its layer DAG.\n\n" +
			"The bare `weave` command prints this help and mutates nothing; run\n" +
			"`weave compile` to actually compile. By default (the Union) it lowers\n" +
			"every harness FACE — CLAUDE.md/AGENTS.md/GEMINI.md (prose) + .claude/skills\n" +
			"(Claude) + .agents/skills (Codex/Gemini); `--target T` lowers only T's face.",
		SilenceUsage:  true,
		SilenceErrors: true,
		// RunE intentionally nil: the bare command is help-only (no compile).
	}
	cmd.AddCommand(buildCompile())
	cmd.AddCommand(buildGolden())
	cmd.AddCommand(buildVerifyComplete())
	cmd.AddCommand(buildSkills())
	cmd.AddCommand(buildSkill())
	cmd.AddCommand(buildLink())
	cmd.AddCommand(buildDependencies())
	return cmd
}

// buildCompile assembles `weave compile [--target <face>] [--dry-run]` — the
// explicit compile verb (M5; was the root's RunE). Default = the Union (every
// harness face); --target T lowers only T's face (entry file + skill dir) and
// prunes the others. No `## Skills` menu — each harness discovers its own skill
// dir (.claude/skills for claude, .agents/skills for codex/gemini). An unknown
// target errors clearly (plan.ParseTarget). --dry-run prints the planned actions
// and mutates nothing.
func buildCompile() *cobra.Command {
	var dryRun bool
	var targetFlag string
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Install dependencies, build layer tools, and compile artifacts",
		Long: "Restores base sources, installs layer Brewfiles, runs owner make tools,\n" +
			"then generates and reconciles artifacts. Default = every harness face.\n" +
			"`--target {claude|codex|gemini}` selects a face. `--dry-run` previews\n" +
			"known operations without cloning, installing, building or generating.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := plan.ParseTarget(targetFlag)
			if err != nil {
				return err
			}
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runCompile(cmd.Context(), weavefs.OSFS{}, root, target, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the planned actions; mutate nothing")
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetAll), "harness target: all | claude | codex | gemini (default all = the Union of every harness's face)")
	return cmd
}

// buildVerifyComplete assembles `weave verify-complete [repoPath...]` — the
// completeness check (the COMPANION to golden). golden classifies the paths
// weave PLANS (catching MIS-production); verify-complete asserts weave's plan
// COVERS every path setup.sh would produce (catching UNDER-production — a
// manifest entry weave's lowering silently drops). For each repo it walks →
// Plans → CheckCompleteness, prints the gaps, and exits non-zero if ANY path is
// under-produced. With seed implemented, ariadne self-walk reports ZERO. Strictly
// read-only.
func buildVerifyComplete() *cobra.Command {
	var targetFlag string
	cmd := &cobra.Command{
		Use:   "verify-complete [repoPath...]",
		Short: "Assert weave's plan covers every path setup.sh would produce (read-only)",
		Long: "Independent completeness check: enumerates every managed path the\n" +
			"walked manifests declare (the setup.sh-equivalent managed set) and\n" +
			"asserts weave's plan covers each. Catches UNDER-production a golden-diff\n" +
			"cannot see (a verb whose lowering drops the entry). Exits non-zero on any\n" +
			"under-produced path. With no args, auto-discovers present sibling repos.\n" +
			"`--target` selects whose plan is checked (default the Union; a skill\n" +
			"intent is covered by its per-harness skill-dir symlinks).",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := plan.ParseTarget(targetFlag)
			if err != nil {
				return err
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runVerifyComplete(weavefs.OSFS{}, cwd, args, target, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetAll), "harness target: all | claude | codex | gemini (default all = the Union of every harness's face)")
	return cmd
}

// runVerifyComplete is the completeness pipeline over a set of repos, for a given
// target: it reuses goldenTargets (the same repo resolution as the golden harness,
// ARCH-DRY) and, per present repo, walks → Plans (planActions — the IDENTICAL plan
// the compile path + golden see for this target, so the lowered skill dirs count
// for skill coverage) → CheckCompleteness → Render. A skill intent is satisfied by
// its per-harness skill-dir symlinks, so both the Union and a lean --target report
// zero under-production. Returns an error iff any repo had an under-produced path.
// Injecting fs + out keeps it testable.
func runVerifyComplete(fs weavefs.FS, cwd string, args []string, target plan.Target, out io.Writer) error {
	repos := goldenTargets(cwd, args)
	anyUnder := false
	for _, repo := range repos {
		if !dirPresent(repo) {
			fmt.Fprintf(out, "== completeness: %s ==\n  SKIP — repo not present\n\n", repo)
			continue
		}
		root := repo
		if resolved, err := filepath.EvalSymlinks(root); err == nil {
			root = resolved
		}
		layers, err := walk.Walk(fs, root)
		if err != nil {
			return fmt.Errorf("verify-complete: walk %s: %w", root, err)
		}
		actions, err := planActions(fs, layers, target)
		if err != nil {
			return fmt.Errorf("verify-complete: plan %s: %w", root, err)
		}
		uncovered := golden.CheckCompleteness(layers, actions)
		fmt.Fprint(out, golden.RenderCompleteness(root, uncovered))
		fmt.Fprintln(out)
		if len(uncovered) > 0 {
			anyUnder = true
		}
	}
	if anyUnder {
		return fmt.Errorf("verify-complete: under-produced path(s) found — see report above")
	}
	return nil
}

// buildLink assembles `weave link <path>` — the directory-agnostic
// substrate-establishment verb. It records `substrate <path>` VERBATIM (the path
// exactly as given, relative or absolute) in the cwd repo's construct/deps,
// idempotently, AND seeds a minimal construct/base.manifest when absent so the
// repo is a valid traversable layer out of the box (#155). This is the
// module-include verb of weave's repo-composition dialect: how a fresh derivative
// declares its dependency on a real ariadne checkout anywhere on disk (the plan's
// "directory-agnostic substrate paths" Revisions entry) — recording the real path,
// not a hardcoded ../ariadne.
func buildLink() *cobra.Command {
	return &cobra.Command{
		Use:           "link <local-path|repo-address>",
		Short:         "Link a base repository, cloning a remote source as a peer when absent",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return linkRepository(cmd.Context(), root, args[0], cmd.OutOrStdout())
		},
	}
}

// runLink appends `substrate <path>` to root/construct/deps, recording path
// VERBATIM (no resolution/relativization — the establishment verb captures the
// real path it was handed). Idempotent: it reuses layergraph.ParseDeps (the same
// grammar the walk + Apply read deps with, ARCH-DRY) to skip when the row is
// already present, and creates construct/deps (+ construct/) when absent. It then
// seeds construct/base.manifest when absent (#155), so the two files it may write
// are construct/deps and construct/base.manifest — nothing else. Injecting fs +
// out keeps it testable.
func runLink(fs weavefs.FS, root, path string, out io.Writer) error {
	return recordLink(fs, root, path, "", out)
}

// ensureBaseManifest seeds root/construct/base.manifest when absent so the repo is
// a traversable weave layer (#155). Idempotent — a present manifest (hand-authored
// or already seeded) is left untouched. Injecting fs + out keeps it testable.
func ensureBaseManifest(fs weavefs.FS, root, substratePath string, out io.Writer) error {
	manifestPath := filepath.Join(root, "construct", "base.manifest")
	if _, err := fs.Stat(manifestPath); err == nil {
		return nil // already a layer — never overwrite a hand-authored manifest
	}
	if err := fs.MkdirAll(filepath.Dir(manifestPath)); err != nil {
		return fmt.Errorf("link: mkdir %s: %w", filepath.Dir(manifestPath), err)
	}
	if err := fs.WriteFile(manifestPath, []byte(seededBaseManifest(filepath.Base(root), substratePath))); err != nil {
		return fmt.Errorf("link: write %s: %w", manifestPath, err)
	}
	fmt.Fprintf(out, "weave: seeded construct/base.manifest (marks %s a traversable layer)\n", filepath.Base(root))
	return nil
}

// seededBaseManifest returns the minimal construct/base.manifest weave link seeds
// for a fresh derivative (#155): a header naming the repo + its substrate, and the
// single `internal prose AGENTS.local.md` row that both declares the repo's own
// constitution fragment and marks it a traversable layer. Mirrors what #95-cutover
// repos hand-authored — the SINGLE source for the seed content (ARCH-DRY), so a
// future change to the stub lands in one place rather than a third hand-copy.
func seededBaseManifest(repoName, substratePath string) string {
	return "# " + repoName + " base.manifest — " + repoName + "'s module declaration for the weave layer compiler.\n" +
		"#\n" +
		"# Seeded by `weave link " + substratePath + "` (ariadne#155). " + repoName + " consumes " + substratePath + "\n" +
		"# as its substrate (construct/deps `substrate " + substratePath + "`) and inherits that layer's\n" +
		"# EXPORTED base layer through the edge. This manifest declares " + repoName + "'s OWN internal artifacts.\n" +
		"#\n" +
		"# Format and verbs: see ariadne's construct/base.manifest. The leading `internal` token\n" +
		"# (ariadne#99) keeps an artifact with this repo (selected only on its own self-walk), never\n" +
		"# leaked into a consumer. NOTE: shipping a base.manifest is ALSO what marks this repo a\n" +
		"# TRAVERSABLE layer — without it a downstream consumer's weave walk stops here.\n" +
		"\n" +
		"# ── Constitution ──────────────────────────────────────────────────────────────\n" +
		"internal  prose AGENTS.local.md\n"
}

// buildSkills assembles `weave skills` — print the agent-agnostic skill listing
// served for this repo: one `name — description` line per skill, foundation-first
// with the downstream cascade. A diagnostic CLI view of the served set (NOT a menu
// compiled into any entry file — harnesses discover the lowered skill dirs
// natively, #107). Read-only.
func buildSkills() *cobra.Command {
	return &cobra.Command{
		Use:           "skills",
		Short:         "Print the skill listing (name — description) served for this repo",
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
	var targetFlag string
	cmd := &cobra.Command{
		Use:   "golden [repoPath...]",
		Short: "Verify weave's intended file-ops match setup.sh's live output (read-only)",
		Long: "Compares weave's planned actions (dry-run, never applied) against the\n" +
			"live repos' current filesystem — which IS setup.sh's output — and\n" +
			"classifies divergences. Exits non-zero on any UNEXPECTED divergence.\n" +
			"With no args, auto-discovers present sibling repos of the cwd's parent.\n" +
			"`--target claude` is the parity check (setup.sh produced claude-shaped\n" +
			".claude/skills); other targets intentionally diverge from setup.sh.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := plan.ParseTarget(targetFlag)
			if err != nil {
				return err
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runGolden(weavefs.OSFS{}, cwd, args, target, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetAll), "harness target: all | claude | codex | gemini (default all = the Union of every harness's face)")
	return cmd
}

// runGolden is the harness pipeline over a set of repos: resolve the target
// repos (explicit args, or auto-discovered present siblings), then for each
// run walk → Plan (NO apply) → Gather (observe live) → Classify → Render. It
// prints each per-repo ledger and returns an error iff any repo had an
// UNEXPECTED divergence (the non-zero exit). Skips an absent repo with a note
// (skip-if-absent). Injecting fs + out keeps it testable.
func runGolden(fs weavefs.FS, cwd string, args []string, target plan.Target, out io.Writer) error {
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
		// Golden classifies weave's FILE-OPS against setup.sh's live output. Use
		// the SAME planActions the compile path uses (ARCH-DRY) so the per-harness
		// skill-dir symlinks weave now emits are classified against the live links —
		// they MATCH the ones sync-local-skills.sh wrote, the M5 cutover's parity
		// check. The entry-file WriteFiles (CLAUDE.md/AGENTS.md/GEMINI.md prose) are
		// classified as bodies; their content intentionally DIVERGES from setup.sh's
		// symlinked AGENTS.md (an expected, hand-checked M5/#107 divergence).
		actions, err := planActions(fs, layers, target)
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

// run retains the test-friendly entry point; commands supply their cancellation context.
func run(fs weavefs.FS, root string, target plan.Target, dryRun bool, out io.Writer) error {
	return runCompile(context.Background(), fs, root, target, dryRun, out)
}

// runCompile restores the graph, prepares each owner, then composes the leaf.
func runCompile(ctx context.Context, fs weavefs.FS, root string, target plan.Target, dryRun bool, out io.Writer) (retErr error) {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	restored, err := acquire.Restore(ctx, root, dryRun)
	for _, missing := range restored.Missing {
		fmt.Fprintf(out, "weave: missing source %s (preview incomplete)\n", missing)
	}
	if err != nil {
		return err
	}
	runner := weavefs.ExecRunner{Context: ctx, Stdout: out, Stderr: out}
	if err := startup.Dependencies(fs, restored.Layers, runner, dryRun, out); err != nil {
		return err
	}
	runner.Env = startup.ToolEnvironment(os.Environ(), restored.Layers)
	if err := startup.Tools(fs, restored.Layers, runner, dryRun, out); err != nil {
		return err
	}
	layers, err := walk.Load(fs, restored.Layers)
	if err != nil {
		return fmt.Errorf("load layers: %w", err)
	}
	dyns, err := walk.DynamicSkills(fs, layers)
	if err != nil {
		return fmt.Errorf("select dynamic skills: %w", err)
	}
	// Data belongs to each declaring owner, independently of its composed artifacts.
	for _, owner := range restored.Layers {
		var mounts []plan.Action
		for _, mount := range restored.Mounts {
			if mount.Owner != owner {
				continue
			}
			dst, err := filepath.Rel(owner, mount.Target)
			if err != nil {
				return err
			}
			mounts = append(mounts, plan.Symlink{Src: mount.Source, Dst: dst})
		}
		if dryRun {
			fmt.Fprint(out, formatActions(mounts))
			continue
		}
		if _, err := plan.ApplyManaged(fs, owner, mounts, plan.ScopeData); err != nil {
			return fmt.Errorf("data mounts for %s: %w", owner, err)
		}
	}
	var generated []plan.Action
	if !dryRun {
		if len(dyns) == 0 {
			if err := plan.ReclaimGenerationStages(fs, root); err != nil {
				return err
			}
		} else {
			stage, err := plan.NewGenerationStage(fs, root)
			if err != nil {
				return err
			}
			defer func() { retErr = errors.Join(retErr, plan.RemoveGenerationStage(stage)) }()
			if err := generateDynamicSkills(fs, dyns, root, stage, runner); err != nil {
				return err
			}
			for _, ds := range dyns {
				outputs, err := plan.StagedActions(fs, root, filepath.Join(stage, ds.Dir), ds.OutputRel)
				if err != nil {
					return fmt.Errorf("dynamic skill %s: %w", ds.Name, err)
				}
				generated = append(generated, outputs...)
			}
		}
	}
	actions, err := planActions(fs, layers, target)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprint(out, formatActions(actions))
		fmt.Fprintln(out, "weave: generation and retirement are not previewed; dry-run changes nothing")
		return nil
	}
	actions = append(actions, generated...)
	retired, err := plan.ApplyManaged(fs, root, actions, plan.ScopeArtifacts)
	if err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	fmt.Fprintf(out, "weave: applied %d action(s) to %s\n", len(actions), root)
	for _, path := range retired {
		fmt.Fprintf(out, "  retired %s\n", path)
	}
	var bins []string
	for _, owner := range restored.Layers {
		bin := filepath.Join(owner, "bin")
		if info, err := fs.Stat(bin); err == nil && info.IsDir() {
			bins = append(bins, bin)
		}
	}
	if len(bins) > 0 {
		fmt.Fprintln(out, "weave: add these layer tool directories to your shell PATH:")
		for _, bin := range bins {
			fmt.Fprintf(out, "  %s\n", bin)
		}
	}
	return nil
}

// generateDynamicSkills keeps the leaf cwd for graph reads, but supplies an
// isolated output directory. Markers opt in before execution; this is a trusted
// layer-code contract, not a sandbox for arbitrary shell programs.
func generateDynamicSkills(fs weavefs.FS, dyns []walk.DynamicSkill, leafRoot, stage string, runner weavefs.Runner) error {
	for _, ds := range dyns {
		body, err := fs.ReadFile(ds.MarkerPath)
		if err != nil {
			return err
		}
		supported := false
		for _, line := range strings.Split(string(body), "\n") {
			if strings.TrimSpace(line) == "# weave-output: argv1" {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("dynamic skill %s: marker must declare '# weave-output: argv1' and write to its supplied output directory", ds.Name)
		}
	}
	for _, ds := range dyns {
		output := filepath.Join(stage, ds.Dir)
		if err := runner.Run(leafRoot, []string{"sh", ds.MarkerPath, output}); err != nil {
			return fmt.Errorf("dynamic skill %s: %w", ds.Name, err)
		}
	}
	return nil
}

// planActions is the full compile lowering for a set of resolved layers, for a
// given target (Option B, #107). Skills lower as per-harness skill-dir symlinks —
// .claude/skills for claude, .agents/skills for codex/gemini (Target.SkillDirs);
// prose lowers as per-harness entry files — CLAUDE.md/AGENTS.md/GEMINI.md
// (Target.EntryFiles). The Union (default) emits every face; a lean --target T
// emits only T's. There is NO `## Skills` menu — each harness discovers its own
// skill dir natively.
//
// Every other file-op (settings merge, scaffold, touch, generic symlink, seed) is
// target-independent. Shared by the compile path (run), the golden harness
// (runGolden), and verify-complete (runVerifyComplete) so all see the IDENTICAL
// action set for a given target (ARCH-DRY).
func planActions(fs weavefs.FS, layers []layer.Layer, target plan.Target) ([]plan.Action, error) {
	// ONE skill discovery (#104): gather → SelectVisible (𝒜(R)). The compile path
	// uses only the selected entries — NO menu (Option B, #107: every harness
	// discovers its own skill dir natively). buildSkillIndex's index serves
	// `weave skills`/`weave skill <name>` elsewhere; here we discard it.
	_, selected, err := buildSkillIndex(fs, layers)
	if err != nil {
		return nil, fmt.Errorf("gather skills: %w", err)
	}
	// Prose → each per-harness ENTRY FILE; the selected skills → each per-harness
	// SKILL DIR. Faces() decides which: the Union (default) writes CLAUDE.md +
	// AGENTS.md + GEMINI.md + .claude/skills + .agents/skills; a lean --target T
	// writes only T's face. codex+gemini share .agents/skills (Target.SkillDirs
	// dedupes).
	actions, err := plan.Plan(layers, target.EntryFiles())
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}
	for _, dir := range target.SkillDirs() {
		for _, l := range plan.SkillSymlinks(selected, dir) {
			actions = append(actions, l)
		}
	}
	actions = append(actions, plan.EnsureGitignore{Entries: plan.GeneratedGitignoreEntries(actions)})
	return actions, nil
}

// buildSkillIndex is weave's SINGLE skill pipeline (#104): walk.GatherSkills (the
// one IO discovery, intent-driven, carrying Visibility+LayerIndex) → skill.SelectVisible
// (the pure 𝒜(R) filter; leaf = the last layer) → skill.Build (the pure body
// lookup). It returns BOTH the index AND the selected entries: the index serves
// `weave skills`/`weave skill <name>`, while the compile path lowers the SAME
// selected entries into each per-harness skill dir (plan.SkillSymlinks) — one scan,
// reused everywhere (ARCH-DRY). No menu (Option B). layers must already be walked.
func buildSkillIndex(fs weavefs.FS, layers []layer.Layer) (skill.SkillIndex, []skill.Entry, error) {
	entries, err := walk.GatherSkills(fs, layers)
	if err != nil {
		return skill.SkillIndex{}, nil, err
	}
	selected := skill.SelectVisible(entries, len(layers)-1)
	return skill.Build(selected), selected, nil
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
	// The skills/skill subcommands serve the COMPOSED set (same select as compile),
	// so a served skill is exactly a lowered one; the entries are discarded here.
	idx, _, err := buildSkillIndex(fs, layers)
	return idx, err
}

// runSkills prints the skill listing (one `name — description` line per skill) for
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
		case plan.Seed:
			b = append(b, fmt.Sprintf("seed      %s -> %s\n", act.Dst, act.Src)...)
		case plan.Touch:
			b = append(b, fmt.Sprintf("touch     %s\n", act.Path)...)
		case plan.MergeSettings:
			b = append(b, fmt.Sprintf("merge     %s -> %s\n", strings.Join(act.Sources, ", "), act.Target)...)
		case plan.EnsureGitignore:
			b = append(b, fmt.Sprintf("gitignore .gitignore (%d entries)\n", len(act.Entries))...)
		default:
			b = append(b, fmt.Sprintf("unknown   %T\n", a)...)
		}
	}
	return string(b)
}

// formatPrunes renders the dry-run preview of the #96 orphan-symlink prune: one
// `prune` line per repo-relative path a real apply WOULD delete. Empty (no
// output) when there is nothing to prune — so a healthy repo's dry-run shows no
// prune noise. Pure (string in/out).
func formatPrunes(paths []string) string {
	var b []byte
	for _, p := range paths {
		b = append(b, fmt.Sprintf("prune     %s (orphaned lowered symlink)\n", p)...)
	}
	return string(b)
}

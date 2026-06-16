// Command weave compiles a repo's agentic context from its layer DAG: it walks
// the layers (construct/deps), composes each layer's intents into an ordered
// []Action (the pure planner), and applies them to the filesystem.
//
//	weave                          (root) print help — the bare command does NOT compile
//	weave compile [--target T]     compile the cwd repo for backend T (default claude)
//	weave compile --dry-run        print the planned []Action; mutate nothing
//	weave golden [--target T]      verify weave's plan matches setup.sh's live output
//	weave verify-complete [--target T]  assert the plan covers every managed path
//	weave skills                   print the skill menu (name — description)
//	weave skill <name>             print a skill's SKILL.md body (served directly)
//	weave link <path>              record a `substrate <path>` dep in construct/deps
//
// One target per invocation (Approach-1): `--target` picks ONE skill backend —
// claude lowers .claude/skills symlinks with a prose-only AGENTS.md; codex/agy
// suppress the symlinks and compose the `## Skills` menu into AGENTS.md instead.
// The two skill backends are mutually exclusive; every other file-op is
// target-independent. See plan.Target.
//
// The pure core (intent/, layer/, plan/, skill/) never touches disk; weave's
// only IO is the walk (reading manifests/deps/prose/skills) and plan.Apply (the
// mutations), behind weavefs.FS (ARCH-PURE). M3 part 1 adds the skill server:
// weave serves skills DIRECTLY (no .claude/skills/ discovery) — the menu is
// compiled into the composed AGENTS.md (always-on), bodies served on demand via
// `weave skill <name>`. M3 part 2 adds `link` (substrate deps). M5 makes the
// compile an explicit subcommand with `--target` backend selection and retires
// the `tool` verb (Go-tool ownership is location-based via dev-aliases.sh, not a
// go.mod edit).
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
// testable. The root command no longer compiles (M5): with RunE nil, a bare
// `weave` prints help/usage and mutates nothing. Compiling is now the explicit
// `weave compile` subcommand, which carries --dry-run and --target.
func buildRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weave",
		Short: "Compile a repo's agentic context from its layer DAG",
		Long: "weave compiles a repo's agentic context from its layer DAG.\n\n" +
			"The bare `weave` command prints this help and mutates nothing; run\n" +
			"`weave compile` to actually compile. One target per invocation:\n" +
			"`--target claude` lowers .claude/skills symlinks (prose-only AGENTS.md),\n" +
			"`--target codex`/`agy` compose the skill menu into AGENTS.md instead.",
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
	return cmd
}

// buildCompile assembles `weave compile [--target <backend>] [--dry-run]` — the
// explicit compile verb (M5; was the root's RunE). --target selects ONE skill
// backend (default claude): claude → .claude/skills symlinks + prose-only
// AGENTS.md; codex/agy → no symlinks + the `## Skills` menu in AGENTS.md. An
// unknown target errors clearly (plan.ParseTarget). --dry-run prints the planned
// actions and mutates nothing.
func buildCompile() *cobra.Command {
	var dryRun bool
	var targetFlag string
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile the cwd repo's agentic context for a backend target",
		Long: "Compiles the current working directory's repo: walk the layer DAG,\n" +
			"plan the file-ops, and apply them. `--target` selects the skill backend\n" +
			"(claude=.claude/skills symlinks + prose-only AGENTS.md; codex/agy=the\n" +
			"`## Skills` menu in AGENTS.md, no symlinks). `--dry-run` prints the plan.",
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
			return run(weavefs.OSFS{}, root, target, dryRun, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the planned actions; mutate nothing")
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetClaude), "skill backend target: claude | codex | agy")
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
			"`--target` selects the backend whose plan is checked (a skill intent is\n" +
			"covered by EITHER backend: claude's symlinks or codex/agy's menu).",
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
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetClaude), "skill backend target: claude | codex | agy")
	return cmd
}

// runVerifyComplete is the completeness pipeline over a set of repos, for ONE
// backend target: it reuses goldenTargets (the same repo resolution as the golden
// harness, ARCH-DRY) and, per present repo, walks → Plans (planActions — the
// IDENTICAL plan the compile path + golden see for this target, so the active
// skill backend counts for skill coverage) → CheckCompleteness → Render. A skill
// intent is satisfied by EITHER backend, so both --target claude (symlinks) and
// --target codex (menu) report zero under-production. Returns an error iff any
// repo had an under-produced path. Injecting fs + out keeps it testable.
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
// idempotently. This is the module-include verb of weave's repo-composition
// dialect: how a fresh derivative declares its dependency on a real ariadne
// checkout anywhere on disk (the plan's "directory-agnostic substrate paths"
// Revisions entry) — recording the real path, not a hardcoded ../ariadne.
func buildLink() *cobra.Command {
	return &cobra.Command{
		Use:           "link <path>",
		Short:         "Record `substrate <path>` (verbatim) in this repo's construct/deps",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			return runLink(weavefs.OSFS{}, root, args[0], cmd.OutOrStdout())
		},
	}
}

// runLink appends `substrate <path>` to root/construct/deps, recording path
// VERBATIM (no resolution/relativization — the establishment verb captures the
// real path it was handed). Idempotent: it reuses layer.ParseDeps (the same
// grammar the walk + Apply read deps with, ARCH-DRY) to skip when the row is
// already present, and creates construct/deps (+ construct/) when absent.
// Injecting fs + out keeps it testable. Read-only on everything but the one deps
// file.
func runLink(fs weavefs.FS, root, path string, out io.Writer) error {
	depsPath := filepath.Join(root, "construct", "deps")

	var existing string
	if data, rerr := fs.ReadFile(depsPath); rerr == nil {
		existing = string(data)
		rows, perr := layer.ParseDeps(existing)
		if perr != nil {
			return fmt.Errorf("link: parse %s: %w", depsPath, perr)
		}
		for _, r := range rows {
			if r == path {
				fmt.Fprintf(out, "weave: substrate %s already present in construct/deps\n", path)
				return nil
			}
		}
	}

	if err := fs.MkdirAll(filepath.Dir(depsPath)); err != nil {
		return fmt.Errorf("link: mkdir %s: %w", filepath.Dir(depsPath), err)
	}
	next := existing
	if next != "" && !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	next += "substrate " + path + "\n"
	if err := fs.WriteFile(depsPath, []byte(next)); err != nil {
		return fmt.Errorf("link: write %s: %w", depsPath, err)
	}
	fmt.Fprintf(out, "weave: declared substrate %s in construct/deps\n", path)
	return nil
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
	cmd.Flags().StringVar(&targetFlag, "target", string(plan.TargetClaude), "skill backend target: claude | codex | agy")
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
		// the SAME planActions the compile path uses (ARCH-DRY) so the .claude/skills
		// symlinks the symlink backend now emits are classified against the live
		// links — they MATCH the ones sync-local-skills.sh wrote, the M5 cutover's
		// parity check. The AGENTS.md WriteFile (prose + the menu backend's `## Skills`
		// section) is classified as one body; its content intentionally DIVERGES from
		// setup.sh's symlinked AGENTS.md (an expected, hand-checked M5 divergence).
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

// run is the compile pipeline: walk → Plan → (Apply | print). Injecting fs +
// out keeps it testable against a t.TempDir-rooted OSFS and a buffer.
//
// root is canonicalized to its physical form up front (filepath.EvalSymlinks ≈
// pwd -P) so it lives in the SAME namespace as the layer Paths the walk
// canonicalizes — without this, on macOS (/tmp → /private/tmp) Apply would
// compute a relative symlink target between a logical dst-dir and a physical
// upstream src that resolves wrong when the OS follows the link (the exact bug
// setup.sh's pwd -P guards against, lines 39-45).
func run(fs weavefs.FS, root string, target plan.Target, dryRun bool, out io.Writer) error {
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	layers, err := walk.Walk(fs, root)
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	actions, err := planActions(fs, layers, target)
	if err != nil {
		return err
	}
	// The lowering source roots for the orphan-symlink prune (#96): the resolved
	// layer roots weave lowers FROM (a weave-owned link's target resolves under
	// one of these). Derived from the walk, never hardcoded.
	sourceRoots := plan.SourceRootsFromPaths(layerPaths(layers))
	if dryRun {
		fmt.Fprint(out, formatActions(actions))
		// Dry-run also previews the prune (read-only scan + the SAME pure decision
		// the apply uses), so a dry-run shows exactly what an apply would delete.
		preview, perr := plan.PrunePreview(fs, root, actions, sourceRoots)
		if perr != nil {
			return fmt.Errorf("prune preview: %w", perr)
		}
		fmt.Fprint(out, formatPrunes(preview))
		return nil
	}
	if err := plan.Apply(fs, root, actions); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	// After Apply, prune ORPHANED lowered symlinks weave no longer produces (#96):
	// renamed/re-prefixed skills + the #95 cutover's dead symlinks. Safety lives in
	// plan.shouldPrune — only a weave-owned symlink absent from this run's produced
	// set, in a managed location, is removed; real files/dirs and non-weave links
	// are never touched.
	pruned, err := plan.PruneOrphans(fs, root, actions, sourceRoots)
	if err != nil {
		return fmt.Errorf("prune: %w", err)
	}
	fmt.Fprintf(out, "weave: applied %d action(s) to %s\n", len(actions), root)
	if len(pruned) > 0 {
		fmt.Fprintf(out, "weave: pruned %d orphaned lowered symlink(s)\n", len(pruned))
		for _, p := range pruned {
			fmt.Fprintf(out, "  pruned %s\n", p)
		}
	}
	return nil
}

// layerPaths extracts the absolute on-disk root of each resolved layer — the
// lowering source roots the prune's weave-owned check tests target containment
// against. Pure.
func layerPaths(layers []layer.Layer) []string {
	out := make([]string, 0, len(layers))
	for _, l := range layers {
		out = append(out, l.Path)
	}
	return out
}

// planActions is the full compile lowering for a set of resolved layers, for ONE
// backend target (Approach-1, M5). The skill intents lower through exactly one of
// the two backends — they are MUTUALLY EXCLUSIVE per target (a harness reads its
// skills one way):
//
//   - target.EmitSkillSymlinks() (claude): emit the .claude/skills/<name> links;
//     the AGENTS.md menu is suppressed (nil menu ⇒ composeAgentsBody yields
//     prose-only, no `## Skills` section).
//   - target.IncludeSkillMenu() (codex/agy): NO .claude/skills links; the
//     `## Skills` menu is composed into AGENTS.md instead.
//
// Every other file-op (prose body, settings merge, scaffold, touch,
// generic symlink, seed) is target-independent. Shared by the compile path (run),
// the golden harness (runGolden), and verify-complete (runVerifyComplete) so all
// see the IDENTICAL action set for a given target (ARCH-DRY).
func planActions(fs weavefs.FS, layers []layer.Layer, target plan.Target) ([]plan.Action, error) {
	// ONE skill discovery (#104): gather → SelectVisible (𝒜(R)). The menu (idx) AND
	// the claude symlinks both read the SAME selected entries — no second scan.
	idx, selected, err := buildSkillIndex(fs, layers)
	if err != nil {
		return nil, fmt.Errorf("gather skills: %w", err)
	}
	// Menu backend (codex/agy only): the `## Skills` section composed into
	// AGENTS.md by the pure planner. A nil menu (claude) ⇒ prose-only AGENTS.md.
	var menu []skill.MenuItem
	if target.IncludeSkillMenu() {
		menu = idx.Menu()
	}
	actions, err := plan.Plan(layers, menu)
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}
	// Symlink backend (claude only): the .claude/skills/<name> links the Claude
	// harness reads. plan.SkillSymlinks is a PURE derivation from the SAME selected
	// entries the menu used (#104 — no separate walk.LowerSkillSymlinks scan).
	if target.EmitSkillSymlinks() {
		for _, l := range plan.SkillSymlinks(selected) {
			actions = append(actions, l)
		}
	}
	// weave OWNS ignoring its own generated-runtime artifacts (gitignore.go): the
	// composed AGENTS.md, the .claude/skills symlinks, the merged
	// .claude/settings.json, the .colima VM tree, vm-log.sh. Append exactly ONE
	// EnsureGitignore per compile (target-independent — every backend produces
	// generated artifacts) so a fresh `weave compile` on ANY derivative leaves a
	// clean `git status` with no per-repo .gitignore hand-edit. The pure planner
	// (plan.Plan) stays free of it — like skillSymlinks, it's appended in this
	// compile lowering and applied through the IO seam (plan.applyEnsureGitignore).
	actions = append(actions, plan.EnsureGitignore{Entries: plan.GeneratedRuntimeGitignoreEntries})
	return actions, nil
}

// buildSkillIndex is weave's SINGLE skill pipeline (#104): walk.GatherSkills (the
// one IO discovery, intent-driven, carrying Visibility+LayerIndex) → skill.SelectVisible
// (the pure 𝒜(R) filter; leaf = the last layer) → skill.Build (the pure menu + body
// lookup). It returns BOTH the index AND the selected entries, so the menu and the
// claude symlinks (plan.SkillSymlinks) lower from the IDENTICAL set — the §A4
// unification (one scan, two renderings, ARCH-DRY). layers must already be walked.
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
		case plan.Seed:
			b = append(b, fmt.Sprintf("seed      %s -> %s\n", act.Dst, act.Src)...)
		case plan.Touch:
			b = append(b, fmt.Sprintf("touch     %s\n", act.Path)...)
		case plan.MergeSettings:
			b = append(b, fmt.Sprintf("merge     %s -> %s\n", act.Source, act.Target)...)
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

// changecode.go — `sdlc change-code --issue N` subcommand.
//
// The planning → implementation transition. Composes four gates:
//
//  1. Structural sanity   — deterministic checks against the issue file
//     (Spec ≥ 50 words, non-empty Plan, etc.).
//  2. Estimate gate       — a positive estimate_hours: (#113), relocated
//     from claim; --no-estimate bypasses.
//  3. Plan-quality judge  — fresh-context LLM review: is this plan
//     executable as-written?
//  4. Branching strategy  — default in-place (#51); --worktree=yes for a
//     worktree, --worktree=ask to be prompted.
//
// Any gate can be skipped with the corresponding --no-* flag, or
// bypassed wholesale with --force <reason>. The --force rationale
// is recorded on stderr so the audit trail captures *why* the gate
// was bypassed.
//
// Branching default (#51): an unset --worktree means in-place — a branch
// in the current checkout, the common case, chosen without nagging.
// --worktree=ask reaches the interactive prompt, or for a non-tty agent
// emits the sizing hint + the ASK_BRANCHING_STRATEGY sentinel on stdout
// and exits 2 (the xx-sdlc skill, #39, turns that into an AskUserQuestion
// and re-invokes with --worktree=yes|no).
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

type changeCodeFlags struct {
	Issue           int
	Name            string
	IssuesDir       string
	PlansDir        string
	Worktree        string // "yes" | "no" | "" (ask)
	Force           string // rationale; non-empty bypasses gate refusal
	NoJudge         bool
	NoStructural    bool
	NoEstimate      bool
	NoEstimateRecon bool
	DryRun          bool
	Agent           string
	AgentExplicit   bool
	Sandbox         bool
}

func NewChangeCodeCmd() *cobra.Command {
	f := changeCodeFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "change-code",
		Short:         "Enter implementation phase (structural + plan-quality gates + branching ask)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"change-code\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.AgentExplicit = cmd.Flags().Changed("agent")
			return runChangeCode(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "ariadne workshop issue ID (derives name from issues/NNNNNN-*.md)")
	cmd.Flags().StringVar(&f.Name, "name", "", "explicit branch name (overrides --issue derivation)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding optional separate plan files")
	cmd.Flags().StringVar(&f.Worktree, "worktree", "", "branching: yes (worktree) | no (in-place) | ask (prompt). Empty = in-place (default).")
	cmd.Flags().StringVar(&f.Force, "force", "", "bypass gate refusals; the value is the rationale (recorded on stderr)")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the plan-quality LLM judge")
	cmd.Flags().BoolVar(&f.NoStructural, "no-structural", false, "skip the structural-sanity checks")
	cmd.Flags().BoolVar(&f.NoEstimate, "no-estimate", false, "skip the estimate_hours gate (#113)")
	cmd.Flags().BoolVar(&f.NoEstimateRecon, "no-estimate-recon", false, "skip the ## Estimate reconciliation gate (#117)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be operations; do nothing")
	cmd.Flags().StringVar(&f.Agent, "agent", "", "agent CLI for plan-quality judge: claude | codex | gemini (default AGENT_CMD, PAIR_AGENT/current agent, or claude)")
	cmd.Flags().BoolVar(&f.Sandbox, "sandbox", isSandbox(), "pass auto-approve flags to codex/gemini")
	return cmd
}

// changeCodeRunner is the test seam shared with other verbs.
var changeCodeRunner gitRunner = execGitRunner{}

// askExitCode is the exit status used when sdlc change-code defers to
// the agent for the branching decision. The xx-sdlc skill keys off
// this code + the ASK_BRANCHING_STRATEGY stdout sentinel.
const askExitCode = 2

// ASK_BRANCHING_STRATEGY is emitted on stdout when stdin is not a tty
// and --worktree is unset. Stable contract — agents grep for it.
const sentinelBranchingStrategy = "ASK_BRANCHING_STRATEGY"

func runChangeCode(stdin io.Reader, stdout, stderr io.Writer, f *changeCodeFlags) error {
	// 1. Resolve issue file path + branch name.
	name, untrackedFile, issuePath, err := resolveChangeCodeName(f, changeCodeRunner)
	if err != nil {
		die(stderr, err.Error())
	}

	// 2. Read issue content (and optional plan file).
	issueBytes, err := os.ReadFile(issuePath)
	if err != nil {
		die(stderr, fmt.Sprintf("read issue file %s: %v", issuePath, err))
	}
	issueContent := string(issueBytes)

	planContent := readOptionalPlanFile(f.PlansDir, name)

	// 3. Structural gates.
	if !f.NoStructural {
		failures := issue.CheckStructural(issueContent)
		if len(failures) > 0 {
			if f.Force == "" {
				fmt.Fprintln(stderr, "structural-sanity gates failed:")
				for _, fail := range failures {
					fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
				}
				cwarn(stderr, "fix the failures above, OR re-run with --force <reason>")
				exitWithCode(1)
			}
			cwarn(stderr, fmt.Sprintf("structural gates bypassed (--force: %s)", f.Force))
			for _, fail := range failures {
				fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
			}
		}
	}

	// 3b. Estimate gate (#113). The universal estimate requirement, relocated
	//     here from `sdlc claim` so claiming stays a cheap early lock. Its own
	//     --no-estimate bypass (per the per-gate --no-<gate> convention),
	//     reusing the pure issue.CheckEstimate split out of CheckStructural.
	if fail := estimateRefusal(issueContent, f.NoEstimate); fail != nil {
		if f.Force == "" {
			fmt.Fprintln(stderr, "estimate gate failed:")
			fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
			cwarn(stderr, "add `estimate_hours: <n>` to the issue frontmatter (set it at start-plan), OR re-run with --no-estimate / --force <reason>")
			exitWithCode(1)
		}
		cwarn(stderr, fmt.Sprintf("estimate gate bypassed (--force: %s)", f.Force))
		fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
	}

	// 3c. Estimate-reconciliation gate (#117). Forces the documented model to be
	//     applied: estimate_hours must reconcile with an itemized ## Estimate
	//     block (no unitemized estimate). Pure decision (estimateReconRefusal);
	//     own --no-estimate-recon bypass per the per-gate convention.
	if fail := estimateReconRefusal(issueContent, f.NoEstimateRecon); fail != nil {
		if f.Force == "" {
			fmt.Fprintln(stderr, "estimate-reconciliation gate failed:")
			fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
			cwarn(stderr, "fix the ## Estimate block so it reconciles, OR re-run with --no-estimate-recon / --force <reason>")
			exitWithCode(1)
		}
		cwarn(stderr, fmt.Sprintf("estimate-reconciliation gate bypassed (--force: %s)", f.Force))
		fmt.Fprintf(stderr, "  [%s] %s\n", fail.Name, fail.Message)
	}

	// 4. Plan-quality + estimate-quality judges (fresh-context LLM).
	if !f.NoJudge {
		if err := runPlanQualityJudge(stdout, stderr, f, name, issueContent, planContent); err != nil {
			// runPlanQualityJudge already printed; honor --force.
			if f.Force == "" {
				exitWithCode(1)
			}
			cwarn(stderr, fmt.Sprintf("plan-quality gate bypassed (--force: %s)", f.Force))
		}
		if err := runEstimateQualityJudge(stdout, stderr, f, name, issueContent); err != nil {
			if f.Force == "" {
				exitWithCode(1)
			}
			cwarn(stderr, fmt.Sprintf("estimate-quality gate bypassed (--force: %s)", f.Force))
		}
	}

	// 5. Branching strategy.
	wt, err := resolveBranchingStrategy(stdin, stdout, stderr, f, name, issueContent)
	if err != nil {
		die(stderr, err.Error())
	}

	// 6. Dry-run pre-empts side effects (apart from the gates already
	//    run above, which are read-only).
	if f.DryRun {
		cinfo(stderr, "dry-run — branch creation skipped")
		if untrackedFile != "" {
			fmt.Fprintf(stdout, "Would commit + push: %s\n", untrackedFile)
		}
		fmt.Fprintf(stdout, "Would create branch %s (mode=%s)\n", name, wt)
		return nil
	}

	// 7. Commit any untracked issue-file changes so the branch starts
	//    clean (mirrors start.go's pre-#39 behavior).
	if err := commitUntrackedIssueFile(stderr, untrackedFile, changeCodeRunner); err != nil {
		die(stderr, err.Error())
	}

	// 8. Create branch.
	switch wt {
	case "yes":
		if _, err := createWorktreeBranch(stdout, stderr, name, changeCodeRunner); err != nil {
			die(stderr, err.Error())
		}
	case "no":
		if _, err := createInPlaceBranch(stdout, stderr, name, changeCodeRunner); err != nil {
			die(stderr, err.Error())
		}
	}
	return nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

// estimateRefusal is the pure decision for change-code's estimate gate (#113):
// nil when the gate is skipped (--no-estimate) or estimate_hours is a positive
// number; the estimate-present failure otherwise. The os.Exit / --force
// handling stays in runChangeCode (the IO shell); this keeps the gate decision
// unit-testable without spawning the command (ARCH-PURE).
func estimateRefusal(issueContent string, noEstimate bool) *issue.StructuralFailure {
	if noEstimate {
		return nil
	}
	return issue.CheckEstimate(issueContent)
}

// estimateReconRefusal is the pure decision for change-code's estimate-
// reconciliation gate (#117): nil when skipped (--no-estimate-recon) or when the
// issue's `## Estimate` block parses and reconciles with frontmatter
// estimate_hours; otherwise a single aggregated failure. All logic is pure
// (issue + estimate parsing); the os.Exit / --force handling stays in
// runChangeCode (ARCH-PURE), so the gate is unit-testable without the command.
func estimateReconRefusal(issueContent string, noRecon bool) *issue.StructuralFailure {
	if noRecon {
		return nil
	}
	fm, body, err := issue.Parse(issueContent)
	if err != nil {
		return &issue.StructuralFailure{Name: "estimate-recon", Message: "issue file has no YAML frontmatter to reconcile estimate_hours against"}
	}
	section, ok := issue.EstimateSection(body)
	if !ok {
		return &issue.StructuralFailure{Name: "estimate-recon", Message: "no `## Estimate` block — derive estimate_hours against the calibrated source (run `sdlc estimate-source` to see the shared method + your repo's calibration doc) and add a fenced ```estimate block (grammar in `sdlc change-code --help`)"}
	}
	block, err := estimate.ParseBlock(section)
	if err != nil {
		return &issue.StructuralFailure{Name: "estimate-recon", Message: "## Estimate block does not parse: " + err.Error()}
	}
	estHoursStr, _ := issue.GetField(fm, "estimate_hours")
	estHours, _ := strconv.ParseFloat(strings.TrimSpace(estHoursStr), 64)
	failures := estimate.Check(block, estHours)
	if len(failures) == 0 {
		return nil
	}
	msgs := make([]string, len(failures))
	for i, ff := range failures {
		msgs[i] = ff.Message
	}
	return &issue.StructuralFailure{Name: "estimate-recon", Message: strings.Join(msgs, "; ")}
}

// resolveChangeCodeName reuses start.go's name-resolution shape but
// also returns the path to the resolved issue file so the gates can
// read it. The untrackedFile return is non-empty when the issue file
// is still untracked and needs to be committed before branch creation.
func resolveChangeCodeName(f *changeCodeFlags, r gitRunner) (name, untrackedFile, issuePath string, err error) {
	nf := &nameFlags{
		Issue:     f.Issue,
		Name:      f.Name,
		IssuesDir: f.IssuesDir,
	}
	name, untrackedFile, err = resolveBranchName(nf, r)
	if err != nil {
		return "", "", "", err
	}

	// Locate the issue file. When --name was used (no --issue), we still
	// need the issue file for gates; derive it from the name prefix.
	if untrackedFile != "" {
		issuePath = untrackedFile
	} else {
		issuePath, err = findIssueFileByName(f.IssuesDir, name)
		if err != nil {
			return "", "", "", err
		}
	}
	return name, untrackedFile, issuePath, nil
}

// findIssueFileByName looks up the issue file for a resolved branch
// name (e.g. "000039-defer-worktree…"). The convention is that the
// branch name equals the issue filename stem.
func findIssueFileByName(issuesDir, name string) (string, error) {
	candidate := filepath.Join(issuesDir, name+".md")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	// Fall back to a glob in case the name has been altered post-claim.
	matches, _ := filepath.Glob(filepath.Join(issuesDir, "*.md"))
	for _, m := range matches {
		if strings.TrimSuffix(filepath.Base(m), ".md") == name {
			return m, nil
		}
	}
	return "", fmt.Errorf("could not locate issue file for name %q under %s", name, issuesDir)
}

// readOptionalPlanFile returns the contents of
// <plansDir>/<name>-plan.md if it exists, else "". Used by the
// plan-quality judge to consume detailed designs that live outside
// the issue file.
func readOptionalPlanFile(plansDir, name string) string {
	path := filepath.Join(plansDir, name+"-plan.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// runPlanQualityJudge dispatches the plan-quality judge against the
// issue + plan content and returns nil on Clean/Info, error on Failure.
// Output is surfaced on stdout exactly as judge.go's dispatch does.
func runPlanQualityJudge(stdout, stderr io.Writer, f *changeCodeFlags, name, issueContent, planContent string) error {
	issueRef := name
	if f.Issue > 0 {
		issueRef = fmt.Sprintf("ariadne#%d", f.Issue)
	}

	prompt := judge.BuildPrompt(judge.PlanQuality, judge.PromptInput{
		IssueRef:     issueRef,
		IssueContent: issueContent,
		PlanContent:  planContent,
	})

	agent := judge.ResolveAgentCLI(f.Agent, f.AgentExplicit, judge.CurrentAgentDefaultEnv())
	tools := judge.PlanQuality.AllowedTools()
	opts := judge.DispatchOptions{
		Agent:        agent,
		Prompt:       prompt,
		AllowedTools: tools,
		IsSandbox:    f.Sandbox,
		Stdout:       stdout,
		Stderr:       stderr,
	}

	if f.DryRun {
		cmdLine, err := judge.FormatCommandLine(opts)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, "── plan-quality prompt ──")
		fmt.Fprintln(stdout, prompt)
		fmt.Fprintln(stdout, "── command (would invoke) ──")
		fmt.Fprintln(stdout, cmdLine)
		return nil
	}

	cinfo(stderr, fmt.Sprintf("invoking %s for plan-quality check …", agent))
	output, dispatchErr := judge.Dispatch(context.Background(), opts)
	if dispatchErr != nil {
		return fmt.Errorf("plan-quality dispatch failed: %v", dispatchErr)
	}

	fmt.Fprint(stdout, output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Fprintln(stdout)
	}

	outcome := judge.Classify(output)
	switch outcome {
	case judge.Clean:
		cok(stderr, "plan-quality: clean")
		return nil
	case judge.Info:
		cinfo(stderr, "plan-quality: info")
		return nil
	case judge.Failure:
		cwarn(stderr, "plan-quality: findings reported — fix the plan, OR re-run with --force <reason>")
		return fmt.Errorf("plan-quality failure")
	}
	return nil
}

// runEstimateQualityJudge dispatches the estimate-quality judge against the issue
// content and returns nil on Clean/Info, error on Failure. Mirrors
// runPlanQualityJudge. Skips silently when there is no `## Estimate` block — the
// reconciliation gate (3c) owns the "block required" enforcement, so with that
// gate skipped (--no-estimate-recon) there is simply nothing to judge.
func runEstimateQualityJudge(stdout, stderr io.Writer, f *changeCodeFlags, name, issueContent string) error {
	if _, body, err := issue.Parse(issueContent); err != nil {
		return nil
	} else if _, ok := issue.EstimateSection(body); !ok {
		return nil
	}

	issueRef := name
	if f.Issue > 0 {
		issueRef = fmt.Sprintf("ariadne#%d", f.Issue)
	}

	prompt := judge.BuildPrompt(judge.EstimateQuality, judge.PromptInput{
		IssueRef:     issueRef,
		IssueContent: issueContent,
	})

	agent := judge.ResolveAgentCLI(f.Agent, f.AgentExplicit, judge.CurrentAgentDefaultEnv())
	opts := judge.DispatchOptions{
		Agent:        agent,
		Prompt:       prompt,
		AllowedTools: judge.EstimateQuality.AllowedTools(),
		IsSandbox:    f.Sandbox,
		Stdout:       stdout,
		Stderr:       stderr,
	}

	if f.DryRun {
		cmdLine, err := judge.FormatCommandLine(opts)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, "── estimate-quality prompt ──")
		fmt.Fprintln(stdout, prompt)
		fmt.Fprintln(stdout, "── command (would invoke) ──")
		fmt.Fprintln(stdout, cmdLine)
		return nil
	}

	cinfo(stderr, fmt.Sprintf("invoking %s for estimate-quality check …", agent))
	output, dispatchErr := judge.Dispatch(context.Background(), opts)
	if dispatchErr != nil {
		return fmt.Errorf("estimate-quality dispatch failed: %v", dispatchErr)
	}

	fmt.Fprint(stdout, output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Fprintln(stdout)
	}

	switch judge.Classify(output) {
	case judge.Clean:
		cok(stderr, "estimate-quality: clean")
		return nil
	case judge.Info:
		cinfo(stderr, "estimate-quality: info")
		return nil
	case judge.Failure:
		cwarn(stderr, "estimate-quality: findings reported — fix the ## Estimate, OR re-run with --no-judge / --force <reason>")
		return fmt.Errorf("estimate-quality failure")
	}
	return nil
}

// resolveBranchingStrategy returns "yes" or "no" for worktree vs
// in-place. Honors --worktree= when set; otherwise asks (tty prompt
// or sentinel-and-exit-2 for agents).
func resolveBranchingStrategy(stdin io.Reader, stdout, stderr io.Writer, f *changeCodeFlags, name, issueContent string) (string, error) {
	switch f.Worktree {
	case "yes", "no":
		return f.Worktree, nil
	case "":
		// Default (ariadne #51): in-place branch, silently. Worktree is
		// opt-in via --worktree=yes; --worktree=ask reaches the interactive
		// prompt / agent sentinel below. A default shouldn't nag, so an unset
		// flag no longer asks.
		cinfo(stderr, "branching: in-place (default; --worktree=yes for an isolated worktree)")
		return "no", nil
	case "ask":
		// fall through to the interactive prompt / agent sentinel
	default:
		return "", fmt.Errorf("--worktree must be 'yes', 'no', or 'ask' (got %q)", f.Worktree)
	}

	// Build the sizing hint either way; both branches print it.
	sizing := issue.ComputeSizingFromContent(issueContent)
	title := issueTitleFromContent(issueContent)
	id := strings.SplitN(name, "-", 2)[0]
	fmt.Fprint(stderr, sizing.Format(id, title))

	if isTTY(stdin) {
		return promptBranchingTTY(stdin, stderr)
	}

	// Non-tty: emit sentinel and exit 2. xx-sdlc skill handles the rest.
	fmt.Fprintln(stdout, sentinelBranchingStrategy)
	cinfo(stderr, "deferring branching decision to operator via agent (exit 2)")
	exitWithCode(askExitCode)
	return "", nil // unreachable
}

func promptBranchingTTY(stdin io.Reader, stderr io.Writer) (string, error) {
	fmt.Fprint(stderr, "\nBranching: [w]orktree / [m]ain in-place / [c]ancel > ")
	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read branching choice: %v", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "w", "worktree", "y", "yes":
		return "yes", nil
	case "m", "main", "n", "no", "in-place":
		return "no", nil
	case "c", "cancel", "":
		return "", fmt.Errorf("cancelled by operator")
	default:
		return "", fmt.Errorf("unrecognized choice %q (want w/m/c)", line)
	}
}

// isTTY reports whether the given reader is a terminal. Returns false
// for non-os.File readers (test pipes) so tests deterministically
// hit the agent-protocol path. Stdlib-only — checks the file mode for
// os.ModeCharDevice rather than pulling in golang.org/x/term.
func isTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// issueTitleFromContent extracts the first H1 heading from the issue
// body, fallback to "(no title)". Used to label the sizing hint.
func issueTitleFromContent(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return "(no title)"
}

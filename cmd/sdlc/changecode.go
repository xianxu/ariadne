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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
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
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
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
	guardIssueNotDone(stderr, issuePath, strconv.Itoa(f.Issue)) // #176 done-issue guard

	// 2. Read issue content (and optional plan file).
	issueBytes, err := os.ReadFile(issuePath)
	if err != nil {
		die(stderr, fmt.Sprintf("read issue file %s: %v", issuePath, err))
	}
	issueContent := string(issueBytes)

	planContent := readOptionalPlanFile(f.PlansDir, name)

	// 3. Run the gate sequence. RUNNING the declaration (rather than hand-sequencing
	//    blocks that happen to match it) is what makes changeCodeGateOrder a real guard:
	//    reordering the literal reorders execution, so the B1 ordering test fails on the
	//    regression it exists to catch instead of passing on a restatement (ARCH-DRY).
	ctx := &changeCodeCtx{
		f: f, stdout: stdout, stderr: stderr,
		name: name, issuePath: issuePath,
		issueContent: issueContent, planContent: planContent,
	}
	for _, g := range changeCodeGates(ctx) {
		if err := g.run(); err != nil {
			// Each gate has already printed its specifics; this is the one shared
			// --force decision, previously copy-pasted across five blocks.
			if f.Force == "" {
				exitWithCode(1)
			}
			cwarn(stderr, fmt.Sprintf("%s gate bypassed (--force: %s)", g.name, f.Force))
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

// ── the gate sequence ───────────────────────────────────────────────────────

// changeCodeCtx carries everything the gate closures need. Bundled so
// changeCodeGates can be a plain data declaration rather than a function with a
// nine-parameter signature.
type changeCodeCtx struct {
	f              *changeCodeFlags
	stdout, stderr io.Writer
	name           string
	issuePath      string
	issueContent   string
	planContent    string
}

// gate is one named step in change-code's sequence. Declaring the gates as data and
// RUNNING that data is the point: the list IS the order.
type gate struct {
	name string
	run  func() error // nil = passed; each closure owns its own --no-<gate> skip
}

// changeCodeGates is the ordered gate sequence.
//
// #187 B1: plan-quality precedes EVERY estimate gate, because the estimate is a function
// of the plan. Demanding one before the plan has been looked at once forces a
// re-derivation per plan revision — pair#127 re-derived its estimate five times, four of
// them forced by plan changes that were still in flight. Costing an unapproved plan is
// waste by construction.
//
// Structural stays first because it is free (no subprocess).
func changeCodeGates(c *changeCodeCtx) []gate {
	return []gate{
		{"structural", c.structural},
		{"plan-quality", c.planQuality},
		{"estimate", c.estimate},
		{"estimate-recon", c.estimateRecon},
		{"estimate-quality", c.estimateQuality},
	}
}

// changeCodeGateOrder returns the gate names in execution order — derived from the same
// declaration the runner iterates, so the ordering test cannot pass on a stale copy.
func changeCodeGateOrder() []string {
	names := []string{}
	for _, g := range changeCodeGates(&changeCodeCtx{f: &changeCodeFlags{}}) {
		names = append(names, g.name)
	}
	return names
}

func (c *changeCodeCtx) structural() error {
	if c.f.NoStructural {
		return nil
	}
	failures := issue.CheckStructural(c.issueContent)
	if len(failures) == 0 {
		return nil
	}
	fmt.Fprintln(c.stderr, "structural-sanity gates failed:")
	for _, fail := range failures {
		fmt.Fprintf(c.stderr, "  [%s] %s\n", fail.Name, fail.Message)
	}
	cwarn(c.stderr, "fix the failures above, OR re-run with --force <reason>")
	return fmt.Errorf("structural failure")
}

// estimate is the universal estimate_hours requirement (#113), relocated here from
// `sdlc claim` so claiming stays a cheap early lock — and, since #187 B1, sequenced
// AFTER plan-quality so the estimate is only ever asked for against an accepted plan.
func (c *changeCodeCtx) estimate() error {
	fail := estimateRefusal(c.issueContent, c.f.NoEstimate)
	if fail == nil {
		return nil
	}
	fmt.Fprintln(c.stderr, "estimate gate failed:")
	fmt.Fprintf(c.stderr, "  [%s] %s\n", fail.Name, fail.Message)
	cwarn(c.stderr, "the plan has cleared plan-quality — derive `estimate_hours: <n>` now and add it to the issue frontmatter, OR re-run with --no-estimate / --force <reason>")
	return fmt.Errorf("estimate failure")
}

// estimateRecon forces the documented model to be applied (#117): estimate_hours must
// reconcile with an itemized ## Estimate block.
func (c *changeCodeCtx) estimateRecon() error {
	fail := estimateReconRefusal(c.issueContent, c.f.NoEstimateRecon)
	if fail == nil {
		return nil
	}
	fmt.Fprintln(c.stderr, "estimate-reconciliation gate failed:")
	fmt.Fprintf(c.stderr, "  [%s] %s\n", fail.Name, fail.Message)
	cwarn(c.stderr, "fix the ## Estimate block so it reconciles, OR re-run with --no-estimate-recon / --force <reason>")
	return fmt.Errorf("estimate-recon failure")
}

func (c *changeCodeCtx) planQuality() error {
	if c.f.NoJudge {
		return nil
	}
	return runPlanQualityJudge(c.stdout, c.stderr, c.f, c.name, c.issuePath, c.issueContent, c.planContent)
}

func (c *changeCodeCtx) estimateQuality() error {
	if c.f.NoJudge {
		return nil
	}
	return runEstimateQualityJudge(c.stdout, c.stderr, c.f, c.name, c.issueContent)
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

// runPlanQualityJudge dispatches the STATEFUL plan-quality judge (#187).
//
// Unlike the pre-#187 version — which built a prompt, printed, and forgot — this reads the
// gate's accumulated ledger, feeds the prior rounds back into the prompt, parses a schema'd
// findings block out of the reply, and computes the block/pass decision as a pure function
// of the accumulated state rather than reading it off the judge's verdict token. That is
// what lets the gate converge: a fresh Minor no longer costs a round-trip, and disposed
// blockers open the gate.
//
// The thin IO seam over internal/gatestate (ARCH-PURE): this function owns the filesystem,
// the clock, and the subprocess; every decision is pure and unit-tested next door.
func runPlanQualityJudge(stdout, stderr io.Writer, f *changeCodeFlags, name, issuePath, issueContent, planContent string) error {
	issueRef := name
	if f.Issue > 0 {
		issueRef = fmt.Sprintf("ariadne#%d", f.Issue)
	}
	issueFile := filepath.Base(issuePath)

	ledger, lerr := readPlanGateLedger(f.PlansDir, issueFile, f.Issue)
	if lerr != nil {
		// A corrupt ledger must HALT, not silently forget — see planreview.go.
		cwarn(stderr, "plan-gate ledger: "+lerr.Error())
		return fmt.Errorf("plan-gate ledger unreadable")
	}

	// Pass-through (#187): unchanged content after a passing round ⇒ no dispatch, no
	// round. This is what makes B1 a net win rather than a net cost. Without it, moving
	// the estimate gates below plan-quality would make EVERY estimate-gate failure pay a
	// fresh multi-minute judge dispatch on the retry — and since B2 tells the agent to
	// derive the estimate only after the plan clears, that retry is guaranteed on every
	// issue. Same mechanism #183 wants at the close boundary (ARCH-DRY).
	contentHash := gatestate.ContentHash(issueContent, planContent)
	if !f.DryRun && gatestate.PassesUnchanged(ledger, contentHash) {
		cok(stderr, fmt.Sprintf("plan-quality: unchanged since round %d — passing through (no re-dispatch)", len(ledger.Rounds)))
		return nil
	}

	prompt := judge.BuildPrompt(judge.PlanQuality, judge.PromptInput{
		IssueRef:      issueRef,
		IssueContent:  issueContent,
		PlanContent:   planContent,
		PriorFindings: gatestate.RenderPriorFindings(ledger),
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

	n := len(ledger.Rounds) + 1
	rr, ok := gatestate.ParseFindingsBlock(output)
	if !ok {
		// Transitional prose fallback (agent-binary-handoff-schema: the schema'd path is
		// authoritative; a fallback may exist transitionally). Warn loudly — this round
		// carries no findings, so the gate cannot converge on it.
		//
		// The round is STILL persisted. If a protocol miss returned early, len(Rounds)
		// would stay 0 forever for an agent CLI that never emits the fence: the prompt
		// would announce "this is the FIRST round" on invocation six, Decide's round cap
		// could never fire to bound the loop it exists to bound, and the close-time
		// gate_rounds metric would report 0 for precisely the most expensive sessions.
		cwarn(stderr, "plan-quality: no valid ```findings block — falling back to the verdict token; this round carries NO findings, so the gate cannot converge on it")
		blocked := classifyFallback(stderr, output) != nil
		persistPlanGateRound(stderr, f, issueFile, ledger, gatestate.Round{
			N: n, Timestamp: nowRFC3339(), Agent: string(agent),
			ProtocolError: "no valid findings block",
			Blocked:       blocked,
			Forced:        forcedRationale(f.Force, blocked),
		})
		if blocked {
			return fmt.Errorf("plan-quality failure")
		}
		return nil
	}

	round := gatestate.AssignIDs(ledger, rr, n, nowRFC3339(), string(agent))
	applied, aerr := gatestate.ApplyChecked(ledger, round)
	if aerr != nil {
		// Same reasoning: a protocol error is a round that happened and cost latency.
		cwarn(stderr, "plan-quality: "+aerr.Error())
		persistPlanGateRound(stderr, f, issueFile, ledger, gatestate.Round{
			N: n, Timestamp: nowRFC3339(), Agent: string(agent),
			ProtocolError: aerr.Error(), Blocked: true,
			Forced: forcedRationale(f.Force, true),
		})
		return fmt.Errorf("plan-quality protocol error: %v", aerr)
	}
	ledger = applied

	d := gatestate.Decide(ledger, roundCapFromEnv())
	last := len(ledger.Rounds) - 1
	ledger.Rounds[last].Blocked = d.Block
	ledger.Rounds[last].Forced = forcedRationale(f.Force, d.Block)
	// Stamp the pass-through key only on a PASSING round — caching a refusal would let a
	// still-blocked plan walk through unchanged on the next invocation.
	if !d.Block {
		ledger.ContentHash = contentHash
	}
	if werr := writePlanGateLedger(f.PlansDir, issueFile, ledger, repoIdentity()); werr != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate ledger not persisted: %v", werr))
	}

	if d.Block {
		cwarn(stderr, "plan-quality: "+d.Reason)
		cwarn(stderr, "address the findings above and re-run — the gate remembers what you fixed; OR re-run with --force <reason>")
		return fmt.Errorf("plan-quality failure")
	}
	cok(stderr, "plan-quality: "+d.Reason)
	return nil
}

// classifyFallback is the pre-#187 tri-state read of the judge's output — judge.Classify's
// switch, extracted verbatim so the schema'd path and the transitional prose path don't
// each carry a copy (ARCH-DRY). Returns nil on Clean/Info, an error on Failure.
func classifyFallback(stderr io.Writer, output string) error {
	switch judge.Classify(output) {
	case judge.Clean:
		cok(stderr, "plan-quality: clean (verdict token)")
		return nil
	case judge.Info:
		cinfo(stderr, "plan-quality: info (verdict token)")
		return nil
	case judge.Failure:
		cwarn(stderr, "plan-quality: findings reported — fix the plan, OR re-run with --force <reason>")
		return fmt.Errorf("plan-quality failure")
	}
	return nil
}

// forcedRationale returns the --force rationale only when this gate actually refused; ""
// otherwise. --force is a GLOBAL bypass consulted by every gate, so stamping it
// unconditionally would mark a plan-gate round "forced" when the operator forced past a
// structural failure — or even when the gate passed cleanly — over-reporting overrides in
// the one number whose whole purpose is to answer "which gates earn their cost".
func forcedRationale(force string, blocked bool) string {
	if blocked {
		return force
	}
	return ""
}

// persistPlanGateRound appends one round and writes the ledger, warning (never failing) on
// a write error. The single persistence call site for the protocol-miss exit paths.
func persistPlanGateRound(stderr io.Writer, f *changeCodeFlags, issueFile string, l gatestate.Ledger, r gatestate.Round) {
	if werr := writePlanGateLedger(f.PlansDir, issueFile, gatestate.Apply(l, r), repoIdentity()); werr != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate ledger not persisted: %v", werr))
	}
}

// roundCapFromEnv reads WF_PLAN_ROUND_CAP, defaulting to gatestate.DefaultRoundCap. Past
// the cap only hard-blocking findings refuse the gate; the rest are carried to the close
// review via the ledger.
func roundCapFromEnv() int {
	if v := os.Getenv("WF_PLAN_ROUND_CAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return gatestate.DefaultRoundCap
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

// isTTY reports whether the given reader is a terminal. Returns false for
// non-os.File readers (test pipes) so tests deterministically hit the
// agent-protocol path. The real terminal test is isTerminal (an ioctl probe,
// stdlib-only, per-OS in tty_*.go): it must NOT be a char-device check, because
// /dev/null — an agent's usual redirected stdin — is a char device but not a
// terminal, and mistaking it for one defeats merge's fail-fast confirm gate (#141).
func isTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(f.Fd())
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

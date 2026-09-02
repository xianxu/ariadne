// changecode.go — `sdlc change-code --issue N` subcommand.
//
// The planning → implementation transition. Composes a GATE SEQUENCE plus the branching
// decision. The sequence is data, not prose — `changeCodeGates` is the single source of the
// order, so this list documents it rather than defining it:
//
//  1. Structural sanity     — deterministic checks against the issue file
//     (Spec ≥ 50 words, non-empty Plan, etc.). Free: no subprocess.
//  2. Plan-quality judge    — fresh-context LLM review: is this plan executable
//     as-written? STATEFUL since #187 — it remembers its own prior findings
//     (see internal/gatestate) and passes through unchanged content.
//  3. Estimate gate         — a positive estimate_hours: (#113).
//  4. Estimate reconcile    — the frontmatter number vs the ## Estimate block.
//  5. Estimate-quality judge — is the estimate DERIVED rather than guessed?
//
// then the branching strategy: default in-place (#51); --worktree=yes for a worktree,
// --worktree=ask to be prompted.
//
// Plan-quality runs BEFORE the estimate gates (#187 B1): the estimate is derived once the
// design is settled, so costing an unapproved plan is work thrown away. The pass-through
// hash is what keeps that reorder from charging a fresh judge dispatch for every
// estimate-gate retry.
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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
	name, issuePath, err := resolveChangeCodeName(f, changeCodeRunner)
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

	// 4. Branching strategy (the gate sequence above is step 3).
	wt, err := resolveBranchingStrategy(stdin, stdout, stderr, f, name, issueContent)
	if err != nil {
		die(stderr, err.Error())
	}

	// 6. Dry-run pre-empts side effects (apart from the gates already
	//    run above, which are read-only).
	if f.DryRun {
		cinfo(stderr, "dry-run — branch creation skipped")
		if id := issueIDFromPath(issuePath); id > 0 {
			// Derive the wording from the SAME test syncIssue branches on, not
			// from a restatement of the outcome: from a branch it commits without
			// publishing, so promising a push here would be a lie in two of the
			// three locations.
			publishes := syncIssuePublishes()
			fmt.Fprintf(stdout, "Would %s issue #%d under %q\n",
				changeCodeSyncVerb(publishes), id,
				issueSyncMessage(id, "spec/plan at change-code"))
			if note := changeCodeSyncNote(publishes); note != "" {
				fmt.Fprintln(stdout, note)
			}
		}
		fmt.Fprintf(stdout, "Would create branch %s (mode=%s)\n", name, wt)
		return nil
	}

	// 7. Land the planning output: commit + publish the issue file so the
	//    branch starts from a tracked state (#206). This is the milestone the
	//    Spec/Plan should become external at — plan-quality has just accepted
	//    the design — so it publishes, unlike the bare `sdlc issue sync`.
	//
	//    Replaces commitUntrackedIssueFile, which was change-code's own private
	//    commit+push of the issue file: a second implementation of this, and one
	//    that only ever handled the UNTRACKED case, leaving a tracked-but-edited
	//    issue file dirty at branch creation.
	syncIssue(stderr, f, issuePath)

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

// syncIssuePublishes reports whether change-code's sync will publish: only from
// main, where "commit here" and "publish to main" coincide. One source for the
// decision and for the --dry-run line that describes it, so the two can't drift
// (the dry-run text used to promise a push unconditionally).
func syncIssuePublishes() bool {
	return gitx.Capture("branch", "--show-current") == "main"
}

// changeCodeSyncVerb names what the sync will do, for human output. Kept to a
// short verb phrase so it reads mid-sentence; the off-main caveat is a separate
// line (changeCodeSyncNote) rather than a parenthetical wedged between the verb
// and its object.
func changeCodeSyncVerb(publishes bool) string {
	if publishes {
		return "sync + push"
	}
	return "sync locally"
}

// changeCodeSyncNote is the follow-up line for the off-main case, or "" on main.
func changeCodeSyncNote(publishes bool) string {
	if publishes {
		return ""
	}
	return "    (not on main — publishing belongs to pr/merge/close)"
}

// syncIssue commits the issue file through the shared sync dispatch (#206), so
// the design that just cleared plan-quality is durable before any code is
// written.
//
// It replaced commitUntrackedIssueFile, and the two review rounds that followed
// were both the same mistake: a swapped helper covering fewer cells than the one
// it replaced. The invariant is therefore stated as a table, over the full
// cross-product the old helper ran under —
//
//	                        │ file untracked        │ file tracked + edited
//	────────────────────────┼───────────────────────┼──────────────────────────
//	on main                 │ commit here + publish │ commit here + publish
//	in-place feature branch │ commit on the branch  │ commit on the branch
//	feature worktree        │ commit on the branch  │ commit on the branch
//
// — crossed with resolveBranchName's three name modes (--issue, --name,
// auto-detect), which the id derivation below collapses: syncIssue reads the
// RESOLVED issuePath, the one thing all three modes produce. Gating on f.Issue
// instead skipped two of the three, and in auto-detect with --worktree=yes left
// the new worktree holding no issue file at all, since `git worktree add` does
// not carry untracked files. TestChangeCodeSyncIssue_ModeMatrix runs the table.
//
// The single property every cell satisfies: THE ISSUE FILE IS COMMITTED IN THIS
// WORKTREE, on the branch about to carry the work. That is what "the branch
// starts from a tracked state" means, and it is why publishing is conditioned on
// already being on main rather than on the caller's intent. From a branch the
// publish route would copy the in-progress body into the main worktree, commit
// it on main and push — putting a half-written Spec on origin/main, leaving the
// branch's own copy dirty, and adding two network round-trips to a milestone
// re-run. main gets the body at `pr`/`merge`/`close`, which is where publishing
// belongs.
//
// BEST-EFFORT, deliberately: change-code's job is to OPEN implementation, and a
// tracker commit that could not land must not stand between the operator and
// starting work. The helper this replaced already warned rather than died on a
// failed push, and the warning names the retry.
func syncIssue(stderr io.Writer, f *changeCodeFlags, issuePath string) {
	// A --name branch can point at a file outside the NNNNNN- convention; there
	// is no id to name in the commit subject, so there is nothing to sync.
	id := issueIDFromPath(issuePath)
	if id == 0 {
		return
	}
	onMain := syncIssuePublishes()
	// DryRun is threaded even though runChangeCode returns before reaching here
	// under --dry-run: a helper that commits must not depend on a caller's early
	// return for its dry-run correctness.
	syncFlags := &claimFlags{
		Issue: id, IssuesDir: f.IssuesDir, NoStart: true,
		DryRun: f.DryRun, NoPush: !onMain,
	}
	msg := issueSyncMessage(id, "spec/plan at change-code")
	if err := syncIssuesToMain(stderr, stderr, syncFlags, changeCodeRunner, msg); err != nil {
		retry := fmt.Sprintf("sdlc issue sync --issue %d", id)
		if onMain {
			retry += " --push"
		}
		cwarn(stderr, fmt.Sprintf("issue file not synced: %v\n"+
			"      the gates passed and the branch is being created anyway;\n"+
			"      re-run `%s` once the cause is cleared", err, retry))
	}
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
	if c.f.Force == "" {
		cwarn(c.stderr, "fix the failures above, OR re-run with --force <reason>")
	}
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

// resolveChangeCodeName reuses start.go's name-resolution shape but also returns
// the path to the resolved issue file so the gates can read it.
//
// resolveBranchName's untrackedFile is consumed here rather than returned: since
// #206 nothing downstream needs "is it untracked?" — the sync stages tracked and
// untracked issue files alike — but it is still the cheapest handle on the issue
// path when the file is brand new and no glob has seen it yet.
func resolveChangeCodeName(f *changeCodeFlags, r gitRunner) (name, issuePath string, err error) {
	nf := &nameFlags{
		Issue:     f.Issue,
		Name:      f.Name,
		IssuesDir: f.IssuesDir,
	}
	name, untrackedFile, err := resolveBranchName(nf, r)
	if err != nil {
		return "", "", err
	}

	// Locate the issue file. When --name was used (no --issue), we still
	// need the issue file for gates; derive it from the name prefix.
	if untrackedFile != "" {
		return name, untrackedFile, nil
	}
	issuePath, err = findIssueFileByName(f.IssuesDir, name)
	if err != nil {
		return "", "", err
	}
	return name, issuePath, nil
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
	contentHash := gatestate.ContentHash(planGateContent(issueContent), planContent)
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
		// KEEP the findings — only the DISPOSITIONS failed validation. `round.New` came
		// through ParseFindingsBlock (severities checked against the model) and got its ids
		// from AssignIDs, so those findings are as valid as any other round's. This differs
		// from the no-findings-block path above, which genuinely has nothing to record.
		//
		// Discarding them was worse than losing data: with no findings recorded,
		// RenderPriorFindings would tell the NEXT round "every prior finding has been
		// disposed" — a positive claim that is false — so a judge that raised three real
		// Criticals alongside one hallucinated disposition id would see a clean slate on
		// re-run, in the artifact whose sole purpose is not losing findings. It also
		// under-reported gate_addressed/gate_open in the close-time metric.
		// #194 M2 review: use ApplyChecked's OWN round, which now keeps the valid
		// dispositions and drops only the invalid ones. Hand-rebuilding it here dropped
		// every disposition, so one typo'd id nullified a round's valid disposals at the
		// gate whose purpose is disposal — the same defect fixed on the boundary side.
		applied.Rounds[len(applied.Rounds)-1].ProtocolError = aerr.Error()
		applied.Rounds[len(applied.Rounds)-1].Blocked = true
		applied.Rounds[len(applied.Rounds)-1].Forced = forcedRationale(f.Force, true)
		persistPlanGateRound(stderr, f, issueFile, ledger, applied.Rounds[len(applied.Rounds)-1])
		return fmt.Errorf("plan-quality protocol error: %v", aerr)
	}
	ledger = applied

	d := gatestate.Decide(ledger, roundCapFromEnv())
	// Stamp the pass-through key only on a PASSING round — caching a refusal would let a
	// still-blocked plan walk through unchanged on the next invocation. Gate-specific, so
	// it happens here rather than in the shared tail.
	if !d.Block {
		ledger.ContentHash = contentHash
	}
	// The SAME tail the boundary gate ends on (#194 close review BR-43). This gate's copy
	// diverged five times before it was extracted; sharing it is what stops a sixth.
	stampAndPersist(stderr, gatePersist{
		Label: "plan-quality",
		Write: func(out gatestate.Ledger) error {
			return writePlanGateLedger(f.PlansDir, issueFile, out, repoIdentity())
		},
	}, ledger, d, f.Force)

	if d.Block {
		cwarn(stderr, "address the findings above and re-run — the gate remembers what you fixed; OR re-run with --force <reason>"+fixTheClassNote())
		return fmt.Errorf("plan-quality failure")
	}
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
		cwarn(stderr, "plan-quality: findings reported — fix the plan, OR re-run with --force <reason>"+fixTheClassNote())
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

// planGateContent returns the part of the issue the PLAN gate is actually asked to review:
// everything except the estimate. It strips the `## Estimate` section and the
// `estimate_hours:` frontmatter line before hashing.
//
// This is what makes the pass-through cover the retry it exists for. B1 moved the estimate
// gates below plan-quality and B2 tells the agent to derive the estimate only once the plan
// clears — so the guaranteed sequence is: plan-quality passes → estimate gate refuses →
// operator adds `estimate_hours:` AND an `## Estimate` block → re-run. Hashing the whole
// issue would see both edits, invalidate the pass-through, and re-dispatch a multi-minute
// judge on exactly the retry the mechanism was built to make cheap.
//
// It is also correct on the merits, not just convenient: B1 removed the estimate from
// plan-quality's remit entirely (the prompt is now forbidden from mentioning it, pinned by
// TestBuildPrompt_PlanQuality_HasContract), so an estimate-only edit cannot change what
// that gate would conclude. Pure (ARCH-PURE).
func planGateContent(issueContent string) string {
	fm, body, err := issue.Parse(issueContent)
	if err != nil {
		return issueContent // no frontmatter to strip; hash it as-is
	}
	var keptFM []string
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "estimate_hours:") {
			continue
		}
		keptFM = append(keptFM, line)
	}
	// Strip the whole `## Estimate` section INCLUDING its heading. issue.EstimateSection
	// returns only the body beneath the heading, so removing that alone would leave an
	// orphan "## Estimate" behind — enough to change the hash and defeat the point.
	// Done line-wise rather than by regex because Go's RE2 has no lookahead, so a pattern
	// spanning to the next `## ` would consume that heading too.
	var kept []string
	inEstimate := false
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			inEstimate = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Estimate"
		}
		if !inEstimate {
			kept = append(kept, line)
		}
	}
	return strings.Join(keptFM, "\n") + "\n" + strings.TrimRight(strings.Join(kept, "\n"), "\n") + "\n"
}

// roundCapFromEnv reads WF_PLAN_ROUND_CAP, defaulting to gatestate.DefaultRoundCap. Past
// the cap only hard-blocking findings refuse the gate; the rest are carried to the close
// review via the ledger.
func roundCapFromEnv() int { return roundCapFromEnvVar("WF_PLAN_ROUND_CAP") }

// roundCapFromEnvVar is the shared reader behind each gate's cap knob (#194 M2 —
// the boundary gate has its own, WF_BOUNDARY_ROUND_CAP).
func roundCapFromEnvVar(name string) int {
	if v := os.Getenv(name); v != "" {
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
		cwarn(stderr, "estimate-quality: findings reported — fix the ## Estimate, OR re-run with --no-judge / --force <reason>"+fixTheClassNote())
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

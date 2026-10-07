// issue.go — `sdlc issue` command group: CRUD/authoring of the issue
// *record*, complementing the flat checkpoint guards that defend workflow
// *transitions* (ariadne#56).
//
// Subcommands: `new` (allocate ID + canonical template; `--from-github N`
// seeds from GitHub), `sync` (commit the body), `set-status`, `list`, `show`.
// `fetch` is a hidden deprecated alias for `new --from-github`; flat
// `set-status` likewise.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// NewIssueCmd returns the `sdlc issue` parent command. Long is a
// placeholder; main.go overrides with helptext.MustGet("issue").
func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "issue",
		Short:         "Create and manage workshop issues",
		Long:          "Placeholder — replaced by helptext.MustGet(\"issue\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newIssueNewCmd())

	// set-status moved under `issue` (#56 M2). The transition guards live in
	// statusDecision / checkTransitionGuards (pure, unit-tested); since #252
	// the status is written to the tracker card. main.go keeps a hidden deprecated
	// flat `sdlc set-status` alias for one cycle.
	setStatus := NewSetStatusCmd()
	setStatus.Long = renderLong("set-status") // #125: derive the lifecycle facts (not add()-wired)
	cmd.AddCommand(setStatus)
	cmd.AddCommand(newIssueSetTitleCmd(), newIssueSetEstimateCmd(), newIssueSetGitHubCmd())
	cmd.AddCommand(newIssueMoveDetailCmd(), newIssueRecoveryCmd(), newIssueMigrateCmd())

	cmd.AddCommand(newIssueSyncCmd())
	cmd.AddCommand(newIssuePublishCmd())
	cmd.AddCommand(newIssueLintIDsCmd())
	cmd.AddCommand(newIssueListCmd())
	cmd.AddCommand(newIssueShowCmd())
	cmd.AddCommand(newIssueValidateCmd())
	return cmd
}

// issueValidateFlags holds the parsed flags for `sdlc issue validate`.
type issueValidateFlags struct {
	Issues    []int
	All       bool
	IssuesDir string
}

func newIssueValidateCmd() *cobra.Command {
	f := issueValidateFlags{}
	cmd := &cobra.Command{
		Use:   "validate [<file>...]",
		Short: "Validate issue file(s) against the #Issue schema (frontmatter + sections)",
		Long: `Check that issue markdown conforms to the issue datatype: frontmatter against
#Issue (cue, via the vocabulary validator) + required-section presence (the SAME
policy the change-code gate uses — issue.CheckSectionsPresence). Well-formedness
only; semantic quality (Spec depth, etc.) is the LLM's job, not this.

  sdlc issue validate --issue 124       # one issue by ID
  sdlc issue validate --issue 1,2,3,4    # several issues by ID
  sdlc issue validate path/to/x.md       # one file
  sdlc issue validate a.md b.md c.md     # several files
  sdlc issue validate --all              # every workshop/issues/*.md

--issue takes a comma-separated list; positional <file> takes one or more paths.
Mixing --issue with positional files is rejected, and --all is mutually exclusive
with explicit targets.

Exits non-zero if any file is nonconforming, printing clear per-field/section
diagnostics — actionable enough to fix and re-validate. On-demand + informative;
the fail-closed boundary is the push/merge gate.`,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args)
		},
	}
	cmd.Flags().IntSliceVar(&f.Issues, "issue", nil, "issue ID(s) to validate (comma-separated: --issue 1,2,3)")
	cmd.Flags().BoolVar(&f.All, "all", false, "validate every issue under issues-dir")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

// runIssueValidate resolves the target file(s), runs full validation on each
// (frontmatter + sections), and returns a non-nil error iff any file is
// nonconforming (so the command exits non-zero).
func runIssueValidate(stdout, stderr io.Writer, f *issueValidateFlags, args []string) error {
	files, err := resolveValidateTargets(f, args)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no issue files to validate (pass <file>, --issue N, or --all)")
	}
	nonconforming := 0
	for _, file := range files {
		probs := validateIssueFull(file)
		if len(probs) == 0 {
			cok(stderr, fmt.Sprintf("%s: conforms", file))
			continue
		}
		nonconforming++
		cwarn(stderr, fmt.Sprintf("%s: %d problem(s)", file, len(probs)))
		for _, p := range probs {
			fmt.Fprintln(stdout, "  - "+p)
		}
	}
	if nonconforming > 0 {
		return fmt.Errorf("%d of %d issue file(s) nonconforming", nonconforming, len(files))
	}
	return nil
}

// resolveValidateTargets picks the files to validate from positional <file>
// paths, --issue ID(s), or --all. Multiple files and a comma-separated --issue
// list are both supported; the three sources are mutually exclusive — --all may
// not combine with explicit targets, and --issue may not mix with positional
// files (one batch form per invocation keeps the contract unambiguous).
func resolveValidateTargets(f *issueValidateFlags, args []string) ([]string, error) {
	hasFiles, hasIssues := len(args) > 0, len(f.Issues) > 0
	switch {
	case f.All:
		if hasFiles || hasIssues {
			return nil, fmt.Errorf("--all is mutually exclusive with explicit <file>/--issue targets")
		}
		return filepath.Glob(filepath.Join(f.IssuesDir, "*.md")) // Glob returns sorted matches
	case hasFiles && hasIssues:
		return nil, fmt.Errorf("specify either <file> path(s) or --issue ID(s), not both")
	case hasIssues:
		files := make([]string, 0, len(f.Issues))
		for _, id := range f.Issues {
			p, err := locateIssueFile(f.IssuesDir, id)
			if err != nil {
				return nil, err
			}
			files = append(files, p)
		}
		return files, nil
	case hasFiles:
		return args, nil
	default:
		return nil, fmt.Errorf("specify <file>, --issue N, or --all")
	}
}

// validateIssueFull runs both halves of conformance on one file: frontmatter (via
// the shared vocabulary-validator seam) + section presence (the shared change-code
// policy). On-demand validation is FULL (not added-only — that grandfather rule is
// the pre-merge gate's concern, not the agent's authoring check).
func validateIssueFull(file string) []string {
	var probs []string
	out, ok, runErr := validateFrontmatterFn("issue", file)
	switch {
	case runErr != nil:
		probs = append(probs, "could not run the frontmatter validator: "+runErr.Error())
	case !ok:
		probs = append(probs, "frontmatter:\n"+indentLines(strings.TrimSpace(out), "      "))
	}
	if data, err := os.ReadFile(file); err == nil {
		for _, sf := range issue.CheckSectionsPresence(string(data)) {
			probs = append(probs, "section: "+sf.Message)
		}
	}
	return probs
}

// issueNewFlags holds the parsed flags for `sdlc issue new`.
type issueNewFlags struct {
	Slug       string
	FromGitHub int
	Deps       []string
	Target     string
	DryRun     bool
	IssuesDir  string
	HistoryDir string
}

func newIssueNewCmd() *cobra.Command {
	f := issueNewFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "new <title>",
		Short: "Create a new workshop issue from the canonical template",
		Long: `Create workshop/issues/NNNNNN-<slug>.md from the canonical template
(see ` + "`sdlc issue --help`" + ` for the field/section contract). Allocates the
next 6-digit ID by scanning issues/ + history/ — the deterministic step the
agent must not do by hand under parallel workstreams — and prints the path.

  sdlc issue new "Some title"              # blank issue
  sdlc issue new "x" --target my-target    # with a target: slug
  sdlc issue new --from-github 42          # title/body from a GitHub issue

With --from-github the title is taken from the GitHub issue (a positional
title overrides it) and the issue body is seeded under ## Problem.`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueNew(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args)
		},
	})
	cmd.Flags().StringVar(&f.Slug, "slug", "", "override the auto-derived slug")
	cmd.Flags().IntVar(&f.FromGitHub, "from-github", 0, "derive title + body from this GitHub issue number")
	cmd.Flags().StringSliceVar(&f.Deps, "deps", nil, "dependency refs, e.g. --deps repo#1,repo#2")
	cmd.Flags().StringVar(&f.Target, "target", "", "target: frontmatter slug")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be path + body; do not write")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	return cmd
}

// runIssueNew files an issue in two places (#252): the card is reserved on the
// tracker branch by its own commit, then the details are written to this
// checkout. Nothing is published to main — details land there through the
// branch's own PR, or `sdlc issue move-detail`, and claim waits for that.
func runIssueNew(ctx context.Context, stdout, stderr io.Writer, f *issueNewFlags, args []string) error {
	if tracked, err := repositoryTracked(ctx, f.IssuesDir); err != nil {
		return err
	} else if !tracked {
		return runLegacyIssueNew(stdout, stderr, f, args)
	}
	title := ""
	if len(args) > 0 {
		title = args[0]
	}

	var ghNum, problemBody string
	if f.FromGitHub > 0 {
		repo, err := detectRepo()
		if err != nil {
			return err
		}
		ghNum = strconv.Itoa(f.FromGitHub)
		ghTitle, ghBody, err := ghClient.TitleAndBody(repo, ghNum)
		if err != nil {
			return fmt.Errorf("fetch GitHub issue %s: %w", ghNum, err)
		}
		if title == "" {
			title = ghTitle
		}
		problemBody = ghBody
	}
	if strings.TrimSpace(title) == "" {
		return errors.New("a title is required (positional arg, or --from-github N to derive it)")
	}
	slug := f.Slug
	if slug == "" {
		slug = issue.Slugify(title)
	}
	if slug == "" {
		return fmt.Errorf("title %q produced an empty slug; pass --slug", title)
	}
	dirs, err := resolveIDDirs(f.IssuesDir, f.HistoryDir)
	if err != nil {
		return err
	}
	home := dirs.Rel[0]

	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	if env.branch == "" {
		return errors.New("issue new writes details on the current branch; check out a branch first")
	}
	today := time.Now().Format("2006-01-02")
	render := func(id string) (tracker.Draft, error) {
		full := issue.Render(issue.ScaffoldSpec{ID: id, Title: title, Today: today, GithubIssue: ghNum,
			ProblemBody: problemBody, Deps: f.Deps, Target: f.Target})
		card, detail, err := issue.SplitCardWithFormat([]byte(full), env.format)
		if err != nil {
			return tracker.Draft{}, err
		}
		return tracker.Draft{CardPath: tracker.CardPath(id, slug), DetailPath: path.Join(home, id+"-"+slug+".md"), Card: card, Detail: detail}, nil
	}
	if f.DryRun {
		snap, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		d, err := render(fmt.Sprintf("%06d", snap.MaxID()+1))
		if err != nil {
			return err
		}
		cinfo(stderr, "dry-run — nothing reserved or written")
		fmt.Fprintf(stdout, "Would reserve: %s (on %s)\nWould create: %s\n", d.CardPath, vocab.Issue().Discovery().Tracker, d.DetailPath)
		fmt.Fprintln(stdout, "─── details ───")
		fmt.Fprint(stdout, string(d.Detail))
		return nil
	}

	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	receipts, err := env.receipts()
	if err != nil {
		return err
	}
	op := tracker.NewCreationOp(ctx, env.repo, env.checkout(mainView.Ref()), render)
	var r tracker.Receipt
	// A peer can take the allocated ID between snapshot and candidate; nothing
	// is published then, so a fresh allocation is safe. Races after the
	// candidate is durable are the receipt engine's (bounded) retries.
	for attempt := 1; ; attempt++ {
		if r, err = op.Start(operationToken("new")); err != nil {
			return err
		}
		r, err = tracker.Drive(r, tracker.CreationStepper, op, receipts)
		invalidateIssueRecords(env.ctx)
		if err == nil || !errors.Is(err, tracker.ErrIDTaken) || !r.Discardable() || attempt == tracker.MaxPublicationAttempts {
			break
		}
		if derr := receipts.Discard(r); derr != nil {
			return derr
		}
	}
	spec := r.Spec()
	if errors.Is(err, tracker.ErrOperationUncertain) {
		return fmt.Errorf("%w\n      nothing is lost: run `sdlc issue recovery reconcile --issue %s` to finish #%s", err, issue.CLIRef(spec.IssueID), spec.IssueID)
	}
	if err != nil {
		return err
	}
	created := fmt.Sprintf("Reserved #%s on %s; created %s", spec.IssueID, vocab.Issue().Discovery().Tracker, spec.DestinationPath)
	if ghNum != "" {
		created += fmt.Sprintf(" (GitHub #%s)", ghNum)
	}
	cok(stderr, created)
	if env.onRest() {
		cinfo(stderr, "details are uncommitted on the resting branch; the issue is claimable once they land on main — "+
			"`sdlc issue move-detail --issue "+issue.CLIRef(spec.IssueID)+"` publishes them")
	} else if err := commitOnly(env, fmt.Sprintf("#%s: issue: new", issue.CLIRef(spec.IssueID)), spec.DestinationPath); err != nil {
		return fmt.Errorf("card reserved and details written, but the local commit failed: %w", err)
	}
	fmt.Fprintln(stdout, spec.DestinationPath)
	return nil
}

// ── issue sync ───────────────────────────────────────────────────────────────

// issueSyncFlags holds the parsed flags for `sdlc issue sync`.
type issueSyncFlags struct {
	Issue     int
	IssuesDir string
	Push      bool
	DryRun    bool
	Context   context.Context
}

// newIssueSyncCmd builds `sdlc issue sync` (#206) — the verb that commits an
// issue's BODY. `claim` and `issue new` publish the reservation (an ID and a
// name); nothing committed the Spec/Plan/Log that follows, so the whole planning
// phase sat uncommitted until an unrelated verb happened to sweep it up, and a
// compaction or a closed terminal lost it.
//
// It WRAPS rather than authors: agents edit markdown incrementally with their
// own tools, so a verb taking body content as an argument would fight how the
// work actually happens. Staging + message + lock is all it adds.
func newIssueSyncCmd() *cobra.Command {
	f := issueSyncFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:   "sync",
		Short: "Commit an issue's body (spec/plan/log) so planning output is durable",
		Long: `Commit workshop/issues/NNNNNN-*.md for one issue, under a message that
names it — the planning-phase counterpart to ` + "`sdlc claim`" + `, which publishes
only the reservation.

  sdlc issue sync --issue 206           # commit here; nothing leaves the machine
  sdlc issue sync --issue 206 --push    # ... and publish to origin/main

DOES NOT PUSH BY DEFAULT. A mid-planning sync is a cheap, frequent, local act;
publishing is an external contract that belongs at the boundaries already owning
it (` + "`sdlc milestone-close`, `sdlc close`, `sdlc change-code`" + `), so the
common case cannot accidentally publish a half-written Spec. The default commit
lands in the CURRENT worktree on the CURRENT branch, needs no network, and works
from an in-place feature branch. --push publishes only the exact new issue commit
created by this invocation. If there is no new commit, select an earlier commit
explicitly with ` + "`sdlc issue publish --commit SHA`" + `. A separate plan or
project document must be deliberately included in the selected commit.

Run it whenever the Spec, Plan or Log has moved — after a brainstorm lands,
after a design decision, before a long-running tool call.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.Context = cmd.Context()
			return runIssueSync(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "workshop issue ID whose files to commit (required)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().BoolVar(&f.Push, "push", false, "also publish to origin/main (off by default — publishing belongs at the milestone verbs)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print what would happen; do not commit/push")
	return cmd
}

func runIssueSync(stdout, stderr io.Writer, f *issueSyncFlags) error {
	if f.Issue <= 0 {
		die(stderr, "--issue N is required: `sdlc issue sync` commits ONE issue's files, and the commit message names it")
	}
	if err := trackedIssueSync(f); err != nil {
		die(stderr, err.Error())
	}
	// Reuses the narrow synchronization dispatch (ARCH-DRY): this verb
	// adds are the subject and the publish choice, both parameters of the shared
	// helper rather than a second sync path.
	syncFlags := &claimFlags{
		Issue:     f.Issue,
		IssuesDir: f.IssuesDir,
		DryRun:    f.DryRun,
		NoStart:   true,
		NoPush:    !f.Push,
	}
	if err := syncIssuesToMain(stdout, stderr, syncFlags, claimRunner, issueSyncMessage(f.Issue, "spec/plan")); err != nil {
		die(stderr, err.Error())
	}
	return nil
}

// issueSyncMessage is the subject an issue-body commit lands under: the tree's
// `#N: <area>: <subject>` shape, so `git log --grep "^#206"` finds the planning
// output alongside the implementation.
//
// The `issue-sync` area is load-bearing, not decorative. Anchoring #N would make
// gitx.IsShippedWorkSubject read a tracker commit as shipped implementation —
// feeding drift detection, milestone review windows and active-time attribution
// — so `issue-sync` is a declared bookkeeping lead-in (gitx/window.go). Changing
// this prefix without changing that list silently re-breaks all three.
func issueSyncMessage(issue int, what string) string {
	return fmt.Sprintf("#%d: issue-sync: %s", issue, what)
}

// ── issue list ───────────────────────────────────────────────────────────────

type issueListFlags struct {
	Status    string
	IssuesDir string
}

func newIssueListCmd() *cobra.Command {
	f := issueListFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workshop issues (ID, status, title)",
		Long: `List issues in workshop/issues/ as "ID  STATUS  TITLE", sorted by ID.
Filter with --status. Broader than 'sdlc state', which surfaces only the
working set + drift.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueList(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().StringVar(&f.Status, "status", "", "filter to this status (open|working|blocked|done|wontfix|punt)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

// runIssueList reuses state.go's listIssues (which reads + sorts by ID)
// rather than re-deriving the scan/sort.
func runIssueList(ctx context.Context, stdout, stderr io.Writer, f *issueListFlags) error {
	if f.Status != "" && !isValidStatus(f.Status) {
		die(stderr, fmt.Sprintf("invalid status %q (valid: %s)", f.Status, strings.Join(vocab.Issue().AllStatuses(), ", ")))
	}
	issues, stale, err := listIssueStates(ctx, f.IssuesDir)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
	}
	if stale {
		cwarn(stderr, staleTrackerNote)
	}
	n := 0
	for _, is := range issues {
		if f.Status != "" && is.Status != f.Status {
			continue
		}
		// width 10 fits the longest real status + the "unreadable"
		// sentinel listIssues emits for broken files.
		fmt.Fprintf(stdout, "%s  %-10s  %s\n", is.ID, valueOr(is.Status, "?"), is.Title)
		n++
	}
	if n == 0 {
		cinfo(stderr, "no issues match")
	}
	return nil
}

// ── issue show ───────────────────────────────────────────────────────────────

type issueShowFlags struct {
	IssuesDir string
	JSON      bool   // #279: emit the observation
	Repo      string // #279: observe another repository's issue
}

func newIssueShowCmd() *cobra.Command {
	f := issueShowFlags{}
	cmd := &cobra.Command{
		Use:   "show <N>",
		Short: "Show an issue's frontmatter + section headers (no bodies)",
		Long: `Print issue <N>'s frontmatter and its body section headers (# / ## lines)
without the section contents — a structured peek for orienting on an issue
without loading the whole file — followed by its observations (#279).

OBSERVATIONS

  --json prints only the observation: a versioned (schema_version 1),
  read-only answer to who owns the issue and where its work is, which
  checkpoints passed, and whether it landed. Consumers reject other versions
  and unknown keys. Sections:
    - tracker: the tracker commit read;
    - card: status, title and revision;
    - assignment: the claimant, its relation to the queried checkout, and
      where the owner's worktree stands;
    - workspaces: local worktrees holding the issue branch, with dirty and
      ahead/behind counts;
    - branch: the branch head and its commits ahead of main;
    - checkpoints: the flow, plan ticks, and each review boundary's verdict
      and open blocking findings;
    - completion: the close's evidence and reviewed commits;
    - landing: whether the work landed, its landed commit, and its archive
      path.

  Every section's "state" is read quality, never the value: present, absent
  (read, none recorded), stale (answered from the last tracker fetch; error
  says why) or unknown (the read failed; error says why). A failed read is
  never reported as absent.

  Authority: "tracker" sections (card, assignment, completion, landing) are
  authoritative for claim, status and landing. "committed" sections
  (branch, checkpoints) read evidence on the issue branch, or on main's
  archive once landed, so squash merges and deleted branches don't lose it.
  "worktree" is activity only. Dirty files and unpushed commits are not
  progress, and a working card proves a claim, not execution.

  Every answer carries observed_at and the tracker commit. The query fetches
  the tracker (updating the remote-tracking ref) and otherwise writes
  nothing.

  It runs from any checkout of the repository. Parked or agentless slots are
  found through git, and another machine's worktree is reported as
  other-machine without being probed. --repo <path> observes an issue of the
  repository containing that path. The repository is always the one
  containing the issues directory, never the shell's current directory.`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueShow(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args[0])
		},
	}
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().BoolVar(&f.JSON, "json", false, "print the issue's observation (schema_version 1) as JSON")
	cmd.Flags().StringVar(&f.Repo, "repo", "", "observe an issue of the repository containing this path")
	return cmd
}

func runIssueShow(ctx context.Context, stdout, stderr io.Writer, f *issueShowFlags, arg string) error {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		die(stderr, fmt.Sprintf("invalid issue id %q (want a positive number, e.g. 56)", arg))
	}
	root, issuesDir, err := issueShowRepo(f)
	if err != nil {
		die(stderr, err.Error())
	}
	if f.JSON { // #279: the observation alone, on stdout
		raw, err := json.MarshalIndent(collectObservation(ctx, root, issuesDir, fmt.Sprintf("%06d", id)), "", "  ")
		if err != nil {
			die(stderr, err.Error())
		}
		fmt.Fprintln(stdout, string(raw))
		return nil
	}
	defer func() { observe.RenderText(stdout, collectObservation(ctx, root, issuesDir, fmt.Sprintf("%06d", id))) }()
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.PreferFresh)
	if err != nil {
		die(stderr, err.Error())
	}
	rec, ok, cardErr := rs.Require(fmt.Sprintf("%06d", id))
	if !ok {
		die(stderr, fmt.Sprintf("no issue #%d: no card and no details under %s", id, f.IssuesDir))
	}
	if rec.Card != nil {
		// The card is authoritative for its fields (#252); shown first.
		stale := ""
		if rs.Stale {
			stale = " (tracker unreachable — as last fetched)"
		}
		fmt.Fprintf(stdout, "card %s on %s%s\n---\n%s---\n", rec.Card.Path, vocab.Issue().Discovery().Tracker, stale, ensureTrailingNewline(rec.Card.Card.Frontmatter))
	}
	if cardErr != nil {
		fmt.Fprintf(stdout, "card unreadable on %s: %v\n", vocab.Issue().Discovery().Tracker, cardErr)
	}
	if rec.DetailPath == "" {
		fmt.Fprintln(stdout, "(details are not in this checkout — card only)")
		fmt.Fprintf(stdout, "# %s\n", rec.Title())
		return nil
	}
	if rec.DetailErr != nil {
		die(stderr, fmt.Sprintf("read %s: %v", rec.DetailPath, rec.DetailErr))
	}
	fmt.Fprintf(stdout, "%s\n---\n%s---\n", filepath.Base(rec.DetailPath), ensureTrailingNewline(rec.DetailFM))
	// Title + section headers only (`# ` / `## `). Deeper headers like
	// `### YYYY-MM-DD` Log entries are intentionally omitted — this is a
	// structure peek, not a content dump.
	for _, line := range strings.Split(rec.DetailBody, "\n") {
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			fmt.Fprintln(stdout, line)
		}
	}
	return nil
}

// ensureTrailingNewline returns s with exactly one terminating newline.
func ensureTrailingNewline(s string) string {
	return strings.TrimRight(s, "\n") + "\n"
}

// trackedIssueSync retires `issue sync` in a repository cut over to the issue
// tracker (#284 audit). There it was only `git commit -- <details>` on the
// issue branch; committing details on a resting branch diverges it from main,
// and copying commits to main (--push) is `issue publish`'s job. It names both
// instead of doing either.
func trackedIssueSync(f *issueSyncFlags) error {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	cut, err := tracker.CutOver(root)
	if err != nil || !cut {
		return err
	}
	n := "N"
	if f.Issue > 0 {
		n = fmt.Sprint(f.Issue)
	}
	return fmt.Errorf("this repository uses the issue tracker, where `issue sync` is retired: commit details with git on the issue branch "+
		"(`git commit -m '#%s: plan: …' -- <details>`), and publish them to main with `sdlc issue publish --issue %s`. "+
		"Never commit on a resting branch: it only fast-forwards to main", n, n)
}

// issueShowRepo resolves what an issue show reads (#279). The repository is
// the one containing the issues dir — the data the verb was given, never the
// process cwd — or, with --repo, the one containing that path, its issues dir
// resolved inside it. An issues dir outside any repository yields root "": its
// details are the whole record, as before.
func issueShowRepo(f *issueShowFlags) (root, issuesDir string, err error) {
	if f.Repo != "" {
		if root = repoRootOf(f.Repo); root == "" {
			return "", "", fmt.Errorf("--repo %s is not inside a repository", f.Repo)
		}
		issuesDir = f.IssuesDir
		if !filepath.IsAbs(issuesDir) {
			issuesDir = filepath.Join(root, issuesDir)
		}
		return root, issuesDir, nil
	}
	if issuesDir, err = filepath.Abs(f.IssuesDir); err != nil {
		return "", "", err
	}
	return repoRootOf(issuesDir), issuesDir, nil
}

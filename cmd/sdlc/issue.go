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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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

	// set-status moved under `issue` (#56 M2). The transition guards live
	// in applyStatus / checkTransitionGuards (returned errors, unit-tested)
	// — only the cobra wiring relocates. main.go keeps a hidden deprecated
	// flat `sdlc set-status` alias for one cycle.
	setStatus := NewSetStatusCmd()
	setStatus.Long = renderLong("set-status") // #125: derive the lifecycle facts (not add()-wired)
	cmd.AddCommand(setStatus)

	cmd.AddCommand(newIssueSyncCmd())
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
			return runIssueNew(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args)
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

// runIssueNew is the entry point for `sdlc issue new`. Hard guardrail
// failures call die(); the happy path prints the created path to stdout.
func runIssueNew(stdout, stderr io.Writer, f *issueNewFlags, args []string) error {
	title := ""
	if len(args) > 0 {
		title = args[0]
	}

	var ghNum, problemBody string
	if f.FromGitHub > 0 {
		repo, err := detectRepo()
		if err != nil {
			die(stderr, err.Error())
		}
		ghNum = strconv.Itoa(f.FromGitHub)
		ghTitle, ghBody, err := ghClient.TitleAndBody(repo, ghNum)
		if err != nil {
			die(stderr, fmt.Sprintf("fetch GitHub issue %s: %v", ghNum, err))
		}
		if title == "" {
			title = ghTitle
		}
		problemBody = ghBody
	}

	if strings.TrimSpace(title) == "" {
		die(stderr, "a title is required (positional arg, or --from-github N to derive it)")
	}

	slug := f.Slug
	if slug == "" {
		slug = issue.Slugify(title)
	}
	if slug == "" {
		die(stderr, fmt.Sprintf("title %q produced an empty slug; pass --slug", title))
	}

	nextID, err := allocateIssueID(stderr, f.IssuesDir, f.HistoryDir, claimRunner)
	if err != nil {
		die(stderr, err.Error())
	}

	today := time.Now().Format("2006-01-02")
	dest := filepath.Join(f.IssuesDir, fmt.Sprintf("%s-%s.md", nextID, slug))
	if _, err := os.Stat(dest); err == nil {
		die(stderr, fmt.Sprintf("issue file already exists: %s", dest))
	}

	rendered := issue.Render(issue.ScaffoldSpec{
		ID:          nextID,
		Title:       title,
		Today:       today,
		GithubIssue: ghNum,
		ProblemBody: problemBody,
		Deps:        f.Deps,
		Target:      f.Target,
	})

	if f.DryRun {
		cinfo(stderr, "dry-run — no files written")
		fmt.Fprintf(stdout, "Would create: %s\n", dest)
		fmt.Fprintln(stdout, "─── body ───")
		fmt.Fprint(stdout, rendered)
		return nil
	}

	if err := os.MkdirAll(f.IssuesDir, 0o755); err != nil {
		die(stderr, fmt.Sprintf("mkdir %s: %v", f.IssuesDir, err))
	}
	if err := os.WriteFile(dest, []byte(rendered), 0o644); err != nil {
		die(stderr, fmt.Sprintf("write %s: %v", dest, err))
	}

	created := fmt.Sprintf("Created %s", dest)
	if ghNum != "" {
		created += fmt.Sprintf(" (GitHub #%s)", ghNum)
	}
	cok(stderr, created)

	// #82 M1: broadcast the new issue to origin/main immediately, so a freshly
	// filed (base) issue is tracker state on main — not untracked working-tree
	// residue that every symlinked derivative reads and that gates trip over.
	// Reuses claim's branch-aware sync with this issue as the `--issue` filter
	// (rides #80's filtered add — unrelated untracked files stay put). nextID is
	// a zero-padded string ("000083"); claimFlags.Issue is an int.
	if id, perr := strconv.Atoi(nextID); perr == nil {
		syncFlags := &claimFlags{Issue: id, IssuesDir: f.IssuesDir, NoStart: true}
		// Route the sync's stdout to stderr: its machine "synced" marker must not
		// pollute `issue new`'s stdout contract (the created path, printed below).
		// "" keeps issue new's historical subject ("issue-sync: update issues");
		// naming the issue is `sdlc issue sync`'s job, not creation's.
		if serr := syncIssuesToMain(stderr, stderr, syncFlags, claimRunner, ""); serr != nil {
			// Best-effort: the file is already written + reported above, so a sync
			// failure (offline, no reachable origin, conflict) must not abort the
			// create — just surface it. `claim` treats the same error as fatal.
			//
			// But fall back to a LOCAL commit first (#206). Publication and
			// durability are separable, and only publication failed here; leaving
			// the new issue as an untracked working-tree file is the hole this
			// issue exists to close. The common trigger is mundane: `issue new`
			// from an in-place feature branch, where the publish route finds no
			// worktree on `main` and there is nothing wrong at all. A no-op when
			// the first attempt already committed and only the push failed.
			syncFlags.NoPush = true
			if lerr := syncIssuesToMain(stderr, stderr, syncFlags, claimRunner, issueSyncMessage(id, "new issue")); lerr != nil {
				cwarn(stderr, fmt.Sprintf("issue created but NOT committed: %v (sync to main also failed: %v)", lerr, serr))
			} else {
				cwarn(stderr, fmt.Sprintf("issue committed locally but not broadcast to main: %v\n"+
					"      peers won't see the reservation yet — publish with `sdlc issue sync --issue %d --push`", serr, id))
			}
		}
	}

	fmt.Fprintln(stdout, dest)
	return nil
}

// ── issue sync ───────────────────────────────────────────────────────────────

// issueSyncFlags holds the parsed flags for `sdlc issue sync`.
type issueSyncFlags struct {
	Issue     int
	IssuesDir string
	Push      bool
	DryRun    bool
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
from an in-place feature branch. --push routes through the same branch-aware
dispatch ` + "`sdlc claim`" + ` uses.

Run it whenever the Spec, Plan or Log has moved — after a brainstorm lands,
after a design decision, before a long-running tool call.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
	// Reuses claim's dispatch wholesale (ARCH-DRY): the only things this verb
	// adds are the subject and the publish choice, both parameters of the shared
	// helper rather than a second sync path.
	syncFlags := &claimFlags{
		Issue:     f.Issue,
		IssuesDir: f.IssuesDir,
		DryRun:    f.DryRun,
		NoStart:   true,
		NoPush:    !f.Push,
		// --push means "make origin/main carry this body", which includes the
		// case where the body is already committed and only the publish is
		// missing — the state the no-push default deliberately creates.
		PublishExisting: f.Push,
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
			return runIssueList(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().StringVar(&f.Status, "status", "", "filter to this status (open|working|blocked|done|wontfix|punt)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

// runIssueList reuses state.go's listIssues (which reads + sorts by ID)
// rather than re-deriving the scan/sort.
func runIssueList(stdout, stderr io.Writer, f *issueListFlags) error {
	if f.Status != "" && !isValidStatus(f.Status) {
		die(stderr, fmt.Sprintf("invalid status %q (valid: %s)", f.Status, strings.Join(vocab.Issue().AllStatuses(), ", ")))
	}
	issues, err := listIssues(f.IssuesDir)
	if err != nil {
		die(stderr, fmt.Sprintf("list issues: %v", err))
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
}

func newIssueShowCmd() *cobra.Command {
	f := issueShowFlags{}
	cmd := &cobra.Command{
		Use:   "show <N>",
		Short: "Show an issue's frontmatter + section headers (no bodies)",
		Long: `Print issue <N>'s frontmatter and its body section headers (# / ## lines)
without the section contents — a structured peek for orienting on an issue
without loading the whole file.`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueShow(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f, args[0])
		},
	}
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

func runIssueShow(stdout, stderr io.Writer, f *issueShowFlags, arg string) error {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		die(stderr, fmt.Sprintf("invalid issue id %q (want a positive number, e.g. 56)", arg))
	}
	path, err := locateIssueFile(f.IssuesDir, id)
	if err != nil {
		die(stderr, err.Error())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		die(stderr, fmt.Sprintf("read %s: %v", path, err))
	}
	fm, body, err := issue.Parse(string(data))
	if err != nil {
		die(stderr, fmt.Sprintf("parse %s: %v", path, err))
	}
	fmt.Fprintf(stdout, "%s\n---\n%s---\n", filepath.Base(path), ensureTrailingNewline(fm))
	// Title + section headers only (`# ` / `## `). Deeper headers like
	// `### YYYY-MM-DD` Log entries are intentionally omitted — this is a
	// structure peek, not a content dump.
	for _, line := range strings.Split(body, "\n") {
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

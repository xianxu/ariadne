// flake_cmd.go — the known-flake registry's verbs (#394). The registry is the
// open issues whose tracker card lists flake_tests: `sdlc flake register`
// writes that field only with same-head fail→pass evidence, and `sdlc flake
// list` is the TL's local daily view.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const flakeTestsField = "flake_tests"

// flakeCards reads every issue's status and flake_tests from records. An
// unreadable card or field refuses: a card that may be a flake issue is
// unknown, never absent (#288), so the registry is never silently smaller.
func flakeCards(rs tracker.Records) ([]flakes.Card, error) {
	if !rs.Tracker {
		return nil, errors.New("the known-flake registry needs an issue tracker (workshop/issue-tracker.json)")
	}
	if rs.Stale {
		return nil, fmt.Errorf("the issue tracker could not be fetched (%v); a stale card may list a closed flake", rs.FetchErr)
	}
	var cards []flakes.Card
	for _, r := range rs.All() {
		if r.Duplicate {
			continue
		}
		if r.CardErr != nil {
			return nil, fmt.Errorf("issue %s: card unreadable: %v", r.ID, r.CardErr)
		}
		if r.Card == nil {
			continue
		}
		tests, _, err := issue.FieldList(r.Card.Card.Frontmatter, flakeTestsField)
		if err != nil {
			return nil, fmt.Errorf("issue %s: %v", r.ID, err)
		}
		if len(tests) > 0 {
			created, _ := r.Field("created")
			cards = append(cards, flakes.Card{ID: r.ID, Status: r.Status(), Tests: tests, Owner: flakeCardOwner(r.Card.Raw), Created: created})
		}
	}
	return cards, nil
}

// flakeCardOwner names who holds a card: the claimant's slot, else its operator.
func flakeCardOwner(raw []byte) string {
	c, ok, err := issue.CardClaimant(raw)
	switch {
	case err != nil:
		return "unreadable claimant"
	case !ok:
		return ""
	case c.Workspace != "":
		return c.Workspace
	}
	return c.Operator
}

// addFlakeTests is the setter's decision, made on the card the compare-and-
// swap hands it (#394 M1 review): the tests already on that card plus add,
// so a registration racing this one is kept, never overwritten from an
// earlier read. A closed issue refuses: its tests would never register.
func addFlakeTests(card tracker.Record, add []string) ([]byte, error) {
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	have, _, err := issue.FieldList(card.Card.Frontmatter, flakeTestsField)
	if err != nil {
		return nil, err
	}
	tests, err := mergeFlakeTests(card.ID, status, have, add)
	if err != nil {
		return nil, err
	}
	return issue.SetCardField(card.Raw, flakeTestsField, issue.FlowList(tests))
}

// mergeFlakeTests is the registration's pure decision on one card's current
// state: its tests plus add, at most flakes.MaxRerun, never on a closed issue.
func mergeFlakeTests(id, status string, have, add []string) ([]string, error) {
	if vocab.Issue().IsTerminal(status) {
		return nil, fmt.Errorf("issue #%s is %s: a closed issue registers nothing; reopen it or file a new one", id, status)
	}
	tests := mergeTests(have, add)
	if len(tests) > flakes.MaxRerun {
		return nil, fmt.Errorf("issue #%s would register %d tests (at most %d): file another issue", id, len(tests), flakes.MaxRerun)
	}
	return tests, nil
}

// removeFlakeTests is unregister's pure decision on the card's current tests
// (#466): name dropped; a name already absent changes nothing (a repeat).
func removeFlakeTests(id, status string, have []string, name string) ([]string, error) {
	if vocab.Issue().IsTerminal(status) {
		return nil, fmt.Errorf("issue #%s is %s: closing it already unregistered its tests", id, status)
	}
	return slices.DeleteFunc(slices.Clone(have), func(t string) bool { return t == name }), nil
}

// narrowFlakeTests is narrow's pure decision (#466): old replaced by to, a
// name strictly under it (one of its subtests), so the registration covers
// less and needs no new evidence. Already narrowed (old gone, to present)
// changes nothing; neither present refuses. Other names are kept as they are.
func narrowFlakeTests(id, status string, have []string, old, to string) ([]string, error) {
	if vocab.Issue().IsTerminal(status) {
		return nil, fmt.Errorf("issue #%s is %s: closing it already unregistered its tests", id, status)
	}
	if to == old || !flakes.Covers(old, to) {
		return nil, fmt.Errorf("%s is not a subtest of %s: narrowing only shrinks a registration; unregister %s and register %s with its own evidence", to, old, old, to)
	}
	hasOld, hasTo := slices.Contains(have, old), slices.Contains(have, to)
	switch {
	case !hasOld && hasTo:
		return slices.Clone(have), nil
	case !hasOld:
		return nil, fmt.Errorf("issue #%s does not register %s: nothing to narrow", id, old)
	}
	var out []string
	for _, t := range have {
		switch {
		case t == old && !hasTo:
			out = append(out, to)
		case t != old:
			out = append(out, t)
		}
	}
	return out, nil
}

// editFlakeTests applies an owner-only edit to the compare-and-swap card: the
// details-publish rule (its owner, or anyone while unclaimed), then the edit
// on the card's current flake_tests, validated as the registry reads them.
func editFlakeTests(env *trackerEnv, card tracker.Record, edit func(id, status string, have []string) ([]string, error)) ([]byte, error) {
	if err := requireOwnedToEdit(env, card, "unregisters or narrows its flakes"); err != nil {
		return nil, err
	}
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	have, _, err := issue.FieldList(card.Card.Frontmatter, flakeTestsField)
	if err != nil {
		return nil, err
	}
	tests, err := edit(card.ID, status, have)
	if err != nil {
		return nil, err
	}
	for _, t := range tests {
		if err := flakes.ValidTest(t); err != nil {
			return nil, err
		}
	}
	if len(tests) == 0 {
		// The last name gone: the field is cleared, as area is; a card
		// refuses an empty list ("[]").
		return issue.SetCardField(card.Raw, flakeTestsField, "")
	}
	return issue.SetCardField(card.Raw, flakeTestsField, issue.FlowList(tests))
}

// flakeIssueOpen is the registry's status rule: anything not terminal.
// codecomplete still registers, since the fix has not landed on main.
func flakeIssueOpen(status string) bool { return !vocab.Issue().IsTerminal(status) }

// loadFlakeRegistry builds the registry from a fresh tracker read, the same
// read that gives each issue's status (#394 D4).
func loadFlakeRegistry(ctx context.Context, issuesDir string) (flakes.Registry, []flakes.Card, error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil {
		return flakes.Registry{}, nil, fmt.Errorf("read the known-flake registry: %w", err)
	}
	cards, err := flakeCards(rs)
	if err != nil {
		return flakes.Registry{}, nil, fmt.Errorf("read the known-flake registry: %w", err)
	}
	r, err := flakes.BuildRegistry(cards, flakeIssueOpen)
	return r, cards, err
}

// suiteStoreDir is this clone's test-stats store.
func suiteStoreDir() (string, error) {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return "", err
	}
	logDir, err := localCheckLogDir(localRunDeps{Root: root, Git: func(args ...string) (string, error) {
		out, err := mergeRunner.GitInDir(root, args...)
		return string(out), err
	}})
	if err != nil {
		return "", err
	}
	return suiteStoreRoot(logDir), nil
}

// flakeDeps are the IO seams the verbs use; tests replace them.
type flakeDeps struct {
	now      func() time.Time
	attempts func() ([]SuiteAttempt, []string, error) // this clone's test-stats store
	logPath  func(SuiteAttempt) string                // where an attempt's log is kept
	subtests func(SuiteAttempt) ([]string, error)     // the failing subtests its saved log names (#466)
	repo     func() (string, error)
	registry func(ctx context.Context) (flakes.Registry, []flakes.Card, error)
	listView func(ctx context.Context) (flakeListView, error)
	contains func(ctx context.Context, ancestor, head string) (bool, error)
	addTests func(ctx context.Context, stdout, stderr io.Writer, id int, add []string) error // merges onto the card's current tests
	// editTests applies an owner-only edit (unregister, narrow) to the card's
	// current tests, compare-and-swap; a repeat writes nothing.
	editTests func(ctx context.Context, stdout, stderr io.Writer, id int, what string, edit func(id, status string, have []string) ([]string, error)) error
	newIssue  func(ctx context.Context, stdout, stderr io.Writer, title, body string) (int, error)
}

func defaultFlakeDeps(issuesDir string) flakeDeps {
	return flakeDeps{
		now: time.Now,
		attempts: func() ([]SuiteAttempt, []string, error) {
			store, err := suiteStoreDir()
			if err != nil {
				return nil, nil, err
			}
			snapshot, err := readSuiteEvidence(store, testStatsRecordLimit)
			return snapshot.flakeAttempts(), append(snapshot.Warnings, snapshot.coverageNotes("")...), err
		},
		logPath: func(a SuiteAttempt) string {
			if store, err := suiteStoreDir(); err == nil {
				if p := suiteAttemptLogFile(store, a); p != "" {
					return p
				}
			}
			return valueOr(a.LogPath, "(none kept)")
		},
		subtests: func(a SuiteAttempt) ([]string, error) {
			store, err := suiteStoreDir()
			if err != nil {
				return nil, err
			}
			return storeSubtests(store)(a)
		},
		repo: detectRepo,
		registry: func(ctx context.Context) (flakes.Registry, []flakes.Card, error) {
			return loadFlakeRegistry(ctx, issuesDir)
		},
		listView: func(ctx context.Context) (flakeListView, error) { return loadFlakeListView(ctx, issuesDir) },
		contains: func(ctx context.Context, ancestor, head string) (bool, error) {
			root, err := gitx.RepoTopLevel()
			if err != nil {
				return false, err
			}
			return flakeGitContains(ctx, root, ancestor, head)
		},
		addTests: func(ctx context.Context, stdout, stderr io.Writer, id int, add []string) error {
			return runCardUpdate(ctx, stdout, stderr, issuesDir, id, "flake_tests + "+issue.FlowList(add), false,
				func(_ *trackerEnv, card tracker.Record, _ string) ([]byte, error) { return addFlakeTests(card, add) })
		},
		editTests: func(ctx context.Context, stdout, stderr io.Writer, id int, what string, edit func(id, status string, have []string) ([]string, error)) error {
			return runCardUpdate(ctx, stdout, stderr, issuesDir, id, what, false,
				func(env *trackerEnv, card tracker.Record, _ string) ([]byte, error) {
					return editFlakeTests(env, card, edit)
				})
		},
		newIssue: func(ctx context.Context, stdout, stderr io.Writer, title, body string) (int, error) {
			return fileFlakeIssue(ctx, stdout, stderr, issuesDir, title, body)
		},
	}
}

type flakeRegisterFlags struct {
	Issue    int
	Tests    []string
	Evidence string
}

func newFlakeCmd() *cobra.Command {
	issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")
	cmd := &cobra.Command{
		Use:   "flake",
		Short: "Register known flaky tests on their issues, and list the daily flake view (#394)",
		Long: `Known-flake landing (#394). A test is a known flake while an open issue's card
lists it in flake_tests. When a recorded suite (local-check run, or merge's
integration suite) fails only on known flakes, it reruns just those tests on
the same head; a pass there is landing evidence, linked to the issue. Any other
failure blocks. Closing the issue unregisters its tests.

  sdlc flake register --test PKG.TestX[/sub…] [--issue N] [--evidence TEXT]
  sdlc flake unregister --issue N --test NAME
  sdlc flake narrow --issue N --test OLD --to OLD/sub…
  sdlc flake list [-v]

A name may be an exact subtest (#466), as go test prints it:
PKG.TestX/plain/dev-aliases.test.sh. It explains only failures of itself and
its own subtests, never a sibling's; its rerun is its whole top-level test.
Rebuild :0 (and refresh the slots) before registering a subtest name: an
older sdlc reads one as malformed, and its known-flake rule is then off
(every failure blocks). It also refuses a test-stats attempt whose
known_flakes names a subtest, so its success guard fails closed.`,
		SilenceUsage: true,
	}
	f := flakeRegisterFlags{}
	register := markMutatingCommand(&cobra.Command{
		Use:   "register --test PKG.TestName[/sub…]... [--issue N]",
		Short: "Register a test as a known flake, given same-head fail→pass evidence",
		Long: `Register tests as known flakes on a flake issue's card (flake_tests).

Each test needs same-exact-head fail→pass evidence (#390): an attempt on this
machine's test-stats store where it failed, followed by a passing suite on the
same commit. A subtest name (#466) needs its top-level test's evidence and the
failed attempt's saved log showing a failing subtest the name covers; rebuild
:0 before registering one (an older sdlc's registry refuses the name). A failure seen elsewhere (a CI run) can be registered with
--evidence TEXT naming the runs. A newly filed issue keeps the text in its
body; with --issue N it is printed for N's Log, since only an issue's owner
publishes its details. Store evidence (attempt IDs) stays in test-stats.

--issue N adds the tests to issue N's card. Any slot may (a card setter has no
owner). Without --issue, a new unclaimed flake issue is filed with the tests
and the evidence, for the TL to dispatch.

A test already registered to another open issue refuses: one owner per test.
That check reads the registry at registration time. Two registrations racing
on different issues can still both land, but the next registry read then
refuses the test as ambiguous, so the known-flake rule is unavailable (every
failure blocks) until one issue drops it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlakeRegister(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), f, defaultFlakeDeps(issuesDir))
		},
	})
	register.Flags().IntVar(&f.Issue, "issue", 0, "the flake issue to register on (default: file a new unclaimed one)")
	register.Flags().StringArrayVar(&f.Tests, "test", nil, "an exact test or subtest name, <import path>.<TestName>[/<sub>…] (repeatable)")
	register.Flags().StringVar(&f.Evidence, "evidence", "", "evidence from outside this machine's test-stats store (a CI run)")
	var verbose bool
	list := &cobra.Command{
		Use:   "list",
		Short: "The daily flake view: the registry, unregistered same-head flakes, unexplained failures",
		Long: `List the open known-flake registry and same-head fail→pass failures.
When one closed issue explicitly covers a test and its landed merge is absent
from the failing head, the default view hides that already-fixed failure.
Use -v to see it with its issue and merge. A failure after the merge is shown
as a recurrence. Ambiguous or unreadable evidence never hides a failure.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlakeListMode(cmd.Context(), cmd.OutOrStdout(), defaultFlakeDeps(issuesDir), verbose)
		},
	}
	var edit flakeEditFlags
	unregister := markMutatingCommand(&cobra.Command{
		Use:   "unregister --issue N --test NAME",
		Short: "Remove one registered name from a flake issue's card",
		Long: `Remove one name from issue N's flake_tests (#466): a compare-and-swap on the
card's current list, so a registration racing this one is never overwritten.
Only the card's owner may, or anyone while the issue is unclaimed (the rule
for publishing its details). A name already absent writes nothing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlakeEdit(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), edit, false, defaultFlakeDeps(issuesDir))
		},
	})
	narrow := markMutatingCommand(&cobra.Command{
		Use:   "narrow --issue N --test OLD --to OLD/sub…",
		Short: "Replace a registered name with one of its subtests",
		Long: `Replace OLD with TO on issue N's flake_tests (#466), where TO is one of OLD's
own subtests: the registration then explains less, so it needs no new
evidence. Use it to correct a registration that covers too much, such as a
whole test registered for one flaky subtest. A compare-and-swap on the card's
current list; the card's owner only, or anyone while the issue is unclaimed.
Already narrowed writes nothing. Rebuild :0 before narrowing to a subtest
(an older sdlc's registry refuses the name).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlakeEdit(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), edit, true, defaultFlakeDeps(issuesDir))
		},
	})
	for _, c := range []*cobra.Command{unregister, narrow} {
		c.Flags().IntVar(&edit.Issue, "issue", 0, "the flake issue whose card to change (required)")
		c.Flags().StringVar(&edit.Test, "test", "", "the registered name (required)")
	}
	narrow.Flags().StringVar(&edit.To, "to", "", "the subtest of --test to register instead (required)")
	list.Flags().BoolVarP(&verbose, "verbose", "v", false, "include flakes fixed by landed issues")
	cmd.AddCommand(register, unregister, narrow, list)
	return cmd
}

type flakeEditFlags struct {
	Issue    int
	Test, To string
}

// runFlakeEdit is unregister (narrow false) and narrow: validate the names,
// then the owner-only compare-and-swap edit of the card.
func runFlakeEdit(ctx context.Context, stdout, stderr io.Writer, f flakeEditFlags, narrow bool, d flakeDeps) error {
	verb := "flake unregister"
	if narrow {
		verb = "flake narrow"
	}
	if f.Issue <= 0 {
		return fmt.Errorf("%s: give --issue N", verb)
	}
	if err := flakes.ValidTest(f.Test); err != nil {
		return fmt.Errorf("%s: --test: %w", verb, err)
	}
	if !narrow {
		return wrapErr(verb, d.editTests(ctx, stdout, stderr, f.Issue, "flake_tests − "+f.Test,
			func(id, status string, have []string) ([]string, error) {
				return removeFlakeTests(id, status, have, f.Test)
			}))
	}
	if err := flakes.ValidTest(f.To); err != nil {
		return fmt.Errorf("%s: --to: %w", verb, err)
	}
	return wrapErr(verb, d.editTests(ctx, stdout, stderr, f.Issue, "flake_tests "+f.Test+" → "+f.To,
		func(id, status string, have []string) ([]string, error) {
			return narrowFlakeTests(id, status, have, f.Test, f.To)
		}))
}

func wrapErr(verb string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", verb, err)
	}
	return nil
}

// flakeEvidence is the registration proof for each test, from the store. A
// subtest (#466) needs its top-level test's same-head fail→pass and the
// failed attempt's saved log showing a failing subtest the name covers:
// records keep top-level failures, so the log is where the subtest is. A
// missing or unreadable log is no evidence.
func flakeEvidence(attempts []SuiteAttempt, repo string, tests []string, subtestsOf func(SuiteAttempt) ([]string, error)) (proof map[string][2]string, missing []string) {
	var mine []SuiteAttempt
	for _, a := range attempts {
		if a.Repo == repo && isFlakeEvidence(a) {
			mine = append(mine, a)
		}
	}
	all := flakeAttempts(mine)
	proof = map[string][2]string{}
	for _, t := range tests {
		n, _ := flakes.ParseName(t)
		f, p, ok := flakes.SameHeadFailPass(all, n.Top())
		if ok && n.Sub != "" {
			ok = subtestFailedIn(mine, f, t, subtestsOf)
		}
		if ok {
			proof[t] = [2]string{f, p}
		} else {
			missing = append(missing, t)
		}
	}
	return proof, missing
}

// subtestFailedIn reports whether attempt id's log shows a failing leaf
// subtest that name covers.
func subtestFailedIn(attempts []SuiteAttempt, id, name string, subtestsOf func(SuiteAttempt) ([]string, error)) bool {
	i := slices.IndexFunc(attempts, func(a SuiteAttempt) bool { return a.ID == id })
	if i < 0 || subtestsOf == nil {
		return false
	}
	subs, err := subtestsOf(attempts[i])
	if err != nil {
		return false
	}
	n, _ := flakes.ParseName(name)
	for _, leaf := range flakes.Leaves(n.Top(), subs) {
		if flakes.Covers(name, leaf) {
			return true
		}
	}
	return false
}

func runFlakeRegister(ctx context.Context, stdout, stderr io.Writer, f flakeRegisterFlags, d flakeDeps) error {
	if len(f.Tests) == 0 {
		return errors.New("flake register: give --test")
	}
	tests := slices.Clone(f.Tests)
	slices.Sort(tests)
	tests = slices.Compact(tests)
	for _, t := range tests {
		if err := flakes.ValidTest(t); err != nil {
			return fmt.Errorf("flake register: %w", err)
		}
	}
	repo, err := d.repo()
	if err != nil {
		return err
	}
	attempts, warnings, err := d.attempts()
	if err != nil {
		return fmt.Errorf("flake register: read test-stats: %w", err)
	}
	for _, w := range warnings {
		cwarn(stderr, "test-stats: "+w)
	}
	proof, missing := flakeEvidence(attempts, repo, tests, d.subtests)
	if len(missing) > 0 && f.Evidence == "" {
		return fmt.Errorf("flake register: no same-head fail→pass evidence in this machine's test-stats for %s.\n"+
			"  A flake is registered only once the same commit failed on it and then passed (#390).\n"+
			"  Rerun the suite on the failing head, or give --evidence naming the runs elsewhere (CI)", strings.Join(missing, ", "))
	}
	reg, _, err := d.registry(ctx)
	if err != nil {
		return fmt.Errorf("flake register: %w", err)
	}
	id := fmt.Sprintf("%06d", f.Issue)
	filed := f.Issue <= 0
	for _, t := range tests {
		for _, e := range reg.Entries() {
			switch {
			case e.Issue == id:
			case e.Test == t:
				return fmt.Errorf("flake register: %s is already registered on issue #%s; add evidence there", t, e.Issue)
			case flakes.Covers(e.Test, t) || flakes.Covers(t, e.Test):
				return fmt.Errorf("flake register: %s overlaps %s, registered on issue #%s; one issue owns a test and its subtests: add evidence there", t, e.Test, e.Issue)
			}
		}
	}
	var lines []string
	for _, t := range tests {
		if p, ok := proof[t]; ok {
			lines = append(lines, fmt.Sprintf("- %s: failed in attempt %s, passed on the same head in attempt %s", t, p[0], p[1]))
		} else {
			lines = append(lines, fmt.Sprintf("- %s: %s", t, f.Evidence))
		}
	}
	if f.Issue <= 0 {
		body := "## Problem\n\nKnown flaky tests (#394), registered with same-head fail→pass evidence:\n\n" + strings.Join(lines, "\n") +
			"\n\n## Done when\n\n- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.\n"
		n, err := d.newIssue(ctx, stdout, stderr, "flake: "+shortTestNames(tests), body)
		if err != nil {
			return fmt.Errorf("flake register: file the flake issue: %w", err)
		}
		f.Issue, id = n, fmt.Sprintf("%06d", n)
	}
	if err := d.addTests(ctx, stdout, stderr, f.Issue, tests); err != nil {
		return fmt.Errorf("flake register: %w", err)
	}
	fmt.Fprintf(stdout, "registered on #%s:\n%s\n", id, strings.Join(lines, "\n"))
	if f.Evidence != "" && f.Issue > 0 && !filed {
		fmt.Fprintf(stdout, "the --evidence text is not on the card: add it to #%s's Log (its owner publishes details)\n", id)
	}
	return nil
}

func mergeTests(have, add []string) []string {
	out := slices.Clone(have)
	for _, t := range add {
		if !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	return out
}

func shortTestNames(tests []string) string {
	var names []string
	for _, t := range tests {
		n, _ := flakes.ParseName(t)
		names = append(names, n.Run())
	}
	return strings.Join(names, ", ")
}

// runFlakeList is the TL's local daily view (#394 D3): the registry; same-head
// fail→pass failures not yet registered (file them today); and failures that
// nothing explains, oldest first. CI is not covered until #451, said so
// rather than shown as zero.
func runFlakeList(ctx context.Context, out io.Writer, d flakeDeps) error {
	return runFlakeListMode(ctx, out, d, false)
}

func runFlakeListMode(ctx context.Context, out io.Writer, d flakeDeps, verbose bool) error {
	repo, err := d.repo()
	if err != nil {
		return err
	}
	view := flakeListView{}
	var regErr error
	if d.listView != nil {
		view, regErr = d.listView(ctx)
	} else {
		view.Registry, view.Cards, regErr = d.registry(ctx)
		view.FixedErr = errors.New("fixed-issue evidence is unavailable")
	}
	reg, cards := view.Registry, view.Cards
	fmt.Fprintln(out, "registry:")
	switch {
	case regErr != nil:
		fmt.Fprintf(out, "  unavailable: %v\n", regErr)
	case len(reg.Entries()) == 0:
		fmt.Fprintln(out, "  (no open flake issue lists a test)")
	}
	byID := map[string]flakes.Card{}
	for _, c := range cards {
		byID[c.ID] = c
	}
	for _, e := range reg.Entries() {
		c := byID[e.Issue]
		fmt.Fprintf(out, "  %s  #%s (%s, owner %s, %s)\n", e.Test, e.Issue, c.Status, valueOr(c.Owner, "unclaimed"), flakeAge(c.Created, d.now()))
	}
	attempts, warnings, err := d.attempts()
	if err != nil {
		return fmt.Errorf("flake list: read test-stats: %w", err)
	}
	var mine []SuiteAttempt
	logs := map[string]SuiteAttempt{}
	for _, a := range attempts {
		if a.Repo == repo && isFlakeEvidence(a) {
			mine = append(mine, a)
			logs[a.ID] = a
		}
	}
	order := flakes.Sort(withLogSubtests(flakeAttempts(mine), mine, d.subtests))
	var unregistered, unexplained, fixedRows []string
	containsCache := map[string]struct {
		yes bool
		err error
	}{}
	contains := func(ancestor, head string) (bool, error) {
		key := ancestor + "\x00" + head
		if hit, ok := containsCache[key]; ok {
			return hit.yes, hit.err
		}
		if d.contains == nil {
			return false, errFixedGitUnknown
		}
		yes, err := d.contains(ctx, ancestor, head)
		containsCache[key] = struct {
			yes bool
			err error
		}{yes, err}
		return yes, err
	}
	seenSHA := map[string]bool{}
	confirmed := map[string]bool{} // listed once, as a flake to file, not also as unexplained
	for i, a := range order {
		if _, ok := flakes.Confirmed(order, i); ok {
			confirmed[a.ID] = true
		}
	}
	for _, a := range order {
		if confirmed[a.ID] {
			// Covered the gate's way: each failing leaf by itself or an
			// ancestor (#466), an open issue's name or a fixed one's (#479).
			for _, t := range flakes.Uncovered(a.Failures, a.Subtests, reg) {
				line := fmt.Sprintf("  %s  attempt %s on %s, %s, log %s", t, a.ID, abbrevSHA(a.SHA), a.FinishedAt.Format(time.RFC3339), d.logPath(logs[a.ID]))
				state, owner, reason := flakeUnknown, fixedFlakeIssue{}, ""
				if regErr == nil && view.FixedErr == nil {
					state, owner, reason = classifyFixedFlake(t, view.Fixed, contains, view.Main, a.SHA)
				}
				switch state {
				case flakeFixed:
					if verbose {
						fixedRows = append(fixedRows, line+fmt.Sprintf(" — fixed by #%s (merge %s)", owner.ID, abbrevSHA(owner.Merge)))
					}
				case flakeRecurred:
					unregistered = append(unregistered, line+fmt.Sprintf(" — recurred after #%s", owner.ID))
				default:
					if reason != "" {
						line += " — coverage unknown: " + reason
					}
					unregistered = append(unregistered, line)
				}
			}
		}
		if seenSHA[a.SHA] {
			continue
		}
		seenSHA[a.SHA] = true
		for _, u := range flakes.Unexplained(order, a.SHA, reg) {
			if !confirmed[u.ID] && d.now().Sub(u.FinishedAt) >= 24*time.Hour {
				unexplained = append(unexplained, fmt.Sprintf("  attempt %s on %s, %s: %s, log %s", u.ID, abbrevSHA(u.SHA), u.FinishedAt.Format(time.RFC3339), orNone(strings.Join(u.Failures, ", ")), d.logPath(logs[u.ID])))
			}
		}
	}
	sort.Strings(unregistered)
	fmt.Fprintln(out, "same-head fail→pass, not registered (file or register today):")
	printOrNone(out, unregistered)
	if verbose {
		sort.Strings(fixedRows)
		fmt.Fprintln(out, "same-head flakes fixed by landed issues:")
		printOrNone(out, fixedRows)
	}
	fmt.Fprintln(out, "unexplained failures older than a day:")
	printOrNone(out, unexplained)
	fmt.Fprintln(out, "CI failures: not covered (#451)")
	for _, w := range warnings {
		fmt.Fprintf(out, "coverage: %s\n", w)
	}
	if regErr != nil {
		fmt.Fprintln(out, "coverage: the registry is unavailable, so every failure above counts as unregistered")
	} else if view.FixedErr != nil {
		fmt.Fprintf(out, "coverage: fixed-issue evidence unavailable (%v); no flake was hidden\n", view.FixedErr)
	}
	return nil
}

func printOrNone(out io.Writer, lines []string) {
	if len(lines) == 0 {
		fmt.Fprintln(out, "  (none)")
		return
	}
	for _, l := range lines {
		fmt.Fprintln(out, l)
	}
}

// fileFlakeIssue files a new unclaimed flake issue through `sdlc issue new`'s
// one creation path, its body passed as a details file, and returns its ID.
func fileFlakeIssue(ctx context.Context, stdout, stderr io.Writer, issuesDir, title, body string) (int, error) {
	dir, err := os.MkdirTemp("", "sdlc-flake-*")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(dir)
	details := filepath.Join(dir, "details.md")
	if err := os.WriteFile(details, []byte(body), 0o600); err != nil {
		return 0, err
	}
	var printed bytes.Buffer
	f := &issueNewFlags{IssuesDir: issuesDir, HistoryDir: envOr("WF_HISTORY_DIR", "workshop/history"), Details: details, NoClaim: true}
	if err := runIssueNew(ctx, io.MultiWriter(&printed, stdout), stderr, f, []string{title}); err != nil {
		return 0, err
	}
	id, _, ok := issue.ParseFilename(filepath.Base(strings.TrimSpace(printed.String())))
	if !ok {
		return 0, fmt.Errorf("issue new printed %q, not an issue path", strings.TrimSpace(printed.String()))
	}
	return strconv.Atoi(id)
}

// flakeAge is how long a flake issue has been open, from its created date.
func flakeAge(created string, now time.Time) string {
	t, err := time.Parse("2006-01-02", created)
	if err != nil {
		return "age unknown"
	}
	return fmt.Sprintf("open %dd", int(now.Sub(t).Hours()/24))
}

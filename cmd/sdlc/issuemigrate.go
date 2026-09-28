// issuemigrate.go — `sdlc issue migrate`: the one-time cutover of a legacy
// repository to the issue tracker (#252).
//
// The command is the thin IO shell around tracker.PlanTrackerMigration: it
// gathers the inventory (pinned main, every branch's issue edits, dirty
// worktrees, legacy close anchors), prints the plan, and on --apply performs
// its two observable phases — bootstrap the tracker, then publish main's details
// conversion with the cutover marker. Each phase is recognized on retry, so an
// interrupted apply resumes; nothing is ever rolled back. --reconcile brings a
// branch created before the cutover onto the tracker's side of it.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type issueMigrateFlags struct {
	Apply, Reconcile bool
	Expect           string
}

func newIssueMigrateCmd() *cobra.Command {
	var f issueMigrateFlags
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "migrate",
		Short:         "Cut a legacy repository over to the issue tracker (dry run unless --apply)",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			guardSpineRepo(cmd.ErrOrStderr())
			return runIssueMigrate(commandContext(cmd.Context()), cmd.OutOrStdout(), cmd.ErrOrStderr(), f)
		},
	})
	cmd.Flags().BoolVar(&f.Apply, "apply", false, "perform the reviewed plan (requires --expect)")
	cmd.Flags().StringVar(&f.Expect, "expect", "", "the digest the dry run printed; --apply refuses any other plan")
	cmd.Flags().BoolVar(&f.Reconcile, "reconcile", false, "bring this pre-cutover branch's details onto the tracker")
	return cmd
}

// migrateEnv is the checkout plus an UNGUARDED tracker: the migration runs
// exactly where the cutover guard would refuse.
type migrateEnv struct {
	*trackerEnv
	tracker *tracker.Repository
	trunk   *gitx.TrunkFile // the issue-tracker branch
}

func openMigrate(ctx context.Context) (*migrateEnv, error) {
	env, err := openTrackerAt(ctx, ".")
	if err != nil {
		return nil, err
	}
	repo, err := tracker.NewRepository(ctx, env.root, env.target.Remote)
	if err != nil {
		return nil, err
	}
	tf, err := gitx.NewTrunkFileContext(ctx, env.root, env.target.Remote, vocab.Issue().Discovery().Tracker)
	if err != nil {
		return nil, err
	}
	return &migrateEnv{trackerEnv: env, tracker: repo, trunk: tf}, nil
}

func runIssueMigrate(ctx context.Context, stdout, stderr io.Writer, f issueMigrateFlags) error {
	if f.Apply && f.Reconcile {
		return errors.New("--apply migrates main; --reconcile brings a branch across afterwards — pick one")
	}
	env, err := openMigrate(ctx)
	if err != nil {
		return err
	}
	if f.Reconcile {
		return runMigrateReconcile(env, stdout, stderr)
	}
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	if _, done, err := migratedAlready(env, mainView); err != nil || done {
		if done {
			cok(stderr, "already migrated: main carries the cutover marker for this repository's issue tracker")
		}
		return err
	}
	in, err := migrationInventory(env, mainView)
	if err != nil {
		return err
	}
	m := tracker.PlanTrackerMigration(in)
	printMigrationPlan(stdout, m)
	if len(m.Refusals) > 0 {
		return fmt.Errorf("%d refusal(s) block the cutover; resolve them under the old workflow and re-run the dry run", len(m.Refusals))
	}
	if !f.Apply {
		cinfo(stderr, fmt.Sprintf("dry run. After freezing every SDLC writer: `sdlc issue migrate --apply --expect %s`", m.Digest))
		return nil
	}
	if f.Expect != m.Digest {
		return fmt.Errorf("the plan is %s, not the reviewed %q; re-run the dry run and review it", m.Digest, f.Expect)
	}
	root, err := applyTrackerBootstrap(env, m)
	if err != nil {
		return err
	}
	cok(stderr, fmt.Sprintf("issue tracker at root %s", shortOID(root)))
	if err := applyMainCutover(env, m, root); err != nil {
		return err
	}
	cok(stderr, fmt.Sprintf("main converted: %d details mirrored, cutover marker %s", len(m.Conversions), tracker.CutoverMarkerPath))
	cinfo(stderr, "next: `git pull` in every resting checkout; `sdlc issue migrate --reconcile` on each pre-cutover branch")
	return nil
}

// migratedAlready reports a completed migration and the tracker root main's
// marker names: main carries a marker whose root is the tracker's. A marker
// naming another tracker refuses.
func migratedAlready(env *migrateEnv, mainView *gitx.TrunkView) (string, bool, error) {
	root, present, err := tracker.ReadCutoverMarkerAt(env.root, mainView.Ref())
	if err != nil || !present {
		return "", false, err
	}
	exists, err := env.tracker.Initialized()
	if err != nil {
		return "", false, err
	}
	if !exists {
		return "", false, fmt.Errorf("main carries %s (tracker root %s) but %s has no issue tracker; restore it rather than migrate again", tracker.CutoverMarkerPath, shortOID(root), env.target.Remote)
	}
	view, err := env.trunk.Snapshot()
	if err != nil {
		return "", false, err
	}
	if ok, err := env.trunk.HasRoot(root, view.Ref()); err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("main's %s names tracker root %s, which the issue tracker's history does not start at", tracker.CutoverMarkerPath, shortOID(root))
		}
		return "", false, err
	}
	return root, true, nil
}

// migrationInventory gathers the plan's input from the pinned main tree, every
// local and publication-remote branch, and this clone's worktrees.
func migrationInventory(env *migrateEnv, mainView *gitx.TrunkView) (tracker.MigrationInput, error) {
	d := vocab.Issue().Discovery()
	in := tracker.MigrationInput{Repository: env.target.Repository, ObjectFormat: env.format, Main: mainView.Ref(), Anchors: map[string][]tracker.MigrationAnchor{}}
	readDir := func(dir string) ([]tracker.MigrationFile, error) {
		present, err := mainView.Exists(dir)
		if err != nil || !present {
			return nil, err
		}
		files, err := mainView.Files(dir)
		if err != nil {
			return nil, err
		}
		var out []tracker.MigrationFile
		for _, f := range files {
			if path.Dir(f.Path) != dir {
				continue
			}
			if _, _, ok := issue.ParseFilename(path.Base(f.Path)); ok {
				out = append(out, tracker.MigrationFile{Path: f.Path, Raw: f.Content})
			}
		}
		return out, nil
	}
	var err error
	if in.Active, err = readDir(d.Home); err != nil {
		return in, err
	}
	if in.Archived, err = readDir(vocab.ArchiveSubdir(d.Archive, vocab.ArchiveIssues)); err != nil {
		return in, err
	}
	refs, err := migrationBranchRefs(env)
	if err != nil {
		return in, err
	}
	for _, ref := range refs {
		files, err := branchIssueEdits(env, ref, mainView.Ref(), d.Home)
		if err != nil {
			return in, err
		}
		in.Branches = append(in.Branches, files...)
	}
	if in.DirtyIssuePaths, err = dirtyIssuePaths(env, d.Home); err != nil {
		return in, err
	}
	for _, f := range in.Active {
		fm, _, perr := issue.Parse(string(f.Raw))
		if status, _ := issue.GetField(fm, "status"); perr != nil || status != "codecomplete" {
			continue
		}
		id, _, _ := issue.ParseFilename(path.Base(f.Path))
		anchors, err := legacyCloseAnchors(env, mainView.Ref(), refs, f.Path)
		if err != nil {
			return in, err
		}
		in.Anchors[id] = anchors
	}
	return in, nil
}

// migrationBranchRefs lists local branches and the publication remote's
// branches, except the resting trunks and the tracker itself.
func migrationBranchRefs(env *migrateEnv) ([]string, error) {
	remote := env.target.Remote
	out, err := env.git("for-each-ref", "--format=%(refname)", "refs/heads", "refs/remotes/"+remote)
	if err != nil {
		return nil, err
	}
	skip := map[string]bool{
		"refs/heads/main": true, "refs/heads/" + env.resting: true,
		"refs/remotes/" + remote + "/main": true, "refs/remotes/" + remote + "/HEAD": true,
		"refs/heads/" + vocab.Issue().Discovery().Tracker: true, "refs/remotes/" + remote + "/" + vocab.Issue().Discovery().Tracker: true,
	}
	var refs []string
	for _, ref := range strings.Fields(out) {
		if !skip[ref] {
			refs = append(refs, ref)
		}
	}
	sort.Strings(refs)
	return refs, nil
}

// branchIssueEdits lists the issue files ref changed relative to its merge base
// with main, with their bytes at ref (nil when ref deleted the file).
func branchIssueEdits(env *migrateEnv, ref, main, home string) ([]tracker.MigrationBranchFile, error) {
	tip, err := env.git("rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return nil, err
	}
	related, err := env.gitTest("merge-base", tip, main)
	if err != nil {
		return nil, err
	}
	if !related {
		return nil, nil // unrelated history: nothing on it can merge into main
	}
	base, err := env.git("merge-base", tip, main)
	if err != nil {
		return nil, err
	}
	names, err := env.git("diff", "--name-only", "--no-renames", "-z", base, tip, "--", home)
	if err != nil {
		return nil, err
	}
	var files []tracker.MigrationBranchFile
	for _, p := range strings.Split(names, "\x00") {
		if p == "" || path.Dir(p) != home {
			continue
		}
		if _, _, ok := issue.ParseFilename(path.Base(p)); !ok {
			continue
		}
		f := tracker.MigrationBranchFile{Branch: strings.TrimPrefix(strings.TrimPrefix(ref, "refs/heads/"), "refs/remotes/"), Path: p}
		present, err := env.has(tip, p)
		if err != nil {
			return nil, err
		}
		if present {
			if f.Raw, err = env.main.ReadAt(tip, p); err != nil {
				return nil, err
			}
		}
		files = append(files, f)
	}
	return files, nil
}

// dirtyIssuePaths lists uncommitted issue edits in every worktree of this clone.
// A worktree git itself reports prunable (its directory or .git link is gone)
// or bare holds no checkout, so nothing there can be lost: it is skipped.
func dirtyIssuePaths(env *migrateEnv, home string) ([]string, error) {
	out, err := env.git("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var dirty []string
	for _, record := range strings.Split(out, "\n\n") {
		wt, skip := "", false
		for _, line := range strings.Split(record, "\n") {
			switch {
			case strings.HasPrefix(line, "worktree "):
				wt = strings.TrimPrefix(line, "worktree ")
			case line == "bare", strings.HasPrefix(line, "prunable"):
				skip = true
			}
		}
		if wt == "" || skip {
			continue
		}
		status, err := env.git("-C", wt, "status", "--porcelain", "--untracked-files=all", "--", home)
		if err != nil {
			return nil, fmt.Errorf("read issue edits in worktree %s: %w", wt, err)
		}
		// porcelainPaths splits on whitespace: the trimmed output loses the
		// first line's leading status column, so fixed-column slicing would not.
		for _, s := range strings.Split(status, "\n") {
			path, dest := porcelainPaths(s)
			if dest != "" {
				path = dest
			}
			if path != "" {
				dirty = append(dirty, wt+": "+path)
			}
		}
	}
	return dirty, nil
}

// legacyCloseAnchors finds, on each ref carrying the issue file, the legacy
// close commit (the one that recorded codecomplete) and whether code followed.
func legacyCloseAnchors(env *migrateEnv, main string, branches []string, issuePath string) ([]tracker.MigrationAnchor, error) {
	runGit := func(args ...string) ([]byte, error) {
		out, err := env.git(args...)
		return []byte(out), err
	}
	var anchors []tracker.MigrationAnchor
	for _, ref := range append([]string{main}, branches...) {
		present, err := env.has(ref, issuePath)
		if err != nil {
			return nil, err
		}
		if !present {
			continue
		}
		anchor, err := codecompleteAnchorCommitAt(ref, issuePath, runGit)
		if err != nil || anchor == "" {
			if err != nil {
				return nil, err
			}
			continue
		}
		parent, err := env.git("rev-parse", "--verify", anchor+"^1")
		if err != nil {
			return nil, fmt.Errorf("the close %s has no parent to have reviewed: %w", shortOID(anchor), err)
		}
		onMain := ref == main
		if !onMain {
			if onMain, err = env.ancestorOf(anchor, main); err != nil {
				return nil, err
			}
		}
		// What code follows the close on this ref: for main's own record, none
		// (its later commits are other work); for a landed branch close, what
		// the branch holds that main does not; otherwise, everything after it.
		codeAfter := false
		if ref != main {
			span := anchor + ".." + ref
			if onMain {
				span = main + "..." + ref
			}
			changed, err := env.git("diff", "--name-only", "-z", span)
			if err != nil {
				return nil, err
			}
			var paths []string
			for _, p := range strings.Split(changed, "\x00") {
				if p != "" {
					paths = append(paths, p)
				}
			}
			codeAfter = publishGateHasCodeSurface(paths)
		}
		anchors = append(anchors, tracker.MigrationAnchor{Ref: ref, Anchor: anchor, Parent: parent, CodeAfter: codeAfter, OnMain: onMain})
	}
	return anchors, nil
}

func printMigrationPlan(w io.Writer, m tracker.MigrationManifest) {
	inferred := 0
	for _, c := range m.Cards {
		if len(c.Inferences) > 0 {
			inferred++
		}
	}
	fmt.Fprintf(w, "issue tracker migration for %s at main %s\n", m.Repository, shortOID(m.Main))
	fmt.Fprintf(w, "  cards: %d (%d with inferred values), details converted: %d, duplicate IDs: %d\n", len(m.Cards), inferred, len(m.Conversions), len(m.Duplicates))
	if inferred > 0 {
		fmt.Fprintln(w, "\ninferred card values (review; the details files are unchanged records):")
		for _, c := range m.Cards {
			if len(c.Inferences) > 0 {
				fmt.Fprintf(w, "  #%s %s: %s\n", c.ID, c.Source, strings.Join(c.Inferences, "; "))
			}
		}
	}
	if len(m.Duplicates) > 0 {
		fmt.Fprintln(w, "\nduplicate IDs (one card each):")
		for _, d := range m.Duplicates {
			fmt.Fprintf(w, "  %s\n", d)
		}
	}
	if len(m.Refusals) > 0 {
		fmt.Fprintln(w, "\nrefusals (the cutover cannot happen until each is resolved):")
		for _, r := range m.Refusals {
			fmt.Fprintf(w, "  %s\n      %s\n      next: %s\n", r.Subject, r.Reason, r.Next)
		}
	}
	fmt.Fprintf(w, "\ndigest: %s\n", m.Digest)
}

// applyTrackerBootstrap creates the issue tracker from the plan by
// expected-absence CAS, or accepts an existing one only when its root commit
// holds exactly the planned files (a resumed apply). It returns the root.
func applyTrackerBootstrap(env *migrateEnv, m tracker.MigrationManifest) (string, error) {
	_, err := env.trunk.Bootstrap(m.TrackerFiles(), "sdlc issue migrate "+m.Digest, func(gitx.BootstrapResult) error { return nil })
	if err != nil && !errors.Is(err, gitx.ErrBootstrapExists) {
		return "", fmt.Errorf("bootstrap the issue tracker: %w\n      re-run the same `--apply --expect %s` to resume", err, m.Digest)
	}
	view, err := env.trunk.Snapshot()
	if err != nil {
		return "", err
	}
	roots, err := env.trunk.Roots(view.Ref())
	if err != nil {
		return "", err
	}
	if len(roots) != 1 {
		return "", fmt.Errorf("the issue tracker has %d root commits; expected one", len(roots))
	}
	if err := sameTrackerFiles(env, roots[0], m); err != nil {
		return "", err
	}
	return roots[0], nil
}

// sameTrackerFiles proves the tracker's root commit holds exactly the plan's
// files (by blob identity), so a resumed apply never adopts another tracker.
func sameTrackerFiles(env *migrateEnv, root string, m tracker.MigrationManifest) error {
	listing, err := env.git("ls-tree", "-r", "-z", "--full-tree", root)
	if err != nil {
		return err
	}
	have := map[string]string{}
	for _, entry := range strings.Split(listing, "\x00") {
		meta, p, ok := strings.Cut(entry, "\t")
		if !ok {
			continue
		}
		fields := strings.Fields(meta)
		if len(fields) == 3 {
			have[p] = fields[2]
		}
	}
	want := m.TrackerFiles()
	if len(have) != len(want) {
		return fmt.Errorf("an issue tracker with a different root (%d files, plan has %d) already exists; it is not this migration's", len(have), len(want))
	}
	for p, raw := range want {
		oid, err := issue.CardBlobOID(raw, env.format)
		if err != nil {
			return err
		}
		if have[p] != oid {
			return fmt.Errorf("an issue tracker whose root differs at %s already exists; it is not this migration's", p)
		}
	}
	return nil
}

// applyMainCutover publishes the details conversion and the cutover marker in
// one commit on the pinned main (the freeze keeps main still; a moved main
// refuses rather than convert details the plan never saw).
func applyMainCutover(env *migrateEnv, m tracker.MigrationManifest, root string) error {
	write := map[string][]byte{tracker.CutoverMarkerPath: tracker.CutoverMarkerBytes(root)}
	for _, c := range m.Conversions {
		write[c.Path] = c.Raw
	}
	msg := fmt.Sprintf("migrate: issue details onto the issue tracker\n\nMigration-Digest: %s\nTracker-Root: %s", m.Digest, root)
	return env.main.UpdateMany(msg, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		if view.Ref() != m.Main {
			return gitx.TrunkWrite{}, fmt.Errorf("main moved from %s to %s since the plan; keep writers frozen and re-run the dry run.\n"+
				"      If main's movement changed no issue files, the new plan matches the tracker already in place and resumes from it;\n"+
				"      otherwise the tracker is not the new plan's: abandon it (delete the issue-tracker branch — nothing has written to it yet) and apply the new plan",
				shortOID(m.Main), shortOID(view.Ref()))
		}
		return gitx.TrunkWrite{Write: write}, nil
	})
}

// runMigrateReconcile brings a branch from before the cutover across by merging
// main's migration commit — not the rest of main — so the migration becomes an
// ancestor and later merges with main start from converted details (committing
// the same bytes without that ancestry conflicts as soon as either side's
// mirror is refreshed). Before merging it proves every details file the branch
// changed kept the imported card's fields and has a card at all; after merging
// it proves every mirror pins its imported card, undoing the merge otherwise.
func runMigrateReconcile(env *migrateEnv, stdout, stderr io.Writer) error {
	if env.branch == "" || env.onRest() || env.branch == "main" {
		return errors.New("reconcile runs on a pre-cutover issue branch; a resting checkout only needs `git pull`")
	}
	mainView, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	root, done, err := migratedAlready(env, mainView)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("main has not been migrated yet; nothing to reconcile against")
	}
	imported, err := importedCards(env, root)
	if err != nil {
		return err
	}
	migration, err := env.git("log", "-1", "--format=%H", "--diff-filter=A", mainView.Ref(), "--", tracker.CutoverMarkerPath)
	if err != nil || migration == "" {
		return fmt.Errorf("find main's migration commit (the one adding %s): %v", tracker.CutoverMarkerPath, err)
	}
	if across, err := env.ancestorOf(migration, "HEAD"); err != nil {
		return err
	} else if across {
		cok(stderr, "nothing to reconcile: this branch already contains the cutover")
		return nil
	}
	home := vocab.Issue().Discovery().Home
	if dirty, err := env.git("status", "--porcelain", "--untracked-files=no"); err != nil {
		return err
	} else if dirty != "" {
		return fmt.Errorf("commit or stash the uncommitted changes first (nothing was changed):\n%s", dirty)
	}
	base, err := env.git("merge-base", "HEAD", migration)
	if err != nil {
		return fmt.Errorf("%s shares no history with main: %w", env.branch, err)
	}
	changed, err := env.git("diff", "--name-only", "-z", base, "HEAD", "--", home)
	if err != nil {
		return err
	}
	listed, err := env.git("ls-tree", "-z", "--name-only", "HEAD", "--", home+"/")
	if err != nil {
		return err
	}
	kept := map[string]bool{} // IDs this branch still has active details for
	for _, rel := range strings.Split(listed, "\x00") {
		if id, _, ok := issue.ParseFilename(path.Base(rel)); ok {
			kept[id] = true
		}
	}
	var refusals []string
	for _, rel := range strings.Split(changed, "\x00") {
		id, _, ok := issue.ParseFilename(path.Base(rel))
		if rel == "" || !ok || path.Dir(rel) != home {
			continue
		}
		present, err := env.has("HEAD", rel)
		if err != nil {
			return err
		}
		if !present {
			card, onTracker := imported[id]
			if !onTracker {
				continue // an issue main never had: nothing to carry
			}
			closed, err := cardClosed(card)
			if err != nil {
				refusals = append(refusals, fmt.Sprintf("%s: the tracker's card for #%s is unreadable (%v); repair it before reconciling", rel, issue.CLIRef(id), err))
				continue
			}
			if !tracker.RemovalArchivesActive(kept[id], closed) {
				continue // a rename, or an issue main closed too
			}
			refusals = append(refusals, fmt.Sprintf("%s: %s (#%s is still open on the issue tracker); land the close on main, or restore the details here, then reconcile again", rel, tracker.ArchivesActiveReason, issue.CLIRef(id)))
			continue
		}
		raw, err := env.main.ReadAt(env.head, rel)
		if err != nil {
			return err
		}
		card, ok := imported[id]
		if !ok {
			refusals = append(refusals, fmt.Sprintf("%s: #%s has no card on the issue tracker (created on this branch before the cutover?); file it with `sdlc issue new` and move this content there", rel, issue.CLIRef(id)))
			continue
		}
		if _, err := issue.ReconcileLegacyDetails(raw, card, env.format); err != nil {
			refusals = append(refusals, fmt.Sprintf("%s: %v.\n"+
				"      This branch changed card fields before the cutover. If the change is wanted, apply it on the card\n"+
				"      from a caught-up checkout (main, after `git pull`) with its verb — claim, set-status, set-title,\n"+
				"      set-estimate, set-github — then revert the fields here to main's values, commit, and reconcile again", rel, err))
		}
	}
	if len(refusals) > 0 {
		return fmt.Errorf("cannot reconcile this branch (nothing was changed):\n  %s", strings.Join(refusals, "\n  "))
	}
	before := env.head
	msg := fmt.Sprintf("migrate: bring %s across the issue tracker cutover", env.branch)
	if out, err := env.git("merge", "--no-ff", "--no-edit", "-m", msg, migration); err != nil {
		if _, aerr := env.git("merge", "--abort"); aerr != nil {
			return fmt.Errorf("merging main's migration commit failed (%v), and aborting that merge failed too: %v\n"+
				"      the checkout is mid-merge: inspect `git status`, then `git merge --abort` by hand", err, aerr)
		}
		return fmt.Errorf("merging main's migration commit conflicts (nothing was changed):\n%s\n"+
			"      Merge origin/main into this branch yourself, resolve the conflicts, and commit; that brings it across too", out)
	}
	if err := verifyMigratedMirrors(env, home, imported); err != nil {
		if _, rerr := env.git("reset", "-q", "--hard", before); rerr != nil {
			return fmt.Errorf("%v\n      undoing the merge failed: %v — the branch holds the merge commit; reset it to %s by hand", err, rerr, shortOID(before))
		}
		return fmt.Errorf("%v\n      the merge was undone; nothing was changed", err)
	}
	cok(stderr, fmt.Sprintf("%s is across the cutover (merged main's migration commit %s)", env.branch, shortOID(migration)))
	fmt.Fprintln(stdout, migration)
	return nil
}

// verifyMigratedMirrors proves every details file in the checkout mirrors the
// card the migration imported for it, with no card-owned field changed: a
// pre-cutover branch can only have gained its mirrors from that merge.
func verifyMigratedMirrors(env *migrateEnv, home string, imported map[string][]byte) error {
	entries, err := os.ReadDir(filepath.Join(env.root, filepath.FromSlash(home)))
	if err != nil {
		return err
	}
	for _, e := range entries {
		id, _, ok := issue.ParseFilename(e.Name())
		if !ok || !e.Type().IsRegular() {
			continue
		}
		rel := path.Join(home, e.Name())
		raw, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		card, carded := imported[id]
		if !issue.HasMirror(raw) || !carded {
			return fmt.Errorf("%s is not a mirrored tracker issue after the merge", rel)
		}
		if _, err := issue.RefreshMirror(raw, card, card); err != nil {
			return fmt.Errorf("%s does not mirror its imported card: %v", rel, err)
		}
	}
	return nil
}

// importedCards reads the cards the migration bootstrapped (the tracker's
// root commit), by ID: the projection every pre-cutover branch was based on.
func importedCards(env *migrateEnv, root string) (map[string][]byte, error) {
	cardsDir := vocab.Issue().Discovery().Cards
	listing, err := env.git("ls-tree", "-z", "--name-only", root, "--", cardsDir+"/")
	if err != nil {
		return nil, err
	}
	cards := map[string][]byte{}
	for _, p := range strings.Split(listing, "\x00") {
		id, _, ok := issue.ParseFilename(path.Base(p))
		if p == "" || !ok {
			continue
		}
		raw, err := env.trunk.ReadAt(root, p)
		if err != nil {
			return nil, err
		}
		cards[id] = raw
	}
	return cards, nil
}

// cardClosed reports whether an imported card's status is terminal.
func cardClosed(card []byte) (bool, error) {
	parsed, err := issue.ParseCard(card)
	if err != nil {
		return false, err
	}
	status, _ := issue.GetField(parsed.Frontmatter, "status")
	return vocab.Issue().IsTerminal(status), nil
}

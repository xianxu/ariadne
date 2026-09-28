// legacymode.go — the pre-#252 issue workflow, for repositories without an
// issue tracker (#252). The #252 binary ships before any repository cuts over,
// so every repository without a tracker must keep working exactly as before:
// the verbs M2 moved onto the tracker (issue new, claim, set-status) dispatch
// here, restored verbatim from pre-252-freeze (the #252 fork point); the new
// tracker-only verbs refuse with the legacy way. Deleted once every fleet
// repository has cut over.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// repositoryTracked reports whether the repository containing dir has an issue
// tracker on its publication remote — the one mode decision, made the way the
// readers make it (LoadRecords): no publication target and nothing fetched, or
// a remote without the tracker branch, is a legacy repository. A checkout whose
// cutover marker names a tracker the remote lacks is an error, never legacy.
//
// An unreachable remote falls back to local evidence (Repository.Presence, the
// same decision the readers make): tracked refuses on the transport error, since
// the verb needs the tracker; legacy keeps a legacy repository's local-only
// verbs working offline, as they did before #252.
func repositoryTracked(ctx context.Context, dir string) (bool, error) {
	root := repoRootOf(dir)
	if root == "" {
		return false, nil
	}
	repo, err := tracker.RepositoryForCheckout(ctx, root)
	if err != nil || repo == nil {
		return false, err
	}
	tracked, _, err := repo.Presence()
	return tracked, err
}

// errLegacyOnly is how a tracker-only verb refuses in a legacy repository.
func errLegacyOnly(verb, legacyWay string) error {
	return fmt.Errorf("%s needs the issue tracker, and this repository has not cut over yet (no issue-tracker branch).\n      In a legacy repository: %s", verb, legacyWay)
}

// runLegacyIssueNew is the entry point for `sdlc issue new`. Hard guardrail
// failures call die(); the happy path prints the created path to stdout.
func runLegacyIssueNew(stdout, stderr io.Writer, f *issueNewFlags, args []string) error {
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

	// Resolve the issue directory ONCE, against the repo top level, and use the
	// resolved location for every step below — allocation, the file write, and
	// the sync pathspec (#213 BR-25). Left cwd-relative, `sdlc issue new` from a
	// subdirectory allocated against an EMPTY id space (docs/sub/workshop/issues
	// does not exist, so both ls-tree and ReadDir truthfully answered nothing),
	// then wrote the file there and pushed it — where no gate looks, and where
	// the natural repair manufactures exactly the collision this issue exists to
	// prevent. Reading and writing must agree on where ids live.
	// Only the WRITE takes the absolute form. f.IssuesDir stays as given, because
	// it also feeds the sync (#213 BR-26). Since #207 BR-16 the sync pins its own
	// git calls to the repo root, so neither form is silently blind from a
	// subdirectory — which is what broke #82's guarantee here before.
	// allocateIssueID resolves for itself, so it needs no help here.
	writeDir, shownDir := f.IssuesDir, f.IssuesDir
	if dirs, derr := resolveIDDirs(f.IssuesDir, f.HistoryDir); derr == nil {
		writeDir, shownDir = dirs.Abs[0], dirs.Rel[0]
	}

	nextID, err := allocateIssueID(stderr, f.IssuesDir, f.HistoryDir, claimRunner)
	if err != nil {
		die(stderr, err.Error())
	}

	today := time.Now().Format("2006-01-02")
	name := fmt.Sprintf("%s-%s.md", nextID, slug)
	dest := filepath.Join(writeDir, name)
	// Reported repo-relative: that is what every gate, commit and human means by
	// an issue path, and from the repo top it is byte-identical to the old output.
	shown := filepath.ToSlash(filepath.Join(shownDir, name))
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
		fmt.Fprintf(stdout, "Would create: %s\n", shown)
		fmt.Fprintln(stdout, "─── body ───")
		fmt.Fprint(stdout, rendered)
		return nil
	}

	if err := os.MkdirAll(writeDir, 0o755); err != nil {
		die(stderr, fmt.Sprintf("mkdir %s: %v", shownDir, err))
	}
	if err := os.WriteFile(dest, []byte(rendered), 0o644); err != nil {
		die(stderr, fmt.Sprintf("write %s: %v", dest, err))
	}

	created := fmt.Sprintf("Created %s", shown)
	if ghNum != "" {
		created += fmt.Sprintf(" (GitHub #%s)", ghNum)
	}
	cok(stderr, created)

	// #82 M1: broadcast the new issue to origin/main immediately, so a freshly
	// filed (base) issue is tracker state on main — not untracked working-tree
	// residue that every symlinked derivative reads and that gates trip over.
	// Reserves only this new issue against fresh remote IDs with the `--issue` filter
	// (rides #80's filtered add — unrelated untracked files stay put). nextID is
	// a zero-padded string ("000083"); claimFlags.Issue is an int.
	if id, perr := strconv.Atoi(nextID); perr == nil {
		syncFlags := &claimFlags{Issue: id, IssuesDir: f.IssuesDir, HistoryDir: f.HistoryDir, NoStart: true, FirstPublication: true}
		// Route the sync's stdout to stderr: its machine "synced" marker must not
		// pollute `issue new`'s stdout contract (the created path, printed below).
		// "" keeps issue new's historical subject ("issue-sync: update issues");
		// naming the issue is `sdlc issue sync`'s job, not creation's.
		serr := syncIssuesToMain(stderr, stderr, syncFlags, claimRunner, "")
		if serr == nil && len(syncFlags.Reallocations) > 0 {
			// The publisher re-allocated, so the file named above is gone: stdout
			// is the CREATED PATH contract, and printing a path finish() deleted
			// hands callers a name that does not exist (#207 BR-1).
			rc := syncFlags.Reallocations[0]
			shown = filepath.Join(filepath.Dir(shown), filepath.Base(rc.NewPath))
			cok(stderr, fmt.Sprintf("id %06d was taken on the trunk — filed as %06d instead: %s",
				rc.OldID, rc.NewID, shown))
		}
		if serr == nil {
			local := *syncFlags
			local.NoPush = true
			if len(syncFlags.Reallocations) > 0 {
				local.Issue = syncFlags.Reallocations[0].NewID
			}
			if err := syncIssuesToMain(stderr, stderr, &local, claimRunner, issueSyncMessage(local.Issue, "new issue")); err != nil {
				cwarn(stderr, fmt.Sprintf("reservation published but local commit failed: %v", err))
			}
		}

		if errors.Is(serr, gitx.ErrPublicationUncertain) {
			cwarn(stderr, fmt.Sprintf("reservation outcome uncertain; source and candidate files preserved without committing a rejected identity. Inspect origin/main before retrying: %v", serr))
		} else if serr != nil {
			// Best-effort: the file is already written + reported above, so a sync
			// failure (offline, no reachable origin, conflict) must not abort the
			// create — just surface it. `claim` treats the same error as fatal.
			//
			// But fall back to a LOCAL commit first (#206). Publication and
			// durability are separable, and only publication failed here; leaving
			// the new issue as an untracked working-tree file is the hole this
			// issue exists to close. Since #207 the common trigger is a genuinely
			// unreachable origin rather than a missing worktree on main — the
			// trunk route needs no checkout. A no-op when the first attempt
			// already committed and only the push failed.
			syncFlags.NoPush = true
			if lerr := syncIssuesToMain(stderr, stderr, syncFlags, claimRunner, issueSyncMessage(id, "new issue")); lerr != nil {
				cwarn(stderr, fmt.Sprintf("issue created but NOT committed: %v (sync to main also failed: %v)", lerr, serr))
			} else if errors.Is(serr, errIDTaken) {
				cwarn(stderr, fmt.Sprintf("issue committed locally but NOT broadcast: %v\n"+
					"      `sdlc issue sync --push` will refuse for the same reason — an id already on the\n"+
					"      trunk is never renumbered (ariadne#188). Delete this file and re-run `sdlc issue\n"+
					"      new`, or rename it AND its `id:` frontmatter to a free id.", serr))
			} else {
				source := gitx.Capture("rev-parse", "HEAD")
				cwarn(stderr, fmt.Sprintf("issue committed locally but not broadcast to main: %v\n"+
					"      verify the ID is still free, then publish with `sdlc issue publish --commit %s`", serr, source))
			}
		}
	}

	fmt.Fprintln(stdout, shown)
	return nil
}

// runLegacyClaim reserves from remote bytes before reconciling local metadata.
func runLegacyClaim(stdout, stderr io.Writer, f *claimFlags) error {
	if f.Issue <= 0 || f.NoStart {
		return fmt.Errorf("claim requires --issue N and an open remote issue; publish documentation with sdlc issue publish --commit SHA")
	}
	paths, err := resolveSyncPaths(f)
	if err != nil {
		return err
	}
	pub, err := newTrunkPublisher(paths.Root)
	if err != nil {
		return err
	}
	var localPath string
	var before []byte
	matches := issueFilesForID(paths.Root, f.IssuesDir, f.Issue)
	if len(matches) > 1 {
		return fmt.Errorf("multiple local issue files for #%d", f.Issue)
	}
	if len(matches) == 1 {
		localPath = matches[0]
		info, err := os.Lstat(localPath)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("local issue is not an ordinary file: %s", localPath)
		}
		before, err = os.ReadFile(localPath)
		if err != nil {
			return err
		}
	}
	var claimed []byte
	dry := errors.New("claim dry run")
	err = pub.UpdateMany(fmt.Sprintf("#%d: issue-sync: claim", f.Issue), func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		space, err := refIDSpace(v.Ref(), paths.Dirs, claimRunner)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		remotePaths := space[f.Issue]
		if len(remotePaths) != 1 {
			return gitx.TrunkWrite{}, fmt.Errorf("origin/main must contain exactly one issue #%d; found %d", f.Issue, len(remotePaths))
		}
		if localPath != "" && filepath.Base(localPath) != filepath.Base(remotePaths[0]) {
			return gitx.TrunkWrite{}, fmt.Errorf("issue #%d has different local and remote names (%s, %s); reconcile identity before claiming", f.Issue, filepath.Base(localPath), remotePaths[0])
		}
		raw, err := v.Read(remotePaths[0])
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		claimed, err = claimDecision(raw, f.Issue, time.Now().Format("2006-01-02"), startedClock())
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if f.DryRun {
			return gitx.TrunkWrite{}, dry
		}
		return gitx.TrunkWrite{Write: map[string][]byte{remotePaths[0]: claimed}}, nil
	})
	if errors.Is(err, dry) {
		cinfo(stderr, "dry-run — remote issue is open; would reserve it")
		return nil
	}
	if err != nil {
		return err
	}
	if localPath != "" {
		current, readErr := os.ReadFile(localPath)
		if readErr != nil || !bytes.Equal(current, before) {
			cwarn(stderr, "claim published; local issue changed meanwhile — reconcile status from origin/main")
		} else {
			fm, body, parseErr := issue.Parse(string(before))
			remoteFM, _, _ := issue.Parse(string(claimed))
			if parseErr != nil {
				cwarn(stderr, "claim published; local issue could not be parsed — reconcile status from origin/main")
			} else {
				for _, field := range []string{"status", "updated", "started"} {
					value, ok := issue.GetField(remoteFM, field)
					if ok {
						fm = issue.SetField(fm, field, value)
					}
				}
				if err := os.WriteFile(localPath, []byte(issue.Compose(fm, body)), 0644); err != nil {
					return fmt.Errorf("claim published, but local reconciliation failed: %w", err)
				}
			}
		}
	}
	cok(stderr, fmt.Sprintf("Issue #%d reserved on origin/main (open → working).", f.Issue))
	fmt.Fprintln(stdout, "claimed")
	return nil
}

// runLegacySetStatus is the entry point for the cobra RunE.
func runLegacySetStatus(stdout, stderr io.Writer, f *setStatusFlags) error {
	path, prev, changed, err := applyLegacyStatus(f.IssuesDir, f.Issue, f.Status, f.Force, f.DryRun)
	if err != nil {
		die(stderr, err.Error())
	}

	// #122 M4: when --force masked the lifecycle gate on an illegal transition,
	// log the override — the escape hatch is explicit and recorded, not silent.
	if f.Force && prev != "" && prev != f.Status && !vocab.Issue().CanTransition(prev, f.Status) {
		cwarn(stderr, fmt.Sprintf("--force: overriding illegal transition %s → %s (not in the lifecycle)", prev, f.Status))
	}

	// No-op when already at the target status (after guards). applyStatus
	// still bumps `updated:` so commits show intent — match `sdlc close`'s
	// posture of always emitting a `updated:` line.
	if prev == f.Status {
		cwarn(stderr, fmt.Sprintf("status already '%s'; updating timestamp only", f.Status))
	}

	if f.DryRun {
		cinfo(stderr, "dry-run — no files written")
		fmt.Fprintf(stdout, "Would update %s: status %s → %s, updated %s\n",
			filepath.Base(path), valueOr(prev, "(unset)"), f.Status, time.Now().Format("2006-01-02"))
		return nil
	}
	if !changed {
		cok(stderr, fmt.Sprintf("no changes to %s", filepath.Base(path)))
		return nil
	}
	cok(stderr, fmt.Sprintf("%s: status %s → %s", filepath.Base(path),
		valueOr(prev, "(unset)"), f.Status))
	fmt.Fprintln(stdout, path)
	return nil
}

// applyLegacyStatus locates issue <id>, enforces the transition guards (unless
// force), and rewrites its status: + updated: frontmatter. On dryRun it
// computes the change without writing. Returns the file path, the previous
// status, and whether the content would change.
//
// Extracted from runSetStatus so `sdlc claim` can fold the → working flip
// into its sync (AGENTS.md §0: one command to claim + start work) without
// duplicating the guard logic or the frontmatter rewrite. Returns errors
// rather than die()-ing so callers compose it; runSetStatus translates the
// error back into a top-level die().

func applyLegacyStatus(issuesDir string, issueID int, status string, force, dryRun bool) (path, prev string, changed bool, err error) {
	if !isValidStatus(status) {
		return "", "", false, fmt.Errorf("invalid status %q (valid: %s)", status, strings.Join(vocab.Issue().AllStatuses(), ", "))
	}
	path, err = locateIssueFile(issuesDir, issueID)
	if err != nil {
		return "", "", false, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return path, "", false, fmt.Errorf("read %s: %v", path, err)
	}
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		return path, "", false, fmt.Errorf("parse frontmatter from %s: %v", path, err)
	}
	prev, _ = issue.GetField(fm, "status")

	if !force {
		if gErr := checkTransitionGuards(prev, status, fm, body); gErr != nil {
			return path, prev, false, gErr
		}
	}

	today := time.Now().Format("2006-01-02")
	newFM := issue.SetField(fm, "status", status)
	newFM = issue.SetField(newFM, "updated", today)
	// #116: stamp `started:` on the open→working transition — the explicit, robust
	// active-time window anchor that supersedes the WorkingTransitionISO git-log
	// heuristic. Local-offset RFC3339 (startedClock) to stay lexically comparable
	// with git's %aI author dates that windowStart compares. Idempotent: never
	// overwrite an existing stamp (re-claim / re-flip must not move the anchor).
	// #122: "open" membership reads from the model; "working" stays literal — it's
	// the specific claim target (the open→working edge stamps started), not a category.
	if vocab.Issue().IsOpen(prev) && status == "working" {
		if cur, _ := issue.GetField(newFM, "started"); strings.TrimSpace(cur) == "" {
			newFM = issue.SetField(newFM, "started", startedClock())
		}
	}
	newText := issue.Compose(newFM, body)
	changed = newText != string(raw)

	if dryRun || !changed {
		return path, prev, changed, nil
	}
	if err := os.WriteFile(path, []byte(newText), 0o644); err != nil {
		return path, prev, changed, fmt.Errorf("write %s: %v", path, err)
	}
	return path, prev, changed, nil
}

// syncLegacyIssue checkpoints the resolved issue locally and publishes only the new
// narrow commit through the same adapter in every checkout. Previously committed
// design packages require explicit issue publish --commit selection. No new
// issue commit is a no-op, never permission to infer an older publication unit.
// Publication remains best-effort: failure preserves the local checkpoint and
// reports the selected SHA so implementation can proceed without losing it.
func syncLegacyIssue(stderr io.Writer, f *changeCodeFlags, issuePath string) {
	// A --name branch can point at a file outside the NNNNNN- convention; there
	// is no id to name in the commit subject, so there is nothing to sync.
	id := issueIDFromPath(issuePath)
	if id == 0 {
		return
	}
	// DryRun is threaded even though runChangeCode returns before reaching here
	// under --dry-run: a helper that commits must not depend on a caller's early
	// return for its dry-run correctness.
	syncFlags := &claimFlags{
		Issue: id, IssuesDir: f.IssuesDir, NoStart: true,
		DryRun: f.DryRun, AllowNoChanges: true,
	}
	msg := issueSyncMessage(id, "spec/plan at change-code")
	if err := syncIssuesToMain(stderr, stderr, syncFlags, changeCodeRunner, msg); err != nil {
		cwarn(stderr, fmt.Sprintf("issue file not synced: %v\n"+
			"      the gates passed and the branch is being created anyway;\n"+
			"      preserve the local commit and use the explicit publication retry above", err))
	}
}

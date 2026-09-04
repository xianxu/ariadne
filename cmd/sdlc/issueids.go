// issueids.go — allocating a new issue id against the PUBLISHED id space (#213).
//
// The defect this exists to fix: `issue.NextID` used to scan the local
// workshop/issues/ only, so a feature branch cut before some issue landed on the
// trunk has a directory that never contained it — and `sdlc issue new` on that
// branch reallocates a published id. Not a race: it reproduces days later,
// because the branch simply does not contain the newer files.
//
// Nothing downstream catches it. Two colliding issues get different slugs, so
// they are different PATHS; git merges both cleanly and no gate objects. Eight
// live collisions across two repos, the oldest from May 2026, were found this
// way — three already merged and archived.
//
// The repo transaction lock is not the answer and demonstrably did not help: the
// colliding allocations ran in linked worktrees sharing one .git/sdlc.lock and
// were serialized. A lock orders access to one view; it cannot reconcile two
// disjoint views. The fix is to read the id space from the trunk ref.
package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// trunkRef is the ref whose tree defines the published id space.
const trunkRef = "origin/main"

// allocateIssueID returns the next id, computed from the UNION of this
// checkout's issue files and the trunk's.
//
// Union, not replacement: an unpushed issue on this branch is real and must be
// excluded too, or two creations on one branch collide with each other.
//
// Offline is not a refusal. Creating an issue when the network is down is
// legitimate, so a failed fetch falls back to the local scan — but LOUDLY, with
// the risk named, because a silent fallback recreates exactly the bug this
// function exists to prevent.
//
// Every degraded read is announced, because the whole family of defects behind
// this issue reduces to one rule: a read feeding the id space must be verified
// FRESH, COMPLETE and ON-TARGET, or said out loud. A non-answer is not an empty
// answer, and unioning zero ids in from a blind read is how each of them
// silently handed out a published id.
func allocateIssueID(stderr io.Writer, issuesDir, historyDir string, r gitRunner) (string, error) {
	dirs, err := resolveIDDirs(issuesDir, historyDir)
	if err != nil {
		// The dirs and the trunk must describe the same repo; without that the
		// published read is meaningless. Fall back to a raw local scan, loudly.
		cwarn(stderr, fmt.Sprintf("id directories unusable (%v) — id allocated from LOCAL files only", err))
		byID, _, serr := issue.LocalPathsByID(issue.IDDirs(issuesDir, historyDir))
		if serr != nil {
			return "", serr
		}
		return issue.NextID(issue.IDsIn(byID)), nil
	}

	localByID, foundLocal, err := issue.LocalPathsByID(dirs.Abs)
	if err != nil {
		return "", err
	}
	local := issue.IDsIn(localByID)

	publishedByID, stale, ferr := publishedIDSpace(dirs, r)
	if ferr == nil && stale != nil {
		// The ref EXISTS but could not be refreshed, so it may be missing ids
		// published since it was last fetched — and allocating from it silently
		// is the original bug arriving through its own fix (#213 BR-23). The
		// earlier warning only fired when the ref was absent entirely, which is
		// the rarer case: a checkout that has ever fetched always has a ref.
		cwarn(stderr, fmt.Sprintf("could not refresh %s (%v) — id allocated against a POSSIBLY STALE "+
			"trunk, which may collide with an id published since the last fetch.\n"+
			"      re-run with the network up, or verify the id before pushing", trunkRef, stale))
	}
	if ferr != nil {
		cwarn(stderr, fmt.Sprintf("%s unreachable — id allocated from LOCAL files only, "+
			"which may collide with an id already published: %v\n"+
			"      if this branch predates issues on the trunk, verify the id before pushing",
			trunkRef, ferr))
		warnIfBlind(stderr, dirs, foundLocal, len(local) > 0)
		return issue.NextID(local), nil
	}
	published := issue.IDsIn(publishedByID)
	warnIfBlind(stderr, dirs, foundLocal, len(local)+len(published) > 0)
	return issue.NextID(local, published), nil
}

// warnIfBlind announces an id space that came back empty from every source.
//
// Empty is a legitimate answer in a fresh repo and a catastrophic one anywhere
// else — and the two are indistinguishable from inside (#213 BR-25). Saying so
// costs a fresh repo one accurate line and gives a misconfigured one the signal
// that used to be missing entirely: `sdlc issue new` run from a subdirectory
// read an empty trunk, allocated 000001, misfiled the issue and pushed it, with
// empty stderr.
func warnIfBlind(stderr io.Writer, dirs idDirs, foundLocal int, sawAnyID bool) {
	if sawAnyID || foundLocal > 0 {
		return
	}
	cwarn(stderr, fmt.Sprintf("no issue files found in %s (relative to %s) — allocating from an EMPTY id space.\n"+
		"      correct in a fresh repo; anywhere else the id directories are misconfigured",
		strings.Join(dirs.Rel, ", "), dirs.Top))
}

// firstLine keeps a git failure to one line: fetch errors are several lines of
// remote diagnostics, and a warning that scrolls reads as noise, not a warning.
func firstLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if ln = strings.TrimSpace(ln); ln != "" {
			return ln
		}
	}
	return "no output"
}

// publishedIDSpace reads the trunk's id space, refreshing the ref first.
//
// Reads the REF, not a checkout: `git ls-tree` needs no working tree and no
// branch switch, and filenames carry the id so no blobs are fetched.
//
// Returns (space, staleErr, err). staleErr non-nil means the ref was READ but
// not refreshed — a stale trunk is strictly better than none, but the caller
// must say so out loud rather than treat it as current.
func publishedIDSpace(dirs idDirs, r gitRunner) (map[int][]string, error, error) {
	// Fetch INTO the remote-tracking ref explicitly. `git fetch origin main`
	// updates FETCH_HEAD and only incidentally refs/remotes/origin/main — in a
	// CI checkout with no configured refspec it does not create it at all, so
	// the read below would find no ref and fall back to a baseline BR-1 proved
	// blind (#213 BR-16). Failure is still tolerated: a stale ref beats none.
	var stale error
	if out, ferr := r.Git("fetch", "--quiet", "origin", "+refs/heads/main:refs/remotes/origin/main"); ferr != nil {
		stale = fmt.Errorf("%v: %s", ferr, firstLine(string(out)))
	}
	space, err := refIDSpace(trunkRef, dirs, r)
	return space, stale, err
}

// refIDSpace reads one ref's id space: id → every path claiming it.
//
// THE reader (#213). Allocation, the merge gate and the lint verb each had
// their own ls-tree loop, and each had its own idea of what a failed read
// meant — which is why the same silent-degradation defect had to be found four
// separate times. One reader, one failure policy.
func refIDSpace(ref string, dirs idDirs, r gitRunner) (map[int][]string, error) {
	if out, err := r.Git("rev-parse", "--verify", "--quiet", ref); err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("no %s ref", ref)
	}
	byID := map[int][]string{}
	for _, dir := range dirs.Rel {
		out, err := r.Git("ls-tree", "--name-only", ref, dir+"/")
		if err != nil {
			// A PARTIAL read is indistinguishable from a complete one once the
			// paths are merged, so it would answer from half the trunk and
			// report success (#213 BR-7). A directory genuinely absent from the
			// ref is not an error — ls-tree exits 0 printing nothing — so
			// reaching here means the read itself failed.
			return nil, fmt.Errorf("ls-tree %s %s/: %v\n%s", ref, dir, err, out)
		}
		for id, paths := range issue.PathsByID(string(out)) {
			for _, p := range paths {
				if !containsPath(byID[id], p) {
					byID[id] = append(byID[id], p)
				}
			}
		}
	}
	return byID, nil
}

// mergedPathsFor computes the id→paths map the MERGE RESULT would have, from
// the three trees a merge actually involves.
//
// This is the question a merge gate has to ask, and two earlier definitions got
// it wrong (#213 close review BR-13, then its own fix):
//
//	"head has a path base lacks"     → every ARCHIVE looks like a collision.
//	                                   `sdlc merge` archives on every close, so
//	                                   this refused nearly every merge.
//	"head contradicts itself"        → correct for one tree, but blind here: the
//	                                   branch never contains the trunk's file,
//	                                   so the collision only materialises WHEN
//	                                   the trees combine.
//
// The merge result keeps everything on the trunk, adds what head added, and
// honours what head deleted:
//
//	merged(id) = (trunk(id) ∪ head(id)) − deletedByEitherSide(id)
//	deletedBySide = base(id) − side(id)      base = merge-base
//
// which is why the runner's merge-base argument is still needed even though it
// is the wrong thing to COMPARE against: it is how a deletion is recognised.
//
// BOTH sides' deletions count, symmetrically (#213 close review BR-18). Honouring
// only head's meant that while a PR was open, an issue archived on MAIN left the
// trunk carrying history/NNN-x.md and the merge-base still carrying
// issues/NNN-x.md — two survivors, falsely refused. That is not an edge case: it
// is every PR open across any close, which is most of them.
//
// Archive by either side → the old path is deleted, one survivor, pass.
// Renumber → same. Two sides each ADDING a file for one id → two survivors,
// refuse: neither deleted anything, so both are live claimants.
func mergedPathsFor(head, base, trunk map[int][]string) map[int][]string {
	merged := map[int][]string{}
	ids := map[int]bool{}
	for id := range head {
		ids[id] = true
	}
	for id := range trunk {
		ids[id] = true
	}
	for id := range ids {
		deleted := map[string]bool{}
		for _, p := range base[id] {
			if !containsPath(head[id], p) || !containsPath(trunk[id], p) {
				deleted[p] = true
			}
		}
		var out []string
		for _, p := range append(append([]string{}, trunk[id]...), head[id]...) {
			if deleted[p] || containsPath(out, p) {
				continue
			}
			out = append(out, p)
		}
		if len(out) > 0 {
			sort.Strings(out)
			merged[id] = out
		}
	}
	return merged
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// introducedCollisions returns ids the MERGE RESULT would have claimed by more
// than one path, excluding those the base already had claimed twice.
//
// Refuse only what the RANGE contributes; report what it inherited — renumbering
// an existing collision is operator work, and blocking every merge until it is
// done is worse than the bug.
//
// "Inherited" means already doubled in EITHER tree the range did not author
// (#213 close review BR-19). Testing only the merge-base blamed a PR for a
// collision that landed on main after its branch was cut: base held one path,
// the trunk held two, both survived into merged, and a PR that touched nothing
// related was refused.
func introducedCollisions(head, base, trunk map[int][]string) []int {
	merged := mergedPathsFor(head, base, trunk)
	var ids []int
	for id, paths := range merged {
		if len(paths) < 2 {
			continue
		}
		if len(base[id]) > 1 || len(trunk[id]) > 1 {
			continue // already broken without this range's help
		}
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// idDirs is where ids live, resolved ONCE for every consumer.
//
// Rel is what `git ls-tree` needs (repo-relative, slash-separated); Abs is what
// the on-disk scan needs. Both come out of the same resolution, so the local
// scan and the trunk read cannot look in different places — which they did.
type idDirs struct {
	Rel []string
	Abs []string
	Top string
}

// resolveIDDirs resolves the id directories against the REPO TOP LEVEL, not the
// process cwd (#213 BR-25).
//
// "workshop/issues" names a place in the repository, not a place relative to
// wherever the operator happens to be standing. Joined onto the cwd it yielded
// docs/sub/workshop/issues from a subdirectory — still inside the repo, so the
// containment guard below passed — and both ls-tree and os.ReadDir then
// truthfully answered "nothing" about a directory that does not exist. The
// result: `cd docs/sub && sdlc issue new` read an EMPTY published id space,
// allocated 000001, wrote the file into the subdirectory and pushed it, with
// empty stderr. All three enforcement layers were blind to it, because each
// only ever looks at the canonical dirs from the top.
//
// An ABSOLUTE dir is still honoured — that is what --issues-dir is for — but
// must resolve inside this repo: `git ls-tree` interprets a path against the
// repo root, so an outside path would make us scan THIS repo's trunk for names
// that mean something else there, reading a stranger's id space as if it were
// ours.
func resolveIDDirs(issuesDir, historyDir string) (idDirs, error) {
	top, err := gitx.RepoTopLevel()
	if err != nil {
		return idDirs{}, fmt.Errorf("not a git repo: %w", err)
	}
	// Resolve symlinks ONCE, on the root, then join — never on the dir itself.
	// EvalSymlinks fails on a path that does not exist yet, which is the normal
	// state of workshop/history in a fresh repo; the first cut fell back to the
	// UNRESOLVED absolute path there, so on macOS (/var → /private/var) it
	// compared an unresolved dir against a resolved root, decided every dir was
	// outside the repo, and silently skipped every layer of this issue's fix.
	if resolved, rerr := filepath.EvalSymlinks(top); rerr == nil {
		top = resolved
	}
	d := idDirs{Top: top}
	for _, dir := range issue.IDDirs(issuesDir, historyDir) {
		abs := dir
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(top, abs)
		} else if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
			abs = resolved
		}
		rel, rerr := filepath.Rel(top, abs)
		if rerr != nil || strings.HasPrefix(rel, "..") {
			return idDirs{}, fmt.Errorf("%s is outside the current repo (%s) — refusing to read its trunk as ours", dir, top)
		}
		d.Rel = append(d.Rel, filepath.ToSlash(rel))
		d.Abs = append(d.Abs, abs)
	}
	return d, nil
}

// renderClashes formats the merge-result collisions a range introduces, one
// report per id. Single renderer (#213 BR-23): the merge gate and the lint verb
// had a copy each, and a third was on the way.
func renderClashes(head, base, trunk map[int][]string) []string {
	merged := mergedPathsFor(head, base, trunk)
	var clashes []string
	for _, id := range introducedCollisions(head, base, trunk) {
		clashes = append(clashes, fmt.Sprintf("  #%06d would be claimed by %d files after merge:\n      %s",
			id, len(merged[id]), strings.Join(merged[id], "\n      ")))
	}
	sort.Strings(clashes)
	return clashes
}

// refuseDuplicateIssueIDs refuses a merge that would land an id already present
// on the trunk (#213).
//
// Allocation reading the trunk prevents NEW collisions; this catches the ones
// already on branches, and any created while offline. Both halves are needed:
// the fix cannot reach a branch cut before it, and the failure is silent
// everywhere else in the lifecycle.
//
// Compares by id AND path, so a file already merged to the trunk (same id, same
// path) is not mistaken for a collision — only a DIFFERENT file claiming a
// live id is.
func refuseDuplicateIssueIDs(stderr io.Writer, issuesDir, historyDir string, r gitRunner) error {
	dirs, err := resolveIDDirs(issuesDir, historyDir)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: %v", err))
		return nil
	}
	// Refresh the trunk ref first (#213 BR-4). merge's own flow has not fetched
	// at step 4.6, so a stale origin/main would miss exactly the collision that
	// landed while this branch was open — the shape of the bug.
	_, _ = r.Git("fetch", "--quiet", "origin", "+refs/heads/main:refs/remotes/origin/main")

	trunk, err := refIDSpace(trunkRef, dirs, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: %v", err))
		return nil
	}
	local, err := refIDSpace("HEAD", dirs, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: %v", err))
		return nil
	}
	// The merge-base is what distinguishes a MOVE from a new claimant, so an
	// unknown base is a NON-ANSWER, not an empty one (#213 BR-18). Defaulting it
	// to {} erased every deletion: `sdlc merge` archives on every close, so an
	// empty base made each archived file look like a second live claimant and
	// the gate refused nearly everything — a false refusal with a confident
	// message, which is worse than the collision it was hunting.
	mergeBase := strings.TrimSpace(gitx.Capture("merge-base", trunkRef, "HEAD"))
	if mergeBase == "" {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: no merge-base between %s and HEAD — "+
			"without it an archive cannot be told from a new claimant", trunkRef))
		return nil
	}
	base, berr := refIDSpace(mergeBase, dirs, r)
	if berr != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: reading merge-base %s: %v", mergeBase[:min(7, len(mergeBase))], berr))
		return nil
	}
	clashes := renderClashes(local, base, trunk)
	if len(clashes) == 0 {
		cok(stderr, "duplicate-id gate: no reused issue ids")
		return nil
	}
	return fmt.Errorf("this branch reuses %d issue id(s) already published on %s:\n%s\n"+
		"  Two files with the same id but different slugs are different PATHS, so git\n"+
		"  merges both and nothing downstream objects — rename this branch's file to a\n"+
		"  fresh id (and its `id:` frontmatter) before merging.\n"+
		"  Bypass with --no-validate only if you are deliberately landing the duplicate.",
		len(clashes), trunkRef, strings.Join(clashes, "\n"))
}

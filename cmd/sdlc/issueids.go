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
	"os"
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
func allocateIssueID(stderr io.Writer, issuesDir, historyDir string, r gitRunner) (string, error) {
	local, err := issue.ScanLocalIDs(issuesDir, historyDir)
	if err != nil {
		return "", err
	}
	published, ferr := publishedIssueIDs(issuesDir, historyDir, r)
	if ferr != nil {
		cwarn(stderr, fmt.Sprintf("%s unreachable — id allocated from LOCAL files only, "+
			"which may collide with an id already published: %v\n"+
			"      if this branch predates issues on the trunk, verify the id before pushing",
			trunkRef, ferr))
		return issue.NextID(local), nil
	}
	return issue.NextID(local, published), nil
}

// publishedIssueIDs reads the ids present in the trunk ref's issue directories.
//
// Reads the REF, not a checkout: `git ls-tree` needs no working tree and no
// branch switch, and filenames carry the id so no blobs are fetched. The fetch
// is best-effort — a repo with no origin still has a usable origin/main ref
// sometimes, and a stale ref is strictly better than none — but a missing ref is
// an error, so the caller warns rather than silently allocating locally.
func publishedIssueIDs(issuesDir, historyDir string, r gitRunner) ([]int, error) {
	// The ref and the directories must describe the SAME repo. git runs relative
	// to the process cwd while the dirs are caller-supplied, so a --issues-dir
	// pointing outside this repo would otherwise scan THIS repo's trunk for paths
	// that mean something else there — reading a stranger's id space as if it
	// were ours. Refuse that (the caller warns and falls back to local) rather
	// than allocating from it.
	dirs, err := repoRelativeIDDirs(issuesDir, historyDir)
	if err != nil {
		return nil, err
	}

	// Fetch INTO the remote-tracking ref explicitly. `git fetch origin main`
	// updates FETCH_HEAD and only incidentally refs/remotes/origin/main — in a
	// CI checkout with no configured refspec it does not create it at all, so
	// the read below would find no ref and fall back to a baseline BR-1 proved
	// blind (#213 BR-16). Failure is still tolerated: a stale ref beats none.
	_, _ = r.Git("fetch", "--quiet", "origin", "+refs/heads/main:refs/remotes/origin/main")

	if out, err := r.Git("rev-parse", "--verify", "--quiet", trunkRef); err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("no %s ref", trunkRef)
	}
	var ids []int
	for _, dir := range dirs {
		out, err := r.Git("ls-tree", "--name-only", trunkRef, dir+"/")
		if err != nil {
			// A PARTIAL read is indistinguishable from a complete one once the
			// ids are unioned, so it would allocate against half the trunk and
			// report success (#213 close review BR-7). A directory genuinely
			// absent from the ref is not an error — ls-tree exits 0 printing
			// nothing — so reaching here means the read itself failed.
			return nil, fmt.Errorf("ls-tree %s %s/: %v\n%s", trunkRef, dir, err, out)
		}
		ids = append(ids, issue.IDsInTreeListing(string(out))...)
	}
	return ids, nil
}

// introducedCollisions returns ids the head tree claims from MORE THAN ONE path
// that the base tree did not already claim twice.
//
// The definition matters and the first cut got it wrong (#213 close review
// BR-13/BR-14). A collision is not "head has a path base lacks" — that is what
// every ARCHIVE looks like, since `sdlc merge` moves
// workshop/issues/NNN-x.md → workshop/history/issues/NNN-x.md on every close.
// Measured: the gate refused a routine archive, which would have broken nearly
// every merge. A rename or renumber has the same shape.
//
// A collision is one TREE contradicting itself: two live paths claiming one id
// at the same time. Diffing that property between base and head separates
// "this range introduced it" (refuse — still cheap to rename) from
// "it was already there" (report — renumbering is operator work, and blocking
// every merge until it is done is worse than the bug).
func introducedCollisions(head, base map[int][]string) []int {
	var ids []int
	for id, paths := range head {
		if len(paths) < 2 {
			continue // one claimant: an archive, a rename, or just normal
		}
		if len(base[id]) < 2 {
			ids = append(ids, id) // base was clean here; this range broke it
		}
	}
	sort.Ints(ids)
	return ids
}

// repoRelativeIDDirs converts the id directories to paths inside the current git
// repo, refusing any that escape it. `git ls-tree` interprets a path against the
// repo root, so an absolute or outside path would silently name the wrong thing.
func repoRelativeIDDirs(issuesDir, historyDir string) ([]string, error) {
	top, err := gitx.RepoTopLevel()
	if err != nil {
		return nil, fmt.Errorf("not a git repo: %w", err)
	}
	if resolved, rerr := filepath.EvalSymlinks(top); rerr == nil {
		top = resolved
	}
	// Resolve symlinks ONCE, on the cwd, then join — never on the dir itself.
	// EvalSymlinks fails on a path that does not exist yet, which is the normal
	// state of workshop/history in a fresh repo; the first cut fell back to the
	// UNRESOLVED absolute path there, so on macOS (/var → /private/var) it
	// compared an unresolved dir against a resolved root, decided every dir was
	// outside the repo, and silently skipped every layer of this issue's fix.
	base, berr := os.Getwd()
	if berr != nil {
		return nil, berr
	}
	if resolved, rerr := filepath.EvalSymlinks(base); rerr == nil {
		base = resolved
	}
	var out []string
	for _, dir := range issue.IDDirs(issuesDir, historyDir) {
		abs := dir
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(base, abs)
		} else if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
			abs = resolved
		}
		rel, err := filepath.Rel(top, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("%s is outside the current repo (%s) — refusing to read its trunk as ours", dir, top)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	return out, nil
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
	// Refresh the trunk ref first (#213 close review BR-4). merge's own flow has
	// not fetched at step 4.6, so a stale origin/main would miss exactly the
	// collision that landed while this branch was open — the shape of the bug.
	_, _ = r.Git("fetch", "--quiet", "origin", "+refs/heads/main:refs/remotes/origin/main")

	trunk, err := issueFilesByID(trunkRef, issuesDir, historyDir, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: %v", err))
		return nil
	}
	local, err := issueFilesByID("HEAD", issuesDir, historyDir, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("duplicate-id gate skipped: %v", err))
		return nil
	}
	var clashes []string
	for _, id := range introducedCollisions(local, trunk) {
		clashes = append(clashes, fmt.Sprintf("  #%06d claimed by %d files:\n      %s",
			id, len(local[id]), strings.Join(local[id], "\n      ")))
	}
	if len(clashes) == 0 {
		cok(stderr, "duplicate-id gate: no reused issue ids")
		return nil
	}
	sort.Strings(clashes)
	return fmt.Errorf("this branch reuses %d issue id(s) already published on %s:\n%s\n"+
		"  Two files with the same id but different slugs are different PATHS, so git\n"+
		"  merges both and nothing downstream objects — rename this branch's file to a\n"+
		"  fresh id (and its `id:` frontmatter) before merging.\n"+
		"  Bypass with --no-validate only if you are deliberately landing the duplicate.",
		len(clashes), trunkRef, strings.Join(clashes, "\n"))
}

// issueFilesByID maps each id in a ref's issue directories to EVERY path
// claiming it.
//
// Every path, not the first (#213 close review BR-2). Collapsing to one made
// detection depend on `ls-tree`'s sort order — i.e. on the slug: when the head
// tree carries BOTH files (a rebased PR, or any branch that pulled main after
// the trunk file landed), head[id] could equal base[id] and nothing was
// reported. Measured on the real repo: `000213-aaa-collision.md` was refused
// while `000213-planted-collision.md`, identical but sorting later, passed.
func issueFilesByID(ref, issuesDir, historyDir string, r gitRunner) (map[int][]string, error) {
	if out, err := r.Git("rev-parse", "--verify", "--quiet", ref); err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("no %s ref", ref)
	}
	dirs, derr := repoRelativeIDDirs(issuesDir, historyDir)
	if derr != nil {
		return nil, derr
	}
	byID := map[int][]string{}
	for _, dir := range dirs {
		out, err := r.Git("ls-tree", "--name-only", ref, dir+"/")
		if err != nil {
			// Same rule as publishedIssueIDs (#213 BR-7/BR-15): a partial read
			// is indistinguishable from a clean tree once parsed, so it would
			// report "no collisions" having looked at half of them. ls-tree
			// exits 0 printing nothing for a directory absent from the ref, so
			// reaching here means the read itself failed.
			return nil, fmt.Errorf("ls-tree %s %s/: %v\n%s", ref, dir, err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			base := line
			if i := strings.LastIndex(base, "/"); i >= 0 {
				base = base[i+1:]
			}
			if n := issue.IDFromFilename(base); n > 0 {
				dup := false
				for _, seen := range byID[n] {
					if seen == line {
						dup = true
						break
					}
				}
				if !dup {
					byID[n] = append(byID[n], line)
				}
			}
		}
	}
	return byID, nil
}

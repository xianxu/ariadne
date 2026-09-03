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

	// Update the ref if we can; ignore failure and use whatever ref exists — a
	// stale trunk is strictly better than none.
	_, _ = r.Git("fetch", "--quiet", "origin", "main")

	if out, err := r.Git("rev-parse", "--verify", "--quiet", trunkRef); err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("no %s ref", trunkRef)
	}
	var ids []int
	for _, dir := range dirs {
		// A directory absent from the ref is not an error: ls-tree prints
		// nothing and the dir contributes no ids.
		out, err := r.Git("ls-tree", "--name-only", trunkRef, dir+"/")
		if err != nil {
			continue
		}
		ids = append(ids, issue.IDsInTreeListing(string(out))...)
	}
	return ids, nil
}

// repoRelativeIDDirs converts the id directories to paths inside the current git
// repo, refusing any that escape it. `git ls-tree` interprets a path against the
// repo root, so an absolute or outside path would silently name the wrong thing.
func repoRelativeIDDirs(issuesDir, historyDir string) ([]string, error) {
	top, err := gitx.RepoTopLevel()
	if err != nil {
		return nil, fmt.Errorf("not a git repo: %w", err)
	}
	top, _ = filepath.EvalSymlinks(top)
	var out []string
	for _, dir := range issue.IDDirs(issuesDir, historyDir) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		// EvalSymlinks only where the path exists; a not-yet-created dir is fine.
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
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
	for id, path := range local {
		if other, ok := trunk[id]; ok && other != path {
			clashes = append(clashes, fmt.Sprintf("  #%06d\n      this branch: %s\n      %s: %s", id, path, trunkRef, other))
		}
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

// issueFilesByID maps each id in a ref's issue directories to its path. A
// duplicate id WITHIN one ref keeps the first path seen; that state is what the
// gate exists to stop reaching the trunk, and it is reported by whichever side
// carries it.
func issueFilesByID(ref, issuesDir, historyDir string, r gitRunner) (map[int]string, error) {
	if out, err := r.Git("rev-parse", "--verify", "--quiet", ref); err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("no %s ref", ref)
	}
	dirs, derr := repoRelativeIDDirs(issuesDir, historyDir)
	if derr != nil {
		return nil, derr
	}
	byID := map[int]string{}
	for _, dir := range dirs {
		out, err := r.Git("ls-tree", "--name-only", ref, dir+"/")
		if err != nil {
			continue
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
				if _, seen := byID[n]; !seen {
					byID[n] = line
				}
			}
		}
	}
	return byID, nil
}

// republish.go — `sdlc issue publish` for details already on main (#284): the
// owner's edits reach main by one narrow commit, never by overwriting a main
// copy that moved since the edits were based on it.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// publishVerdict is republishDecision's answer for one issue.
type publishVerdict int

const (
	publishNothing publishVerdict = iota // main already has these details
	publishWrite                         // main is still at the base: publish
	publishRefuse                        // main moved since the base: bring it in first
)

// republishDecision judges one issue's details bodies — the frontmatter
// carries the card mirror, which differs by design: local (this checkout),
// base (at the merge base of HEAD and main; ok false when absent there) and
// main. Pure.
func republishDecision(local, base string, baseOK bool, main string) publishVerdict {
	switch {
	case local == main:
		return publishNothing
	case baseOK && base == main:
		return publishWrite
	default:
		return publishRefuse
	}
}

// republishItem is one owned issue's publication, read before any effect.
type republishItem struct {
	id, path string
	read     []byte // the local file as read, to prove it unchanged at the finish
	publish  []byte // the bytes main receives: read, with the card mirror refreshed
	base     string // the details body at the merge base of HEAD and main
	baseOK   bool
	write    bool // false: main already has these details (finish only)
}

// republishOwned publishes owned issues' edited details to main in one narrow
// commit, then finishes this checkout (#284). It never writes main for an issue
// this workspace does not own — re-checked against a fresh tracker read before
// the push — nor over a main copy that moved since the checkout's base, judged
// again on every attempt. A rerun after an uncertain push finds main already
// holding the details and only finishes the checkout.
func republishOwned(env *trackerEnv, stdout, stderr io.Writer, ids []string, issuesRel string, dryRun bool) error {
	return republishOwnedOnce(env, stdout, stderr, ids, issuesRel, dryRun, true)
}

// republishOwnedOnce is republishOwned; bringIn allows one pass of bringing a
// moved main into a resting branch's edits before judging again.
func republishOwnedOnce(env *trackerEnv, stdout, stderr io.Writer, ids []string, issuesRel string, dryRun, bringIn bool) error {
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	mergeBase, err := env.git("merge-base", "HEAD", view.Ref())
	if err != nil {
		return fmt.Errorf("this checkout shares no history with main: %w", err)
	}
	var items []republishItem
	var written []string
	var moved []movedDetail
	for _, id := range ids {
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		if err := requireOwnedToPublish(env, card); err != nil {
			return err
		}
		rel, abs := localDetail(env, issuesRel, card.Path)
		it := republishItem{id: id, path: rel}
		if info, err := os.Lstat(abs); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("#%s: no local details at %s to publish", issue.CLIRef(id), it.path)
		}
		if it.read, err = os.ReadFile(abs); err != nil {
			return err
		}
		if it.publish, err = refreshMirror(env, id, it.read); err != nil {
			return fmt.Errorf("%s: %w — nothing was published", it.path, err)
		}
		baseRaw, baseOK, err := detailsAt(env, mergeBase, it.path)
		if err != nil {
			return err
		}
		if baseOK {
			if _, it.base, err = issue.Parse(string(baseRaw)); err != nil {
				return fmt.Errorf("%s at the merge base: %w", it.path, err)
			}
			it.baseOK = true
		}
		mainRaw, err := view.Read(it.path)
		if err != nil {
			return err
		}
		_, localBody, err := issue.Parse(string(it.publish))
		if err != nil {
			return fmt.Errorf("%s: %w", it.path, err)
		}
		if hasConflictMarkers(localBody) {
			return fmt.Errorf("%s still has merge conflict markers; resolve them, then rerun `sdlc issue publish --issue %s` — nothing was published", it.path, issue.CLIRef(id))
		}
		_, mainBody, err := issue.Parse(string(mainRaw))
		if err != nil {
			return fmt.Errorf("main's %s: %w", it.path, err)
		}
		switch republishDecision(localBody, it.base, it.baseOK, mainBody) {
		case publishRefuse:
			if !env.onRest() || !bringIn || dryRun {
				return movedMainRefusal(env, id, it.path)
			}
			moved = append(moved, movedDetail{item: it, base: baseRaw, main: mainRaw})
			continue
		case publishWrite:
			it.write = true
			written = append(written, id)
		}
		items = append(items, it)
	}
	if len(moved) > 0 {
		return bringMainIn(env, stdout, stderr, ids, issuesRel, view.Ref(), moved)
	}
	if dryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would publish %s to main in one commit; %d already there", issue.JoinRefs(written, "#", ", "), len(items)-len(written)))
		return nil
	}
	if len(written) > 0 {
		msg := fmt.Sprintf("%s: issue: publish details\n\nTracker-Operation: %s", issue.JoinRefs(written, "#", ","), operationToken("publish"))
		err := mainPublish(env, msg, func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
			w := map[string][]byte{}
			for _, it := range items {
				if !it.write {
					continue
				}
				cur, err := v.Read(it.path)
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				_, curBody, err := issue.Parse(string(cur))
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				if curBody != it.base {
					return gitx.TrunkWrite{}, fmt.Errorf("#%s: main's details changed while publishing; rerun `sdlc issue publish --issue %s` to judge them again — nothing was published", issue.CLIRef(it.id), issue.CLIRef(it.id))
				}
				w[it.path] = it.publish
			}
			return gitx.TrunkWrite{Write: w, ExactBytes: true}, nil
		}, func(string, string) error {
			fresh, err := env.repo.Snapshot()
			if err != nil {
				return err
			}
			for _, id := range written {
				card, err := fresh.Require(id)
				if err != nil {
					return err
				}
				if err := requireOwnedToPublish(env, card); err != nil {
					return err
				}
			}
			return nil
		})
		invalidateIssueRecords(env.ctx)
		if err != nil {
			return uncertainCardWrite(err, "sdlc issue publish --issue "+issue.JoinRefs(ids, "", ","))
		}
		cok(stderr, fmt.Sprintf("%s's details published to main in one commit", issue.JoinRefs(written, "#", ", ")))
	}
	finishPublished(env, stderr, items)
	for _, it := range items {
		fmt.Fprintln(stdout, it.path)
	}
	return nil
}

// mainPublish is republishing's one write to main — a seam so a test can make
// the push land and its response be lost.
var mainPublish = func(env *trackerEnv, msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error), beforePush func(base, candidate string) error) error {
	return env.main.UpdateManyPrepared(msg, prepare, beforePush)
}

// requireOwnedToPublish refuses republishing a card this workspace does not
// hold: an unowned card is claimed first, another workspace's is theirs.
func requireOwnedToPublish(env *trackerEnv, card tracker.Record) error {
	own, recorded, _, err := ownership(env, card)
	if err != nil {
		return err
	}
	switch own {
	case issue.OwnershipMine:
		return nil
	case issue.OwnershipForeign:
		return fmt.Errorf("#%s is owned by %s; only its owner republishes its details", issue.CLIRef(card.ID), describeClaimant(recorded))
	default:
		return fmt.Errorf("#%s has no owner; claim it first (`sdlc claim --issue %s`), then publish", issue.CLIRef(card.ID), issue.CLIRef(card.ID))
	}
}

// finishPublished makes this checkout agree with what main now holds. A
// resting branch never carries the edits as commits: each file, proven
// unchanged since it was read, is restored and the rest fast-forwards to main.
// On an issue's own branch the published bytes are committed narrowly, so its
// later merge is a no-op on that file; on any other branch (another issue's,
// #272) the edit is taken back to that branch's copy — main has it. A file
// changed since it was read is never overwritten. Problems warn: main already
// has the details.
func finishPublished(env *trackerEnv, stderr io.Writer, items []republishItem) {
	view, err := env.main.Snapshot()
	if err != nil {
		cwarn(stderr, fmt.Sprintf("checkout not finished: reading main failed: %v", err))
		return
	}
	var ready []republishItem
	for _, it := range items {
		cur, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(it.path)))
		if err != nil || !bytes.Equal(cur, it.read) {
			cwarn(stderr, fmt.Sprintf("%s changed while publishing; left as it is — main has the published copy", it.path))
			continue
		}
		ready = append(ready, it)
	}
	restore := func(it republishItem) bool {
		if err := restoreToHead(env, it.path); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not restored: %v", it.path, err))
			return false
		}
		return true
	}
	if env.onRest() {
		for _, it := range ready {
			if !restore(it) {
				return
			}
		}
		if warn := fastForwardRest(env, view.Ref()); warn != "" {
			cwarn(stderr, warn)
		}
		return
	}
	for _, it := range ready {
		if env.branch != issue.BranchName(it.path) {
			restore(it) // another issue's branch carries none of #it's edits
			continue
		}
		if err := os.WriteFile(filepath.Join(env.root, filepath.FromSlash(it.path)), it.publish, 0o644); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not written: %v", it.path, err))
			continue
		}
		if dirty, err := env.git("status", "--porcelain", "--", it.path); err != nil || dirty == "" {
			continue
		}
		if _, err := env.git("add", "--", it.path); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not staged: %v", it.path, err))
			continue
		}
		msg := "#" + issue.CLIRef(it.id) + ": issue: details as published"
		if _, err := env.git("commit", "-q", "--only", "-m", msg, "--", it.path); err != nil {
			cwarn(stderr, fmt.Sprintf("the published details were not committed on %s: %v", env.branch, err))
		}
	}
}

// localDetail is an issue's details path in this checkout: relative to the
// root (as main names it) and absolute.
func localDetail(env *trackerEnv, issuesRel, cardPath string) (rel, abs string) {
	rel = path.Join(issuesRel, path.Base(cardPath))
	return rel, filepath.Join(env.root, filepath.FromSlash(rel))
}

// movedDetail is a resting branch's edit whose main copy moved since its base.
type movedDetail struct {
	item       republishItem
	base, main []byte
}

// bringMainIn is the remedy for main having moved under a resting branch's
// edits (#284 BR-8): each edit is merged three ways (base, local, main) — the
// same merge `git stash; git merge; git stash pop` would do — the rest
// fast-forwards to main, and the merged details are written back. A clean
// merge publishes, now based on main; a conflict leaves the markers in the file
// and refuses with the one remaining step, which runs from that state.
func bringMainIn(env *trackerEnv, stdout, stderr io.Writer, ids []string, issuesRel, mainRef string, moved []movedDetail) error {
	merged := map[string][]byte{}
	var conflicted []string
	for _, m := range moved {
		out, conflicts, err := mergeDetails(env, m.item.read, m.base, m.main)
		if err != nil {
			return fmt.Errorf("#%s: merging main into the local details failed: %w — nothing was changed", issue.CLIRef(m.item.id), err)
		}
		merged[m.item.path] = out
		if conflicts {
			conflicted = append(conflicted, m.item.path)
		}
	}
	// Set every edit aside, all or nothing. Each edit is first copied under the
	// git directory, so no failure below can lose it; a failure puts every
	// touched file back exactly as it was read, and says where the copies are
	// if even that fails.
	gitDir, err := env.git("rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	aside := filepath.Join(gitDir, "sdlc", "publish-aside")
	if err := os.MkdirAll(aside, 0o700); err != nil {
		return fmt.Errorf("keeping a copy of the edits failed: %w — nothing was changed", err)
	}
	for _, m := range moved {
		if err := os.WriteFile(filepath.Join(aside, path.Base(m.item.path)), m.item.read, 0o600); err != nil {
			return fmt.Errorf("keeping a copy of %s failed: %w — nothing was changed", m.item.path, err)
		}
	}
	putBack := func(cause string) error {
		var failed []string
		for _, m := range moved {
			if err := writeLocal(filepath.Join(env.root, filepath.FromSlash(m.item.path)), m.item.read, 0o644); err != nil {
				failed = append(failed, fmt.Sprintf("%s (%v)", m.item.path, err))
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("%s, and putting the edits back failed for %s; copies of every edit are in %s", cause, strings.Join(failed, ", "), aside)
		}
		removeAside(aside, moved)
		return fmt.Errorf("%s — the edits are as they were", cause)
	}
	for _, m := range moved {
		if err := restoreToHead(env, m.item.path); err != nil {
			return putBack(fmt.Sprintf("main's details changed since this checkout's base, and setting %s aside to bring main in failed: %v", m.item.path, err))
		}
	}
	if warn := fastForwardRest(env, mainRef); warn != "" {
		return putBack("main's details changed since this checkout's base, and bringing main in failed: " + warn)
	}
	for p, b := range merged {
		if err := writeLocal(filepath.Join(env.root, filepath.FromSlash(p)), b, 0o644); err != nil {
			return fmt.Errorf("main was brought in, but writing the merged %s failed: %w; copies of every edit are in %s", p, err, aside)
		}
	}
	removeAside(aside, moved)
	if len(conflicted) > 0 {
		return fmt.Errorf("main's details changed since this checkout's base: main was brought in and merged into your edits, with conflicts in %s; resolve the markers, then rerun `sdlc issue publish --issue %s` — nothing was published", strings.Join(conflicted, ", "), issue.JoinRefs(ids, "", ","))
	}
	cinfo(stderr, "main had moved; it was brought in and merged cleanly into the edits")
	return republishOwnedOnce(env, stdout, stderr, ids, issuesRel, false, false)
}

// mergeDetails merges local and main over their base, as `git merge-file`
// does, through the env's git; conflicts reports whether markers were left. An
// absent base is empty.
func mergeDetails(env *trackerEnv, local, base, main []byte) ([]byte, bool, error) {
	dir, err := os.MkdirTemp("", "sdlc-merge-")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)
	for name, b := range map[string][]byte{"local": local, "base": base, "main": main} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
			return nil, false, err
		}
	}
	out, err := env.gitRaw(nil, "merge-file", "-p", "-L", "yours", "-L", "base", "-L", "main",
		filepath.Join(dir, "local"), filepath.Join(dir, "base"), filepath.Join(dir, "main"))
	var exit *exec.ExitError
	switch {
	case err == nil:
		return out, false, nil
	case errors.As(err, &exit) && exit.ExitCode() > 0 && exit.ExitCode() < 128:
		return out, true, nil // the exit code counts the conflicts
	default:
		return nil, false, err
	}
}

// hasConflictMarkers reports whether a details body still holds a merge
// conflict: an opening and a closing marker line (a lone "=======" is an
// ordinary setext heading underline).
func hasConflictMarkers(body string) bool {
	open, closing := false, false
	for _, line := range strings.Split(body, "\n") {
		open = open || strings.HasPrefix(line, "<<<<<<< ")
		closing = closing || strings.HasPrefix(line, ">>>>>>> ")
	}
	return open && closing
}

// movedMainRefusal is the main-moved refusal off a resting branch, naming steps
// that run from this state (#284 BR-8): on the issue's own branch, commit the
// edit and merge main, which git allows once the edit is committed; on any
// other branch, publishing belongs on a resting branch or the issue's own.
func movedMainRefusal(env *trackerEnv, id, rel string) error {
	issueStr := issue.CLIRef(id)
	switch {
	case env.onRest():
		return fmt.Errorf("#%s: main's details changed since this checkout's base; rerun `sdlc issue publish --issue %s` (not a dry run): on a resting branch it brings main in and merges it into the edit — nothing was published", issueStr, issueStr)
	case env.branch == issue.BranchName(rel):
		return fmt.Errorf("#%s: main's details changed since this branch's base. Commit the edit (`git commit -m '#%s: plan: …' -- %s`), merge main (`git fetch %s && git merge %s/main`, resolving any conflict), then rerun `sdlc issue publish --issue %s` — nothing was published", issueStr, issueStr, rel, env.target.Remote, env.target.Remote, issueStr)
	default:
		return fmt.Errorf("#%s: main's details changed since this branch's base, and %s is not #%s's branch. Switch to a resting branch, where the uncommitted edit comes along (`git switch %s`), and rerun `sdlc issue publish --issue %s`: it brings main in there — nothing was published", issueStr, env.branch, issueStr, env.resting, issueStr)
	}
}

// detailsAt reads a details file at a commit; present is false when the
// commit has none. A failed probe is an error, never absence.
func detailsAt(env *trackerEnv, commit, rel string) ([]byte, bool, error) {
	present, err := env.has(commit, rel)
	if err != nil || !present {
		return nil, false, err
	}
	raw, err := env.gitRaw(nil, "show", commit+":"+rel)
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

// restoreToHead sets a details file back to HEAD's copy, or removes it when
// HEAD has none. A seam, so a test can fail it part-way through a set.
var restoreToHead = func(env *trackerEnv, rel string) error {
	inHead, err := env.has("HEAD", rel)
	if err != nil {
		return err
	}
	if inHead {
		_, err = env.git("checkout", "HEAD", "--", rel)
		return err
	}
	return os.Remove(filepath.Join(env.root, filepath.FromSlash(rel)))
}

// writeLocal writes a details file in the checkout; a seam for failure tests.
var writeLocal = os.WriteFile

// removeAside deletes the bring-in's copies once the edits are safe again.
func removeAside(aside string, moved []movedDetail) {
	for _, m := range moved {
		_ = os.Remove(filepath.Join(aside, path.Base(m.item.path)))
	}
}

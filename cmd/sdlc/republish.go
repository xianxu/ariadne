// republish.go — `sdlc issue publish` for details already on main (#284): the
// edits (the owner's, or anyone's while unclaimed, #274) reach main by one narrow commit, never by overwriting a main
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
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
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

// republishItem is one issue's publication, read before any effect.
type republishItem struct {
	id, path                       string
	assets, baseAssets, mainAssets issueassets.Bundle
	inventory                      map[string][]byte
	read                           []byte // the local file as read, to prove it unchanged at the finish
	publish                        []byte // the bytes main receives: read, with the card mirror refreshed
	base                           string // the details body at the merge base of HEAD and main
	baseOK                         bool
	write                          bool // false: main already has these details (finish only)
}

// republishOwned publishes issues' edited details (mayNotEditDetails) to main in one narrow
// commit, then finishes this checkout (#284). It never writes main for another
// workspace's issue (#274: an unclaimed one publishes) — re-checked against a fresh tracker read before
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
		it.mainAssets, err = assetFilesFromView(view, it.path, id, mainRaw)
		if err != nil {
			return err
		}
		it.baseAssets, err = assetFilesFromView(env.main.ViewOf(mergeBase), it.path, id, baseRaw)
		if err != nil {
			return err
		}
		it.inventory, err = readAssetFiles(env.root, it.path, id)
		if err != nil {
			return err
		}
		localFiles := map[string][]byte{}
		refs, err := issueassets.ParseRefs(id, issueAssetBody(it.publish))
		if err != nil {
			return err
		}
		wanted := map[string]bool{}
		for _, p := range refs {
			wanted[p] = true
		}
		for p, b := range it.inventory {
			if !wanted[p] {
				old, ok := it.baseAssets.Files[p]
				if !ok || !bytes.Equal(old, b) {
					return fmt.Errorf("unlinked asset %s: link it in the body or delete it", p)
				}
				continue
			}
			localFiles[p] = b
		}
		it.assets, err = issueassets.BuildBundle(id, issueAssetBody(it.publish), localFiles)
		if err != nil {
			return err
		}
		for p := range it.assets.Files {
			dest := path.Join(path.Dir(it.path), p)
			if err := verifyAssetAttributes(env.ctx, env.root, dest); err != nil {
				return err
			}
			if err := verifyAssetAttributesAt(env.ctx, env.root, dest, view.Ref()); err != nil {
				return err
			}
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
		decision := republishDecision(localBody, it.base, it.baseOK, mainBody)
		if decision == publishNothing && !assetBundlesEqual(it.assets, it.mainAssets) {
			decision = publishWrite
		}
		if decision == publishWrite && !assetBundlesEqual(it.baseAssets, it.mainAssets) {
			decision = publishRefuse
		}
		switch decision {
		case publishRefuse:
			if !env.onRest() || !bringIn || dryRun {
				return movedMainRefusal(env, id, it.path)
			}
			moved = append(moved, movedDetail{item: it, base: baseRaw, main: mainRaw})
			continue
		case publishWrite:
			if len(issueassets.BundleDelta(it.mainAssets, it.assets).Writes) > 0 {
				policy, err := resolveAssetPolicy(env.ctx, env.root)
				if err != nil {
					return err
				}
				if err = requireAssetPolicy(policy); err != nil {
					return err
				}
			}
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
			var deletes []string
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
				currentAssets, err := assetFilesFromView(v, it.path, it.id, cur)
				if err != nil {
					return gitx.TrunkWrite{}, err
				}
				if !assetBundlesEqual(currentAssets, it.mainAssets) {
					return gitx.TrunkWrite{}, fmt.Errorf("#%s assets changed on main: merge main and rerun issue publish", it.id)
				}
				delta := issueassets.BundleDelta(currentAssets, it.assets)
				if len(delta.Writes) > 0 {
					policy, err := resolveAssetPolicy(env.ctx, env.root)
					if err != nil {
						return gitx.TrunkWrite{}, err
					}
					if err = requireAssetPolicy(policy); err != nil {
						return gitx.TrunkWrite{}, err
					}
				}
				for p, b := range delta.Writes {
					dest := path.Join(path.Dir(it.path), p)
					if err := verifyAssetAttributesAt(env.ctx, env.root, dest, v.Ref()); err != nil {
						return gitx.TrunkWrite{}, err
					}
					w[dest] = b
				}
				for _, p := range delta.Deletes {
					deletes = append(deletes, path.Join(path.Dir(it.path), p))
				}
				w[it.path] = it.publish
			}
			return gitx.TrunkWrite{Write: w, Delete: deletes, ExactBytes: true}, nil
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

// requireOwnedToPublish refuses republishing another workspace's card; this
// workspace's own and an unclaimed one publish (mayNotEditDetails, #274), with
// main's movement since the read as the concurrency check.
func requireOwnedToPublish(env *trackerEnv, card tracker.Record) error {
	own, recorded, _, err := ownership(env, card)
	if err != nil {
		return err
	}
	if mayNotEditDetails(own) {
		return fmt.Errorf("#%s is owned by %s; only its owner republishes its details", issue.CLIRef(card.ID), describeClaimant(recorded))
	}
	return nil
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
		now, readErr := readAssetFiles(env.root, it.path, it.id)
		if readErr != nil || !assetBundlesEqual(issueassets.Bundle{Files: now}, issueassets.Bundle{Files: it.inventory}) {
			cwarn(stderr, fmt.Sprintf("%s assets changed while publishing; left unchanged — main has published bytes; reconcile and rerun issue publish", it.path))
			continue
		}
		ready = append(ready, it)
	}
	restore := func(it republishItem) bool {
		paths := assetBundlePaths(it.path, it.assets, it.baseAssets, issueassets.Bundle{Files: it.inventory})
		for _, p := range paths {
			if err := restoreToHead(env, p); err != nil {
				cwarn(stderr, fmt.Sprintf("%s not restored: %v", p, err))
				return false
			}
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
		paths := assetBundlePaths(it.path, it.assets, it.baseAssets, issueassets.Bundle{Files: it.inventory})
		for p := range it.inventory {
			if _, keep := it.assets.Files[p]; !keep {
				if err := os.Remove(filepath.Join(env.root, filepath.FromSlash(path.Join(path.Dir(it.path), p)))); err != nil && !errors.Is(err, os.ErrNotExist) {
					cwarn(stderr, err.Error())
					return
				}
			}
		}
		dirty, err := env.git(append([]string{"status", "--porcelain", "--"}, paths...)...)
		if err != nil {
			cwarn(stderr, fmt.Sprintf("%s: reading its status failed, so the published bytes were not committed on %s: %v", it.path, env.branch, err))
			continue
		}
		if dirty == "" {
			continue
		}
		// A removal committed before publish is absent from both HEAD and
		// the index. It needs no staging, and Git rejects that missing pathspec.
		var stagePaths []string
		for _, p := range paths {
			_, statErr := os.Lstat(filepath.Join(env.root, filepath.FromSlash(p)))
			tracked, probeErr := env.has("HEAD", p)
			if probeErr != nil || (statErr != nil && !errors.Is(statErr, os.ErrNotExist)) {
				cwarn(stderr, fmt.Sprintf("%s not staged: cannot verify local presence; rerun issue publish", p))
				return
			}
			if tracked || statErr == nil {
				stagePaths = append(stagePaths, p)
			}
		}
		paths = stagePaths
		if _, err := env.git(append([]string{"add", "-A", "--"}, paths...)...); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not staged: %v", it.path, err))
			continue
		}
		msg := "#" + issue.CLIRef(it.id) + ": issue: details as published"
		if _, err := env.git(append([]string{"commit", "-q", "--only", "-m", msg, "--"}, paths...)...); err != nil {
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
	originals := map[string][]byte{}
	affected := map[string]bool{}
	var conflicted []string
	var removeAfter []string
	for _, m := range moved {
		current, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(m.item.path)))
		inventory, assetErr := readAssetFiles(env.root, m.item.path, m.item.id)
		if err != nil || assetErr != nil || !bytes.Equal(current, m.item.read) || !assetBundlesEqual(issueassets.Bundle{Files: inventory}, issueassets.Bundle{Files: m.item.inventory}) {
			return fmt.Errorf("%s or its assets changed while preparing publication; nothing changed; rerun issue publish", m.item.path)
		}
		out, conflicts, err := mergeDetails(env, m.item.read, m.base, m.main)
		if err != nil {
			return fmt.Errorf("#%s: merging main into the local details failed: %w — nothing was changed", issue.CLIRef(m.item.id), err)
		}
		merged[m.item.path] = out
		originals[m.item.path] = m.item.read
		for p, b := range assetBundleWrite(issueassets.Bundle{Files: m.item.inventory}, m.item.path) {
			originals[p] = b
		}
		for _, p := range assetBundlePaths(m.item.path, m.item.baseAssets, m.item.assets, issueassets.Bundle{Files: m.item.inventory}) {
			affected[p] = true
		}
		union := map[string][]byte{}
		for p, b := range m.item.mainAssets.Files {
			union[p] = b
		}
		for p, b := range m.item.assets.Files {
			if old, ok := union[p]; ok && !bytes.Equal(old, b) {
				return fmt.Errorf("asset %s has conflicting bytes; reconcile it before publication", p)
			}
			union[p] = b
		}
		keep := map[string]bool{}
		if conflicts {
			conflicted = append(conflicted, m.item.path)
			for p := range union {
				keep[p] = true
			}
		} else {
			refs, err := issueassets.ParseRefs(m.item.id, issueAssetBody(out))
			if err != nil {
				return err
			}
			for _, p := range refs {
				keep[p] = true
				if _, ok := union[p]; !ok {
					return fmt.Errorf("merged details reference missing asset %s; restore it before publication", p)
				}
			}
		}
		for p, b := range union {
			if keep[p] {
				merged[path.Join(path.Dir(m.item.path), p)] = b
			}
		}
		for p := range m.item.mainAssets.Files {
			if !keep[p] {
				removeAfter = append(removeAfter, path.Join(path.Dir(m.item.path), p))
			}
		}
	}
	gitDir, err := env.git("rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	aside := filepath.Join(gitDir, "sdlc", "publish-aside")
	backupPath := func(p string) string {
		if assetIssueID(p) != "" {
			return filepath.Join(aside, filepath.FromSlash(p))
		}
		return filepath.Join(aside, path.Base(p))
	}
	for p, b := range originals {
		backup := backupPath(p)
		if err := os.MkdirAll(filepath.Dir(backup), 0700); err != nil {
			return fmt.Errorf("keeping a copy of %s failed: %w — nothing changed", p, err)
		}
		if err := os.WriteFile(backup, b, 0600); err != nil {
			return fmt.Errorf("keeping a copy of %s failed: %w — nothing changed", p, err)
		}
	}
	cleanup := func() {
		for p := range originals {
			backup := backupPath(p)
			_ = os.Remove(backup)
			for dir := filepath.Dir(backup); dir != aside && strings.HasPrefix(dir, aside+string(filepath.Separator)); dir = filepath.Dir(dir) {
				if err := os.Remove(dir); err != nil {
					break
				}
			}
		}
	}
	putBack := func(cause string) error {
		var failed []string
		for p := range affected {
			abs := filepath.Join(env.root, filepath.FromSlash(p))
			var err error
			if b, present := originals[p]; present {
				err = os.MkdirAll(filepath.Dir(abs), 0755)
				if err == nil {
					err = writeLocal(abs, b, 0644)
				}
			} else {
				err = os.Remove(abs)
				if errors.Is(err, os.ErrNotExist) {
					err = nil
				}
			}
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (%v)", p, err))
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("%s, and putting edits back failed for %s; copies of every edit are in %s", cause, strings.Join(failed, ", "), aside)
		}
		cleanup()
		return fmt.Errorf("%s — the edits are as they were", cause)
	}
	// Keep the established document ordering; asset paths then join the same
	// set-aside transaction, so untracked identical bytes cannot obstruct Git.
	for _, m := range moved {
		for _, p := range assetBundlePaths(m.item.path, m.item.baseAssets, m.item.assets, issueassets.Bundle{Files: m.item.inventory}) {
			if err := restoreToHead(env, p); err != nil {
				return putBack(fmt.Sprintf("main's details changed since this checkout's base, and setting %s aside to bring main in failed: %v", p, err))
			}
		}
	}
	if warn := fastForwardRest(env, mainRef); warn != "" {
		return putBack("main's details changed since this checkout's base, and bringing main in failed: " + warn)
	}
	keys := make([]string, 0, len(merged))
	for p := range merged {
		keys = append(keys, p)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := assetIssueID(keys[i]) != "", assetIssueID(keys[j]) != ""
		if a != b {
			return a
		}
		return keys[i] < keys[j]
	})
	for _, p := range keys {
		abs := filepath.Join(env.root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			return fmt.Errorf("main was brought in, but preparing %s failed: %w; copies of every edit are in %s", p, err, aside)
		}
		if err := writeLocal(abs, merged[p], 0644); err != nil {
			return fmt.Errorf("main was brought in, but writing the merged %s failed: %w; copies of every edit are in %s", p, err, aside)
		}
	}
	for _, p := range removeAfter {
		if err := os.Remove(filepath.Join(env.root, filepath.FromSlash(p))); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("main was brought in, but removing unlinked asset %s failed: %w; copies of every edit are in %s", p, err, aside)
		}
	}
	cleanup()
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
	err = os.Remove(filepath.Join(env.root, filepath.FromSlash(rel)))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// writeLocal writes a details file in the checkout; a seam for failure tests.
var writeLocal = os.WriteFile

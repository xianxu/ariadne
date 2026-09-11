// synctrunk.go — publish changed issue files straight to the trunk, with no
// working tree involved (ariadne#207).
//
// Replaces syncViaMainWorktree, which drove SOMEONE ELSE'S CHECKOUT: find the
// worktree on main, refuse if it is dirty, pull --rebase it, detect files
// changed on both sides, copy across, commit, push. Every one of those guards
// exists to make a shared working directory safe, and each is a way to fail —
// main can be dirty, mid-rebase, another actor's tree, or (the common case,
// since change-code branches in place) not checked out at all.
//
// Here there is no working directory. gitx.TrunkFile builds the commit in the
// object database and pushes it as a compare-and-swap, which is a strictly
// stronger concurrency primitive than the cleanliness check it replaces: it
// sees the remote, where the check could only see local divergence.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// trunkPublisher is the seam, declared in the consumer (Go idiom) so the verb's
// logic is testable with a fake.
type trunkPublisher interface {
	UpdateMany(msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error)) error
}

// newTrunkPublisher builds the publisher the dispatch uses. A package-level var
// so tests can drive this arm against a controlled trunk — the same shim pattern
// gitx.run uses. Without it the publisher is constructed inline and the
// re-allocation path has no end-to-end test at all (#207 BR-1).
var newTrunkPublisher = func(root string) (trunkPublisher, error) {
	return gitx.NewTrunkFile(root, "origin", "main")
}

// errIDTaken reports that this issue's id belongs to a different slug on the
// trunk. Callers branch on it to give advice that will actually work: `issue
// sync` cannot fix a taken id, it refuses for the same reason.
var errIDTaken = errors.New("issue id is taken on the trunk")

// publishResult is what one publish attempted, across ALL of its attempts.
type publishResult struct {
	// reallocs are the id changes the FINAL attempt decided, reset per attempt:
	// a decision taken against a base that has since moved is stale.
	reallocs []*reallocation
	// candidates are every local path ANY attempt wrote — NOT reset, because a
	// rejected attempt's file is still on disk and this list is the only thing
	// that can find it again (#207 BR-2).
	candidates []string
	// collisions are every trunk path that forced a re-allocation, so exhaustion
	// can name them all rather than reporting generic contention.
	collisions []string
	onTrunk    bool
}

// finish applies the local half of a SUCCESSFUL publish: originals that were
// re-allocated away are removed, and so is every candidate the final attempt
// did not publish.
func (res *publishResult) finish() error {
	keep := map[string]bool{}
	for _, rc := range res.reallocs {
		keep[rc.NewPath] = true
	}
	var errs []string
	drop := func(path string) {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err.Error())
		}
	}
	for _, c := range res.candidates {
		if !keep[c] {
			drop(c) // written by an attempt whose push was rejected
		}
	}
	for _, rc := range res.reallocs {
		if rc.OldPath != rc.NewPath {
			drop(rc.OldPath)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("re-allocate cleanup: %s", strings.Join(errs, "; "))
	}
	return nil
}

// discardCandidates undoes the local half of a FAILED publish: nothing reached
// the trunk, so every candidate is unreferenced. The ORIGINALS are untouched —
// removing those is the inversion that deleted an operator's issue (#207 BR-2).
func (res *publishResult) discardCandidates(stderr io.Writer) {
	for _, c := range res.candidates {
		if err := os.Remove(c); err != nil && !os.IsNotExist(err) {
			cwarn(stderr, fmt.Sprintf("could not remove %s after a failed publish: %v", c, err))
		}
	}
}

// syncViaTrunk publishes the changed issue files in ONE commit on the trunk.
func syncViaTrunk(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string, pub trunkPublisher) error {
	res, err := syncViaTrunkWithRealloc(stdout, stderr, f, r, msg, pub)
	if err != nil {
		res.discardCandidates(stderr)
		if len(res.reallocs) > 0 {
			// A re-allocation was needed and the publish still failed, so the id
			// is taken and the local file still carries it. Say so, or the
			// caller's recovery advice sends the operator at a verb that will
			// refuse for exactly this reason (#207 BR-13).
			err = fmt.Errorf("%w: id %06d is held by %s; %w",
				errIDTaken, res.reallocs[0].OldID, strings.Join(res.reallocs[0].Foreign, ", "), err)
		}
		return err
	}
	for _, rc := range res.reallocs {
		cwarn(stderr, fmt.Sprintf("id %06d was taken on the trunk by %s — published as %06d (%s)",
			rc.OldID, strings.Join(rc.Foreign, ", "), rc.NewID, filepath.Base(rc.NewPath)))
	}
	if cerr := res.finish(); cerr != nil {
		cwarn(stderr, cerr.Error())
	}
	f.Reallocations = res.reallocs
	// `synced` is the machine-readable contract callers parse, and the in-place
	// arm emits it only after a real commit or push. Printed here only when
	// UpdateMany returned nil — at which point the trunk carries exactly this
	// content, whether we pushed it or it already matched. The no-change and
	// dry-run exits stay silent rather than naming a publication that never
	// happened.
	if res.onTrunk {
		cok(stderr, "Issue changes are on the trunk.")
		fmt.Fprintln(stdout, "synced")
	}
	return nil
}

// syncViaTrunkWithRealloc publishes and reports what it did. The result is
// always non-nil, so a failed publish can still clean up after itself.
//
// The collision check runs INSIDE prepare, which UpdateMany re-calls on every
// attempt. That placement is the correctness: a peer landing between our read
// and our push moves the ref, the push is rejected, prepare re-runs against the
// new base, and the check re-evaluates. Hoisting it out would let the retry
// re-push a colliding id as a clean fast-forward — ariadne#188's hole.
func syncViaTrunkWithRealloc(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string, pub trunkPublisher) (*publishResult, error) {
	res := &publishResult{}
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return res, err
	}
	// Same rule as the arm this replaces: nothing to COPY is not nothing to
	// publish. A body already committed here still needs routing to the trunk.
	if len(changed) == 0 {
		if !f.PublishExisting || f.Issue <= 0 {
			cok(stderr, "No issue changes to sync.")
			return res, nil
		}
		changed = issueFilesForID(f.IssuesDir, f.Issue)
		if len(changed) == 0 {
			cok(stderr, "No issue changes to sync.")
			return res, nil
		}
	}
	sort.Strings(changed)

	root, err := gitx.RepoTopLevel()
	if err != nil {
		return res, err
	}
	cinfo(stderr, "Publishing issue changes to the trunk:")
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — skipping publish")
		return res, nil
	}
	dirs, err := resolveIDDirs(f.IssuesDir, "workshop/history")
	if err != nil {
		return res, err
	}

	err = pub.UpdateMany(syncMessage(msg, defaultSyncSubject), func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		res.reallocs = nil // per-attempt decision; candidates persist
		set := gitx.TrunkWrite{Write: map[string][]byte{}}
		space, err := refIDSpace(v.Ref(), dirs, r)
		if err != nil {
			// A partial id-space read would answer from half the trunk and report
			// success, which is the #213 BR-7 defect. Fail instead.
			return set, err
		}
		free, err := unionIDSpace(space, dirs)
		if err != nil {
			return set, err
		}
		// Drop OUR OWN candidates from rejected attempts. They are on disk, so the
		// local scan counts them as taken and the retry walks the id forward —
		// 000700 to 000702 with 000701 free. Nothing published them, so they are
		// not reservations (#207 BR-2).
		for _, c := range res.candidates {
			rel, ok := repoRel(root, c)
			if !ok {
				continue
			}
			id := issueIDFromPath(rel)
			if id <= 0 {
				continue
			}
			var kept []string
			for _, held := range free[id] {
				if held != rel && held != c {
					kept = append(kept, held)
				}
			}
			if len(kept) == 0 {
				delete(free, id)
			} else {
				free[id] = kept
			}
		}
		claimed := map[int]string{}
		for _, rel := range changed {
			data, rerr := os.ReadFile(filepath.Join(root, rel))
			if rerr != nil {
				if !os.IsNotExist(rerr) {
					return set, fmt.Errorf("read %s: %v", rel, rerr)
				}
				// Absent locally is a DELETE only when git says the path is
				// tracked and now gone. An unexplained not-exist — a mis-parsed
				// path, say — must fail rather than publish a removal nobody
				// asked for (#207 BR-9).
				tracked, terr := pathTrackedAtHEAD(r, rel)
				if terr != nil {
					return set, terr
				}
				if !tracked {
					return set, fmt.Errorf(
						"%s is reported changed but is neither on disk nor tracked at HEAD; "+
							"refusing to guess that it was deleted", rel)
				}
				set.Delete = append(set.Delete, rel)
				continue
			}
			id := issueIDFromPath(rel)
			if id <= 0 {
				set.Write[rel] = data
				continue
			}
			// Two files in ONE publish claiming one id is a collision the trunk
			// cannot arbitrate — it would simply see one path win (#207 BR-15).
			if prev, dup := claimed[id]; dup {
				return set, fmt.Errorf(
					"two files in this publish claim id %06d: %s and %s — rename one first", id, prev, rel)
			}
			switch verdict, foreign := decideCollision(id, rel, space, f.FirstPublication); verdict {
			case verdictRefuse:
				return set, collisionRefusal(id, rel, foreign)
			case verdictReallocate:
				res.collisions = appendUnique(res.collisions, foreign...)
				newID := nextFreeID(id+1, free)
				newRel, newData, rerr := rewriteIdentity(rel, data, newID)
				if rerr != nil {
					return set, rerr
				}
				abs := filepath.Join(root, newRel)
				// Written BEFORE the push, so a crash never leaves the trunk ahead
				// of the working tree. finish() removes the original once the push
				// lands; discardCandidates removes this if it never does.
				if werr := os.WriteFile(abs, newData, 0o644); werr != nil {
					return set, fmt.Errorf("write %s: %v", newRel, werr)
				}
				res.candidates = appendUnique(res.candidates, abs)
				res.reallocs = append(res.reallocs, &reallocation{
					OldID: id, NewID: newID,
					OldPath: filepath.Join(root, rel), NewPath: abs, Foreign: foreign,
				})
				free[newID] = append(free[newID], newRel) // taken within this set too
				claimed[newID] = newRel
				set.Write[newRel] = newData
			default:
				claimed[id] = rel
				set.Write[rel] = data
			}
		}
		return set, nil
	})
	res.onTrunk = err == nil
	if err != nil && errors.Is(err, gitx.ErrTrunkMoved) && len(res.collisions) > 0 {
		// Spec step 5: on exhaustion name every collision seen, not just the last.
		err = fmt.Errorf("%w; ids contended by: %s", err, strings.Join(res.collisions, ", "))
	}
	return res, err
}

// pathTrackedAtHEAD reports whether git has this path at HEAD.
func pathTrackedAtHEAD(r gitRunner, rel string) (bool, error) {
	root, rerr := gitx.RepoTopLevel()
	if rerr != nil {
		return false, rerr
	}
	out, err := r.GitInDir(root, "ls-tree", "--full-tree", "--name-only", "--end-of-options", "HEAD", "--", rel)
	if err != nil {
		return false, fmt.Errorf("ls-tree HEAD -- %s: %v\n%s", rel, err, out)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func appendUnique(dst []string, add ...string) []string {
	for _, a := range add {
		found := false
		for _, d := range dst {
			if d == a {
				found = true
				break
			}
		}
		if !found {
			dst = append(dst, a)
		}
	}
	return dst
}

// unionIDSpace merges the trunk's id space with the local working tree's.
//
// refIDSpace answers "what is published"; a re-allocation also has to avoid what
// is merely WRITTEN here and not yet pushed. ariadne#213 settled this for
// allocation and the same reasoning applies: an unpushed local issue is a real
// reservation, and stepping onto it mints the collision.
//
// A failed local scan is an ERROR, not a shrug. The first version swallowed it
// reasoning that "the CAS will reject anyway" — which is false: the CAS compares
// REFS and knows nothing about a local id (#207 BR-6).
func unionIDSpace(trunk map[int][]string, dirs idDirs) (map[int][]string, error) {
	out := make(map[int][]string, len(trunk))
	for id, paths := range trunk {
		out[id] = paths
	}
	localByID, _, err := issue.LocalPathsByID(dirs.Abs)
	if err != nil {
		return nil, fmt.Errorf("scan local issue ids: %w", err)
	}
	for id, paths := range localByID {
		out[id] = append(out[id], paths...)
	}
	return out, nil
}

// collisionRefusal is a next-action spec, not a generic contention error: it
// names both paths and says why renumbering is not offered here.
func collisionRefusal(id int, mine string, foreign []string) error {
	return fmt.Errorf(
		"issue id %06d is already published under a different name:\n"+
			"      ours:  %s\n"+
			"      trunk: %s\n"+
			"    Not renumbering: this id is published, so it is already referenced by the\n"+
			"    branch name, commit subjects, deps: in sibling issues, and review sidecars\n"+
			"    (ariadne#188). Rename one side by hand and re-run",
		id, mine, strings.Join(foreign, "\n             "))
}

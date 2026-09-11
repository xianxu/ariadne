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

// syncViaTrunk publishes the changed issue files in ONE commit on the trunk.
//
// The collision check runs INSIDE prepare, which UpdateMany re-calls on every
// attempt. That placement is the correctness: a peer landing between our read
// and our push moves the ref, the push is rejected, prepare re-runs against the
// new base, and the check re-evaluates. Hoisting it out would let the retry
// re-push a colliding id as a clean fast-forward — ariadne#188's hole.
func syncViaTrunk(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string, pub trunkPublisher) error {
	rc, onTrunk, err := syncViaTrunkWithRealloc(stdout, stderr, f, r, msg, pub)
	if err != nil {
		// NEVER clean up after a failed publish. finish() removes the ORIGINAL
		// local file, so running it here would delete the operator's issue while
		// nothing reached the trunk — the Spec's ordering inverted, and data loss.
		// Candidate paths a rejected attempt wrote are left in place; they are
		// visible, and the next run reconciles them.
		return err
	}
	if rc != nil {
		cwarn(stderr, fmt.Sprintf("issue %06d was re-allocated to %06d — local file is now %s",
			rc.OldID, rc.NewID, filepath.Base(rc.NewPath)))
		if cerr := rc.finish(); cerr != nil {
			cwarn(stderr, cerr.Error())
		}
	}
	// `synced` is the machine-readable contract callers parse, and the in-place
	// arm emits it only after a real commit or push (claim.go). It is printed
	// here only when UpdateMany returned nil — at which point the trunk carries
	// exactly this content, whether we pushed it or it already matched. The
	// no-change and dry-run exits return onTrunk=false and stay silent: a marker
	// on those would name a publication that never happened.
	if onTrunk {
		cok(stderr, "Issue changes are on the trunk.")
		fmt.Fprintln(stdout, "synced")
	}
	return nil
}

// syncViaTrunkWithRealloc returns the id change, if any, so the caller can clean
// up the superseded local files AFTER the push has succeeded, and whether the
// trunk now carries this content (false on the no-change and dry-run exits).
func syncViaTrunkWithRealloc(stdout, stderr io.Writer, f *claimFlags, r gitRunner, msg string, pub trunkPublisher) (*reallocation, bool, error) {
	changed, err := changedIssueFiles(f, r)
	if err != nil {
		return nil, false, err
	}
	// Same rule as the arm this replaces: nothing to COPY is not nothing to
	// publish. A body already committed here still needs routing to the trunk.
	if len(changed) == 0 {
		if !f.PublishExisting || f.Issue <= 0 {
			cok(stderr, "No issue changes to sync.")
			return nil, false, nil
		}
		changed = issueFilesForID(f.IssuesDir, f.Issue)
		if len(changed) == 0 {
			cok(stderr, "No issue changes to sync.")
			return nil, false, nil
		}
	}
	sort.Strings(changed)

	root, err := gitx.RepoTopLevel()
	if err != nil {
		return nil, false, err
	}
	cinfo(stderr, "Publishing issue changes to the trunk:")
	for _, c := range changed {
		fmt.Fprintf(stderr, "  %s\n", c)
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — skipping publish")
		return nil, false, nil
	}

	dirs, err := resolveIDDirs(f.IssuesDir, "workshop/history")
	if err != nil {
		return nil, false, err
	}

	var rc *reallocation
	// BR-1: "" means "each arm's own default" (claim.go), and `issue new` and
	// `claim` both pass it. syncInPlace applies the default through syncMessage;
	// passing "" straight to commit-tree instead published an EMPTY subject on the
	// trunk from every feature branch — which, since change-code branches in
	// place, is the common case. One default, shared by both arms.
	err = pub.UpdateMany(syncMessage(msg, defaultSyncSubject), func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		// Reset per attempt: a re-allocation decided on a base that has since
		// moved is stale, and carrying it forward would report an id change the
		// final push never made.
		rc = nil
		set := gitx.TrunkWrite{Write: map[string][]byte{}}
		space, err := refIDSpace(v.Ref(), dirs, r)
		if err != nil {
			// A partial id-space read would answer from half the trunk and report
			// success, which is the #213 BR-7 defect. Fail instead.
			return set, err
		}
		for _, rel := range changed {
			data, rerr := os.ReadFile(filepath.Join(root, rel))
			if os.IsNotExist(rerr) {
				// Deleted locally -> remove from the trunk. The arm this replaces
				// fails loudly here (os.ReadFile); silently skipping would report
				// success while the file stayed published.
				set.Delete = append(set.Delete, rel)
				continue
			}
			if rerr != nil {
				return set, fmt.Errorf("read %s: %v", rel, rerr)
			}
			// issueIDFromPath (issuefiles.go) reads the id off the FILENAME, which
			// is what the tooling actually collides on: measured 2026-09-09, a
			// `000207-*` file with NO frontmatter still blocked `sdlc claim`, and a
			// frontmatter-keyed check would have missed it. Reused rather than
			// rewritten — the convention is parsed in one place (ARCH-DRY).
			id := issueIDFromPath(rel)
			if id <= 0 {
				set.Write[rel] = data
				continue
			}
			switch verdict, foreign := decideCollision(id, rel, space, f.FirstPublication); verdict {
			case verdictRefuse:
				return set, collisionRefusal(id, rel, foreign)
			case verdictReallocate:
				// Re-derived on EVERY attempt from this attempt's base, so a peer
				// landing mid-window produces a different id rather than a re-push
				// of the colliding one.
				// UNION the trunk's id space with the local scan before picking.
				// The trunk alone is not the id space: unpublished local issues
				// are real, and re-allocating onto one would mint exactly the
				// duplicate this arm exists to prevent. Same union ariadne#213
				// established for allocation (`issue.NextID(local, published)`).
				newID := nextFreeID(id+1, unionIDSpace(space, dirs))
				newRel, newData, rerr := rewriteIdentity(rel, data, newID)
				if rerr != nil {
					return set, rerr
				}
				// Written BEFORE the push, so a crash never leaves the trunk ahead
				// of the working tree. finish() removes the old path and any
				// orphans from rejected attempts once the push has landed.
				if werr := os.WriteFile(filepath.Join(root, newRel), newData, 0o644); werr != nil {
					return set, fmt.Errorf("write %s: %v", newRel, werr)
				}
				if rc == nil {
					rc = &reallocation{OldID: id, OldPath: filepath.Join(root, rel)}
				}
				rc.NewID, rc.NewPath, rc.Foreign = newID, filepath.Join(root, newRel), foreign
				rc.written = append(rc.written, filepath.Join(root, newRel))
				cwarn(stderr, fmt.Sprintf(
					"id %06d was taken on the trunk by %s — this issue is now %06d (%s)",
					id, strings.Join(foreign, ", "), newID, newRel))
				set.Write[newRel] = newData
			default:
				set.Write[rel] = data
			}
		}
		return set, nil
	})
	return rc, err == nil, err
}

// unionIDSpace merges the trunk's id space with the local working tree's.
//
// refIDSpace answers "what is published"; a re-allocation also has to avoid what
// is merely WRITTEN here but not yet pushed. ariadne#213 settled this for
// allocation and the same reasoning applies to re-allocation: an unpushed local
// issue is a real reservation, and stepping onto it mints the collision.
//
// A failed local scan is not fatal — the trunk half still constrains the choice,
// and refusing to re-allocate because the filesystem hiccuped would be worse than
// a slightly-too-low id that the CAS will reject anyway.
func unionIDSpace(trunk map[int][]string, dirs idDirs) map[int][]string {
	out := make(map[int][]string, len(trunk))
	for id, paths := range trunk {
		out[id] = paths
	}
	localByID, _, err := issue.LocalPathsByID(dirs.Abs)
	if err != nil {
		return out
	}
	for id, paths := range localByID {
		out[id] = append(out[id], paths...)
	}
	return out
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

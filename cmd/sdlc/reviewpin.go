// reviewpin.go — #304 D4. A finalized review's head H_r is the identity of the reviewed
// branch patch; the next window and the publish gate replay it onto today's main. A
// rebase leaves H_r unreferenced, so each boundary pins the newest commit its finalized
// review produced (the evidence commit, else H_r — it descends from H_r either way) at
// refs/sdlc/reviewed/<id>/<boundary>. Refs are shared across linked worktrees, so a move
// between slots keeps them.
//
// Lifecycle (ARCH-FUNERAL): created/advanced at boundary finalize; every ref under the
// issue's prefix is removed when the card goes done, when the issue is abandoned, and by
// the legacy archive; settle and recovery reconcile also sweep the pins of issues whose
// cards are terminal or gone. At most (milestones + 1) refs per in-flight issue.
//
// The pin is retention, never authority: every failure is a warning, not a refusal.
package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const reviewedPinPrefix = "refs/sdlc/reviewed/"

// reviewedPinRef names a boundary's pin; the whole-issue close is "close".
func reviewedPinRef(id, boundary string) string {
	if boundary == "" {
		boundary = "close"
	}
	return reviewedPinPrefix + id + "/" + boundary
}

// pinReviewed points the boundary's pin at commit. Returns a warning, or "".
func pinReviewed(id, boundary, commit string) string {
	if id == "" || !isResolvedSHA(commit) {
		return fmt.Sprintf("reviewed-head pin for #%s not written: no resolved commit", id)
	}
	if out, err := gitx.RunGit("update-ref", "-m", "sdlc: reviewed head (#304)", reviewedPinRef(id, boundary), commit); err != nil {
		return fmt.Sprintf("reviewed-head pin %s not written: %v %s", reviewedPinRef(id, boundary), err, strings.TrimSpace(string(out)))
	}
	return ""
}

// unpinReviewed deletes every pin of the issue in the checkout at dir ("" = cwd). No
// pins is a no-op. Returns a warning, or "".
func unpinReviewed(dir, id string) string {
	refs, err := reviewedPins(dir, reviewedPinPrefix+id+"/")
	if err != nil {
		return err.Error()
	}
	return deletePins(dir, refs)
}

// sweepReviewedPins deletes the pins of every issue `live` rejects — the end for pins
// whose issue landed and settled elsewhere, or was archived by hand. Returns a warning, or "".
func sweepReviewedPins(dir string, live func(id string) bool) string {
	refs, err := reviewedPins(dir, reviewedPinPrefix)
	if err != nil {
		return err.Error()
	}
	var dead []string
	for _, ref := range refs {
		id, _, _ := strings.Cut(strings.TrimPrefix(ref, reviewedPinPrefix), "/")
		if !live(id) {
			dead = append(dead, ref)
		}
	}
	return deletePins(dir, dead)
}

func reviewedPins(dir, prefix string) ([]string, error) {
	out, err := gitx.RunGit(inDir(dir, "for-each-ref", "--format=%(refname)", prefix)...)
	if err != nil {
		return nil, fmt.Errorf("list reviewed-head pins under %s: %v", prefix, err)
	}
	return strings.Fields(string(out)), nil
}

func deletePins(dir string, refs []string) string {
	var failed []string
	for _, ref := range refs {
		if _, err := gitx.RunGit(inDir(dir, "update-ref", "-d", ref)...); err != nil {
			failed = append(failed, ref)
		}
	}
	if len(failed) > 0 {
		return "reviewed-head pin(s) not removed: " + strings.Join(failed, ", ")
	}
	return ""
}

// inDir prefixes a git argv with -C dir when a caller names a checkout.
func inDir(dir string, args ...string) []string {
	if dir == "" {
		return args
	}
	return append([]string{"-C", dir}, args...)
}

// unpinArchived ends the pins of exactly the issues an archive moved. Refs are shared
// by every linked worktree, so liveness judged from THIS checkout's issues dir could
// delete a pin another slot's branch still needs; the moved set is this archive's own.
func unpinArchived(stderr io.Writer, dir string, moves []preparedArchiveMove) {
	seen := map[int]bool{}
	for _, m := range moves {
		id := issueIDFromPath(m.IssuePath)
		if id <= 0 || seen[id] {
			continue // a plan artifact or sidecar move, or an issue already handled
		}
		seen[id] = true
		if w := unpinReviewed(dir, fmt.Sprintf("%06d", id)); w != "" {
			cwarn(stderr, w)
		}
	}
}

// sweepTrackedPins ends the pins of every issue whose card is terminal or gone — the
// end for pins of an issue landed and settled elsewhere, or archived by hand.
func sweepTrackedPins(env *trackerEnv, rs tracker.Records) string {
	live := map[string]bool{}
	for _, rec := range rs.All() {
		if rec.Card != nil && !vocab.Issue().IsTerminal(rec.Status()) {
			live[rec.ID] = true
		}
	}
	return sweepReviewedPins(env.root, func(id string) bool { return live[id] })
}

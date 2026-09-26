// transferguard.go — protect handed-off details at publication (#252). After
// `issue move-detail`, the source branch's add-then-remove is net-zero, but a
// branch that re-adds, deletes or rewrites the details (an unfinished removal,
// a conflicting resolution, an edit from the wrong branch) would silently undo
// the new owner's work when it lands. The prospective merge must leave every
// handed-off issue's details exactly as main has them.
package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// guardTransferredDetailsFn is the gates' seam: publish-gate unit fixtures model
// anchor logic in remote-less repositories and stub it; the guard itself is
// exercised against real remotes in transferguard_test.go.
var guardTransferredDetailsFn = guardTransferredDetails

// guardTransferredDetails checks HEAD's prospective merge into fresh main. The
// issue's own branch is exempt: edits there are the claimed owner's work.
func guardTransferredDetails(ctx context.Context) error {
	env, err := openTracker(ctx)
	if err != nil {
		return err
	}
	if ok, err := env.repo.Initialized(); err != nil {
		return err
	} else if !ok {
		return nil // no tracker, so no handoffs to protect
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		return err
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return err
	}
	mainTip := view.Ref()
	var paths []string
	for _, rec := range snap.Records() {
		h, ok, err := issue.CardHandoff(rec.Raw)
		if err == nil && ok {
			err = validHandoffDestination(h.Destination, rec.Path)
		}
		if err != nil {
			return fmt.Errorf("tracker card #%s is malformed: %w\n"+
				"      every PR, push and merge in this repository is refused until the card is repaired", rec.ID, err)
		}
		if !ok || env.branch == strings.TrimSuffix(path.Base(h.Destination), ".md") {
			continue
		}
		// An interrupted handoff may have published to main before recording it
		// on the card; the details on main are protected from that moment.
		if h.MainCommit == "" {
			onMain, err := env.has(mainTip, h.Destination)
			if err != nil {
				return err
			}
			if !onMain {
				continue
			}
		}
		paths = append(paths, h.Destination)
	}
	if len(paths) == 0 {
		return nil
	}
	return checkTransferredPaths(env, mainTip, paths)
}

// checkTransferredPaths compares each handed-off path in the merge result of
// main and HEAD against main. A conflict, deletion, re-addition or rewrite refuses.
func checkTransferredPaths(env *trackerEnv, mainTip string, paths []string) error {
	out, err := env.git("merge-tree", "--write-tree", "--name-only", "--no-messages", mainTip, "HEAD")
	lines := strings.Split(out, "\n")
	if lines[0] == "" {
		return fmt.Errorf("compute the prospective merge with main: %v", err)
	}
	result := lines[0]
	if err != nil {
		// Exit 1: conflicts, listed after the tree. A conflict on a handed-off
		// path means the landing would need a resolution that could lose it.
		for _, conflicted := range lines[1:] {
			for _, p := range paths {
				if conflicted == p {
					return transferRefusal(p, "conflicts with main's copy")
				}
			}
		}
	}
	for _, p := range paths {
		onMain, err := env.has(mainTip, p)
		if err != nil {
			return err
		}
		inResult, err := env.has(result, p)
		if err != nil {
			return err
		}
		var want, got string
		if onMain {
			if want, err = env.git("rev-parse", "--verify", mainTip+":"+p); err != nil {
				return err
			}
		}
		if inResult {
			if got, err = env.git("rev-parse", "--verify", result+":"+p); err != nil {
				return err
			}
		}
		switch {
		case !onMain && !inResult:
			continue // absent on both (archived by its owner)
		case !inResult:
			return transferRefusal(p, "would be deleted")
		case !onMain:
			return transferRefusal(p, "would be re-added after its owner archived it")
		case want != got:
			return transferRefusal(p, "would be overwritten")
		}
	}
	return nil
}

// validHandoffDestination checks untrusted card data before it is used as a
// path and as the exemption key: a clean repository-relative details path
// named after the card.
func validHandoffDestination(dest, cardPath string) error {
	if dest == "" || path.IsAbs(dest) || path.Clean(dest) != dest || strings.HasPrefix(dest, "../") || dest == ".." ||
		path.Base(dest) != path.Base(cardPath) || strings.ContainsAny(dest, "\\:*?[") {
		return fmt.Errorf("handoff destination %q is not a details path for %s", dest, path.Base(cardPath))
	}
	return nil
}

var errTransferredDetails = errors.New("landing would change handed-off issue details")

func transferRefusal(p, what string) error {
	return fmt.Errorf("%w: %s %s by this branch.\n"+
		"      Its details were handed off to main (`sdlc issue move-detail`) and may have a new owner.\n"+
		"      Finish an interrupted handoff with `sdlc issue recovery reconcile --issue N`, or remove this\n"+
		"      branch's change to %s (restore it to main's version)", errTransferredDetails, p, what, p)
}

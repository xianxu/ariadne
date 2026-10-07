// republish.go — `sdlc issue publish` for details already on main (#284): the
// owner's edits reach main by one narrow commit, never by overwriting a main
// copy that moved since the edits were based on it.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

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
	for _, id := range ids {
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		if err := requireOwnedToPublish(env, card); err != nil {
			return err
		}
		it := republishItem{id: id, path: path.Join(issuesRel, path.Base(card.Path))}
		abs := filepath.Join(env.root, filepath.FromSlash(it.path))
		if info, err := os.Lstat(abs); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("#%s: no local details at %s to publish", issue.CLIRef(id), it.path)
		}
		if it.read, err = os.ReadFile(abs); err != nil {
			return err
		}
		if it.publish, err = refreshMirror(env, id, it.read); err != nil {
			return fmt.Errorf("%s: %w — nothing was published", it.path, err)
		}
		if raw, err := env.gitRaw(nil, "show", mergeBase+":"+it.path); err == nil {
			if _, it.base, err = issue.Parse(string(raw)); err == nil {
				it.baseOK = true
			}
		}
		mainRaw, err := view.Read(it.path)
		if err != nil {
			return err
		}
		_, localBody, err := issue.Parse(string(it.publish))
		if err != nil {
			return fmt.Errorf("%s: %w", it.path, err)
		}
		_, mainBody, err := issue.Parse(string(mainRaw))
		if err != nil {
			return fmt.Errorf("main's %s: %w", it.path, err)
		}
		switch republishDecision(localBody, it.base, it.baseOK, mainBody) {
		case publishRefuse:
			return fmt.Errorf("#%s: main's details changed since this checkout's base; bring main in first (a resting branch: `sdlc claim --issue %s` fast-forwards it; a branch: merge main) — nothing was published", issue.CLIRef(id), issue.CLIRef(id))
		case publishWrite:
			it.write = true
			written = append(written, id)
		}
		items = append(items, it)
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
					return gitx.TrunkWrite{}, fmt.Errorf("#%s: main's details changed while publishing; bring main in first — nothing was published", issue.CLIRef(it.id))
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
// Any other branch commits the published bytes narrowly, so its later merge is
// a no-op on those files. Problems warn: main already has the details.
func finishPublished(env *trackerEnv, stderr io.Writer, items []republishItem) {
	view, err := env.main.Snapshot()
	if err != nil {
		cwarn(stderr, fmt.Sprintf("checkout not finished: reading main failed: %v", err))
		return
	}
	var paths []string
	for _, it := range items {
		paths = append(paths, it.path)
	}
	if env.onRest() {
		for _, it := range items {
			cur, err := os.ReadFile(filepath.Join(env.root, filepath.FromSlash(it.path)))
			if err != nil || !bytes.Equal(cur, it.read) {
				cwarn(stderr, fmt.Sprintf("%s changed while publishing; left as it is — main has the published copy", it.path))
				return
			}
		}
		for _, p := range paths {
			if inHead, _ := env.gitTest("cat-file", "-e", "HEAD:"+p); inHead {
				if _, err := env.git("checkout", "HEAD", "--", p); err != nil {
					cwarn(stderr, fmt.Sprintf("%s not restored: %v", p, err))
					return
				}
			} else if err := os.Remove(filepath.Join(env.root, filepath.FromSlash(p))); err != nil {
				cwarn(stderr, fmt.Sprintf("%s not cleared: %v", p, err))
				return
			}
		}
		if warn := fastForwardRest(env, view.Ref()); warn != "" {
			cwarn(stderr, warn)
		}
		return
	}
	for _, it := range items {
		if err := os.WriteFile(filepath.Join(env.root, filepath.FromSlash(it.path)), it.publish, 0o644); err != nil {
			cwarn(stderr, fmt.Sprintf("%s not written: %v", it.path, err))
			return
		}
	}
	if same, _ := env.gitTest(append([]string{"diff", "--quiet", "HEAD", "--"}, paths...)...); same {
		return
	}
	var ids []string
	for _, it := range items {
		ids = append(ids, it.id)
	}
	msg := issue.JoinRefs(ids, "#", ",") + ": issue: details as published"
	if _, err := env.git(append([]string{"commit", "-q", "--only", "-m", msg, "--"}, paths...)...); err != nil {
		cwarn(stderr, fmt.Sprintf("the published details were not committed on %s: %v", env.branch, err))
	}
}

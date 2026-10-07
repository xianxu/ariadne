// handoff.go — started work changing hands (#284). Unclaiming started work
// pushes its branch and records the tip on the card; a later claim, in another
// slot, machine or operator's checkout, fetches exactly that tip and resumes
// there. The card is the lock; the branch is the work.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// cleanTree returns this checkout's uncommitted changes, untracked files
// included ("" when clean): a handoff carries only committed work.
func cleanTree(env *trackerEnv) (string, error) {
	return env.git("status", "--porcelain", "--untracked-files=all")
}

// appendUnclaimNote files one dated note line in a details file's ## Log,
// once: a rerun finds it there and adds nothing. changed reports a write.
func appendUnclaimNote(abs, note string) (changed bool, err error) {
	raw, err := os.ReadFile(abs)
	if err != nil {
		return false, err
	}
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		return false, err
	}
	line := fmt.Sprintf("- %s: unclaimed: %s", time.Now().Format("2006-01-02"), note)
	if strings.Contains(body, line) {
		return false, nil
	}
	return true, os.WriteFile(abs, []byte(issue.Compose(fm, insertLogLine(body, line))), 0o644)
}

// runHandoff releases this workspace's started issue: the note (if any) is
// committed on the issue branch, the branch is pushed, the card records the
// release with its branch and tip, and the checkout returns to rest. Order is
// push, card, switch: a lost step is settled by rerunning, which recognises
// its own release and only finishes the switch.
func runHandoff(env *trackerEnv, stdout, stderr io.Writer, card tracker.Record, issuesRel, note string, me issue.Claimant, dryRun bool) error {
	id := card.ID
	branch := issue.BranchName(card.Path)
	if _, err := unclaimDecision(card.Raw, id, me, "", ""); errors.Is(err, errAlreadyReleased) {
		cok(stderr, fmt.Sprintf("#%s already handed off by this workspace; nothing to release", issue.CLIRef(id)))
		if !dryRun {
			finishHandoff(env, stderr, card.Raw, branch)
		}
		fmt.Fprintln(stdout, "released")
		return nil
	} else if err != nil {
		return err
	}
	if env.branch != branch {
		return fmt.Errorf("#%s is handed off from its branch: run `sdlc unclaim` in the checkout on %s (this one is on %q)", issue.CLIRef(id), branch, env.branch)
	}
	if dirty, err := cleanTree(env); err != nil {
		return err
	} else if dirty != "" {
		return fmt.Errorf("#%s: a handoff carries only committed work; commit or remove these first (nothing was changed):\n%s", issue.CLIRef(id), dirty)
	}
	if dryRun {
		cinfo(stderr, fmt.Sprintf("dry-run — would push %s and release #%s with its tip", branch, issue.CLIRef(id)))
		return nil
	}
	if note != "" {
		rel, abs := localDetail(env, issuesRel, card.Path)
		if changed, err := appendUnclaimNote(abs, note); err != nil {
			return err
		} else if changed {
			if _, err := env.git("commit", "-q", "-m", "#"+issue.CLIRef(id)+": log: handoff", "--", rel); err != nil {
				return fmt.Errorf("committing the handoff note: %w", err)
			}
		}
	}
	head, err := env.git("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if err := pushIssueBranch(env, branch); err != nil {
		return fmt.Errorf("pushing %s failed: %w — the claim is kept; rerun `sdlc unclaim --issue %s`", branch, err, issue.CLIRef(id))
	}
	decide := func(current map[string]tracker.Record) (map[string][]byte, error) {
		rec, ok := current[id]
		if !ok {
			return nil, fmt.Errorf("no readable card #%s on the tracker", id)
		}
		next, err := unclaimDecision(rec.Raw, id, me, branch, head)
		if err != nil {
			return nil, err
		}
		return map[string][]byte{id: next}, nil
	}
	err = cardsPublish(env, []string{id}, operationToken("unclaim"), nil, decide, nil)
	invalidateIssueRecords(env.ctx)
	if errors.Is(err, errAlreadyReleased) {
		err = nil
	}
	if err = uncertainCardWrite(err, "sdlc unclaim --issue "+issue.CLIRef(id)); err != nil {
		return err
	}
	cok(stderr, fmt.Sprintf("#%s handed off: %s pushed at %s; no owner, status unchanged. A claim elsewhere resumes there.", issue.CLIRef(id), branch, shortOID(head)))
	if _, err := env.git("switch", "-q", env.resting); err != nil {
		cwarn(stderr, fmt.Sprintf("this checkout stays on %s: switching to %s failed: %v", branch, env.resting, err))
	}
	fmt.Fprintln(stdout, "released")
	return nil
}

// finishHandoff completes a handoff whose card already landed: the checkout
// returns to rest when it is still on the released tip and clean.
func finishHandoff(env *trackerEnv, stderr io.Writer, card []byte, branch string) {
	rel, ok, err := issue.CardRelease(card)
	if err != nil || !ok || env.branch != branch {
		return
	}
	head, err := env.git("rev-parse", "HEAD")
	if err != nil || head != rel.Head {
		return
	}
	if dirty, err := cleanTree(env); err != nil || dirty != "" {
		return
	}
	if _, err := env.git("switch", "-q", env.resting); err != nil {
		cwarn(stderr, fmt.Sprintf("this checkout stays on %s: switching to %s failed: %v", branch, env.resting, err))
	}
}

// pushIssueBranch publishes the issue branch to the publication remote. The
// owner is its only writer (the claim, and no stacking, #272), so a rewritten
// branch is pushed with a lease on the tip the remote had.
func pushIssueBranch(env *trackerEnv, branch string) error {
	remoteTip, err := env.git("ls-remote", env.target.Remote, "refs/heads/"+branch)
	if err != nil {
		return err
	}
	expect := ""
	if fields := strings.Fields(remoteTip); len(fields) > 0 {
		expect = fields[0]
	}
	ref := "refs/heads/" + branch
	_, err = env.git("push", "-q", "--force-with-lease="+ref+":"+expect, env.target.Remote, ref+":"+ref)
	return err
}

// takeover is the git side of claiming a handed-off issue: the branch and the
// tip it must be fetched at, checked before the card is written.
type takeover struct {
	branch, head string
}

// prepareTakeover checks, before any effect, that this checkout can resume a
// released issue: the release names this issue's own branch, the checkout is a
// clean resting branch, the remote still has the branch at the recorded tip,
// and no local copy of the branch has diverged from it. nil, nil means the
// card holds no handoff branch (a plain claim).
func prepareTakeover(env *trackerEnv, card tracker.Record) (*takeover, error) {
	rel, ok, err := issue.CardRelease(card.Raw)
	if err != nil || !ok || rel.Branch == "" {
		return nil, err
	}
	if _, has, err := issue.CardClaimant(card.Raw); err != nil || has {
		return nil, err
	}
	id := issue.CLIRef(card.ID)
	want := issue.BranchName(card.Path)
	if rel.Branch != want {
		return nil, fmt.Errorf("#%s's release names branch %q, not its own %s; refusing to fetch it", id, rel.Branch, want)
	}
	if _, err := env.git("check-ref-format", "--branch", rel.Branch); err != nil {
		return nil, fmt.Errorf("#%s's release names an invalid branch %q: %w", id, rel.Branch, err)
	}
	if !env.onRest() {
		return nil, fmt.Errorf("#%s is taken over by checking out %s: run the claim from a resting branch (this checkout is on %q)", id, rel.Branch, env.branch)
	}
	if dirty, err := cleanTree(env); err != nil {
		return nil, err
	} else if dirty != "" {
		return nil, fmt.Errorf("#%s is taken over by checking out %s: this checkout must be clean first (nothing was changed):\n%s", id, rel.Branch, dirty)
	}
	tracking := "refs/remotes/" + env.target.Remote + "/" + rel.Branch
	if _, err := env.git("fetch", "-q", env.target.Remote, "+refs/heads/"+rel.Branch+":"+tracking); err != nil {
		return nil, fmt.Errorf("#%s: fetching %s failed: %w", id, rel.Branch, err)
	}
	tip, err := env.git("rev-parse", tracking)
	if err != nil {
		return nil, err
	}
	if tip != rel.Head {
		return nil, fmt.Errorf("#%s: %s is at %s, not the %s it was handed off at — someone pushed after the release; inspect it (`git log %s`) before taking over", id, rel.Branch, shortOID(tip), shortOID(rel.Head), tracking)
	}
	if local, err := env.git("rev-parse", "-q", "--verify", "refs/heads/"+rel.Branch); err == nil && local != "" && local != rel.Head {
		if ahead, err := env.gitTest("merge-base", "--is-ancestor", local, rel.Head); err != nil {
			return nil, err
		} else if !ahead {
			return nil, fmt.Errorf("#%s: the local %s has commits the handed-off tip lacks; reconcile them before taking over", id, rel.Branch)
		}
	}
	return &takeover{branch: rel.Branch, head: rel.Head}, nil
}

// finishTakeover puts the handed-off tip in this checkout: the local branch is
// created (or fast-forwarded) at it and checked out. A failure warns: the card
// already names this workspace, and a rerun of the claim finishes it.
func finishTakeover(env *trackerEnv, stderr io.Writer, t *takeover) {
	if _, err := env.git("branch", "-f", t.branch, t.head); err != nil {
		cwarn(stderr, fmt.Sprintf("%s not set at %s: %v; rerun `sdlc claim` to finish", t.branch, shortOID(t.head), err))
		return
	}
	if _, err := env.git("switch", "-q", t.branch); err != nil {
		cwarn(stderr, fmt.Sprintf("%s not checked out: %v; rerun `sdlc claim` to finish", t.branch, err))
		return
	}
	cok(stderr, fmt.Sprintf("resumed on %s at %s", t.branch, shortOID(t.head)))
}

// finishOwnedTakeover completes a takeover whose card landed but whose switch
// did not: a claim repeat on started work, from a clean resting branch, with
// the issue's branch present locally, checks it out.
func finishOwnedTakeover(env *trackerEnv, stderr io.Writer, card tracker.Record) {
	status, _ := issue.GetField(card.Card.Frontmatter, "status")
	if !env.onRest() || vocab.Issue().IsOpen(status) {
		return
	}
	branch := issue.BranchName(card.Path)
	if has, err := env.gitTest("rev-parse", "-q", "--verify", "refs/heads/"+branch); err != nil || !has {
		return
	}
	if dirty, err := cleanTree(env); err != nil || dirty != "" {
		return
	}
	if _, err := env.git("switch", "-q", branch); err != nil {
		cwarn(stderr, fmt.Sprintf("%s not checked out: %v", branch, err))
		return
	}
	cok(stderr, fmt.Sprintf("resumed on %s", branch))
}

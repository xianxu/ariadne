package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const handoffDetail = "workshop/issues/000009-s09.md"

// startedHere claims and starts #9 in r and commits work on its branch.
func startedHere(t *testing.T) (*trackerRepo, map[string]string, issue.Claimant) {
	t.Helper()
	r, paths := claimSetRepo(t)
	claimFor(t, 9)
	var out bytes.Buffer
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, "cmd/nine.go", "package nine\n")
	r.git("add", "-A")
	r.git("commit", "-qm", "#9: work in progress")
	owner, _ := ownerOf(t, r, paths["000009"])
	return r, paths, owner
}

// anotherMachine clones origin and returns its root and an identity on
// another machine, for the same repository.
func anotherMachine(t *testing.T, r *trackerRepo, owner issue.Claimant) (string, issue.Claimant) {
	t.Helper()
	peer := filepath.Join(t.TempDir(), "peer")
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	testfix.Git(t, peer, "config", "user.name", "Peer")
	testfix.Git(t, peer, "config", "user.email", "peer@example.com")
	them := owner
	them.Operator, them.Machine, them.MachineName, them.Workspace, them.Worktree = "Peer", issue.MachineFingerprint("machine-b"), "box-b", "", canonRoot(peer)
	return peer, them
}

func claimHere(t *testing.T, ids ...int) (string, error) {
	t.Helper()
	var out, errs bytes.Buffer
	err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: ids, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"})
	return out.String() + errs.String(), err
}

// #284: started work changes machines. Unclaim pushes the branch, records its
// tip and returns this checkout to rest; a claim on another machine fetches
// the branch and resumes at exactly that tip, the note travelling with it.
func TestHandoffAndTakeover(t *testing.T) {
	r, paths, owner := startedHere(t)
	if out, err := unclaim(t, "parser done; lexer next", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	card := r.card(paths["000009"])
	rel, ok, _ := issue.CardRelease([]byte(card))
	if _, held, _ := issue.CardClaimant([]byte(card)); held || !ok || rel.Branch != "000009-s09" || !strings.Contains(card, "status: working") {
		t.Fatalf("card not handed off:\n%s", card)
	}
	if tip := strings.Fields(r.git("ls-remote", "origin", "refs/heads/000009-s09"))[0]; tip != rel.Head {
		t.Fatalf("origin's branch %s is not the recorded head %s", tip, rel.Head)
	}
	if r.git("branch", "--show-current") != "main" {
		t.Fatal("the handing-off checkout did not return to rest")
	}

	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	out, err := claimHere(t, 9)
	if err != nil {
		t.Fatalf("takeover: %v\n%s", err, out)
	}
	if got, _ := ownerOf(t, r, paths["000009"]); got != them {
		t.Fatalf("owner %+v, want the other machine", got)
	}
	if _, ok, _ := issue.CardRelease([]byte(r.card(paths["000009"]))); ok {
		t.Fatal("the takeover did not spend the release")
	}
	if strings.TrimSpace(testfix.Capture(t, peer, "branch", "--show-current")) != "000009-s09" || strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD")) != rel.Head {
		t.Fatal("the taker is not on the branch at the handed-off tip")
	}
	if !strings.Contains(testfix.Capture(t, peer, "show", "HEAD:"+handoffDetail), "unclaimed: parser done; lexer next") {
		t.Fatal("the handoff note did not travel with the branch")
	}
}

// #284: a handoff carries only committed work from the issue's own branch.
func TestHandoffRefusals(t *testing.T) {
	r, _, _ := startedHere(t)
	writeRepoFile(t, r.root, "cmd/nine.go", "package nine // edited\n")
	if out, err := unclaim(t, "", 9); err == nil || !strings.Contains(err.Error(), "only committed work") {
		t.Fatalf("dirty tracked: %v\n%s", err, out)
	}
	r.git("checkout", "--", "cmd/nine.go")
	writeRepoFile(t, r.root, "scratch.txt", "x\n")
	if out, err := unclaim(t, "", 9); err == nil || !strings.Contains(err.Error(), "scratch.txt") {
		t.Fatalf("untracked: %v\n%s", err, out)
	}
	r.git("clean", "-qf")
	r.git("switch", "-q", "main")
	if out, err := unclaim(t, "", 9); err == nil || !strings.Contains(err.Error(), "run `sdlc unclaim` in the checkout on 000009-s09") {
		t.Fatalf("off the branch: %v\n%s", err, out)
	}
}

// #284: a lost card write from the handoff is settled by the rerun, which
// recognises its own release and only finishes the switch to rest.
func TestHandoffRerunAfterALostResponse(t *testing.T) {
	r, paths, _ := startedHere(t)
	restore := loseResponses(t)
	out, err := unclaim(t, "", 9)
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command (sdlc unclaim --issue 9)") {
		t.Fatalf("lost response: %v\n%s", err, out)
	}
	if out, err := unclaim(t, "", 9); err != nil || !strings.Contains(out, "already handed off by this workspace") {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	if r.git("branch", "--show-current") != "main" {
		t.Fatal("the rerun did not finish the switch")
	}
	if _, held, _ := issue.CardClaimant([]byte(r.card(paths["000009"]))); held {
		t.Fatal("still held")
	}
}

// #284: a takeover checks everything before writing the card: a tip pushed
// after the release, a checkout off rest or dirty, a diverged local branch.
func TestTakeoverRefusals(t *testing.T) {
	r, paths, owner := startedHere(t)
	if out, err := unclaim(t, "", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	before := r.card(paths["000009"])

	testfix.Git(t, peer, "switch", "-q", "-c", "elsewhere")
	if out, err := claimHere(t, 9); err == nil || !strings.Contains(err.Error(), "run the claim from a resting branch") || r.card(paths["000009"]) != before {
		t.Fatalf("off rest: %v\n%s", err, out)
	}
	testfix.Git(t, peer, "switch", "-q", "main")
	writeRepoFile(t, peer, "scratch.txt", "x\n")
	if out, err := claimHere(t, 9); err == nil || !strings.Contains(err.Error(), "must be clean first") || r.card(paths["000009"]) != before {
		t.Fatalf("dirty: %v\n%s", err, out)
	}
	testfix.Git(t, peer, "clean", "-qf")

	testfix.Git(t, peer, "branch", "000009-s09", "main") // a stale local copy that diverged
	testfix.Git(t, peer, "switch", "-q", "000009-s09")
	writeRepoFile(t, peer, "other.txt", "x\n")
	testfix.Git(t, peer, "add", "other.txt")
	testfix.Git(t, peer, "commit", "-qm", "#9: diverged")
	testfix.Git(t, peer, "switch", "-q", "main")
	if out, err := claimHere(t, 9); err == nil || !strings.Contains(err.Error(), "has commits the handed-off tip lacks") || r.card(paths["000009"]) != before {
		t.Fatalf("diverged: %v\n%s", err, out)
	}
	testfix.Git(t, peer, "branch", "-D", "000009-s09")

	// The original machine pushes after releasing: the tip moved.
	r.git("switch", "-q", "000009-s09")
	writeRepoFile(t, r.root, "late.txt", "x\n")
	r.git("add", "late.txt")
	r.git("commit", "-qm", "#9: pushed after the release")
	r.git("push", "-q", "origin", "000009-s09")
	if out, err := claimHere(t, 9); err == nil || !strings.Contains(err.Error(), "someone pushed after the release") || r.card(paths["000009"]) != before {
		t.Fatalf("moved tip: %v\n%s", err, out)
	}
}

// #284: the takeover's switch is lost (the card already names the taker): a
// rerun of the claim finishes it.
func TestTakeoverRerunFinishesTheSwitch(t *testing.T) {
	r, _, owner := startedHere(t)
	if out, err := unclaim(t, "", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	if out, err := claimHere(t, 9); err != nil {
		t.Fatalf("takeover: %v\n%s", err, out)
	}
	testfix.Git(t, peer, "switch", "-q", "main") // as if the switch never happened
	if out, err := claimHere(t, 9); err != nil || !strings.Contains(out, "resumed on 000009-s09") {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	if strings.TrimSpace(testfix.Capture(t, peer, "branch", "--show-current")) != "000009-s09" {
		t.Fatal("the rerun did not finish the switch")
	}
}

// #284: a card claimed before #277 — codecomplete, no owner, no release — is
// taken over by a plain claim; its status stays codecomplete.
func TestPlainClaimTakesOverAPre277Codecomplete(t *testing.T) {
	r, paths := claimSetRepo(t)
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000009", paths["000009"], "legacy", operationToken("set"), func(c []byte) ([]byte, error) {
		c, err := issue.SetCardField(c, "actual_hours", "1")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(c, "status", "codecomplete")
	}); err != nil {
		t.Fatal(err)
	}
	if out, err := claimHere(t, 9); err != nil {
		t.Fatalf("takeover: %v\n%s", err, out)
	}
	card := r.card(paths["000009"])
	if owner, ok := ownerOf(t, r, paths["000009"]); !ok || owner.Worktree != canonRoot(r.root) || !strings.Contains(card, "status: codecomplete") {
		t.Fatalf("not taken over with its status kept:\n%s", card)
	}
}

// #284 BR-23: the takeover's card write lands but its response is lost; the
// release (and its tip) is spent and this fresh clone has no local branch. The
// rerun resumes on the branch as the remote has it.
func TestTakeoverLostResponseRerunResumes(t *testing.T) {
	r, _, owner := startedHere(t)
	if out, err := unclaim(t, "", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	head := strings.Fields(r.git("ls-remote", "origin", "refs/heads/000009-s09"))[0]
	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	restore := loseResponses(t)
	out, err := claimHere(t, 9)
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command") {
		t.Fatalf("lost response: %v\n%s", err, out)
	}
	if out, err := claimHere(t, 9); err != nil || !strings.Contains(out, "resumed on 000009-s09") {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	if strings.TrimSpace(testfix.Capture(t, peer, "branch", "--show-current")) != "000009-s09" || strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD")) != head {
		t.Fatal("the rerun did not resume at the handed-off tip")
	}
}

// #284 BR-24: the card is released again at a new tip after the takeover
// fetched the old one: the claim refuses and nothing is checked out.
func TestTakeoverRefusesAReReleaseBeforeTheWrite(t *testing.T) {
	r, paths, owner := startedHere(t)
	if out, err := unclaim(t, "", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	prev := cardsPublish
	t.Cleanup(func() { cardsPublish = prev })
	cardsPublish = func(env *trackerEnv, ids []string, token string, trailers []string, decide func(map[string]tracker.Record) (map[string][]byte, error), before func(string, string) error) error {
		calls := 0
		return prev(env, ids, token, trailers, decide, func(base, candidate string) error {
			if calls++; calls == 1 {
				if err := env.repo.ChangeCard("000009", paths["000009"], "release", operationToken("set"), func(c []byte) ([]byte, error) {
					rel, _, _ := issue.CardRelease(c)
					rel.Head = strings.Repeat("e", 40)
					return issue.SetCardRelease(c, &rel)
				}); err != nil {
					return err
				}
			}
			if before != nil {
				return before(base, candidate)
			}
			return nil
		})
	}
	if out, err := claimHere(t, 9); err == nil || !strings.Contains(err.Error(), "was released again") {
		t.Fatalf("a re-release must refuse: %v\n%s", err, out)
	}
	if _, held, _ := issue.CardClaimant([]byte(r.card(paths["000009"]))); held {
		t.Fatal("the claim landed over a re-release")
	}
	if strings.TrimSpace(testfix.Capture(t, peer, "branch", "--show-current")) != "main" {
		t.Fatal("something was checked out")
	}
}

// #284 BR-24: the handoff's push leases on the copy this checkout last
// fetched: a tip pushed meanwhile by someone else is not overwritten.
func TestHandoffLeaseRefusesAnUnseenRemoteTip(t *testing.T) {
	r, _, owner := startedHere(t)
	if out, err := unclaim(t, "", 9); err != nil {
		t.Fatalf("handoff: %v\n%s", err, out)
	}
	peer, them := anotherMachine(t, r, owner)
	t.Chdir(peer)
	withClaimant(t, them)
	if out, err := claimHere(t, 9); err != nil {
		t.Fatalf("takeover: %v\n%s", err, out)
	}
	r.git("switch", "-q", "000009-s09") // the old machine pushes without the lock
	writeRepoFile(t, r.root, "rogue.txt", "x\n")
	r.git("add", "rogue.txt")
	r.git("commit", "-qm", "#9: rogue")
	r.git("push", "-q", "origin", "000009-s09")
	rogue := r.git("rev-parse", "HEAD")
	writeRepoFile(t, peer, "mine.txt", "x\n")
	testfix.Git(t, peer, "add", "mine.txt")
	testfix.Git(t, peer, "commit", "-qm", "#9: mine")
	if out, err := unclaim(t, "", 9); err == nil || !strings.Contains(err.Error(), "last fetched") {
		t.Fatalf("the lease must refuse an unseen tip: %v\n%s", err, out)
	}
	if tip := strings.Fields(r.git("ls-remote", "origin", "refs/heads/000009-s09"))[0]; tip != rogue {
		t.Fatal("the unseen tip was overwritten")
	}
}

// #284 BR-23: a note already filed under another date is not filed again.
func TestUnclaimNoteDedupesAcrossDates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "d.md")
	writeRepoFile(t, dir, "d.md", "---\nid: 000009\n---\n\n# t\n\n## Log\n\n### 2026-01-01\n\n- 2026-01-01: unclaimed: hand back\n")
	if changed, err := appendUnclaimNote(p, "hand back"); err != nil || changed {
		t.Fatalf("an earlier-dated note was filed again: %v %v", changed, err)
	}
	if changed, err := appendUnclaimNote(p, "something else"); err != nil || !changed {
		t.Fatalf("a new note was not filed: %v %v", changed, err)
	}
}

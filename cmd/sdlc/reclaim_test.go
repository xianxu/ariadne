package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func reclaimCard(t *testing.T, status string, owner *issue.Claimant) []byte {
	t.Helper()
	raw := []byte("---\nid: 000031\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\n" +
		map[bool]string{true: "actual_hours: 1\n", false: ""}[status == "codecomplete" || status == "done"] +
		"---\n\n# t\n\n## Problem\n\nx\n")
	if owner != nil {
		var err error
		if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
			t.Fatal(err)
		}
	}
	return raw
}

// #278: reclaim's decision over status × ownership × expect × reason. Only an
// owned card of another workspace, inspected at this revision, with a one-line
// reason, transfers; the owner's repeat is the no-op that reconciles a retry.
func TestReclaimDecision(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m2"), MachineName: "box2", Worktree: "/w/new", Repository: "r"}
	old := issue.Claimant{Operator: "Them", Machine: issue.MachineFingerprint("m1"), MachineName: "box1", Workspace: "r:1", Worktree: "/w/old", Repository: "r"}
	const rev = "1111111111111111111111111111111111111111"
	for _, c := range []struct {
		name, status   string
		owner          *issue.Claimant
		expect, reason string
		want           string // "transfer", "mine", or a refusal fragment
	}{
		{"working foreign", "working", &old, rev, "box1 died; operator moved the work", "transfer"},
		{"blocked foreign", "blocked", &old, rev, "r", "transfer"},
		{"codecomplete foreign", "codecomplete", &old, rev, "r", "transfer"},
		{"already mine (retry)", "working", &me, rev, "r", "mine"},
		{"mine, stale expect still no-op", "working", &me, "2222222222222222222222222222222222222222", "", "mine"},
		{"open", "open", nil, rev, "r", "sdlc claim --issue 31"},
		{"done", "done", &old, rev, "r", "no live responsibility"},
		{"unattributed", "working", nil, rev, "r", "--adopt"},
		{"no expect", "working", &old, "", "r", "--expect is required"},
		{"stale expect", "working", &old, "2222222222222222222222222222222222222222", "r", "changed since you inspected"},
		{"empty reason", "working", &old, rev, "  ", "--reason is required"},
		{"multi-line reason", "working", &old, rev, "a\nb", "--reason is required"},
	} {
		next, from, err := reclaimDecision(reclaimCard(t, c.status, c.owner), rev, c.expect, c.reason, me)
		switch c.want {
		case "transfer":
			got, ok, cerr := issue.CardClaimant(next)
			if err != nil || cerr != nil || !ok || got != me || from != old {
				t.Errorf("%s: want transfer from %+v, got %+v (from %+v) %v %v", c.name, old, got, from, err, cerr)
			}
		case "mine":
			if !errors.Is(err, errAlreadyMine) {
				t.Errorf("%s: want the owner's no-op, got %v", c.name, err)
			}
		default:
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: want refusal %q, got %v", c.name, c.want, err)
			}
		}
	}
}

func TestReclaimTrailersRoundTrip(t *testing.T) {
	from := issue.Claimant{Operator: "Them", MachineName: "box1", Workspace: "r:1", Worktree: "/w/old"}
	to := issue.Claimant{Operator: "Me", MachineName: "box2", Worktree: "/w/new"}
	msg := "#31: tracker: update card\n\nTracker-Operation: reclaim-x\n" + strings.Join(reclaimTrailers(from, to, " box1: disk failed ✓ "), "\n")
	f, tt, reason, ok := parseReclaimTrailers(msg)
	if !ok || f != describeClaimant(from) || tt != describeClaimant(to) || reason != "box1: disk failed ✓" {
		t.Fatalf("round trip: %q %q %q %v", f, tt, reason, ok)
	}
	if _, _, _, ok := parseReclaimTrailers("#31: tracker: update card\n\nTracker-Operation: claim-x"); ok {
		t.Fatal("a claim commit parsed as a reclaim")
	}
}

// reclaimFixture: #N claimed by a slot worktree (the old owner) holding its
// branch with uncommitted work; the primary checkout is the new workspace.
func reclaimFixture(t *testing.T, n int) (r *trackerRepo, slot, cardPath, detailPath string) {
	t.Helper()
	id := itoa6(n)
	cardPath, card, detailPath, detail := seededIssue(t, id, "reclaim")
	r = newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	slot = filepath.Join(t.TempDir(), "slot")
	testfix.Git(t, r.root, "worktree", "add", "-q", "-b", id+"-reclaim", slot)
	t.Chdir(slot)
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(n)); err != nil {
		t.Fatalf("claim in the slot: %v\n%s", err, errs.String())
	}
	t.Chdir(r.root)
	return r, slot, cardPath, detailPath
}

func reclaimRun(t *testing.T, n int, expect, reason string) (string, error) {
	t.Helper()
	var out, errs bytes.Buffer
	err := runReclaim(context.Background(), &out, &errs, &reclaimFlags{Issue: n, Expect: expect, Reason: reason, IssuesDir: "workshop/issues"})
	return out.String() + errs.String(), err
}

func trackerTip(t *testing.T, r *trackerRepo) string {
	t.Helper()
	return strings.Fields(testfix.Capture(t, r.root, "ls-remote", "origin", "refs/heads/issue-tracker"))[0]
}

var confirmRE = regexp.MustCompile(`sdlc reclaim --issue (\d+) --expect ([0-9a-f]{40,64}) --reason`)

// #278: inspect writes nothing and prints a working confirm command; confirm
// moves responsibility here, records old/new/reason in tracker history, and the
// old workspace can no longer continue. No worktree is touched on either side.
// An identical rerun is a no-op that writes nothing.
func TestReclaimTransfersResponsibility(t *testing.T) {
	r, slot, cardPath, detailPath := reclaimFixture(t, 380)
	old, _ := ownerOf(t, r, cardPath)
	writeRepoFile(t, slot, "wip.txt", "the old worker's uncommitted work\n")
	writeRepoFile(t, r.root, "here.txt", "the new worker's scratch\n")
	statusBefore := testfix.Capture(t, slot, "status", "--porcelain") + testfix.Capture(t, r.root, "status", "--porcelain")
	tip := trackerTip(t, r)
	view, err := reclaimRun(t, 380, "", "")
	if err != nil || trackerTip(t, r) != tip {
		t.Fatalf("inspect wrote or failed: %v\n%s", err, view)
	}
	for _, want := range []string{"current owner:  " + old.Operator, old.Worktree, "proposed owner: ", canonRoot(r.root), "(this workspace)"} {
		if !strings.Contains(view, want) {
			t.Errorf("inspect lacks %q:\n%s", want, view)
		}
	}
	m := confirmRE.FindStringSubmatch(view)
	if m == nil || m[1] != "380" {
		t.Fatalf("no confirm command:\n%s", view)
	}
	if out, err := reclaimRun(t, 380, m[2], "slot machine retired; operator agreed the move"); err != nil {
		t.Fatalf("confirm: %v\n%s", err, out)
	}
	now, _ := ownerOf(t, r, cardPath)
	if now.Worktree != canonRoot(r.root) {
		t.Fatalf("owner %+v", now)
	}
	msg := testfix.Capture(t, r.root, "log", "-1", "--format=%B", trackerTip(t, r))
	for _, want := range []string{"Reclaim-From: " + describeClaimant(old), "Reclaim-To: " + describeClaimant(now), "Reclaim-Reason: slot machine retired; operator agreed the move"} {
		if !strings.Contains(msg, want) {
			t.Errorf("tracker commit lacks %q:\n%s", want, msg)
		}
	}
	// Wrong-owner resume: the old workspace is refused; the new one owns it.
	t.Chdir(slot)
	if refusal, _ := ownershipGate(t, r, "start-plan", 380, detailPath); !strings.Contains(refusal, "owned by "+now.Operator) {
		t.Fatalf("old workspace continued after reclaim: %q", refusal)
	}
	t.Chdir(r.root)
	if statusAfter := testfix.Capture(t, slot, "status", "--porcelain") + testfix.Capture(t, r.root, "status", "--porcelain"); statusAfter != statusBefore {
		t.Fatalf("reclaim touched a worktree:\nbefore %q\nafter  %q", statusBefore, statusAfter)
	}
	if raw, _ := os.ReadFile(filepath.Join(slot, "wip.txt")); string(raw) != "the old worker's uncommitted work\n" {
		t.Fatal("the old worker's uncommitted work changed")
	}
	tip = trackerTip(t, r)
	if out, err := reclaimRun(t, 380, m[2], "slot machine retired; operator agreed the move"); err != nil || !strings.Contains(out, "nothing to reclaim") || trackerTip(t, r) != tip {
		t.Fatalf("identical rerun: %v\n%s", err, out)
	}
	view, _ = reclaimRun(t, 380, "", "")
	if !strings.Contains(view, "past reclaims:") || !strings.Contains(view, "slot machine retired") || !strings.Contains(view, "already owns it") {
		t.Fatalf("history not shown:\n%s", view)
	}
}

// #278: a card that changed after inspection refuses the stale confirm and
// keeps its owner; a lost publication response is reconciled by rerunning.
func TestReclaimStaleAndLostResponse(t *testing.T) {
	r, _, cardPath, _ := reclaimFixture(t, 381)
	m := confirmRE.FindStringSubmatch(func() string { v, _ := reclaimRun(t, 381, "", ""); return v }())
	var out, errs bytes.Buffer
	if err := runCardUpdate(context.Background(), &out, &errs, "workshop/issues", 381, "estimate", false, func(_ *trackerEnv, card tracker.Record, _ string) ([]byte, error) {
		return issue.SetCardField(card.Raw, "estimate_hours", "3")
	}); err != nil {
		t.Fatalf("intervening change: %v\n%s", err, errs.String())
	}
	before := r.card(cardPath)
	if got, err := reclaimRun(t, 381, m[2], "late"); err == nil || !strings.Contains(err.Error(), "changed since you inspected") || r.card(cardPath) != before {
		t.Fatalf("stale confirm: %v\n%s", err, got)
	}
	m = confirmRE.FindStringSubmatch(func() string { v, _ := reclaimRun(t, 381, "", ""); return v }())
	prev := reclaimEffect
	reclaimEffect = func(env *trackerEnv, card tracker.Record, next []byte, trailers []string) error {
		if err := prev(env, card, next, trailers); err != nil {
			return err
		}
		return fmt.Errorf("%w: push response lost", gitx.ErrPublicationUncertain)
	}
	got, err := reclaimRun(t, 381, m[2], "lost response")
	reclaimEffect = prev
	if err == nil || !strings.Contains(err.Error(), "rerun the same command") {
		t.Fatalf("lost response not reported: %v\n%s", err, got)
	}
	tip := trackerTip(t, r)
	if got, err := reclaimRun(t, 381, m[2], "lost response"); err != nil || !strings.Contains(got, "nothing to reclaim") || trackerTip(t, r) != tip {
		t.Fatalf("rerun after a lost response: %v\n%s", err, got)
	}
}

// #278: two clones confirm the same inspected revision at once: exactly one
// takes responsibility; the other refuses and publishes nothing.
func TestReclaimRaceHasExactlyOneWinner(t *testing.T) {
	if _, err := machineID(); err != nil {
		t.Skipf("reclaim needs the host machine ID, unreadable here: %v", err)
	}
	binary := buildFleetE2EBinary(t)
	cardPath, card, detailPath, detail := seededIssue(t, "000382", "reclaimrace")
	away := issue.Claimant{Operator: "Gone", Machine: issue.MachineFingerprint("dead-box"), MachineName: "dead-box", Worktree: "/gone/ariadne", Repository: "placeholder"}
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	away.Repository = env.target.Repository
	if err := env.repo.ChangeCard("000382", cardPath, "owner", operationToken("set"), func(c []byte) ([]byte, error) {
		w := strings.Replace(string(c), "status: open", "status: working\nstarted: 2026-10-01T09:00:00-07:00", 1)
		return issue.SetCardClaimant([]byte(w), away)
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	snap, _ := env.repo.Snapshot()
	rec, _ := snap.Card("000382")
	peer := filepath.Join(t.TempDir(), "peer")
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "Peer")
	git(t, peer, "config", "user.email", "peer@example.com")
	results := raceBuiltBinary(t, binary, []string{r.root, peer}, "reclaim", "--issue", "382", "--expect", rec.BlobOID, "--reason", "dead-box retired")
	wins, winner := 0, ""
	for _, res := range results {
		if res.err == nil {
			wins++
			winner = res.dir
		} else if !strings.Contains(res.out, "changed while reclaiming") && !strings.Contains(res.out, "changed since you inspected") {
			t.Errorf("loser was not a stale/CAS refusal: %v %s", res.err, res.out)
		}
	}
	if wins != 1 {
		t.Fatalf("reclaim winners=%d; want exactly one", wins)
	}
	if owner, _ := ownerOf(t, r, cardPath); owner.Worktree != canonRoot(winner) {
		t.Fatalf("owner %+v, want the winner %s", owner, winner)
	}
}

// #278: nothing but the operator's `sdlc reclaim` reaches the transfer — no
// timeout, reachability, move or recovery path calls it. Every production
// reference to the reclaim entry points must sit in its one allowed caller.
func TestReclaimIsOnlyOperatorInvoked(t *testing.T) {
	allowed := map[string]string{ // symbol → the only function that may reference it
		"reclaimEffect":   "runReclaim",
		"reclaimDecision": "runReclaim|inspectReclaim",
		"runReclaim":      "NewReclaimCmd",
		"NewReclaimCmd":   "buildRoot",
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool { return !strings.HasSuffix(info.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if !ok {
						return true
					}
					callers, guarded := allowed[id.Name]
					if !guarded {
						return true
					}
					seen[id.Name] = true
					if !regexp.MustCompile(`^(` + callers + `)$`).MatchString(fn.Name.Name) {
						t.Errorf("%s:%s references %s; only %s may (reclaim is operator-invoked only)", name, fn.Name.Name, id.Name, callers)
					}
					return true
				})
			}
		}
	}
	for sym := range allowed {
		if !seen[sym] {
			t.Errorf("guard is vacuous: no reference to %s found", sym)
		}
	}
}

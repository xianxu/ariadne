package main

import (
	"strings"
	"testing"
)

func TestSessionMovePlanLinks(t *testing.T) {
	old := "teams/dev/tl-status-op.md"
	doc := "teams/dev/tl-status-op.side/01-a.md"
	s := sessionMoveSnapshot{Old: old, Files: map[string]moveSource{
		old: {Bytes: []byte("---\ntype: session-sync\ncouch-owner: ops:0\nfuture:\n  - doc: leave/me.md\ndiscussions:\n  - doc: tl-status-op.side/01-a.md\n---\n[x](tl-status-op.side/01-a.md) [project](../../team.md)\n")},
		doc: {Bytes: []byte("---\ntype: side-doc\nparent: ../tl-status-op.md\n---\n[parent](../tl-status-op.md) [project](../../../team.md)\n")},
	}, Existing: map[string]bool{old: true, doc: true, "team.md": true}}
	p, err := planSessionMove(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Files) != 2 {
		t.Fatalf("files: %+v", p.Files)
	}
	for _, f := range p.Files {
		text := string(f.After)
		if f.New == "session.md" {
			for _, want := range []string{"doc: session.side/01-a.md", "[project](team.md)", "future:\n  - doc: leave/me.md"} {
				if !strings.Contains(text, want) {
					t.Errorf("session missing %q: %s", want, text)
				}
			}
		} else if !strings.Contains(text, "parent: ../session.md") || !strings.Contains(text, "[project](../team.md)") {
			t.Errorf("side doc not rebased: %s", text)
		}
	}
}

func TestSessionMoveTransition(t *testing.T) {
	for n := 1; n <= 5; n++ {
		state := sessionMoveState{Total: n, Effect: movePrepare}
		verified := false
		copies, switches, removed := 0, 0, 0
		for state.Effect != moveDone {
			// Every interruption stops authority, including a published-but-unconfirmed rename.
			for _, outcome := range []moveOutcome{moveFailed, moveUnconfirmed} {
				stopped, err := transitionMove(state, outcome)
				if err != nil || stopped.Effect != moveStop {
					t.Fatalf("unsafe failure: %+v %v", stopped, err)
				}
				if _, err := transitionMove(stopped, moveSucceeded); err == nil {
					t.Fatal("resumed without fresh observations")
				}
			}
			switch state.Effect {
			case moveCopy:
				copies++
			case moveSwitch:
				if copies != n {
					t.Fatal("switch before all copies")
				}
				switches++
			case moveVerify:
				if switches != n {
					t.Fatal("verification before all bindings")
				}
				verified = true
			case moveRemove:
				if !verified || copies != n || switches != n {
					t.Fatal("deletion without verification")
				}
				removed++
			}
			var err error
			state, err = transitionMove(state, moveSucceeded)
			if err != nil {
				t.Fatal(err)
			}
		}
		if removed != n {
			t.Fatalf("cleanup count %d", removed)
		}
	}
}

func TestSessionMoveIgnores(t *testing.T) {
	before := []byte("# keep\nold-pattern\n")
	after := moveIgnores(before)
	if !strings.HasPrefix(string(after), string(before)) || string(moveIgnores(after)) != string(after) {
		t.Fatalf("not idempotent: %s", after)
	}
	for _, want := range []string{"/session.md\n", "/session.side/\n", "/.session.md.ss-seen\n"} {
		if !strings.Contains(string(after), want) {
			t.Errorf("missing %s", want)
		}
	}
}

func TestMoveSessionMetadata(t *testing.T) {
	for _, text := range []string{
		"---\ntype: session-sync\ntype: session-sync\n---\n",
		"---\ntype: session-sync\ncouch-owner: 123\n---\n",
		"---\ntype: session-sync\noperator: true\n---\n",
		"---\ntype: session-sync\n",
	} {
		if _, err := moveSessionFields(text); err == nil {
			t.Errorf("accepted %q", text)
		}
	}
	f, err := moveSessionFields("---\ntype: \"session-sync\"\ncouch-owner: 'ops:0'\nfuture: {n: 1e400}\n---\n")
	if err != nil || f["couch-owner"] != "ops:0" {
		t.Fatalf("%+v %v", f, err)
	}
}

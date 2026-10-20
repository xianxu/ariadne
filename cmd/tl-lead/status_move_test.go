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
	for _, phase := range []sessionMovePhase{movePreparing, moveSwitching, moveCleanup, moveComplete} {
		if _, err := transitionMove(phase, "cleanup"); (phase != moveSwitching) != (err != nil) {
			t.Fatalf("cleanup from %v: %v", phase, err)
		}
	}
	p := movePreparing
	for _, e := range []string{"copied", "cleanup", "complete"} {
		var err error
		p, err = transitionMove(p, e)
		if err != nil {
			t.Fatal(err)
		}
	}
	if p != moveComplete {
		t.Fatal(p)
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

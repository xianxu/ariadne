package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// #633: set-kind marks a card bug or feature. The weekly trend reads a closed
// card's updated as its close date, so marking a closed card (the TL's
// back-fill) must leave updated alone.
func TestIssueSetKind(t *testing.T) {
	cardPath, card, _, _ := seededIssue(t, "000001", "x")
	card = strings.Replace(card, "status: open", "status: done\nactual_hours: N/A", 1)
	card = strings.Replace(card, "updated: 2026-09-01", "updated: 2026-09-03", 1)
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)

	if out, err := runSdlc(t, "issue", "set-kind", "--issue", "1", "bug"); err != nil {
		t.Fatalf("set-kind: %v\n%s", err, out)
	}
	got := r.card(cardPath)
	if !strings.Contains(got, "kind: bug") {
		t.Fatalf("card lacks kind: bug:\n%s", got)
	}
	if !strings.Contains(got, "updated: 2026-09-03") {
		t.Fatalf("set-kind moved a closed card's updated (its close date):\n%s", got)
	}
	out, err := runSdlc(t, "issue", "set-kind", "--issue", "1", "bug")
	if err != nil || !strings.Contains(out, "already has that kind") {
		t.Fatalf("rerun: err %v, want a no-op\n%s", err, out)
	}
	for _, bad := range []struct {
		args []string
		want string
	}{
		{[]string{"issue", "set-kind", "--issue", "1", "chore"}, "bug, feature"},
		{[]string{"issue", "set-kind", "--issue", "1"}, "--clear"},
	} {
		out, err := runSdlc(t, bad.args...)
		if err == nil || !strings.Contains(err.Error()+out, bad.want) {
			t.Fatalf("%v: err %v, want %q\n%s", bad.args, err, bad.want, out)
		}
	}
	if after := r.card(cardPath); after != got {
		t.Fatalf("a refused set-kind changed the card:\n%s", after)
	}
	if out, err := runSdlc(t, "issue", "set-kind", "--issue", "1", "--clear"); err != nil {
		t.Fatalf("clear: %v\n%s", err, out)
	}
	parsed, err := issue.ParseCard([]byte(r.card(cardPath)))
	if err != nil {
		t.Fatal(err)
	}
	if kind, _ := issue.FieldString(parsed.Frontmatter, "kind"); kind != "" {
		t.Fatalf("kind not cleared: %q", kind)
	}
}

// A kind outside the enum is refused when a card is read, as severity is.
func TestParseCardRefusesUnknownKind(t *testing.T) {
	_, card, _, _ := seededIssue(t, "000001", "x")
	bad := strings.Replace(card, "status: open", "status: open\nkind: chore", 1)
	if _, err := issue.ParseCard([]byte(bad)); err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("ParseCard accepted kind: chore: %v", err)
	}
}

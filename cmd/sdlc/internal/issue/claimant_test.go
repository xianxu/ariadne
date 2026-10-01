package issue

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

const claimantCard = "---\nid: 000277\nstatus: open\ncreated: 2026-10-01\nupdated: 2026-10-01\n---\n\n# Ownership\n\n## Problem\n\nx\n"

func testClaimant() Claimant {
	return Claimant{Operator: "Xian Xu", Machine: MachineFingerprint("UUID-1"), MachineName: "MacBook Pro",
		Workspace: "ariadne:1", Worktree: "/w/ariadne-slot1/ariadne", Repository: "github.com/xianxu/ariadne"}
}

func TestSetCardClaimantRoundTripsAndKeepsOtherBytes(t *testing.T) {
	c := testClaimant()
	out, err := SetCardClaimant([]byte(claimantCard), c)
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := CardClaimant(out)
	if err != nil || !ok || got != c {
		t.Fatalf("round trip: %+v %v %v\n%s", got, ok, err, out)
	}
	if !strings.HasPrefix(string(out), "---\nid: 000277\nstatus: open\ncreated: 2026-10-01\nupdated: 2026-10-01\nclaimant:\n") ||
		!strings.HasSuffix(string(out), "---\n\n# Ownership\n\n## Problem\n\nx\n") {
		t.Fatalf("other bytes disturbed:\n%s", out)
	}
	// Replacement, and a slot-less (non-Couch, plain clone) claimant.
	plain := c
	plain.Workspace, plain.Worktree = "", "/home/u/src/ariadne"
	again, err := SetCardClaimant(out, plain)
	if err != nil {
		t.Fatal(err)
	}
	if got, _, _ := CardClaimant(again); got != plain || strings.Count(string(again), "claimant:") != 1 || strings.Contains(string(again), "workspace:") {
		t.Fatalf("replace: %+v\n%s", got, again)
	}
	if _, ok, err := CardClaimant([]byte(claimantCard)); ok || err != nil {
		t.Fatalf("absent claimant: %v %v", ok, err)
	}
}

func TestClaimantValidationFailsClosed(t *testing.T) {
	head := "---\nid: 000277\nstatus: working\ncreated: 2026-10-01\nupdated: 2026-10-01\n"
	tail := "---\n\n# Ownership\n\n## Problem\n\nx\n"
	full := "claimant:\n    operator: a\n    machine: 0123456789abcdef0123456789abcdef\n    machine_name: m\n    worktree: /w\n    repository: r\n"
	if _, err := ParseCard([]byte(head + full + tail)); err != nil {
		t.Fatalf("valid claimant refused: %v", err)
	}
	for name, block := range map[string]string{
		"extra key":        full + "    pid: '1'\n",
		"missing worktree": strings.Replace(full, "    worktree: /w\n", "", 1),
		"non-string":       strings.Replace(full, "operator: a", "operator: [a]", 1),
		"integer":          strings.Replace(full, "machine_name: m", "machine_name: 7", 1),
		"null value":       strings.Replace(full, "operator: a", "operator:", 1),
		"empty string":     strings.Replace(full, "operator: a", "operator: ''", 1),
		"scalar":           "claimant: me\n",
		"flow nonsense":    "claimant: {operator: a}\n",
		"bad fingerprint":  strings.Replace(full, "0123456789abcdef0123456789abcdef", "UUID-RAW", 1),
	} {
		_, err := ParseCard([]byte(head + block + tail))
		var v *ValidationError
		if !errors.As(err, &v) || v.Field != ClaimantField {
			t.Errorf("%s: accepted or wrong error: %v", name, err)
		}
	}
}

func TestMatchClaimant(t *testing.T) {
	me := testClaimant()
	other := func(f func(*Claimant)) Claimant { c := me; f(&c); return c }
	for name, c := range map[string]struct {
		recorded *Claimant
		want     Ownership
	}{
		"absent":                       {nil, OwnershipUnknown},
		"same workspace":               {&me, OwnershipMine},
		"another operator, same place": {ptr(other(func(c *Claimant) { c.Operator = "Someone" })), OwnershipMine},
		"slot label differs":           {ptr(other(func(c *Claimant) { c.Workspace = "" })), OwnershipMine},
		"other worktree, same machine": {ptr(other(func(c *Claimant) { c.Worktree = "/w/ariadne-slot2/ariadne" })), OwnershipForeign},
		"other machine, same path":     {ptr(other(func(c *Claimant) { c.Machine = MachineFingerprint("UUID-2") })), OwnershipForeign},
		"other repository":             {ptr(other(func(c *Claimant) { c.Repository = "github.com/x/fork" })), OwnershipForeign},
	} {
		if got := MatchClaimant(c.recorded, me); got != c.want {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}
}

func TestRelocationAllowed(t *testing.T) {
	old := testClaimant()
	here := old
	here.Workspace, here.Worktree = "ariadne:0", "/w/ariadne"
	for name, c := range map[string]struct {
		recorded           Claimant
		onBranch, oldHolds bool
		want               bool
	}{
		"owner moved its branch":          {old, true, false, true},
		"old worktree still holds it":     {old, true, true, false},
		"not on the issue branch here":    {old, false, false, false},
		"another machine":                 {func() Claimant { c := old; c.Machine = MachineFingerprint("UUID-2"); return c }(), true, false, false},
		"another repository":              {func() Claimant { c := old; c.Repository = "r2"; return c }(), true, false, false},
		"already here (convergent no-op)": {here, true, false, false},
	} {
		if got := RelocationAllowed(c.recorded, here, c.onBranch, c.oldHolds); got != c.want {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}
}

func TestMachineFingerprintIsKeyedAndStable(t *testing.T) {
	a := MachineFingerprint("2F1C-UUID")
	if a != MachineFingerprint("2F1C-UUID") || !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(a) {
		t.Fatalf("fingerprint %q", a)
	}
	if a == MachineFingerprint("2F1C-UUIE") || strings.Contains(strings.ToLower(a), "2f1c") {
		t.Fatal("fingerprint does not separate machines or leaks the raw ID")
	}
}

func ptr(c Claimant) *Claimant { return &c }

// #277: the claimant is a mirrored card field — it appears in details, is
// replaced and removed by refresh, and a hand edit in details is refused.
func TestClaimantMirrorsIntoDetails(t *testing.T) {
	baseline, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := SetCardClaimant([]byte(strings.Replace(string(baseline), "status: open", "status: working", 1)), testClaimant())
	if err != nil {
		t.Fatal(err)
	}
	withOwner, err := RefreshMirror(detail, baseline, claimed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(withOwner), "claimant:\n    operator: Xian Xu\n") || !strings.Contains(string(withOwner), "worktree: /w/ariadne-slot1/ariadne") {
		t.Fatalf("claimant not mirrored:\n%s", withOwner)
	}
	moved := testClaimant()
	moved.Workspace, moved.Worktree = "ariadne:0", "/w/ariadne"
	relocated, err := SetCardClaimant(claimed, moved)
	if err != nil {
		t.Fatal(err)
	}
	replaced, err := RefreshMirror(withOwner, claimed, relocated)
	if err != nil || strings.Contains(string(replaced), "ariadne-slot1") || !strings.Contains(string(replaced), "worktree: /w/ariadne\n") {
		t.Fatalf("claimant not replaced: %v\n%s", err, replaced)
	}
	removed, err := RefreshMirror(replaced, relocated, baseline)
	if err != nil || strings.Contains(string(removed), "claimant") {
		t.Fatalf("claimant not removed: %v\n%s", err, removed)
	}
	edited := strings.Replace(string(withOwner), "operator: Xian Xu", "operator: Someone Else", 1)
	var own *OwnershipError
	if _, err := RefreshMirror([]byte(edited), claimed, relocated); !errors.As(err, &own) || own.Field != ClaimantField || own.Setter != "sdlc claim" {
		t.Fatalf("hand-edited claimant: %v", err)
	}
}

package issue

import (
	"strings"
	"testing"
)

func releaseCard(t *testing.T) []byte {
	t.Helper()
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	return card
}

var releaser = Claimant{Operator: "Me", Machine: MachineFingerprint("m1"), MachineName: "box", Workspace: "r:2", Worktree: "/w/a", Repository: "r"}

// #284: an envelope rewrite keeps keys this binary does not know, so an older
// sdlc can never silently drop a newer record (and this one never drops a
// future one).
func TestEnvelopeKeepsUnknownKeys(t *testing.T) {
	card := releaseCard(t)
	withFuture := []byte(strings.Replace(string(card), "---\n\n#", "tracker:\n  version: 1\n  future:\n    x: 1\n    note: kept\n---\n\n#", 1))
	if _, err := ParseCard(withFuture); err != nil {
		t.Fatal(err)
	}
	out, err := SetCardCompletion(withFuture, Completion{Token: "close-x", Repository: "r", ReviewedHEAD: strings.Repeat("a", 40), EvidenceCommit: strings.Repeat("b", 40)})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"future:", "x: 1", "note: kept", "completion:"} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("rewrite lost %q:\n%s", want, out)
		}
	}
}

// #284: the release record round-trips; a started release names its branch and
// head; nil clears it; malformed ones fail closed — the record is read on
// another machine, so its branch is held to an issue branch name.
func TestReleaseRecord(t *testing.T) {
	card := releaseCard(t)
	open, err := SetCardRelease(card, &Release{By: ReleasedBy(releaser)})
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := CardRelease(open)
	if err != nil || !ok || got.By.Claimant() != releaser || got.Branch != "" {
		t.Fatalf("open release: %+v %v %v", got, ok, err)
	}
	started, err := SetCardRelease(card, &Release{By: ReleasedBy(releaser), Branch: "000284-claims-handoff-takeover", Head: strings.Repeat("c", 40)})
	if err != nil {
		t.Fatal(err)
	}
	if got, _, _ := CardRelease(started); got.Branch != "000284-claims-handoff-takeover" || got.Head != strings.Repeat("c", 40) {
		t.Fatalf("started release: %+v", got)
	}
	cleared, err := SetCardRelease(started, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := CardRelease(cleared); ok {
		t.Fatal("nil did not clear the release")
	}
	for name, r := range map[string]Release{
		"branch without head": {By: ReleasedBy(releaser), Branch: "000284-x"},
		"head without branch": {By: ReleasedBy(releaser), Head: strings.Repeat("c", 40)},
		"short head":          {By: ReleasedBy(releaser), Branch: "000284-x", Head: "abc"},
		"dot-dot branch":      {By: ReleasedBy(releaser), Branch: "000284-..x", Head: strings.Repeat("c", 40)},
		"leading-dash branch": {By: ReleasedBy(releaser), Branch: "-x", Head: strings.Repeat("c", 40)},
		"not an issue branch": {By: ReleasedBy(releaser), Branch: "main", Head: strings.Repeat("c", 40)},
		"path branch":         {By: ReleasedBy(releaser), Branch: "000284-x/../../y", Head: strings.Repeat("c", 40)},
		"raw machine id":      {By: releasedByRaw(releaser), Branch: "", Head: ""},
		"no releaser":         {},
	} {
		if _, err := SetCardRelease(card, &r); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func releasedByRaw(c Claimant) ReleaseBy {
	b := ReleasedBy(c)
	b.Machine = "RAW-MACHINE-ID"
	return b
}

// #284: clearing the claimant removes exactly that field; clearing an
// unclaimed card changes nothing.
func TestClearCardClaimant(t *testing.T) {
	card := releaseCard(t)
	owned, err := SetCardClaimant(card, releaser)
	if err != nil {
		t.Fatal(err)
	}
	cleared, err := ClearCardClaimant(owned)
	if err != nil {
		t.Fatal(err)
	}
	if string(cleared) != string(card) {
		t.Fatalf("clear is not the inverse of set:\n%s\nwant\n%s", cleared, card)
	}
	if again, err := ClearCardClaimant(card); err != nil || string(again) != string(card) {
		t.Fatalf("clearing an unclaimed card: %v", err)
	}
}

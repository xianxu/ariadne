package issue

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// modeKeyLines returns the frontmatter lines of doc that write a review mode
// key, which #575 retired and #576 stops every writer emitting.
func modeKeyLines(doc string) []string {
	var hits []string
	for _, l := range strings.Split(doc, "\n") {
		for _, k := range append([]string{"collaboration-mode"}, sidedoc.RetiredKeys...) {
			if strings.HasPrefix(l, k+":") {
				hits = append(hits, l)
			}
		}
	}
	return hits
}

// #576: new issues carry no mode key, from either creation path.
func TestNewIssuesWriteNoModeKeys(t *testing.T) {
	if hits := modeKeyLines(Render(ScaffoldSpec{ID: "000057", Title: "x", Today: "2026-05-31", Area: []string{"a"}, Severity: "low"})); len(hits) != 0 {
		t.Errorf("Render writes %q", hits)
	}
	card, _ := SetCardHandoff([]byte(setterCard), testHandoff())
	details, err := DetailsFromCard(card, "sha1", "2026-09-25")
	if err != nil {
		t.Fatal(err)
	}
	if hits := modeKeyLines(string(details)); len(hits) != 0 {
		t.Errorf("DetailsFromCard writes %q", hits)
	}
}

// #576: an issue written before #575 keeps its legacy key line through a
// frontmatter rewrite: readers never require it and writers never drop it.
func TestLegacyModeKeySurvivesSetField(t *testing.T) {
	fm := "id: 000001\nstatus: open\ncollaboration-mode: collaboration\npresentation-mode: propose\nupdated: 2026-01-01"
	got := SetField(fm, "updated", "2026-10-23")
	want := strings.Replace(fm, "updated: 2026-01-01", "updated: 2026-10-23", 1)
	if got != want {
		t.Errorf("SetField rewrote a legacy key:\n got %q\nwant %q", got, want)
	}
}

// #576: a card publish (SplitCard) keeps an older issue's legacy mode keys
// in its details byte for byte, and the card never carries them.
func TestLegacyModeKeysSurviveCardPublish(t *testing.T) {
	legacy := strings.Replace(cardDetailFixture, "deps: [pair#17]\n", "deps: [pair#17]\ncollaboration-mode: collaboration\npresentation-mode: propose\nprocess-instruction: \"[copy-edit] x\"\n", 1)
	card, details, err := SplitCard([]byte(legacy))
	if err != nil {
		t.Fatal(err)
	}
	if hits := modeKeyLines(string(card)); len(hits) != 0 {
		t.Errorf("the card carries %q", hits)
	}
	fm, _, _ := Parse(legacy)
	gotFM, _, _ := Parse(string(details))
	if !strings.HasPrefix(gotFM, fm) {
		t.Errorf("details frontmatter lost bytes:\n got %q\nwant prefix %q", gotFM, fm)
	}
	for _, k := range []string{"collaboration-mode: collaboration\n", "presentation-mode: propose\n", "process-instruction: \"[copy-edit] x\"\n"} {
		if !strings.Contains(string(details), k) {
			t.Errorf("details lost %q", k)
		}
	}
}

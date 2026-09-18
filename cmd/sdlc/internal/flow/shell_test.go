package flow

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// TestMeasure: lines come from the numstat rows of code files only — tests and
// docs never count — and however many code files a window spreads across, only
// their added lines are summed (pair#283's shape: seven files, a few lines each).
func TestMeasure(t *testing.T) {
	stats := []churn.FileStat{
		{Path: "cmd/a.go", Insertions: 30},
		{Path: "cmd/a_test.go", Insertions: 400},
		{Path: "README.md", Insertions: 90},
		{Path: "tests/x_spec.lua", Insertions: 70},
		{Path: "construct/vocabulary/issue.cue", Insertions: 5},
	}
	for i := 0; i < 7; i++ {
		stats = append(stats, churn.FileStat{Path: "internal/f" + string(rune('a'+i)) + ".go", Insertions: 4})
	}
	got := Measure(stats, []string{"M1"})
	if got.AddedLines != 63 {
		t.Errorf("AddedLines = %d, want 63 (30 + 5 + 7×4; tests and docs excluded)", got.AddedLines)
	}
	if strings.Join(got.Milestones, ",") != "M1" {
		t.Errorf("Milestones = %v", got.Milestones)
	}
}

// TestCrossings pins each limit at its edge: exactly at the limit stays inside
// the shell, one past it crosses — and every crossing names itself.
func TestCrossings(t *testing.T) {
	cases := []struct {
		name string
		size Size
		want []string // substrings, one per expected reason
	}{
		{"at the line limit", Size{AddedLines: MaxAddedLines}, nil},
		{"empty", Size{}, nil},
		{"one line too many", Size{AddedLines: MaxAddedLines + 1}, []string{"added lines"}},
		{"Mx milestones", Size{Milestones: []string{"M1"}}, []string{"milestones"}},
		{"an earlier full round", Size{EarlierFullReview: true}, []string{"earlier round"}},
		{"unreadable ledger", Size{LedgerErr: errors.New("corrupt")}, []string{"ledger"}},
		{"everything", Size{AddedLines: 500, Milestones: []string{"M1"},
			EarlierFullReview: true, LedgerErr: errors.New("e")},
			[]string{"added lines", "milestones", "earlier round", "ledger"}},
	}
	for _, c := range cases {
		got := c.size.Crossings()
		if len(got) != len(c.want) {
			t.Errorf("%s: %d crossings %v, want %d", c.name, len(got), got, len(c.want))
			continue
		}
		for i, w := range c.want {
			if !strings.Contains(got[i], w) {
				t.Errorf("%s: crossing %d = %q, want it to name %q", c.name, i, got[i], w)
			}
		}
	}
}

// TestUpgrade: crossing the shell makes the record full/inferred — whoever
// declared quick — and drops the contract hashes, which only the quick flow uses.
func TestUpgrade(t *testing.T) {
	q, err := Parse(`{kind: quick, provenance: operator, spec: "1a2b3c4d", done: "5e6f7a8b"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := Upgrade(q); got != (Flow{kind: Full, provenance: Inferred}) {
		t.Errorf("Upgrade = %+v, want full/inferred with no hashes", got)
	}
}

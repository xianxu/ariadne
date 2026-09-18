package flow

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// TestMeasure: code files come from the file list (so a binary, which numstat
// cannot count, still counts as a file), lines from the numstat rows of code
// files only; tests and docs are neither.
func TestMeasure(t *testing.T) {
	files := []string{"cmd/a.go", "cmd/a_test.go", "README.md", "assets/logo.png", "tests/x_spec.lua", "construct/vocabulary/issue.cue"}
	stats := []churn.FileStat{
		{Path: "cmd/a.go", Insertions: 30},
		{Path: "cmd/a_test.go", Insertions: 400},
		{Path: "README.md", Insertions: 90},
		{Path: "tests/x_spec.lua", Insertions: 70},
		{Path: "construct/vocabulary/issue.cue", Insertions: 5},
	}
	got := Measure(files, stats, []string{"M1"})
	if strings.Join(got.CodeFiles, ",") != "cmd/a.go,assets/logo.png,construct/vocabulary/issue.cue" {
		t.Errorf("CodeFiles = %v", got.CodeFiles)
	}
	if got.AddedLines != 35 {
		t.Errorf("AddedLines = %d, want 35 (tests and docs excluded)", got.AddedLines)
	}
	if strings.Join(got.Milestones, ",") != "M1" {
		t.Errorf("Milestones = %v", got.Milestones)
	}
}

func files(n int) []string {
	var out []string
	for i := 0; i < n; i++ {
		out = append(out, "cmd/f"+string(rune('a'+i))+".go")
	}
	return out
}

// TestCrossings pins each limit at its edge: exactly at the limit stays inside
// the shell, one past it crosses — and every crossing names itself.
func TestCrossings(t *testing.T) {
	cases := []struct {
		name string
		size Size
		want []string // substrings, one per expected reason
	}{
		{"at both limits", Size{CodeFiles: files(MaxCodeFiles), AddedLines: MaxChangedLines}, nil},
		{"empty", Size{}, nil},
		{"one file too many", Size{CodeFiles: files(MaxCodeFiles + 1)}, []string{"code files"}},
		{"one line too many", Size{AddedLines: MaxChangedLines + 1}, []string{"added lines"}},
		{"Mx milestones", Size{Milestones: []string{"M1"}}, []string{"milestones"}},
		{"an earlier full round", Size{EarlierFullReview: true}, []string{"earlier round"}},
		{"unreadable ledger", Size{LedgerErr: errors.New("corrupt")}, []string{"ledger"}},
		{"everything", Size{CodeFiles: files(3), AddedLines: 500, Milestones: []string{"M1"},
			EarlierFullReview: true, LedgerErr: errors.New("e")},
			[]string{"code files", "added lines", "milestones", "earlier round", "ledger"}},
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

// TestCrossingsCapTheFileList: a crossing's reason lands on one Log line, so a
// large diff names a bounded sample and the count, not every path.
func TestCrossingsCapTheFileList(t *testing.T) {
	got := Size{CodeFiles: files(12)}.Crossings()
	if len(got) != 1 || !strings.Contains(got[0], "12 code files") || !strings.Contains(got[0], "and 7 more") {
		t.Errorf("Crossings = %v, want the count and a capped list", got)
	}
}

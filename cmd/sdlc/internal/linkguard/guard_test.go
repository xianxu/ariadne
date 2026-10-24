package linkguard

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCensus(t *testing.T) {
	files := map[string]string{
		"w/i.side/2026-10/01-a.md": "---\nparent: ../../i.md\n---\n[nested](02-b.md) [back](01-a.md)",
		"w/i.side/2026-10/02-b.md": "[loop](01-a.md) [draft](../draft.md)",
	}
	states := map[string]LinkState{
		"w/i.side/2026-10/01-a.md": Committed,
		"w/i.side/2026-10/02-b.md": Committed,
		"w/i.side/draft.md":        Uncommitted,
		"w/i.md":                   Committed,
		"w/private.md":             Ignored,
		"w/gone.md":                Absent,
		"w/ok.md":                  Committed,
	}
	asked := map[string]int{}
	state := func(p string) (LinkState, error) {
		asked[p]++
		st, ok := states[p]
		if !ok {
			return 0, errors.New("unexpected " + p)
		}
		return st, nil
	}
	read := func(p string) (string, error) { return files[p], nil }
	docs := []Doc{{Path: "w/i.md", Text: "[q](i.side/2026-10/01-a.md)\n[p](private.md) [g](gone.md) [o](ok.md) [o2](ok.md)"}}
	got, err := Census(docs, state, read)
	if err != nil {
		t.Fatal(err)
	}
	want := []Refusal{
		{Doc: "w/i.md", Line: 2, Target: "private.md", Path: "w/private.md", State: Ignored},
		{Doc: "w/i.side/2026-10/02-b.md", Line: 1, Target: "../draft.md", Path: "w/i.side/draft.md", State: Uncommitted},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
	for p, n := range asked {
		if n != 1 {
			t.Errorf("%s asked %d times", p, n)
		}
	}
	msg := FormatRefusals(got)
	for _, w := range []string{`w/i.md:2 links "private.md", which is git-ignored`, `fold what it settled into the artifact`,
		`w/i.side/2026-10/02-b.md:1 links "../draft.md", which is not committed: commit "w/i.side/draft.md" with the artifact`} {
		if !strings.Contains(msg, w) {
			t.Errorf("missing %q in\n%s", w, msg)
		}
	}
}

func TestCensusPassesCleanAndPropagatesErrors(t *testing.T) {
	got, err := Census([]Doc{{Path: "i.md", Text: "[a](a.md) [u](https://x)"}},
		func(string) (LinkState, error) { return Committed, nil }, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
	if _, err := Census([]Doc{{Path: "i.md", Text: "[a](a.md)"}},
		func(string) (LinkState, error) { return 0, errors.New("git broke") }, nil); err == nil || !strings.Contains(err.Error(), "i.md:1: a.md: git broke") {
		t.Fatalf("want the state error with its link: %v", err)
	}
}

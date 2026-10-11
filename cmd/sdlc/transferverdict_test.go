package main

import "testing"

// Every combination of the facts: published details are refused to another
// workspace, and to anyone else behind main's latest version. Row i sets
// Published=bit0, Foreign=bit1, Based=bit2; A accept, F foreign, H behind.
func TestDetailsVerdict(t *testing.T) {
	const want = "AHAFAAAF"
	code := map[verdict]byte{verdictAccept: 'A', verdictForeign: 'F', verdictBehind: 'H'}
	for i := 0; i < 8; i++ {
		f := detailsFacts{Published: i&1 != 0, Foreign: i&2 != 0, Based: i&4 != 0}
		if got := code[detailsVerdict(f)]; got != want[i] {
			t.Errorf("%+v: got %c, want %c", f, got, want[i])
		}
	}
}

package main

import "testing"

// Every combination of the facts: only an owner based on main's latest
// version may change published details. Row i sets Published=bit0,
// Owner=bit1, Based=bit2; A accept, N not owner, H behind.
func TestDetailsVerdict(t *testing.T) {
	const want = "ANAHANAA"
	code := map[verdict]byte{verdictAccept: 'A', verdictNotOwner: 'N', verdictBehind: 'H'}
	for i := 0; i < 8; i++ {
		f := detailsFacts{Published: i&1 != 0, Owner: i&2 != 0, Based: i&4 != 0}
		if got := code[detailsVerdict(f)]; got != want[i] {
			t.Errorf("%+v: got %c, want %c", f, got, want[i])
		}
	}
}

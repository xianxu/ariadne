package main

import "testing"

// Every combination of the four facts: only an owner based on main's latest
// version, with no conflict, may change published details. Row i sets
// Published=bit0, Owner=bit1, Based=bit2, Conflict=bit3; A accept, N not
// owner, H behind.
func TestDetailsVerdict(t *testing.T) {
	const want = "ANAHANAAANAHANAH"
	code := map[verdict]byte{verdictAccept: 'A', verdictNotOwner: 'N', verdictBehind: 'H'}
	for i := 0; i < 16; i++ {
		f := detailsFacts{Published: i&1 != 0, Owner: i&2 != 0, Based: i&4 != 0, Conflict: i&8 != 0}
		if got := code[detailsVerdict(f)]; got != want[i] {
			t.Errorf("%+v: got %c, want %c", f, got, want[i])
		}
	}
}

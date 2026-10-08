package main

import "testing"

// #284: republishDecision over every equality pattern of (local, base, main),
// with the base present or absent. Publishing never overwrites a main copy
// that moved since the base; identical details publish nothing.
func TestRepublishDecision(t *testing.T) {
	values := []string{"a", "b", "c"}
	for _, local := range values {
		for _, base := range values {
			for _, main := range values {
				for _, baseOK := range []bool{true, false} {
					got := republishDecision(local, base, baseOK, main)
					var want publishVerdict
					switch {
					case local == main:
						want = publishNothing
					case baseOK && main == base:
						want = publishWrite
					default:
						want = publishRefuse
					}
					if got != want {
						t.Errorf("local=%s base=%s(%v) main=%s: got %d, want %d", local, base, baseOK, main, got, want)
					}
					if got == publishWrite && (!baseOK || main != base) {
						t.Errorf("local=%s base=%s(%v) main=%s: would overwrite a main copy that moved", local, base, baseOK, main)
					}
				}
			}
		}
	}
}

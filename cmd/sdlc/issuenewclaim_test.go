package main

import "testing"

func TestCreatorClaims(t *testing.T) {
	for slot, want := range map[string]bool{
		"ariadne:4": true,  // a numbered slot takes the work it files
		"pair:1":    true,  // filed from pair:1's environment into a dependency
		"ariadne:0": false, // the operator's resting brainstorm
		"":          false, // a checkout outside the slot layout
		"ariadne":   false, // no slot
		":3":        false, // no repository
		"x:y":       false, // unparsable
	} {
		if got := creatorClaims(slot); got != want {
			t.Errorf("creatorClaims(%q) = %v, want %v", slot, got, want)
		}
	}
}

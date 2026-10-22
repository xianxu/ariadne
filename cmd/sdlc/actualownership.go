package main

import "fmt"

// unownedActualError is actionable only when the selected current working
// card and its complete pinned history contain no tracker owner.
type unownedActualError struct{ Issue string }

func (e *unownedActualError) recovery() string {
	return fmt.Sprintf("run sdlc claim --issue %s from the slot doing the work", e.Issue)
}

func (e *unownedActualError) Error() string {
	return "no owner since cutover: " + e.recovery() + "; work between cutover and claim is unowned and not counted"
}

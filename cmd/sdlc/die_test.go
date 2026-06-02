package main

import (
	"io"
	"testing"
)

// dieSignal is the panic payload expectDie uses to unwind a die() call
// without exiting the test process. It carries the message die was given
// so the test can assert on the refusal text.
type dieSignal struct{ msg string }

// expectDie runs fn with the package-level `die` swapped for a panic that
// expectDie itself recovers, then returns whether die fired and the message
// it carried. This is the reusable seam for testing ANY `run*` verb's
// refusal path (#63): in production die calls os.Exit(1) (unrecoverable in
// a test); here it panics, preserving die's "halt the flow right here"
// semantics — code after the die() call does not run, exactly as in prod —
// while letting the test inspect the outcome.
//
//	msg, died := expectDie(t, func() { runMerge(out, errw, &flags) })
//	if !died { t.Fatal("expected a refusal") }
//	if !strings.Contains(msg, "...") { ... }
//
// A panic that is NOT a *dieSignal (a real bug in the code under test) is
// re-raised so the test fails loudly rather than swallowing it.
func expectDie(t *testing.T, fn func()) (msg string, died bool) {
	t.Helper()
	prev := die
	die = func(_ io.Writer, m string) { panic(&dieSignal{m}) }
	defer func() {
		die = prev
		if r := recover(); r != nil {
			if ds, ok := r.(*dieSignal); ok {
				msg, died = ds.msg, true
				return
			}
			panic(r) // not ours — a genuine panic in the code under test
		}
	}()
	fn()
	return "", false
}

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
//
// Semantic note for future adopters: the panic unwinds the stack and runs
// any deferred funcs between the die() call and here, whereas production's
// os.Exit(1) does NOT run defers. runMerge holds no relevant defers, so the
// two are equivalent today — but a run* verb that defers cleanup could see a
// test-only side effect. Prefer keeping refusal paths defer-free above die().
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

// exitSignal is the panic payload expectExit uses to unwind an exitWithCode call.
type exitSignal struct{ code int }

// expectExit is expectDie's counterpart for code that exits through exitWithCode
// directly rather than through die (#191: runChangeCode's gate loop). It swaps the
// exitWithCode var for a panic it recovers, so code after the exit does not run,
// exactly as in production, and returns the exit code. The same defer caveat as
// expectDie applies, and the swap does not run the die cleanups: a test that needs
// them registered and fired is testing the registry, not the exit.
func expectExit(t *testing.T, fn func()) (code int, exited bool) {
	t.Helper()
	prev := exitWithCode
	exitWithCode = func(c int) { panic(&exitSignal{c}) }
	defer func() {
		exitWithCode = prev
		if r := recover(); r != nil {
			if es, ok := r.(*exitSignal); ok {
				code, exited = es.code, true
				return
			}
			panic(r) // not ours — a genuine panic in the code under test
		}
	}()
	fn()
	return 0, false
}

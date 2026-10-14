package issueassets

import (
	"bytes"
	"testing"
)

func TestAttachTransitionsAndOwnership(t *testing.T) {
	for _, existing := range []bool{false, true} {
		p := member(t, []byte("x"))
		body := []byte("body")
		source := []byte("x")
		s, e := PrepareAttach(p, body, source, existing)
		if e != nil {
			t.Fatal(e)
		}
		body[0] = 'X'
		source[0] = 'Y'
		if !bytes.Equal(s.ExpectedBody(), []byte("body")) || !bytes.Equal(s.Source(), []byte("x")) {
			t.Fatal("aliased")
		}
		event := InstallNew
		if existing {
			event = InstallExisting
		}
		s, effect, e := NextAttach(s, event)
		if e != nil || effect != CompareBodyAndLink {
			t.Fatal(s, effect, e)
		}
		success, effect, e := NextAttach(s, LinkSucceeded)
		if e != nil || success.Stage() != Linked || effect != None {
			t.Fatal(success, effect, e)
		}
		success, effect, e = NextAttach(success, Retry)
		if e != nil || success.Stage() != Linked || effect != None {
			t.Fatal("retry not idempotent")
		}
		failed, effect, e := NextAttach(s, LinkFailed)
		if e != nil {
			t.Fatal(e)
		}
		if existing {
			if failed.Stage() != Retained || effect != Report {
				t.Fatal("preexisting cleanup requested")
			}
		} else {
			if effect != RemoveOwnedBytes {
				t.Fatal("new cleanup not requested")
			}
			for _, ev := range []AttachEvent{RemovalConfirmed, RemovalFailed} {
				end, eff, e := NextAttach(failed, ev)
				if e != nil || eff != Report {
					t.Fatal(end, eff, e)
				}
				if (end.Stage() == RolledBack) != (ev == RemovalConfirmed) {
					t.Fatal(end)
				}
			}
		}
	}
}
func FuzzAttachEventSequences(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, events []byte) {
		for _, existing := range []bool{true, false} {
			s, e := PrepareAttach(member(t, []byte("x")), nil, []byte("x"), existing)
			if e != nil {
				t.Fatal(e)
			}
			for _, b := range events {
				before := s
				next, effect, e := NextAttach(s, AttachEvent(b%9))
				if e != nil {
					if next.Stage() != before.Stage() {
						t.Fatal("invalid event mutated state")
					}
					continue
				}
				if existing && effect == RemoveOwnedBytes {
					t.Fatal("preexisting bytes scheduled for removal")
				}
				s = next
			}
		}
	})
}

package queue

import (
	"strings"
	"testing"
)

const sampleDoc = `# Queue

Advisory ordering. Read through ` + "`sdlc queue`" + `.

- ariadne#207 — after #206 merges, same dispatch [sdlc]
- pair#171 — floor under attention

<!-- a comment -->
`

func TestDoc_RoundTrip(t *testing.T) {
	for _, tc := range []struct{ name, in string }{
		{"sample", sampleDoc},
		{"empty", ""},
		{"just newline", "\n"},
		{"no trailing newline", "- a#1 — why"},
		{"blank lines preserved", "\n\n- a#1 — why\n\n\n"},
		{"prose only", "just some prose\nand more\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(Parse([]byte(tc.in)).Render()); got != tc.in {
				t.Errorf("round-trip = %q, want %q", got, tc.in)
			}
		})
	}
}

func TestDoc_Entries(t *testing.T) {
	d := Parse([]byte(sampleDoc))
	got := d.Entries()
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(got), got)
	}
	if got[0].Ref != "ariadne#207" || got[1].Ref != "pair#171" {
		t.Errorf("entries out of order or wrong: %+v", got)
	}
	if got[0].Tag != "sdlc" {
		t.Errorf("tag not parsed: %+v", got[0])
	}
	// Prose is not an entry, but it is still in the document.
	if !strings.Contains(string(d.Render()), "<!-- a comment -->") {
		t.Error("non-entry content must survive")
	}
}

// FuzzDocRoundTrip is the mechanical guard behind "a malformed line is kept
// as-is, never discarded".
//
// Doc parses a human-editable file arriving from the trunk — input this process
// did not produce (ARCH-SECURE) — so chosen examples are not enough: they prove
// the shapes I thought of, and the adversarial class here is arbitrary bytes.
func FuzzDocRoundTrip(f *testing.F) {
	for _, seed := range []string{
		sampleDoc, "", "\n", "\r\n", "- ", "-", "- a#1 — why",
		"- a#1 — why\r\n- b#2 — why2\n", "— — —", "- project: — ",
		"- a#1 — why [", "- a#1 — why []", "\x00\x01", "– en dash –",
		strings.Repeat("- a#1 — why\n", 50),
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		got := Parse(b).Render()
		if string(got) != string(b) {
			t.Fatalf("round-trip changed the document:\n in: %q\nout: %q", b, got)
		}
	})
}

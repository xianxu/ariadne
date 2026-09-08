package queue

import (
	"strings"
	"testing"
)

func TestParseLine_RoundTrips(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want Line
		ok   bool
	}{
		{"issue with tag", "- ariadne#207 — after #206 merges, same dispatch [sdlc]",
			Line{Ref: "ariadne#207", WhyNow: "after #206 merges, same dispatch", Tag: "sdlc", Kind: KindIssue, parsed: true}, true},
		{"issue no tag", "- pair#171 — floor under attention",
			Line{Ref: "pair#171", WhyNow: "floor under attention", Kind: KindIssue, parsed: true}, true},
		{"project line", "- project:sdlc-fleet — the whole policy area is next [sdlc]",
			Line{Ref: "sdlc-fleet", WhyNow: "the whole policy area is next", Tag: "sdlc", Kind: KindProject, parsed: true}, true},
		{"bracket inside prose is not a tag", "- a#1 — see [RFC] for why it matters",
			Line{Ref: "a#1", WhyNow: "see [RFC] for why it matters", Kind: KindIssue, parsed: true}, true},

		// Unparsed forms are PRESERVED, never dropped.
		{"prose", "Some heading text", Line{raw: "Some heading text"}, false},
		{"blank", "", Line{raw: ""}, false},
		{"bullet without separator", "- just a bullet", Line{raw: "- just a bullet"}, false},
		{"bullet with hyphen not em-dash", "- a#1 - why", Line{raw: "- a#1 - why"}, false},
		{"empty why", "- a#1 — ", Line{raw: "- a#1 — "}, false},
		{"empty ref", "-  — why", Line{raw: "-  — why"}, false},
		{"bare project prefix", "- project: — why", Line{raw: "- project: — why"}, false},
		// A bracket with no prose before it is why-now text, not a tag: the tag
		// opener must have something to the left of it. Keeping this line PARSED
		// is deliberate — ValidateWhyNow stops the verb from ever creating one, so
		// it can only arrive by hand, and an unparsed line is one `sdlc queue
		// remove a#1` cannot find. Preserving it as an entry keeps it removable.
		{"leading bracket is prose, not a tag", "- a#1 — [tag]",
			Line{Ref: "a#1", WhyNow: "[tag]", Kind: KindIssue, parsed: true}, true},
		{"trailing dash at EOF", "- ", Line{raw: "- "}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseLine(tc.in)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (got %+v)", ok, tc.ok, got)
			}
			if got != tc.want {
				t.Errorf("parsed = %+v, want %+v", got, tc.want)
			}
			if rt := got.String(); rt != tc.in {
				t.Errorf("round-trip = %q, want %q", rt, tc.in)
			}
		})
	}
}

func TestValidateWhyNow(t *testing.T) {
	for _, tc := range []struct {
		name, in string
		wantErr  bool
	}{
		{"ordinary", "after #206 merges, same dispatch", false},
		{"unicode fine", "needs the café measurement first", false},
		{"empty", "   ", true},
		{"newline", "first\nsecond", true},
		{"carriage return", "first\rsecond", true},
		{"control char", "a\x00b", true},
		{"em-dash separator", "a — b", true},
		{"open bracket", "see [RFC]", true},
		{"close bracket", "see RFC]", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateWhyNow(tc.in); (err != nil) != tc.wantErr {
				t.Errorf("ValidateWhyNow(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
		})
	}
}

func TestValidateRef(t *testing.T) {
	for _, tc := range []struct {
		in      string
		wantErr bool
	}{
		{"ariadne#207", false}, {"pair#171", false}, {"sdlc-fleet", false},
		{"", true}, {"  ", true}, {" a#1", true}, {"a #1", true},
		{"a\n1", true}, {"a[1]", true}, {"a — b", true},
	} {
		if err := ValidateRef(tc.in); (err != nil) != tc.wantErr {
			t.Errorf("ValidateRef(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
		}
	}
}

// A newline in why-now would render as a second queue entry — the concrete
// breakage the validator exists to stop, pinned rather than asserted in prose.
func TestValidateWhyNow_NewlineWouldForgeAnEntry(t *testing.T) {
	evil := "why\n- forged#1 — injected"
	if err := ValidateWhyNow(evil); err == nil {
		t.Fatal("must reject")
	}
	// Demonstrate the breakage the rejection prevents.
	forged := Line{Ref: "a#1", WhyNow: evil, Kind: KindIssue, parsed: true}.String()
	if n := len(Parse([]byte(forged)).Entries()); n != 2 {
		t.Fatalf("expected the unvalidated line to forge a 2nd entry, got %d", n)
	}
	if !strings.Contains(forged, "forged#1") {
		t.Error("sanity: the forged entry should be present")
	}
}

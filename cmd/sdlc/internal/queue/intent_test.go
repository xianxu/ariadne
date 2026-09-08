package queue

import (
	"errors"
	"strings"
	"testing"
)

// base is a three-entry queue used across the table.
const base = "# Queue\n\n- a#1 — first\n- b#2 — second\n- c#3 — third\n"

func refsOf(t *testing.T, d *Doc) []string {
	t.Helper()
	var out []string
	for _, l := range d.Entries() {
		out = append(out, l.Ref)
	}
	return out
}

// The Spec's interleaving table IS this test list. Every row has a case, and
// each names the concurrent event it models.
func TestIntent_Apply_InterleavingTable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		doc      string
		in       Intent
		wantRefs []string
		wantNote string
		wantErr  error
	}{
		{
			name: "add onto a base a peer grew — both survive",
			doc:  base + "- peer#9 — landed while we were deciding\n",
			in:   Intent{Op: OpAdd, Ref: "d#4", WhyNow: "mine"},
			// The peer's line is still there: this is the replay, not a re-push.
			wantRefs: []string{"a#1", "b#2", "c#3", "peer#9", "d#4"},
		},
		{
			name:     "remove converges when a peer already removed it",
			doc:      base,
			in:       Intent{Op: OpRemove, Ref: "zzz#9"},
			wantRefs: []string{"a#1", "b#2", "c#3"},
			wantNote: "nothing to remove",
		},
		{
			name:     "add converges when a peer already added it; the note names what changed",
			doc:      base,
			in:       Intent{Op: OpAdd, Ref: "b#2", WhyNow: "sharper reason"},
			wantRefs: []string{"a#1", "b#2", "c#3"},
			wantNote: "updated why-now",
		},
		{
			name:    "move refuses when a peer deleted the anchor",
			doc:     base,
			in:      Intent{Op: OpMove, Ref: "a#1", Anchor: "gone#9"},
			wantErr: ErrAnchorMissing,
		},
		{
			name:    "move refuses when a peer deleted the subject",
			doc:     base,
			in:      Intent{Op: OpMove, Ref: "gone#9", Anchor: "a#1"},
			wantErr: ErrSubjectMissing,
		},
		{
			name:     "move before, from below the anchor",
			doc:      base,
			in:       Intent{Op: OpMove, Ref: "c#3", Anchor: "a#1"},
			wantRefs: []string{"c#3", "a#1", "b#2"},
		},
		{
			name:     "move before, from ABOVE the anchor (the index-shift case)",
			doc:      base,
			in:       Intent{Op: OpMove, Ref: "a#1", Anchor: "c#3"},
			wantRefs: []string{"b#2", "a#1", "c#3"},
		},
		{
			name:     "move after, from above",
			doc:      base,
			in:       Intent{Op: OpMove, Ref: "a#1", Anchor: "c#3", After: true},
			wantRefs: []string{"b#2", "c#3", "a#1"},
		},
		{
			name:     "move after, from below",
			doc:      base,
			in:       Intent{Op: OpMove, Ref: "c#3", Anchor: "a#1", After: true},
			wantRefs: []string{"a#1", "c#3", "b#2"},
		},
		{
			name:    "move relative to itself is refused, not a silent no-op",
			doc:     base,
			in:      Intent{Op: OpMove, Ref: "a#1", Anchor: "a#1"},
			wantErr: nil, // checked below as a non-nil generic error
		},
		{
			name:    "add of a malformed ref fails before any git call",
			doc:     base,
			in:      Intent{Op: OpAdd, Ref: "bad ref", WhyNow: "x"},
			wantErr: nil,
		},
		{
			name:    "add with a newline in why-now fails before any git call",
			doc:     base,
			in:      Intent{Op: OpAdd, Ref: "d#4", WhyNow: "why\n- forged#1 — injected"},
			wantErr: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := Parse([]byte(tc.doc))
			got, applied, err := tc.in.Apply(in)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if tc.wantRefs == nil { // expecting some error
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotRefs := refsOf(t, got); !equal(gotRefs, tc.wantRefs) {
				t.Errorf("refs = %v, want %v", gotRefs, tc.wantRefs)
			}
			if tc.wantNote != "" && !strings.Contains(applied.Note, tc.wantNote) {
				t.Errorf("note = %q, want it to mention %q", applied.Note, tc.wantNote)
			}
			// The input document is never mutated: a transform that mutated
			// shared state would corrupt the base on a CAS retry.
			if gotIn := refsOf(t, in); !equal(gotIn, refsOf(t, Parse([]byte(tc.doc)))) {
				t.Errorf("Apply mutated its input: %v", gotIn)
			}
		})
	}
}

// Non-entry content survives every operation. The file is co-authored by hand;
// an edit that ate the prose would make it untrustworthy.
func TestIntent_Apply_PreservesProse(t *testing.T) {
	doc := "# Queue\n\nSome prose.\n\n- a#1 — first\n- b#2 — second\n\n<!-- tail -->\n"
	for _, in := range []Intent{
		{Op: OpAdd, Ref: "c#3", WhyNow: "third"},
		{Op: OpRemove, Ref: "a#1"},
		{Op: OpMove, Ref: "b#2", Anchor: "a#1"},
	} {
		got, _, err := in.Apply(Parse([]byte(doc)))
		if err != nil {
			t.Fatalf("%v: %v", in.Op, err)
		}
		for _, want := range []string{"# Queue", "Some prose.", "<!-- tail -->"} {
			if !strings.Contains(string(got.Render()), want) {
				t.Errorf("op %v dropped %q", in.Op, want)
			}
		}
	}
}

// Adding to an empty (or absent) file works — the path M2's seed takes, since
// the trunk read returns empty for a path that does not exist yet.
func TestIntent_Apply_AddToEmptyDoc(t *testing.T) {
	got, _, err := Intent{Op: OpAdd, Ref: "a#1", WhyNow: "first"}.Apply(Parse(nil))
	if err != nil {
		t.Fatal(err)
	}
	if s := string(got.Render()); s != "- a#1 — first\n" {
		t.Errorf("got %q", s)
	}
}

// A project line is distinguishable from an issue line and round-trips.
func TestIntent_Apply_ProjectLine(t *testing.T) {
	got, _, err := Intent{Op: OpAdd, Ref: "sdlc-fleet", WhyNow: "area is next",
		Tag: "sdlc", Kind: KindProject}.Apply(Parse(nil))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got.Render())
	if s != "- project:sdlc-fleet — area is next [sdlc]\n" {
		t.Fatalf("got %q", s)
	}
	e := Parse([]byte(s)).Entries()
	if len(e) != 1 || e[0].Kind != KindProject || e[0].Ref != "sdlc-fleet" {
		t.Errorf("project line did not round-trip: %+v", e)
	}
}

// An unknown Op must fail loudly rather than silently returning the document
// unchanged, which would report success for an edit that never happened.
func TestIntent_Apply_UnknownOpFails(t *testing.T) {
	if _, _, err := (Intent{Op: Op(99), Ref: "a#1"}).Apply(Parse([]byte(base))); err == nil {
		t.Error("an unknown operation must fail, not no-op")
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Converge merges rather than replaces, and an add that changes NOTHING says so
// rather than claiming an update it did not make.
func TestIntent_Apply_ConvergeMergeSemantics(t *testing.T) {
	doc := "- a#1 — original [sdlc]\n"
	for _, tc := range []struct {
		name     string
		in       Intent
		wantLine string
		wantNote string
	}{
		{"omitting the tag keeps it", Intent{Op: OpAdd, Ref: "a#1", WhyNow: "newer"},
			"- a#1 — newer [sdlc]\n", "updated why-now"},
		{"an explicit tag replaces it", Intent{Op: OpAdd, Ref: "a#1", WhyNow: "newer", Tag: "couch"},
			"- a#1 — newer [couch]\n", "and tag"},
		{"identical add is reported as unchanged", Intent{Op: OpAdd, Ref: "a#1", WhyNow: "original", Tag: "sdlc"},
			"- a#1 — original [sdlc]\n", "unchanged"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, applied, err := tc.in.Apply(Parse([]byte(doc)))
			if err != nil {
				t.Fatal(err)
			}
			if s := string(got.Render()); s != tc.wantLine {
				t.Errorf("line = %q, want %q", s, tc.wantLine)
			}
			if !strings.Contains(applied.Note, tc.wantNote) {
				t.Errorf("note = %q, want it to mention %q", applied.Note, tc.wantNote)
			}
		})
	}
}

// Every user-supplied field the line format interpolates is validated. The
// enumeration is the point: Tag was the one missed when guards were added
// per-field, and a newline in it published a forged entry with exit 0.
func TestIntent_Validate_CoversEveryInterpolatedField(t *testing.T) {
	forge := "x\n- forged#9 — injected"
	for _, tc := range []struct {
		name string
		in   Intent
	}{
		{"ref", Intent{Op: OpAdd, Ref: "bad ref", WhyNow: "x"}},
		{"why-now", Intent{Op: OpAdd, Ref: "a#1", WhyNow: forge}},
		{"tag", Intent{Op: OpAdd, Ref: "a#1", WhyNow: "x", Tag: forge}},
		{"move anchor", Intent{Op: OpMove, Ref: "a#1", Anchor: "bad anchor"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.in.Validate(); err == nil {
				t.Fatalf("%s must be validated", tc.name)
			}
			// And the forged entry must not survive into a document either.
			if _, _, err := tc.in.Apply(Parse(nil)); err == nil {
				t.Errorf("Apply must refuse too (defence in depth)")
			}
		})
	}
}

// BR-33: per-field validators can each pass while their COMBINATION renders a
// line that reads back as a different record — an entry that duplicates on
// re-add and cannot be removed. Validate answers the real question by rendering
// and re-parsing, so a case nobody enumerated is still caught.
func TestIntent_Validate_RejectsRecordsThatDoNotSurviveARoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   Intent
	}{
		{"issue ref carrying the project prefix", Intent{Op: OpAdd, Ref: "project:foo", WhyNow: "why"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.in.Validate(); err == nil {
				// Demonstrate the damage the rejection prevents.
				d, _, _ := tc.in.Apply(Parse(nil))
				e := Parse(d.Render()).Entries()
				t.Fatalf("accepted, and it round-trips to ref %q instead of %q", e[0].Ref, tc.in.Ref)
			}
		})
	}
}

// The check is PRECISE, not a blanket ban on the prefix. `Ref: "project:foo"`
// with `Kind: KindProject` renders "- project:project:foo — why" and reads back
// unchanged: odd-looking, but consistent and removable, so it is not a defect and
// is not rejected. Only the record that CHANGES under a round trip is — which is
// the difference between asking the real question and banning a substring.
func TestIntent_Validate_AllowsOddButConsistentRecords(t *testing.T) {
	in := Intent{Op: OpAdd, Ref: "project:foo", WhyNow: "why", Kind: KindProject}
	if err := in.Validate(); err != nil {
		t.Fatalf("a consistent record must pass even if it renders oddly: %v", err)
	}
	d, _, err := in.Apply(Parse(nil))
	if err != nil {
		t.Fatal(err)
	}
	if e := Parse(d.Render()).Entries(); e[0].Ref != "project:foo" || e[0].Kind != KindProject {
		t.Errorf("record changed: %+v", e)
	}
}

// A legitimate project line still works.
func TestIntent_Validate_AllowsHonestProjectLines(t *testing.T) {
	in := Intent{Op: OpAdd, Ref: "sdlc-fleet", WhyNow: "area is next", Tag: "sdlc", Kind: KindProject}
	if err := in.Validate(); err != nil {
		t.Fatalf("a normal project line must pass: %v", err)
	}
	d, _, err := in.Apply(Parse(nil))
	if err != nil {
		t.Fatal(err)
	}
	e := Parse(d.Render()).Entries()
	if len(e) != 1 || e[0].Ref != "sdlc-fleet" || e[0].Kind != KindProject {
		t.Errorf("round-trip changed the record: %+v", e)
	}
}

// A converging re-add that never mentions kind must not destroy a peer's
// project marker. KindUnspecified is what makes "not mentioned" representable.
func TestIntent_Apply_ConvergeDoesNotFlipUnmentionedKind(t *testing.T) {
	doc := "- project:sdlc-fleet — the whole area is next [sdlc]\n"
	got, applied, err := Intent{Op: OpAdd, Ref: "sdlc-fleet", WhyNow: "sharper reason"}.Apply(Parse([]byte(doc)))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got.Render())
	if !strings.Contains(s, "project:sdlc-fleet") {
		t.Errorf("the peer's project marker was destroyed by an edit that never mentioned kind: %q", s)
	}
	if strings.Contains(applied.Note, "kind") {
		t.Errorf("note claims a kind change that was never requested: %q", applied.Note)
	}
}

// Appending to a CRLF document must not introduce mixed line endings.
func TestIntent_Apply_AppendMatchesDocumentLineEnding(t *testing.T) {
	got, _, err := Intent{Op: OpAdd, Ref: "b#2", WhyNow: "second"}.Apply(Parse([]byte("- a#1 — first\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	if s := string(got.Render()); s != "- a#1 — first\r\n- b#2 — second\r\n" {
		t.Errorf("mixed line endings: %q", s)
	}
}

package main

import (
	"strings"
	"testing"
)

func TestWorkingdocMetadataRoundTrip(t *testing.T) {
	text := "---\ntype: workingdoc\nowner: ops:0\ntopic: Direction\nfuture:\n  nested: [one, two] # retain\n---\n# Direction\nBody.\n"
	m, err := parseWorkingdoc(text)
	if err != nil {
		t.Fatal(err)
	}
	if m.Owner != "ops:0" || m.Topic != "Direction" || m.Pointer != nil {
		t.Fatalf("metadata: %+v", m)
	}
	p := settlementPointer{Path: "workshop/workingdoc/direction.md", Commit: strings.Repeat("a", 40)}
	stamped, err := stampWorkingdoc(text, p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stamped, "future:\n  nested: [one, two] # retain\n") {
		t.Fatalf("lost unknown metadata: %s", stamped)
	}
	got, err := parseWorkingdoc(stamped)
	if err != nil || got.Pointer == nil || *got.Pointer != p {
		t.Fatalf("pointer: %+v %v", got, err)
	}
	clean, _, err := projectWorkingdoc(stamped, "session.side/01-direction.md", p.Path, func(string) bool { return true })
	if err != nil || clean != text {
		t.Fatalf("projection: %q, %v", clean, err)
	}
}

func TestWorkingdocMetadataRefusesAmbiguity(t *testing.T) {
	for _, field := range []string{"owner: [a]\n", "topic: one\ntopic: two\n", "settled-into: nope\n", "settled-into: |\n  nope\n", "type: project\n"} {
		if _, err := parseWorkingdoc("---\n" + field + "---\nBody\n"); err == nil {
			t.Errorf("accepted %q", field)
		}
	}
}

func TestSettlementPointerRejectsUntrustedPaths(t *testing.T) {
	for _, p := range []string{"../x.md", "/x.md", "-x.md", "a/../../x.md", "x.md @ bad", "session.side/x.md"} {
		if _, err := parseSettlementPointer(p + " @ " + strings.Repeat("b", 40)); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
}

func TestWorkingdocStampRefusesUnrepresentablePlainPointers(t *testing.T) {
	for _, target := range []string{"a #b.md", "a: b.md", "!tag.md", "&anchor.md", "*alias.md", "[flow].md", "{flow}.md", "#comment.md", "? key.md", ": value.md", "'quoted.md", "\"quoted.md", "|block.md", ">block.md", "@reserved.md", "`reserved.md"} {
		t.Run(target, func(t *testing.T) {
			if !workingdocPath(target) {
				t.Fatal("fixture must be an otherwise permitted target path")
			}
			for _, existing := range []bool{false, true} {
				source := "---\ntype: workingdoc\nfuture: [one, two] # retained\n"
				if existing {
					source += "settled-into: old.md @ " + strings.Repeat("a", 40) + "\n"
				}
				source += "---\nBody.\n"
				if stamped, err := stampWorkingdoc(source, settlementPointer{target, strings.Repeat("b", 40)}); err == nil {
					t.Errorf("existing=%v: accepted unreadable pointer: %q", existing, stamped)
				}
			}
		})
	}
}

func TestWorkingdocStampPlainPointerRoundTripPreservesBytes(t *testing.T) {
	for _, target := range []string{"workshop/workingdoc/topic.md", "a#b.md", "a:b.md", "dir/!tag.md", "dir/[flow].md", "日本語 and spaces.md"} {
		for _, eol := range []string{"\n", "\r\n"} {
			for _, existing := range []bool{false, true} {
				source := "---\ntype: workingdoc\nfuture: [one, two] # retained\n"
				if existing {
					source += "settled-into: old.md @ " + strings.Repeat("a", 40) + "\n"
				}
				source += "---\nBody: literal # punctuation.\n"
				source = strings.ReplaceAll(source, "\n", eol)
				pointer := settlementPointer{target, strings.Repeat("c", 64)}
				stamped, err := stampWorkingdoc(source, pointer)
				if err != nil {
					t.Fatalf("%q existing=%v: %v", target, existing, err)
				}
				meta, err := parseWorkingdoc(stamped)
				if err != nil || meta.Pointer == nil || *meta.Pointer != pointer {
					t.Fatalf("%q roundtrip: %+v %v", target, meta, err)
				}
				before, err := parseWorkingdoc(source)
				if err != nil {
					t.Fatal(err)
				}
				want := source
				if existing {
					want = source[:before.pointerStart] + source[before.pointerEnd:]
				}
				if got := stamped[:meta.pointerStart] + stamped[meta.pointerEnd:]; got != want {
					t.Fatalf("%q changed unowned bytes: %q != %q", target, got, want)
				}
			}
		}
	}
}

func TestSettlementCannotStampBeforePublication(t *testing.T) {
	s := settlementState{Phase: settlementPlanned}
	if _, err := transitionSettlement(s, settlementStamped); err == nil {
		t.Fatal("stamped before publication")
	}
	s, err := transitionSettlement(s, settlementPublished)
	if err != nil {
		t.Fatal(err)
	}
	s, err = transitionSettlement(s, settlementStamped)
	if err != nil || s.Phase != settlementDone {
		t.Fatalf("completion: %+v %v", s, err)
	}
	if _, err = transitionSettlement(s, settlementPublished); err == nil {
		t.Fatal("terminal phase accepted effect")
	}
}

func TestWorkingdocClassificationUsesBytes(t *testing.T) {
	for _, tc := range []struct{ source, base, target, want string }{{"a", "a", "a", "settled"}, {"b", "a", "a", "drift"}, {"a", "a", "b", "conflict"}} {
		if got := classifyWorkingdoc([]byte(tc.source), []byte(tc.base), []byte(tc.target)); got != tc.want {
			t.Errorf("%+v: %s", tc, got)
		}
	}
}

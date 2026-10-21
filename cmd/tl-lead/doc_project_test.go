package main

import (
	"strings"
	"testing"
)

func TestWorkingdocProjectionRefusesUnrewritableParent(t *testing.T) {
	for _, parent := range []string{"parent: ../session.md # note", "parent: [../session.md]", "parent: |\n  ../session.md"} {
		source := "---\ntype: side-doc\n" + parent + "\n---\n# Topic\n"
		if _, _, err := projectWorkingdoc(source, "session.side/01-topic.md", "workshop/workingdoc/topic.md", func(string) bool { return true }); err == nil {
			t.Errorf("accepted parent %q", parent)
		}
	}
}

func FuzzWorkingdocPointerRoundTrip(f *testing.F) {
	f.Add("A thought.")
	f.Add("Unicode 日本語 🤖 and inline `code`.")
	f.Fuzz(func(t *testing.T, body string) {
		source := "---\ntype: workingdoc\nfuture: {nested: [one, two]} # retained\n---\n" + body
		p := settlementPointer{"workshop/workingdoc/topic.md", strings.Repeat("c", 40)}
		stamped, err := stampWorkingdoc(source, p)
		if err != nil {
			t.Fatal(err)
		}
		meta, err := parseWorkingdoc(stamped)
		if err != nil {
			t.Fatal(err)
		}
		if meta.Pointer == nil || *meta.Pointer != p {
			t.Fatal("pointer changed")
		}
		restored := stamped[:meta.pointerStart] + stamped[meta.pointerEnd:]
		if restored != source {
			t.Fatal("stamp changed unowned document bytes")
		}
	})
}

func FuzzWorkingdocPointerPathRoundTrip(f *testing.F) {
	for _, target := range []string{"topic.md", "a #b.md", "a: b.md", "!tag.md", "&anchor.md", "日本語.md"} {
		f.Add(target)
	}
	f.Fuzz(func(t *testing.T, target string) {
		source := "---\r\ntype: workingdoc\r\nfuture: {nested: [one, two]} # retained\r\n---\r\nBody.\r\n"
		pointer := settlementPointer{target, strings.Repeat("c", 40)}
		stamped, err := stampWorkingdoc(source, pointer)
		if err != nil {
			return // Unsupported paths must fail before effects.
		}
		meta, err := parseWorkingdoc(stamped)
		if err != nil || meta.Pointer == nil || *meta.Pointer != pointer {
			t.Fatalf("successful stamp did not round-trip: %+v %v", meta, err)
		}
		if stamped[:meta.pointerStart]+stamped[meta.pointerEnd:] != source {
			t.Fatal("stamp changed unowned document bytes")
		}
	})
}

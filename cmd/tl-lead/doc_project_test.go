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

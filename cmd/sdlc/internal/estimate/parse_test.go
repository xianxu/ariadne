package estimate

import (
	"strings"
	"testing"
)

const fence = "```"

// greenSection is the canonical reconciling fixture — the same 5-item / 3.4 block
// used in the plan's Core concepts and dogfooded on #117 itself. Reused by the
// parse and check tests so example, dogfood, and tests agree.
var greenSection = fence + "estimate\n" +
	"model: estimate-logic-v2\n" +
	"familiarity: 1.0\n" +
	"item: greenfield-go-module   design=0.3 impl=0.6\n" +
	"item: smaller-go-module      design=0.2 impl=0.6\n" +
	"item: smaller-go-module      design=0.2 impl=0.5\n" +
	"item: atlas-docs             design=0.0 impl=0.2\n" +
	"item: milestone-review       design=0.0 impl=0.6\n" +
	"design-buffer: 0.30\n" +
	"total: 3.4\n" +
	fence + "\n"

func TestParseBlock_Green(t *testing.T) {
	b, err := ParseBlock("some prose\n\n" + greenSection + "\ntrailing prose\n")
	if err != nil {
		t.Fatalf("ParseBlock returned error: %v", err)
	}
	if b.Model != "estimate-logic-v2" {
		t.Errorf("Model = %q, want estimate-logic-v2", b.Model)
	}
	if b.Familiarity != 1.0 {
		t.Errorf("Familiarity = %v, want 1.0", b.Familiarity)
	}
	if b.DesignBuffer != 0.30 {
		t.Errorf("DesignBuffer = %v, want 0.30", b.DesignBuffer)
	}
	if len(b.Items) != 5 {
		t.Fatalf("len(Items) = %d, want 5", len(b.Items))
	}
	if b.Total != 3.4 {
		t.Errorf("Total = %v, want 3.4", b.Total)
	}
	if got := b.Recomputed(); got < 3.4 || got > 3.42 {
		t.Errorf("Recomputed = %v, want ~3.41", got)
	}
}

func TestParseBlock_Defaults(t *testing.T) {
	// No familiarity / design-buffer lines → defaults 1.0 / 0.30.
	section := fence + "estimate\n" +
		"model: estimate-logic-v2\n" +
		"item: atlas-docs design=0.0 impl=0.2\n" +
		"total: 0.2\n" +
		fence + "\n"
	b, err := ParseBlock(section)
	if err != nil {
		t.Fatalf("ParseBlock error: %v", err)
	}
	if b.Familiarity != 1.0 || b.DesignBuffer != 0.30 {
		t.Errorf("defaults wrong: familiarity=%v design-buffer=%v", b.Familiarity, b.DesignBuffer)
	}
}

func TestParseBlock_Errors(t *testing.T) {
	cases := map[string]string{
		"no fence":         "## Estimate\n\njust prose, no block\n",
		"non-numeric impl": fence + "estimate\nmodel: estimate-logic-v2\nitem: atlas-docs design=0.1 impl=lots\ntotal: 0.2\n" + fence + "\n",
		"missing total":          fence + "estimate\nmodel: estimate-logic-v2\nitem: atlas-docs design=0.1 impl=0.2\n" + fence + "\n",
		"unknown field":          fence + "estimate\nmodel: estimate-logic-v2\nbogus: 1\ntotal: 0.2\n" + fence + "\n",
		"non-numeric familiarity": fence + "estimate\nmodel: estimate-logic-v2\nfamiliarity: lots\nitem: atlas-docs design=0.1 impl=0.2\ntotal: 0.2\n" + fence + "\n",
		"non-numeric total":       fence + "estimate\nmodel: estimate-logic-v2\nitem: atlas-docs design=0.1 impl=0.2\ntotal: soon\n" + fence + "\n",
	}
	for name, section := range cases {
		if _, err := ParseBlock(section); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestParseBlock_FenceIsolated(t *testing.T) {
	// Ensure the closing fence terminates the block (prose after isn't consumed).
	b, err := ParseBlock(greenSection + "\nitem: SHOULD-NOT-PARSE design=9 impl=9\n")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	for _, it := range b.Items {
		if strings.Contains(it.Slug, "SHOULD-NOT-PARSE") {
			t.Error("parser consumed an item line outside the fence")
		}
	}
}

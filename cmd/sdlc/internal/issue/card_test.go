package issue

import (
	"strings"
	"testing"
)

const cardDetailFixture = "---\nid: 000252\nstatus: open\ncreated: 2026-09-25\nupdated: 2026-09-25\nestimate_hours:\ngithub_issue:\ndeps: [pair#17]\n# preserve this comment\ntarget: 'stable'\nflow: {kind: quick, provenance: operator}\ncustom:\n  nested: [a, b]\n---\n\n# Card title\n\n## Problem\nOriginal report.\n\n```md\n## Not a section\n```\n\n## Spec\nKeep  these bytes.\n\n## Log\n- entry\n"

func TestSplitCardPreservesDetailBytes(t *testing.T) {
	card, details, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseCard(card)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ID != "000252" || parsed.Title != "Card title" || !strings.Contains(parsed.Problem, "## Not a section") {
		t.Fatalf("card: %+v", parsed)
	}
	for _, branch := range []string{"deps:", "target:", "flow:", "custom:", "## Spec", "## Log"} {
		if strings.Contains(string(card), branch) {
			t.Fatalf("card leaked %q", branch)
		}
	}
	_, body, _ := Parse(cardDetailFixture)
	_, gotBody, _ := Parse(string(details))
	if gotBody != body {
		t.Fatal("split changed body bytes")
	}
	if !strings.Contains(string(details), "deps: [pair#17]\n# preserve this comment\ntarget: 'stable'\nflow: {kind: quick, provenance: operator}\ncustom:\n  nested: [a, b]\n") {
		t.Fatal("split changed unowned YAML bytes")
	}
	if !strings.Contains(string(card), "id: 000252\n") {
		t.Fatal("legacy ID spelling lost")
	}
}

func TestParseCardRejectsInvalidSchema(t *testing.T) {
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"duplicate":         strings.Replace(string(card), "status: open", "status: open\nstatus: working", 1),
		"status type":       strings.Replace(string(card), "status: open", "status: [open]", 1),
		"unknown status":    strings.Replace(string(card), "status: open", "status: waiting", 1),
		"bad id":            strings.Replace(string(card), "id: 000252", "id: false", 1),
		"bad estimate":      strings.Replace(string(card), "estimate_hours:", "estimate_hours: '2'", 1),
		"nonfinite":         strings.Replace(string(card), "estimate_hours:", "estimate_hours: .inf", 1),
		"negative":          strings.Replace(string(card), "estimate_hours:", "estimate_hours: -1", 1),
		"closed no actual":  strings.Replace(string(card), "status: open", "status: codecomplete", 1),
		"duplicate title":   string(card) + "\n# Another title\n",
		"duplicate problem": string(card) + "\n## Problem\nOther report\n",
		"alias":             strings.Replace(string(card), "status: open", "status: &s open\nstarted: *s", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCard([]byte(raw)); err == nil {
				t.Fatal("accepted invalid card")
			}
		})
	}
}

func TestSplitCardRejectsDuplicateUnknownKeys(t *testing.T) {
	for _, extra := range []string{"custom: x\n", "other: {x: 1, x: 2}\n"} {
		raw := strings.Replace(cardDetailFixture, "---\n\n#", extra+"---\n\n#", 1)
		if _, _, err := SplitCard([]byte(raw)); err == nil {
			t.Fatal("duplicate YAML keys accepted")
		}
	}
}

func TestParseCardTransactionEnvelope(t *testing.T) {
	card, detail, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	valid := strings.Replace(string(card), "---\n\n#", "tracker:\n  version: 1\n  completion: {head: abc, stage: pending}\n---\n\n#", 1)
	if _, err := ParseCard([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	refreshed, err := RefreshMirror(detail, card, []byte(valid))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(refreshed), "tracker:") {
		t.Fatal("internal transaction data leaked into mirror")
	}
	for _, version := range []string{"2", "'1'", "null"} {
		if _, err := ParseCard([]byte(strings.Replace(valid, "version: 1", "version: "+version, 1))); err == nil {
			t.Fatal("accepted unsupported envelope version", version)
		}
	}
}

func TestParseCardRejectsDetailSections(t *testing.T) {
	if _, err := ParseCard([]byte(cardDetailFixture)); err == nil {
		t.Fatal("accepted branch detail fields")
	}
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseCard(append(card, []byte("\n## Spec\nnot card content\n")...)); err == nil {
		t.Fatal("accepted branch sections on card")
	}
}

func TestSplitCardLegacyDecimalIDs(t *testing.T) {
	for _, id := range []string{"000001", "000008", "000099", "000252", "999999"} {
		card, _, err := SplitCard([]byte(strings.Replace(cardDetailFixture, "000252", id, 1)))
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		parsed, err := ParseCard(card)
		if err != nil || parsed.ID != id {
			t.Fatalf("identity %s: %+v, %v", id, parsed, err)
		}
	}
}

func FuzzSplitCard(f *testing.F) {
	f.Add(cardDetailFixture)
	f.Add("---\nid: 000001\nstatus: open\n---\n# T\n\n## Problem\nP\n")
	f.Add("---\n[broken\n---\n# T\n")
	f.Fuzz(func(t *testing.T, raw string) {
		card, detail, err := SplitCard([]byte(raw))
		if err != nil {
			return
		}
		if _, err := ParseCard(card); err != nil {
			t.Fatalf("split produced invalid card: %v", err)
		}
		_, oldBody, _ := Parse(raw)
		_, newBody, _ := Parse(string(detail))
		if oldBody != newBody {
			t.Fatal("split changed body")
		}
		got, err := RefreshMirror(detail, card, card)
		if err != nil || string(got) != string(detail) {
			t.Fatalf("identity refresh changed detail: %v", err)
		}
	})
}

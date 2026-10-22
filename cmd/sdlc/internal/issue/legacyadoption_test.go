package issue

import (
	"strings"
	"testing"
)

func testLegacyAdoption() LegacyAdoption {
	return LegacyAdoption{Repository: "file:/r", SourceMain: strings.Repeat("a", 40), SourcePath: "workshop/history/issues/000001-old.md", SourceBlob: strings.Repeat("b", 40), TerminalCommit: strings.Repeat("c", 40), CompletionToken: "close-1", WindowClass: LegacyWindowlessClose}
}

func legacyCard(t *testing.T) []byte {
	t.Helper()
	card, err := SetCardCompletion([]byte(setterCard), testCompletion("close-1"))
	if err != nil {
		t.Fatal(err)
	}
	card, err = SetCardLegacyAdoption(card, testLegacyAdoption())
	if err != nil {
		t.Fatal(err)
	}
	return card
}

func TestLegacyAdoptionRoundTripAndUnknownFields(t *testing.T) {
	if _, ok, err := CardLegacyAdoption([]byte(setterCard)); err != nil || ok {
		t.Fatalf("absent record: %v %v", ok, err)
	}
	card := legacyCard(t)
	if got, ok, err := CardLegacyAdoption(card); err != nil || !ok || got != testLegacyAdoption() {
		t.Fatalf("roundtrip: %+v %v %v", got, ok, err)
	}
	raw := strings.Replace(string(card), "        window_class: windowless-close\n", "        window_class: windowless-close\n        future: {nested: [one, two]}\n", 1)
	if raw == string(card) {
		t.Fatal("missing fixture member")
	}
	for name, fn := range map[string]func([]byte) ([]byte, error){
		"legacy": func(b []byte) ([]byte, error) { return SetCardLegacyAdoption(b, testLegacyAdoption()) },
		"completion": func(b []byte) ([]byte, error) {
			c := testCompletion("close-1")
			c.LandedCommit = strings.Repeat("d", 40)
			return SetCardCompletion(b, c)
		},
		"scalar":  func(b []byte) ([]byte, error) { return SetCardField(b, "updated", "2026-10-22") },
		"handoff": func(b []byte) ([]byte, error) { return SetCardHandoff(b, testHandoff()) },
	} {
		t.Run(name, func(t *testing.T) {
			out, err := fn([]byte(raw))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "future: {nested: [one, two]}") {
				t.Fatalf("lost future data: %s", out)
			}
		})
	}
}

func TestLegacyAdoptionStrictValidation(t *testing.T) {
	for name, mutate := range map[string]func(*LegacyAdoption){
		"repository":      func(a *LegacyAdoption) { a.Repository = "\n" },
		"source main":     func(a *LegacyAdoption) { a.SourceMain = "abc" },
		"source blob":     func(a *LegacyAdoption) { a.SourceBlob = strings.Repeat("G", 40) },
		"terminal commit": func(a *LegacyAdoption) { a.TerminalCommit = "" },
		"class":           func(a *LegacyAdoption) { a.WindowClass = "unknown" },
		"missing token":   func(a *LegacyAdoption) { a.CompletionToken = "" },
		"bad token":       func(a *LegacyAdoption) { a.CompletionToken = "a\nb" },
		"terminal token":  func(a *LegacyAdoption) { a.WindowClass = LegacyWindowlessTerminal },
	} {
		t.Run(name, func(t *testing.T) {
			a := testLegacyAdoption()
			mutate(&a)
			if err := a.validate(); err == nil {
				t.Fatal("accepted invalid record")
			}
		})
	}
	for _, p := range []string{"", "/absolute", "../escape", "a/../b", "a/./b", "a//b", "a/", "a\\b", "a\nb", ".", "C:/absolute"} {
		t.Run("path "+p, func(t *testing.T) {
			a := testLegacyAdoption()
			a.SourcePath = p
			if err := a.validate(); err == nil {
				t.Fatal("accepted unsafe path")
			}
		})
	}
	for _, class := range []string{LegacyRetained, LegacyWindowlessClose, LegacyWindowlessTerminal} {
		t.Run(class, func(t *testing.T) {
			a := testLegacyAdoption()
			a.WindowClass = class
			if class == LegacyWindowlessTerminal {
				a.CompletionToken = ""
			}
			a.SourceMain = strings.Repeat("e", 64)
			if err := a.validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
	card := string(legacyCard(t))
	for _, field := range []string{"repository", "source_main", "source_path", "source_blob", "terminal_commit", "completion_token", "window_class"} {
		t.Run("known type "+field, func(t *testing.T) {
			// Restrict the replacement to the provenance member, so a malformed
			// completion cannot accidentally be the reason this test passes.
			start := strings.Index(card, "    legacy_adoption:\n")
			if start < 0 {
				t.Fatal("missing provenance")
			}
			for _, value := range []string{"42", "true", "null", "[]", "{}"} {
				lines := strings.Split(card[start:], "\n")
				for i, line := range lines {
					if strings.HasPrefix(line, "        "+field+":") {
						lines[i] = "        " + field + ": " + value
						break
					}
				}
				raw := []byte(card[:start] + strings.Join(lines, "\n"))
				if _, err := ParseCard(raw); err == nil {
					t.Fatalf("accepted %s: %s", field, value)
				}
			}
		})
	}
	for name, raw := range map[string]string{
		"numeric":           strings.Replace(card, "        source_main: "+strings.Repeat("a", 40)+"\n", "        source_main: 42\n", 1),
		"boolean class":     strings.Replace(card, "window_class: windowless-close", "window_class: true", 1),
		"null":              strings.Replace(card, "window_class: windowless-close", "window_class: null", 1),
		"duplicate":         strings.Replace(card, "window_class: windowless-close", "window_class: windowless-close\n        window_class: retained", 1),
		"duplicate unknown": strings.Replace(card, "window_class: windowless-close", "window_class: windowless-close\n        future: {x: one, x: two}", 1),
		"wrong token":       strings.Replace(card, "completion_token: close-1", "completion_token: close-2", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCard([]byte(raw)); err == nil {
				t.Fatal("parsed invalid record")
			}
			if _, _, err := CardLegacyAdoption([]byte(raw)); err == nil {
				t.Fatal("accessor accepted invalid record")
			}
			if _, err := SetCardLegacyAdoption([]byte(raw), testLegacyAdoption()); err == nil {
				t.Fatal("setter rewrote invalid record")
			}
		})
	}
}

func TestLegacyAdoptionCompletionGeneration(t *testing.T) {
	card := legacyCard(t)
	c := testCompletion("close-1")
	c.LandedCommit = strings.Repeat("e", 40)
	landed, err := SetCardCompletion(card, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := CardLegacyAdoption(landed); err != nil || !ok {
		t.Fatalf("same generation lost provenance: %v", err)
	}
	landed = []byte(strings.Replace(string(landed), "        window_class: windowless-close\n", "        window_class: windowless-close\n        future: {nested: [one, two]}\n", 1))
	next, err := SetCardCompletion(landed, testCompletion("close-2"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := CardLegacyAdoption(next); err != nil || ok {
		t.Fatalf("new generation retained provenance: %v", err)
	}
	if strings.Contains(string(next), "future:") {
		t.Fatal("removed provenance resurrected future data")
	}
	a := testLegacyAdoption()
	a.WindowClass = LegacyWindowlessTerminal
	a.CompletionToken = ""
	terminal, err := SetCardLegacyAdoption([]byte(setterCard), a)
	if err != nil {
		t.Fatal(err)
	}
	next, err = SetCardCompletion(terminal, testCompletion("close-2"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := CardLegacyAdoption(next); err != nil || ok {
		t.Fatalf("new close retained terminal provenance: %v", err)
	}
	if _, err := SetCardLegacyAdoption([]byte(setterCard), testLegacyAdoption()); err == nil {
		t.Fatal("close provenance without completion")
	}
	if _, err := SetCardLegacyAdoption(card, a); err == nil {
		t.Fatal("terminal provenance with completion")
	}
}

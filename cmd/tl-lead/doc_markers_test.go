package main

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
)

func TestDocMarkerGrammar(t *testing.T) {
	for _, tc := range []struct{ text, kind string }{
		{"🤖[question]", "commentary"},
		{"🤖<anchor>[question]", "commentary"},
		{"🤖<anchor>{replacement}", "proposal"},
		{"🤖{insertion}", "proposal"},
		{"🤖~deletion~", "proposal"},
		{"🤖~deletion~{replacement}", "proposal"},
		{"🤖~deletion~[replacement]", "proposal"},
		{"🤖[human]{robot}[human again]", "commentary"},
		{"🤖{robot}[human]", "commentary"},
		{"🤖<anchor>[human]{robot}", "commentary"},
		{"🤖{first<br>second\\<br>literal}", "proposal"},
		{"🤖[see [link](path)]", "commentary"},
		{"🤖{balanced {inner} value}", "proposal"},
		{"🤖[escaped \\] closer]", "commentary"},
		{"🤖[", "malformed"},
		{"🤖{missing", "malformed"},
		{"🤖<missing", "malformed"},
		{"🤖~missing", "malformed"},
		{"🤖[ok]{missing", "malformed"},
		{"🤖<anchor><another>", "malformed"},
		{"🤖<anchor>", ""},
		{"a literal 🤖 robot", ""},
		{"🤖", ""},
	} {
		t.Run(tc.text, func(t *testing.T) {
			got := workingdocMarkers(tc.text)
			if tc.kind == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected markers: %#v", got)
				}
				return
			}
			want := []docMarker{{Line: 1, Column: 1, Kind: tc.kind, Text: tc.text}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

func TestDocMarkerLiteralContextsAndLocations(t *testing.T) {
	for _, literal := range []string{
		"`🤖[literal]`", "``a ` 🤖{literal} ``", "~~~markdown\n🤖[literal]\n~~~", "```\n🤖[literal]\n```", "   ````md\n🤖[literal]\n```\n````", "`multiline\n🤖[literal]\nspan`",
	} {
		if got := workingdocMarkers(literal); len(got) != 0 {
			t.Errorf("%q: %#v", literal, got)
		}
	}
	text := "é text\nλ 🤖[ask] and 🤖{new}\n` unmatched 🤖[active]"
	want := []docMarker{{2, 3, "commentary", "🤖[ask]"}, {2, 14, "proposal", "🤖{new}"}, {3, 13, "commentary", "🤖[active]"}}
	if got := workingdocMarkers(text); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if got := workingdocMarkers("🤖[broken\nnext]"); len(got) != 1 || got[0].Kind != "malformed" {
		t.Fatalf("multiline marker: %#v", got)
	}
}

func docMarkerBlockBoundaries() []struct{ name, body string } {
	return []struct{ name, body string }{
		{"paragraph", "`opener\n\n🤖[open question]\n\ncloser`\n"},
		{"heading", "`opener\n# 🤖[open question]\ncloser`\n"},
		{"list_items", "- `opener\n- 🤖[open question]\n- closer`\n"},
		{"blockquote", "`opener\n> 🤖[open question]\ncloser`\n"},
	}
}

func TestDocMarkersCannotCrossBlockBoundaries(t *testing.T) {
	for _, tc := range docMarkerBlockBoundaries() {
		t.Run(tc.name, func(t *testing.T) {
			got := workingdocMarkers(tc.body)
			if len(got) != 1 || got[0].Kind != "commentary" || got[0].Text != "🤖[open question]" {
				t.Fatalf("unmatched delimiters hid active commentary: %#v", got)
			}
		})
	}
}

func TestDocMarkersActualCodeAndEscapedDelimiters(t *testing.T) {
	for _, body := range []string{
		"    🤖[indented code]\n",
		"- ```\n  🤖[nested fenced code]\n  ```\n",
		"> ```\n> 🤖[quoted fenced code]\n> ```\n",
		"- `multiline\n  🤖[literal]\n  code`\n",
		"\\`🤖[active]`",
		"`literal \\` 🤖[active]",
	} {
		got := workingdocMarkers(body)
		active := strings.Contains(body, "[active]")
		if (active && (len(got) != 1 || got[0].Text != "🤖[active]")) || (!active && len(got) != 0) {
			t.Errorf("%q: %#v", body, got)
		}
	}
}

func FuzzDocMarkers(f *testing.F) {
	for _, s := range []string{"plain", "🤖[ask]", "`🤖[literal]`", "🤖<x>{y}", "🤖[broken", "🤖[a]{b}[c]"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		before := strings.Clone(body)
		got := workingdocMarkers(body)
		if body != before {
			t.Fatal("scanner changed source")
		}
		for _, marker := range got {
			if marker.Line < 1 || marker.Column < 1 || !strings.Contains(body, marker.Text) || !strings.HasPrefix(marker.Text, "🤖") {
				t.Fatalf("invalid diagnostic: %#v", marker)
			}
		}
		// Encode arbitrary fuzz bytes as delimiter-free content to independently
		// construct valid and malformed grammar, and literal-code contexts.
		payload := hex.EncodeToString([]byte(body))
		for _, marker := range []string{"🤖[" + payload + "]", "🤖{" + payload + "}", "🤖~" + payload + "~", "🤖[" + payload} {
			if out := workingdocMarkers(marker); len(out) != 1 || out[0].Text != marker {
				t.Fatalf("lost active marker %q: %#v", marker, out)
			}
			for _, literal := range []string{"`" + marker + "`", "~~~\n" + marker + "\n~~~"} {
				if out := workingdocMarkers(literal); len(out) != 0 {
					t.Fatalf("literal marker %q: %#v", literal, out)
				}
			}
		}
	})
}

func TestDocMarkerLinearWork(t *testing.T) {
	for _, unit := range []string{
		"🤖[ask] and 🤖{proposal} ",
		"🤖[unterminated ",
		"`unmatched \\\\\\` mixed ``code 🤖[literal]``\n",
		"~~~\n🤖[literal]\n~~~\n",
		"`unmatched\n\n🤖[open question]\n\ncloser`\n",
		"- `unmatched\n- 🤖[open question]\n- closer`\n",
		"    🤖[indented literal]\n\n",
	} {
		_, n := scanWorkingdocMarkers(strings.Repeat(unit, 512))
		input := strings.Repeat(unit, 1024)
		_, twice := scanWorkingdocMarkers(input)
		if twice > 2*n+16 || twice > 16*len(input) {
			t.Fatalf("nonlinear work for %q: N=%d 2N=%d bytes=%d", unit, n, twice, len(input))
		}
	}
}

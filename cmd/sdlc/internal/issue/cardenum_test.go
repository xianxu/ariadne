package issue

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// #348: card shape for the classification fields. Membership in a repository's
// declaration is never ParseCard's job, so an undeclared-but-well-formed area
// parses.
func TestParseCardAreaSeverityShape(t *testing.T) {
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	with := func(extra string) []byte {
		return []byte(strings.Replace(string(card), "status: open", "status: open\n"+extra, 1))
	}
	for _, ok := range []string{
		"severity: high",
		"severity:",
		"area: [couch/broadcast/web, sdlc]",
		"area: []",
		"area:",
		"area: [never/declared]",
		"area: [a, b, c, d, e, f, g, h]",
	} {
		if _, err := ParseCard(with(ok)); err != nil {
			t.Errorf("%q refused: %v", ok, err)
		}
	}
	for _, bad := range []string{
		"severity: urgent",
		"severity: [high]",
		"severity: 3",
		"area: couch",
		"area: [Couch]",
		"area: [a//b]",
		"area: [a/b/c/d/e]",
		"area: [a, a]",
		"area: [a, b, c, d, e, f, g, h, i]",
		"area: [a, [b]]",
		"area: [1]",
		"area:\n  - a",
		"area: {a: b}",
	} {
		if _, err := ParseCard(with(bad)); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestSetCardFieldAreaSeverity(t *testing.T) {
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		t.Fatal(err)
	}
	out, err := SetCardField(card, "area", "[couch/broadcast/web, sdlc]")
	if err != nil {
		t.Fatal(err)
	}
	if got, ok, err := FieldList(mustFM(t, out), "area"); err != nil || !ok || strings.Join(got, ",") != "couch/broadcast/web,sdlc" {
		t.Fatalf("%v %v %v", got, ok, err)
	}
	if out, err = SetCardField(out, "area", ""); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := FieldList(mustFM(t, out), "area"); ok {
		t.Fatal("cleared area still present")
	}
	if out, err = SetCardField(out, "severity", "critical"); err != nil || !strings.Contains(string(out), "severity: critical") {
		t.Fatalf("%v %s", err, out)
	}
	if _, err := SetCardField(out, "severity", "urgent"); err == nil {
		t.Fatal("invalid severity set")
	}
}

func mustFM(t *testing.T, raw []byte) string {
	t.Helper()
	fm, _, err := Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	return fm
}

func TestFieldList(t *testing.T) {
	for fm, want := range map[string]string{
		"area: [a/b, c]":      "a/b,c",
		"area:\n- a\n- b":     "a,b",
		"area:\n  - a/b":      "a/b",
		"area: []":            "",
		"other: 1\narea: [x]": "x",
	} {
		got, ok, err := FieldList(fm, "area")
		if err != nil || !ok || strings.Join(got, ",") != want {
			t.Errorf("%q: %v %v %v", fm, got, ok, err)
		}
	}
	for _, fm := range []string{"other: 1", "area:", "area: null"} {
		if _, ok, err := FieldList(fm, "area"); ok || err != nil {
			t.Errorf("%q: ok %v err %v", fm, ok, err)
		}
	}
	if _, _, err := FieldList("area: {a: b}", "area"); err == nil {
		t.Error("a mapping decoded as a list")
	}
}

// R1 PQ-1: sequences compare by content, not by the (empty) scalar Value.
func TestSameCardValueSequences(t *testing.T) {
	node := func(s string) *yaml.Node {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte("v: "+s), &doc); err != nil {
			t.Fatal(err)
		}
		return doc.Content[0].Content[1]
	}
	if !sameCardValue(node("[a, b]"), node("[a,  b]")) {
		t.Fatal("equal lists compared unequal")
	}
	if sameCardValue(node("[a, b]"), node("[a, c]")) || sameCardValue(node("[a]"), node("[a, b]")) {
		t.Fatal("different lists compared equal")
	}
}

func TestSetFieldDropsBlockSequence(t *testing.T) {
	fm := "id: 1\narea:\n- a\n- b\nstatus: open"
	got := SetField(fm, "area", "[c]")
	if got != "id: 1\narea: [c]\nstatus: open" {
		t.Fatalf("%q", got)
	}
}

// R1 PQ-1: a card area change reaches the details mirror; an unchanged one is
// a no-op.
func TestRefreshMirrorAreaChange(t *testing.T) {
	seeded := strings.Replace(cardDetailFixture, "status: open", "status: open\narea: [a/b]", 1)
	baseline, detail, err := SplitCard([]byte(seeded))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(baseline), "area: [a/b]") {
		t.Fatalf("split lost area:\n%s", baseline)
	}
	current, err := SetCardField(baseline, "area", "[a/b, c]")
	if err != nil {
		t.Fatal(err)
	}
	got, err := RefreshMirror(detail, baseline, current)
	if err != nil {
		t.Fatal(err)
	}
	if v, _, _ := FieldList(mustFM(t, got), "area"); strings.Join(v, ",") != "a/b,c" {
		t.Fatalf("mirror area %v:\n%s", v, got)
	}
	again, err := RefreshMirror(got, current, current)
	if err != nil || string(again) != string(got) {
		t.Fatalf("unchanged refresh rewrote the file: %v", err)
	}
}

// R1 PQ-1: migration carries valid values (a hand-written block list in the
// card's one-line form) and drops invalid ones with an inference.
func TestMigrateCarriesAreaSeverity(t *testing.T) {
	raw := "---\nid: 000007\nstatus: done\nactual_hours: 1\nseverity: high\narea:\n  - couch/broadcast\n  - sdlc\n---\n\n# Seven\n\n## Problem\nx\n"
	card, inferences, err := ArchivedCard([]byte(raw), "000007-seven.md", "sha1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(card), "severity: high") || !strings.Contains(string(card), "area: [couch/broadcast, sdlc]") || len(inferences) != 0 {
		t.Fatalf("card:\n%s\ninferences %v", card, inferences)
	}
	raw = strings.Replace(strings.Replace(raw, "severity: high", "severity: urgent", 1), "  - sdlc", "  - Bad", 1)
	card, inferences, err = ArchivedCard([]byte(raw), "000007-seven.md", "sha1")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(inferences, "\n")
	if strings.Contains(string(card), "severity") || strings.Contains(string(card), "area") || !strings.Contains(joined, "severity") || !strings.Contains(joined, "area") {
		t.Fatalf("card:\n%s\ninferences %v", card, inferences)
	}
}

// FuzzRepoEnumCardShape: any YAML value for area either refuses or is a flow
// list of valid paths; ParseCard never panics.
func FuzzRepoEnumCardShape(f *testing.F) {
	card, _, err := SplitCard([]byte(cardDetailFixture))
	if err != nil {
		f.Fatal(err)
	}
	for _, s := range []string{"[a/b]", "[]", "", "a", "[a, [b]]", "{a: 1}", "[A]", "[a, a]", "\n  - a", "[\"a/b\", 'c']", "[a/b/c/d]"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, value string) {
		raw := strings.Replace(string(card), "status: open", "status: open\narea: "+value, 1)
		c, err := ParseCard([]byte(raw))
		if err != nil {
			return
		}
		got, ok, ferr := FieldList(c.Frontmatter, "area")
		if ferr != nil {
			t.Fatalf("accepted card's area unreadable: %v (%q)", ferr, value)
		}
		spec := vocab.Issue().RepoEnums.Fields["area"]
		for _, v := range got {
			if !repoenum.ValidValue(spec, v) {
				t.Fatalf("invalid value %q accepted from %q", v, value)
			}
		}
		if ok && len(got) > spec.MaxItems {
			t.Fatalf("%d values accepted from %q", len(got), value)
		}
	})
}

// BR-2: FlowList quotes exactly the values YAML would read as another type, and
// every rendered list parses back to the same strings.
func TestFlowList(t *testing.T) {
	values := []string{"a/b", "2026", "true", "null", "1e5", "sdlc", "0x10", "on"}
	got := FlowList(values)
	if !strings.Contains(got, "a/b, \"2026\", \"true\", \"null\", \"1e5\", sdlc") {
		t.Fatalf("%s", got)
	}
	back, ok, err := FieldList("area: "+got, "area")
	if err != nil || !ok || strings.Join(back, ",") != strings.Join(values, ",") {
		t.Fatalf("round trip %v %v %v", back, ok, err)
	}
	var n yaml.Node
	if err := yaml.Unmarshal([]byte("v: "+got), &n); err != nil {
		t.Fatal(err)
	}
	for _, item := range n.Content[0].Content[1].Content {
		if item.Tag != "!!str" {
			t.Fatalf("%q decoded as %s", item.Value, item.Tag)
		}
	}
}

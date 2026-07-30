package issueref

import (
	"reflect"
	"regexp"
	"testing"
)

func TestFindSeparatesLocalFromForeign(t *testing.T) {
	cases := []struct {
		text string
		want []Ref
	}{
		// The convention (312 of ariadne's last 400 subjects).
		{"#187 M2: churn — four-bucket classification", []Ref{{Num: "187"}}},
		{"fixes #127, #128", []Ref{{Num: "127"}, {Num: "128"}}},
		{"(#127)", []Ref{{Num: "127"}}},
		{"PR #106", []Ref{{Num: "106"}}},
		// A RANGE stays two LOCAL refs: `-` is outside \b's word class, so the second is
		// not read as `174-`-qualified. This is the false positive a hand-written
		// preceded-by class would have introduced, and it appears in real history.
		{"#174-#176", []Ref{{Num: "174"}, {Num: "176"}}},
		// The bug: one subject carrying both a local and a foreign ref.
		{"#187 M2: pair#127 replay harness + round 1 evidence",
			[]Ref{{Num: "187"}, {Qualifier: "pair", Num: "127"}}},
		// Every real repo-name shape in the workspace.
		{"pair#127", []Ref{{Qualifier: "pair", Num: "127"}}},
		{"brain-family#12", []Ref{{Qualifier: "brain-family", Num: "12"}}},
		{"parley.nvim#12", []Ref{{Qualifier: "parley.nvim", Num: "12"}}},
		{"42shots#12", []Ref{{Qualifier: "42shots", Num: "12"}}},
		{"xianxu.dev#3", []Ref{{Qualifier: "xianxu.dev", Num: "3"}}},
		// Self-qualified: parsed WITH its qualifier; localness is the caller's call.
		{"ariadne#180", []Ref{{Qualifier: "ariadne", Num: "180"}}},
		{"no refs here", nil},
	}
	for _, c := range cases {
		if got := Find(c.text); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Find(%q) = %+v, want %+v", c.text, got, c.want)
		}
	}
}

// The {1,6} bound is inherited from refScanRE/parseRef, and RE2's trailing \b makes a
// 7+-digit run match NOTHING rather than a truncated 6-digit prefix. TestRewriteRefs pins
// this for migrate; pin it here too, since this is now the source.
func TestFindRejectsOverlongIDs(t *testing.T) {
	if got := Find("#1234567"); got != nil {
		t.Errorf("Find(#1234567) = %+v, want nil (7 digits is not a ref)", got)
	}
	if got := Find("#123456"); len(got) != 1 || got[0].Num != "123456" {
		t.Errorf("Find(#123456) = %+v, want one ref", got)
	}
}

func TestLocalNums(t *testing.T) {
	const subject = "#187 M2: pair#127 replay harness; also #187 and ariadne#180"
	if got, want := LocalNums(subject, "ariadne"), []string{"187", "180"}; !reflect.DeepEqual(got, want) {
		t.Errorf("LocalNums(selfRepo=ariadne) = %v, want %v (deduped, first-seen order)", got, want)
	}
	// Unknown self-repo → only bare refs are local, and nothing panics.
	if got, want := LocalNums(subject, ""), []string{"187"}; !reflect.DeepEqual(got, want) {
		t.Errorf(`LocalNums(selfRepo="") = %v, want %v`, got, want)
	}
	// The #190 defect itself: a foreign ref must never appear, even when its number names a
	// real local issue (ariadne#127 exists and absorbed 46 minutes of #187's work).
	for _, n := range LocalNums("pair#127", "ariadne") {
		if n == "127" {
			t.Error("pair#127 resolved to local 127 — the #190 defect")
		}
	}
}

func TestCountLocal(t *testing.T) {
	text := "working #187, replaying pair#127, more #187, and #190"
	tracked := map[string]bool{"187": true, "127": true, "190": true}
	if got, want := CountLocal(text, "ariadne", tracked), map[string]int{"187": 2, "190": 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("CountLocal = %v, want %v — 127 is pair's, not ours", got, want)
	}
	// The Compute contract: an empty tracked set matches nothing (previously expressed as a
	// nil *regexp). Untracked issues are excluded even when local.
	if got := CountLocal(text, "ariadne", nil); len(got) != 0 {
		t.Errorf("an empty tracked set must yield no mentions, got %v", got)
	}
	if got := CountLocal("", "ariadne", tracked); len(got) != 0 {
		t.Errorf("empty text must yield no mentions, got %v", got)
	}
	if got := CountLocal("#999 untracked", "ariadne", tracked); len(got) != 0 {
		t.Errorf("an untracked local ref must not count, got %v", got)
	}
}

// IsLocal is EXACT, deliberately unlike resolveRepoDir's exact-then-unique-prefix matching
// (resolve.go:193-199). Prefix matching is a navigation convenience; here it would be a
// correctness bug, re-introducing the very cross-repo bleed this package removes.
func TestIsLocalIsExactNotPrefix(t *testing.T) {
	if (Ref{Qualifier: "brain", Num: "1"}).IsLocal("brain-family") {
		t.Error("prefix matching would re-introduce cross-repo bleed")
	}
	if (Ref{Qualifier: "ariadne", Num: "1"}).IsLocal("ariadne") == false {
		t.Error("an exact self-qualified match must be local")
	}
	if (Ref{Num: "1"}).IsLocal("") == false {
		t.Error("a bare ref is always local")
	}
	if (Ref{Qualifier: "pair", Num: "1"}).IsLocal("") {
		t.Error(`selfRepo "" must not make every qualifier local`)
	}
}

// QualifiedIDPattern is exported so migrate.go's ANCHORED spanRefRE can compose from the
// same grammar instead of restating it. A compiled regexp cannot be re-anchored, so the
// fragment is the shareable unit — this is what makes the consolidation 5 → 1 rather than
// 5 → 2. Pin that it composes.
func TestQualifiedIDPatternComposesAnchored(t *testing.T) {
	anchored := mustCompileAnchored(t)
	for _, s := range []string{"#171", "ariadne#171", "pair#127", "#171 M4", "#171 M4b"} {
		if !anchored.MatchString(s) {
			t.Errorf("anchored pattern should match a whole-span ref %q", s)
		}
	}
	// The #179 corruption cases: a quoted command must NOT read as a whole-span ref.
	for _, s := range []string{`git log --grep "^#15"`, "see #171 for details", "#171 and #172"} {
		if anchored.MatchString(s) {
			t.Errorf("anchored pattern must not match mixed content %q", s)
		}
	}
}

// mustCompileAnchored builds the whole-span discriminator the way migrate.go's spanRefRE
// does, proving the exported fragment is composable.
func mustCompileAnchored(t *testing.T) *regexp.Regexp {
	t.Helper()
	return regexp.MustCompile(`^` + QualifiedIDPattern + `( M[0-9]+[a-z]?)?$`)
}

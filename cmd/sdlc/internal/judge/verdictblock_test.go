package judge

import (
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestParseVerdictBlock(t *testing.T) {
	cases := []struct {
		name, in, wantTok, wantConf string
		wantOK                      bool
	}{
		{"valid + confidence", "preamble\n```verdict\nverdict: FIX-THEN-SHIP\nconfidence: high\n```\nbody", "FIX-THEN-SHIP", "high", true},
		{"no confidence", "```verdict\nverdict: SHIP\n```", "SHIP", "", true},
		{"lowercase token uppercased", "```verdict\nverdict: rework\n```", "REWORK", "", true},
		{"missing block", "just prose, no fenced verdict block", "", "", false},
		{"model-invalid token", "```verdict\nverdict: MAYBE\n```", "", "", false},
		{"last block wins", "```verdict\nverdict: SHIP\n```\n...\n```verdict\nverdict: REWORK\n```", "REWORK", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tok, conf, ok := ParseVerdictBlock(c.in)
			if ok != c.wantOK || tok != c.wantTok || (c.wantOK && conf != c.wantConf) {
				t.Errorf("ParseVerdictBlock = (%q,%q,%v), want (%q,%q,%v)", tok, conf, ok, c.wantTok, c.wantConf, c.wantOK)
			}
		})
	}
}

func TestParseVerdict_BlockBeatsProse(t *testing.T) {
	// The session's failure mode (#143/#137): a prose verdict + a structured block
	// resolves correctly FROM THE BLOCK — would have degraded to `unknown` pre-#147.
	out := "My review is complete and the verdict stands: **FIX-THEN-SHIP**.\n\n" +
		"```verdict\nverdict: FIX-THEN-SHIP\nconfidence: high\n```\n"
	if v := ParseVerdict(out); v != VerdictFixThenShip {
		t.Errorf("block should win over prose: got %v, want FIX-THEN-SHIP", v)
	}
	// Prose-only (no block) still parses via the transitional fallback (no regression).
	if v := ParseVerdict("VERDICT: SHIP (confidence: high)\n\nlooks good"); v != VerdictShip {
		t.Errorf("prose fallback regressed: got %v", v)
	}
	// Block alone.
	if v := ParseVerdict("```verdict\nverdict: REWORK\n```"); v != VerdictRework {
		t.Errorf("block alone: got %v", v)
	}
	// Neither a block nor a parseable prose verdict → unknown.
	if v := ParseVerdict("just some prose, no verdict anywhere"); v != VerdictUnknown {
		t.Errorf("no verdict should be unknown: got %v", v)
	}
}

// TestVerdictDriftGuard pins the SHIP-family consumers in classify.go to the model
// (#147 disposition table) so verdict.cue stays the single source.
func TestVerdictDriftGuard(t *testing.T) {
	emitted := vocab.Verdict().Emitted()

	// Equality: the Go Verdict enum's emitted family == the model's Emitted().
	enum := []string{string(VerdictShip), string(VerdictFixThenShip), string(VerdictRework)}
	if !sameStringSet(enum, emitted) {
		t.Errorf("Verdict enum %v != model Emitted() %v — drift", enum, emitted)
	}

	// Derive: verdictFor maps every emitted token to itself; a non-emitted → unknown.
	for _, tok := range emitted {
		if got := verdictFor(tok); string(got) != tok {
			t.Errorf("verdictFor(%q)=%v, want %q", tok, got, tok)
		}
	}
	if verdictFor("CLEAN") != VerdictUnknown {
		t.Error("a non-emitted token must map to VerdictUnknown")
	}

	// Subset: every emitted token is matched by the prose-fallback regex, so the
	// transitional path covers the whole model.
	for _, tok := range emitted {
		if !verdictTokenRE.MatchString(tok) {
			t.Errorf("prose fallback verdictTokenRE does not match emitted token %q (drift)", tok)
		}
	}
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]int{}
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

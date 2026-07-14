package processmanual

import "testing"

// classifyOutputLine over REAL captured lines (copied verbatim from
// ~/.claude/projects) plus the documented-shape G2/refusal cases and — the whole
// point — the contamination cases the classifier MUST reject. Inventing fixtures is
// exactly the trap the #172 plan reviews caught; these are ground truth.
func TestClassifyOutputLine(t *testing.T) {
	esc := "\x1b"
	cases := []struct {
		name    string
		line    string
		verb    string
		wantOK  bool
		wantK   GateEventKind
		wantG   string
		wantObs Observability
	}{
		// ---- real runtime bypass ACKs (reset-bearing) ----
		{"g1 no-verdict ack", "  " + esc + "[1;33m[!]" + esc + "[0m --no-verdict (or --force): skipping Review-Verdict check for 1 milestone(s): M1",
			"close", true, GateBypass, "no-verdict", ObsFull},
		{"cinfo no-judge ack", esc + "[1;36m==>" + esc + "[0m skipping issue boundary review per --no-judge (or --force)",
			"close", true, GateBypass, "no-judge", ObsFull},
		{"g3 no-validate ack (⚠️ + double space)", "  " + esc + "[1;33m[!]" + esc + "[0m ⚠️  --no-validate: SKIPPING the instance-conformance gate (#124) — issue frontmatter NOT verified",
			"merge", true, GateBypass, "no-validate", ObsFull},
		// G2 change-code: force-only observability, ViaForce
		{"g2 plan-quality force-bypass", "  " + esc + "[1;33m[!]" + esc + "[0m plan-quality gate bypassed (--force: needed to iterate)",
			"change-code", true, GateBypass, "no-judge", ObsForceOnly},

		// ---- real runtime refusal (plain string, NO reset) ----
		{"no-actual refusal", "  Pass --no-actual (or --force) only when measurement is not applicable; close records actual_hours: N/A.",
			"close", true, GateRefusal, "no-actual", ObsFull},

		// ---- contamination: MUST classify as none ----
		{"warmup success line (not a refusal)", "             Pass --no-actual (or --force) only if there's genuinely nothing",
			"close", false, 0, "", 0},
		{"close.go source read (cat-n + append)", "944\t\ttail = append(tail, \"  Pass --no-actual (or --force) only when measurement is not applicable\")",
			"close", false, 0, "", 0},
		{"Sprintf source line", "        id := fmt.Sprintf(\"%06d\", issueNum)",
			"close", false, 0, "", 0},
		{"cat-n shell read", "226\t            --force) force=1; shift;;",
			"close", false, 0, "", 0},
		// verb anchoring: a close-gate ACK line seen under `merge` is not a merge gate
		{"wrong verb → none", "  " + esc + "[1;33m[!]" + esc + "[0m --no-verdict (or --force): skipping Review-Verdict check",
			"merge", false, 0, "", 0},
	}
	for _, c := range cases {
		ev, ok := classifyOutputLine(c.line, c.verb)
		if ok != c.wantOK {
			t.Errorf("%s: ok=%v want %v (ev=%+v)", c.name, ok, c.wantOK, ev)
			continue
		}
		if !ok {
			continue
		}
		if ev.Kind != c.wantK || ev.Gate != c.wantG || ev.Observability != c.wantObs {
			t.Errorf("%s: got {kind:%d gate:%q obs:%d} want {kind:%d gate:%q obs:%d}",
				c.name, ev.Kind, ev.Gate, ev.Observability, c.wantK, c.wantG, c.wantObs)
		}
	}
}

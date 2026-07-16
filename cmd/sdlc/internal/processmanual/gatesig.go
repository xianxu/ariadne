package processmanual

import (
	"regexp"
	"sort"
)

// Gate bypass/refusal signature catalog (#172 friction audit).
//
// The `sdlc` spine has 14 bypass gates across six commands. This catalog is the
// single source the friction classifier, the report, and the cross-command drift
// guard (cmd/sdlc/gates_test.go) all derive from. Its ground truth is the master
// table in workshop/plans/000172-sdlc-painpoint-audit-plan.md, itself enumerated
// from source + verified against the real transcript corpus.
//
// Three ACK grammars — one regex will not do (M1 Task 2 attaches the regexes):
//   G1 close/mclose : "[!] --no-<gate> (or --force): <verb…>"     (paren + colon)
//   cinfo           : "==> skipping … per --no-judge (or --force)" (no colon)
//   G2 change-code  : "<gate> gate bypassed (--force: <reason>)"   (silent unless --force)
//   G3 merge/push   : "[!] --no-<gate>: <verb…>"                   (no "(or --force)")

type ackGrammar int

const (
	grammarG1    ackGrammar = iota // close/milestone-close paren+colon
	grammarCinfo                   // close/milestone-close no-judge (cinfo skip)
	grammarG2                      // change-code (force-only ACK; silent alone)
	grammarG3                      // merge/push colon, no "(or --force)"
)

// GateSig is one catalog row: which sdlc verbs a gate applies to, its flag name,
// and its ACK grammar + observability metadata. Regexes (M1 Task 2) attach later.
type GateSig struct {
	Commands []string // sdlc verbs this signature applies to
	Flag     string   // e.g. "no-actual" (the registered --<flag> name, without "--")
	Grammar  ackGrammar

	// SilentAlone: change-code gates used WITHOUT --force skip silently (no ACK) —
	// a bypass is observable only when --force was used.
	SilentAlone bool
	// HasRefusal: false for auto-dispatch skips (close/mclose no-judge just skip a
	// review dispatch; there is no gate refusal).
	HasRefusal bool
	// RefusalNamesFlag: false when the refusal text never names the flag
	// (merge/push publish-gate) — refusal→retry attribution is best-effort there.
	RefusalNamesFlag bool

	// AckPat / RefusalPat: the classifier's patterns, matched against an ANSI-stripped
	// output line. AckPat is gated on the runtime reset by the classifier; RefusalPat
	// is grammar+digit-anchored (no reset — runtime refusals are plain strings) and
	// keyed on the exact per-gate tail so the printSemanticWarmup success line and the
	// catalog-read-as-transcript never match. Empty RefusalPat ⇒ no refusal.
	AckPat     string
	RefusalPat string

	ackRE, refusalRE *regexp.Regexp
}

func init() {
	for i := range GateCatalog {
		g := &GateCatalog[i]
		if g.AckPat != "" {
			g.ackRE = regexp.MustCompile(g.AckPat)
		}
		if g.RefusalPat != "" {
			g.refusalRE = regexp.MustCompile(g.RefusalPat)
		}
	}
}

// GateCatalog — the 18 signature rows over the 14 distinct spine gates.
var GateCatalog = []GateSig{
	// close / milestone-close — G1 (shared computeClose emits these). ACK = the
	// paren+colon form; refusal = the exact per-gate tail (NOT the shared prefix —
	// the printSemanticWarmup `only if there's genuinely nothing` must not match).
	{Commands: closeMclose, Flag: "no-actual", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-actual \(or --force\): closing with actual_hours`,
		RefusalPat: `Pass --no-actual \(or --force\) only when measurement is not applicable`},
	{Commands: closeMclose, Flag: "no-verified", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-verified \(or --force\): closing with NO verification evidence`,
		RefusalPat: `Pass --no-verified \(or --force\) only if there's genuinely no behavior`},
	{Commands: closeMclose, Flag: "no-reclose-guard", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-reclose-guard \(or --force\): re-closing`,
		RefusalPat: `is already status: done — pass --no-reclose-guard \(or --force\) to re-close`},
	{Commands: closeMclose, Flag: "no-atlas", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-atlas \(or --force\): skipping atlas/ change check`,
		RefusalPat: `pass --no-atlas \(or --force\) with the rationale`},
	{Commands: closeMclose, Flag: "no-verdict", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-verdict \(or --force\): skipping Review-Verdict check`,
		RefusalPat: `Or pass --no-verdict \(or --force\); record`},
	{Commands: closeMclose, Flag: "no-plan-check", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-plan-check \(or --force\): closing .* with \d+ unchecked`,
		RefusalPat: `pass --no-plan-check, or --force, to close anyway`},
	{Commands: closeMclose, Flag: "no-project", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-project \(or --force\): skipping detail-block`,
		RefusalPat: `--no-project, or --force, if it's`},
	{Commands: closeMclose, Flag: "no-judge", Grammar: grammarCinfo, HasRefusal: false,
		AckPat: `skipping (issue boundary review|milestone-review) per --no-judge \(or --force\)`},

	// project close — G1. Nested commands remain full catalog keys so transcript
	// attribution cannot conflate this boundary with issue close.
	{Commands: []string{"project close"}, Flag: "no-retro", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-retro \(or --force\): closing without a recorded project retro`,
		RefusalPat: `run .*project retro.*, or pass --no-retro \(or --force\)`},
	{Commands: []string{"project close"}, Flag: "no-ledger", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-ledger \(or --force\): skipping fog-factor ledger`,
		RefusalPat: `or pass --no-ledger \(or --force\)`},

	// change-code — G2, silent unless --force. ACK = "<base> gate[s] bypassed (--force:";
	// the base differs from the flag (no-judge → plan-quality/estimate-quality).
	{Commands: []string{"change-code"}, Flag: "no-judge", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `(plan-quality|estimate-quality) gate bypassed \(--force:`,
		RefusalPat: `(plan-quality|estimate-quality): findings reported`},
	{Commands: []string{"change-code"}, Flag: "no-structural", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `structural gates bypassed \(--force:`,
		RefusalPat: `structural-sanity gates failed:`},
	{Commands: []string{"change-code"}, Flag: "no-estimate", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `estimate gate bypassed \(--force:`,
		RefusalPat: `estimate gate failed:`},
	{Commands: []string{"change-code"}, Flag: "no-estimate-recon", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `estimate-reconciliation gate bypassed \(--force:`,
		RefusalPat: `estimate-reconciliation gate failed:`},

	// merge / push — G3, colon, no "(or --force)". no-validate ACK carries a ⚠️ +
	// double space (tolerated by not anchoring to line start). publish-gate refusals
	// never name the flag (RefusalNamesFlag=false) → best-effort attribution.
	{Commands: []string{"merge"}, Flag: "no-judge", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: false,
		AckPat:     `--no-judge: skipping the pre-merge publish gate`,
		RefusalPat: `publish gate: \d+ commit\(s\) landed after`},
	{Commands: []string{"merge"}, Flag: "no-validate", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-validate: SKIPPING the instance-conformance gate`,
		RefusalPat: `instance-conformance gate: \d+ nonconforming`},
	{Commands: []string{"push"}, Flag: "no-judge", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: false,
		AckPat:     `--no-judge: skipping the pre-push publish gate`,
		RefusalPat: `publish gate: \d+ commit\(s\) landed after`},
	{Commands: []string{"push"}, Flag: "no-validate", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: true,
		AckPat:     `--no-validate: SKIPPING the instance-conformance gate`,
		RefusalPat: `instance-conformance gate: \d+ nonconforming`},
}

var closeMclose = []string{"close", "milestone-close"}

// GateFlagNames returns the distinct gate flag names, sorted — the closed
// vocabulary of the 14 spine bypass gates.
func GateFlagNames() []string {
	seen := map[string]bool{}
	var out []string
	for _, g := range GateCatalog {
		if !seen[g.Flag] {
			seen[g.Flag] = true
			out = append(out, g.Flag)
		}
	}
	sort.Strings(out)
	return out
}

// GateFlagsFor returns the catalog's gate flags for one sdlc verb (unsorted).
func GateFlagsFor(command string) []string {
	var out []string
	for _, g := range GateCatalog {
		for _, c := range g.Commands {
			if c == command {
				out = append(out, g.Flag)
				break
			}
		}
	}
	return out
}

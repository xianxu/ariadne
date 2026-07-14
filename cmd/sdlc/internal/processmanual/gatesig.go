package processmanual

import "sort"

// Gate bypass/refusal signature catalog (#172 friction audit).
//
// The `sdlc` spine has 12 bypass gates across five commands. This catalog is the
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
}

// GateCatalog — the 16 signature rows over the 12 distinct spine gates.
var GateCatalog = []GateSig{
	// close / milestone-close — G1 (shared computeClose emits these)
	{Commands: closeMclose, Flag: "no-actual", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-verified", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-reclose-guard", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-atlas", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-verdict", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-plan-check", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-project", Grammar: grammarG1, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: closeMclose, Flag: "no-judge", Grammar: grammarCinfo, HasRefusal: false},

	// change-code — G2, silent unless --force
	{Commands: []string{"change-code"}, Flag: "no-judge", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: []string{"change-code"}, Flag: "no-structural", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: []string{"change-code"}, Flag: "no-estimate", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: []string{"change-code"}, Flag: "no-estimate-recon", Grammar: grammarG2, SilentAlone: true, HasRefusal: true, RefusalNamesFlag: true},

	// merge / push — G3, colon, no "(or --force)"
	{Commands: []string{"merge"}, Flag: "no-judge", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: false},
	{Commands: []string{"merge"}, Flag: "no-validate", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: true},
	{Commands: []string{"push"}, Flag: "no-judge", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: false},
	{Commands: []string{"push"}, Flag: "no-validate", Grammar: grammarG3, HasRefusal: true, RefusalNamesFlag: true},
}

var closeMclose = []string{"close", "milestone-close"}

// GateFlagNames returns the distinct gate flag names, sorted — the closed
// vocabulary of the 12 spine bypass gates.
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

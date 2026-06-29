package estimate

import "sort"

const currentModel = "estimate-logic-v3.1"

// primitives is the closed vocabulary of estimate-logic-v2 primitive slugs — the
// canonical source of truth. helptext/estimate.md documents it and a drift-guard
// test asserts the two match; the brain estimate-logic-v2.md primitive table is
// the human narrative this mirrors. Keep them reconciled if the set changes.
var primitives = map[string]bool{
	"pensive":                   true,
	"issue-spec":                true,
	"typed-data-prototype":      true,
	"skill-or-dispatcher":       true,
	"smaller-go-module":         true,
	"greenfield-go-module":      true,
	"api-integration":           true,
	"greenfield-service":        true,
	"tui-screen":                true,
	"cross-cutting-refactor":    true,
	"cross-repo-refactor-small": true,
	"cross-repo-refactor-large": true,
	"atlas-docs":                true,
	"lua-neovim":                true,
	"milestone-review":          true,
	"real-api-discovery":        true,
	"scope-pivot":               true,
	"ux-rename-iteration":       true,
	"method-b-decisions":        true,
}

// models is the recognized set of estimate model versions a provenance line may
// name. It widens as new models are calibrated (e.g. an operator-attention model
// from #112) — the grammar/guard/judge/ledger are unchanged when it does.
var models = map[string]bool{
	"estimate-logic-v2":   true,
	"estimate-logic-v2.1": true,
	"estimate-logic-v3":   true,
	"estimate-logic-v3.1": true,
}

// KnownPrimitive reports whether slug is in the closed v2 vocabulary.
func KnownPrimitive(slug string) bool { return primitives[slug] }

// KnownModel reports whether m is a recognized estimate model version.
func KnownModel(m string) bool { return models[m] }

// CurrentModel is the model new estimates should use by default.
func CurrentModel() string { return currentModel }

// Primitives returns the canonical slug set, sorted (for helptext + drift tests).
func Primitives() []string { return sortedKeys(primitives) }

// Models returns the recognized model versions, sorted.
func Models() []string { return sortedKeys(models) }

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

package layer

import "github.com/xianxu/ariadne/cmd/weave/internal/intent"

// Layer is a resolved layer: an on-disk sibling dir (Path) with a Name and the
// typed Intents parsed from its construct/base.manifest. A plain data struct —
// no IO (ARCH-PURE). The Resolver produces layer names in foundation-first
// order; the IO seam (part 2) reads each layer's base.manifest, parses it via
// intent.ParseManifest, and fills Intents. The Planner then lowers a
// foundation-first []Layer into filesystem Actions.
//
// Resolve emits root LAST and self-included, so in a fully-built []Layer the
// final element is the consuming repo itself; lowering that accounts for
// root-is-last-and-self (the self-reference filter is an IO-walk concern,
// deferred to part 2 — see plan.go's TODO).
//
// ProseFragments holds the RESOLVED text of this layer's `prose` intents, in
// intent order. The intents carry only relpaths (intent.Intent.Source);
// reading each fragment file is an IO-seam concern (part 2), so the seam fills
// ProseFragments when it builds the Layer. Keeping the content here (not on the
// Intent) lets the Planner stay pure — it composes fragments across layers
// without touching disk. Empty when the layer declares no prose.
type Layer struct {
	Name           string
	Path           string
	Intents        []intent.Intent
	ProseFragments []string
}

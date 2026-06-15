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
// ProseFragment is one resolved `prose` intent: its visibility (export|internal,
// from the manifest token — see intent.Visibility) paired with the resolved file
// Content. The visibility travels WITH the content so the Planner can select
// 𝒜(R) — every layer's export prose plus the leaf's internal prose — without
// re-reading the manifest. (Carrying it here, not on the bare string, is what
// lets the pure Planner compose visibility-aware prose without touching disk.)
type ProseFragment struct {
	Visibility intent.Visibility
	Content    string
}

// ProseFragments holds the RESOLVED `prose` intents of this layer, in intent
// order, each tagged with its visibility (see ProseFragment). The intents carry
// only relpaths (intent.Intent.Source); reading each fragment file is an IO-seam
// concern (part 2), so the seam fills ProseFragments when it builds the Layer.
// Keeping the content here (not on the Intent) lets the Planner stay pure — it
// composes fragments across layers without touching disk. Empty when the layer
// declares no prose.
type Layer struct {
	Name           string
	Path           string
	Intents        []intent.Intent
	ProseFragments []ProseFragment
}

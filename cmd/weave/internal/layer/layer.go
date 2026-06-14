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
type Layer struct {
	Name    string
	Path    string
	Intents []intent.Intent
}

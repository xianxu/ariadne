package plan

import "strings"

// composeProse concatenates prose fragments foundation-first (in the slice
// order given) into the AGENTS.md body. Pure. Fragments arrive already ordered
// by the planner (foundation layer first, the consuming repo's own fragment
// last), so this just joins them.
//
// This is the structural fix for the @AGENTS.local.md bug: setup.sh symlinks a
// single AGENTS.md whose body @-imports AGENTS.local.md, which silently resolves
// to the FOUNDATION's local file in a derivative. weave instead concatenates
// every layer's fragment in directly — no @-import — so each repo gets its own
// prose. The flip from symlinked AGENTS.md to a composed real file is the
// expected golden-diff divergence (see plan Core concepts).
//
// Fragments are separated by a blank line and the body ends with a trailing
// newline; an empty fragment list yields "" (the planner then emits no
// AGENTS.md WriteFile rather than an empty file).
func composeProse(fragments []string) string {
	if len(fragments) == 0 {
		return ""
	}
	return strings.Join(fragments, "\n\n") + "\n"
}

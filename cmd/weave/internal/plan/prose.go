package plan

import (
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
)

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

// composeSkillMenu renders the always-on `## Skills` section appended to the
// composed AGENTS.md: a note pointing the agent at `weave skill <name>` for a
// body, then one `name — description` line per skill in menu order. This is the
// FIRST of the skill server's two faces (the menu compiled into the system
// prompt); `weave skill <name>` serves the bodies on demand. Agent-agnostic by
// construction — it rides AGENTS.md, not .claude/skills/ discovery. Pure; empty
// menu ⇒ "" (the planner then appends nothing).
func composeSkillMenu(menu []skill.MenuItem) string {
	if len(menu) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Skills\n\n")
	b.WriteString("These skills are served directly by `weave` (no `.claude/skills/` discovery). ")
	b.WriteString("Run `weave skill <name>` to load a skill's full body on demand.\n\n")
	for _, m := range menu {
		b.WriteString("- ")
		b.WriteString(m.Name)
		if m.Description != "" {
			b.WriteString(" — ")
			b.WriteString(m.Description)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// composeAgentsBody assembles the full AGENTS.md body: the foundation-first
// prose, then (if any skills) the `## Skills` menu section, separated by a
// blank line. Returns "" only when there is neither prose nor a menu — the
// planner then emits no AGENTS.md. Pure.
func composeAgentsBody(fragments []string, menu []skill.MenuItem) string {
	prose := composeProse(fragments)
	skills := composeSkillMenu(menu)
	switch {
	case prose != "" && skills != "":
		return prose + "\n" + skills
	case prose != "":
		return prose
	case skills != "":
		return skills
	default:
		return ""
	}
}

package plan

import "fmt"

// Target is the backend a single `weave compile` invocation lowers for — one
// target per invocation (the Approach-1 design). It is pure data (a string
// enum) and lives here, beside Plan, because the one decision a Target drives is
// how the skill intents lower: the .claude/skills symlink backend versus the
// AGENTS.md `## Skills` menu backend. The two are MUTUALLY EXCLUSIVE per target
// (a harness reads its skills ONE way), so the Target picks exactly one:
//
//   - claude → .claude/skills/<name> symlinks; AGENTS.md prose-only (NO menu).
//     The Claude harness discovers skills from .claude/skills, so the always-on
//     menu would be redundant noise in its system prompt.
//   - codex / agy → NO .claude/skills symlinks; AGENTS.md carries the `## Skills`
//     menu. These harnesses have no skill-discovery dir, so the menu IS the
//     discovery face (bodies still served on demand via `weave skill <name>`).
//
// All non-skill file-ops (prose body, settings merge, tool, scaffold, touch,
// generic symlink, seed) are target-INDEPENDENT.
type Target string

const (
	// TargetClaude lowers skills as .claude/skills symlinks (menu suppressed).
	TargetClaude Target = "claude"
	// TargetCodex lowers skills as the AGENTS.md menu (no symlinks).
	TargetCodex Target = "codex"
	// TargetAgy lowers skills as the AGENTS.md menu (no symlinks), like codex.
	TargetAgy Target = "agy"
)

// ParseTarget validates a backend name against the known set, returning a clear
// error on an unknown name. Pure.
func ParseTarget(s string) (Target, error) {
	switch Target(s) {
	case TargetClaude, TargetCodex, TargetAgy:
		return Target(s), nil
	default:
		return "", fmt.Errorf("unknown target %q; expected one of: claude, codex, agy", s)
	}
}

// EmitSkillSymlinks reports whether this target lowers the .claude/skills
// symlink backend. True for claude; false for the menu-driven harnesses.
func (t Target) EmitSkillSymlinks() bool {
	return t == TargetClaude
}

// IncludeSkillMenu reports whether this target composes the `## Skills` menu
// into AGENTS.md. The complement of EmitSkillSymlinks — the two skill backends
// are mutually exclusive per target.
func (t Target) IncludeSkillMenu() bool {
	return !t.EmitSkillSymlinks()
}

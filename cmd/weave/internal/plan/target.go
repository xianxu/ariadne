package plan

import "fmt"

// Target selects which per-harness FACES a single `weave compile` lowers
// (Option B, #107). A harness FACE is its pure-prose ENTRY FILE + its skill
// DISCOVERY DIR — disjoint paths per harness, so one checkout serves every
// harness with no contested file:
//
//	claude → CLAUDE.md  + .claude/skills/   (Claude Code reads CLAUDE.md only)
//	codex  → AGENTS.md  + .agents/skills/   (OpenAI Codex reads AGENTS.md)
//	gemini → GEMINI.md  + .agents/skills/   (Gemini CLI reads GEMINI.md)
//
// codex + gemini SHARE the .agents/skills dir (the Agent Skills neutral path).
// There is NO `## Skills` menu: every harness discovers its skill dir NATIVELY
// (Codex/Gemini auto-compose their own menu from .agents/skills; weave emitting
// one would double-expose). The entry files are the SAME composed prose under
// three names.
//
//	weave compile  (default, TargetAll) = the UNION: every face at once.
//	weave compile --target T            = the LEAN subset: only T's face.
//
// All non-skill, non-prose file-ops (settings merge, generic symlinks, scaffolds,
// seeds) are target-INDEPENDENT — always present regardless of target.
type Target string

const (
	TargetAll    Target = "all"    // the Union — every harness's face (the default)
	TargetClaude Target = "claude" // CLAUDE.md + .claude/skills
	TargetCodex  Target = "codex"  // AGENTS.md + .agents/skills
	TargetGemini Target = "gemini" // GEMINI.md + .agents/skills
)

// Face is one harness's Option B artifact set: its prose ENTRY FILE + its skill
// DISCOVERY DIR. Pure data.
type Face struct {
	Target    Target
	EntryFile string
	SkillDir  string
}

// faces is the per-harness face registry — the single source of truth for "what
// does harness T put where". Order is the Union's lowering order (claude first).
var faces = []Face{
	{TargetClaude, "CLAUDE.md", ".claude/skills"},
	{TargetCodex, "AGENTS.md", ".agents/skills"},
	{TargetGemini, "GEMINI.md", ".agents/skills"},
}

// ParseTarget validates a target name. The empty string (no --target) is the
// Union default. Pure.
func ParseTarget(s string) (Target, error) {
	switch Target(s) {
	case "", TargetAll:
		return TargetAll, nil
	case TargetClaude, TargetCodex, TargetGemini:
		return Target(s), nil
	default:
		return "", fmt.Errorf("unknown target %q; expected one of: all, claude, codex, gemini", s)
	}
}

// Faces returns the faces this target lowers: every harness for the Union
// (TargetAll), or the single selected harness for a lean compile. Pure.
func (t Target) Faces() []Face {
	if t == TargetAll {
		return append([]Face(nil), faces...)
	}
	for _, f := range faces {
		if f.Target == t {
			return []Face{f}
		}
	}
	return nil
}

// EntryFiles + SkillDirs are the deduped paths the target's faces write to (codex
// + gemini share .agents/skills, so the Union has ONE .agents/skills dir). The
// EntryFiles all receive the SAME composed prose; the SkillDirs each receive the
// full selected skill set. Pure.
func (t Target) EntryFiles() []string {
	return dedupe(t.Faces(), func(f Face) string { return f.EntryFile })
}
func (t Target) SkillDirs() []string {
	return dedupe(t.Faces(), func(f Face) string { return f.SkillDir })
}

func dedupe(fs []Face, key func(Face) string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range fs {
		k := key(f)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

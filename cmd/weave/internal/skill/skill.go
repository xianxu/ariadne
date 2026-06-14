// Package skill is weave's agent-agnostic skill server core: a PURE index over
// the skills gathered across a repo's resolved layers. It backs the two faces of
// the skill server — the always-on menu compiled into the composed AGENTS.md,
// and `weave skill <name>` serving a body on demand — WITHOUT relying on the
// harness's .claude/skills/ discovery (that symlink delivery is setup.sh's
// mechanism; weave serves skills directly).
//
// Build takes already-parsed Entries (name, description, body-path) gathered
// foundation-first by the IO seam (the walk reads each layer's SKILL.md files;
// see walk.GatherSkills) and produces an ordered, collision-free menu plus a
// name→body-path lookup. Pure — no IO (ARCH-PURE); reading SKILL.md off disk is
// the seam's job.
package skill

// Entry is one parsed skill, as gathered by the IO seam: its NAMESPACED name
// (the discovery-time symlink name — `<prefix><dir>` for a local skill, the
// bare `<dir>` for an adapted one, ported from sync-local-skills.sh, ARCH-DRY),
// the description from its SKILL.md frontmatter, and the absolute path to that
// SKILL.md (the body weave serves). Entries arrive foundation-first.
type Entry struct {
	Name        string
	Description string
	BodyPath    string
}

// MenuItem is one line of the always-on skill menu compiled into AGENTS.md:
// just the name + description (the body is served on demand via the lookup).
type MenuItem struct {
	Name        string
	Description string
}

// SkillIndex is the resolved, collision-free skill index: an ordered menu and a
// name→body-path lookup. Construct it via Build; the zero value is an empty
// index. Immutable after Build (read-only accessors).
type SkillIndex struct {
	menu   []MenuItem
	bodies map[string]string
}

// Build folds foundation-first Entries into a SkillIndex, applying the cascade:
// the FIRST appearance of a name fixes its slot in the menu; a later
// (downstream) layer re-declaring that name OVERRIDES the description + body in
// place (without reordering), so a derivative can customize a base skill. This
// mirrors the layer cascade weave applies everywhere — downstream wins, order
// preserved. Pure: a single pass, no IO.
func Build(entries []Entry) SkillIndex {
	idx := SkillIndex{bodies: map[string]string{}}
	pos := map[string]int{} // name → index into idx.menu (for in-place override)
	for _, e := range entries {
		idx.bodies[e.Name] = e.BodyPath // last writer wins ⇒ downstream override
		if i, seen := pos[e.Name]; seen {
			idx.menu[i].Description = e.Description // override in original slot
			continue
		}
		pos[e.Name] = len(idx.menu)
		idx.menu = append(idx.menu, MenuItem{Name: e.Name, Description: e.Description})
	}
	return idx
}

// Menu returns the ordered, collision-free menu (foundation-first by first
// appearance). Safe to render straight into the AGENTS.md `## Skills` section.
func (s SkillIndex) Menu() []MenuItem { return s.menu }

// BodyPath returns the absolute SKILL.md path for a skill name, and whether the
// name is known. The path resolves to the OVERRIDING (most-downstream) body when
// multiple layers declared the same name.
func (s SkillIndex) BodyPath(name string) (string, bool) {
	p, ok := s.bodies[name]
	return p, ok
}

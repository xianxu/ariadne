package walk

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// skills.go is the skill-discovery IO seam — the read side of weave's
// agent-agnostic skill server. It ports sync-local-skills.sh's discovery
// (ARCH-DRY: that hook is the source of truth for WHERE skills live and HOW
// they're named/prefixed):
//
//   - construct/local/<dir>/SKILL.md   → namespaced name "<prefix><dir>"
//   - construct/adapted/<dir>/SKILL.md → namespaced name "<dir>" (no prefix)
//
// The prefix comes from the layer's own construct/config.json `localPrefix`
// (default "xx-", matching the hook's fallback). Names are derived from the
// DIRECTORY name + prefix — the same string the hook computes for its
// .claude/skills symlink — NOT from the SKILL.md frontmatter `name:` (which is
// inconsistent across skills); the description IS read from frontmatter.
//
// The result feeds skill.Build (the pure index). Reading SKILL.md/config.json
// off disk is confined here (the seam); the index reasoning stays pure.

// adaptedSkillRel is the one skill dir whose skills stay BARE — external skills
// (superpowers) keep their published names; every OTHER skill dir gets the layer's
// prefix (skillPrefix: config.json localPrefix, else the repo name).
const adaptedSkillRel = "construct/adapted"

// GatherSkills is weave's SINGLE skill discovery (skill-system v2, #104): walking
// the resolved layers foundation-first, for each layer it reads that layer's
// `skill <dir>` INTENTS (not a hardcoded dir pair) and scans each declared dir,
// emitting one skill.Entry per skill. Every entry carries the composition-algebra
// inputs — the declaring row's Visibility and the layer's index — so the menu
// (skill.Build) and the claude symlinks (plan.SkillSymlinks) both derive from the
// SAME selected set (skill.SelectVisible); there is no second scan (ARCH-DRY,
// closes #104 §A1/§A4). IO confined to this seam (ARCH-PURE).
//
// Naming: a skill in `construct/adapted` keeps its bare dir name (external skills
// preserve their published names); every other dir (`construct/local`,
// `construct/skill`, …) gets the layer's prefix. Order: layers foundation-first,
// intents in manifest order, skills within a dir alphabetical (ReadDir sorted),
// so a downstream re-declaration of a name overrides (skill.Build's cascade).
func GatherSkills(fs weavefs.FS, layers []layer.Layer) ([]skill.Entry, error) {
	var entries []skill.Entry
	for i, l := range layers {
		prefix := skillPrefix(fs, l.Path)
		for _, in := range l.Intents {
			if in.Kind != intent.Skill {
				continue
			}
			rowPrefix := prefix
			if in.Source == adaptedSkillRel { // external skills keep their bare names
				rowPrefix = ""
			}
			es, err := scanSkillDir(fs, filepath.Join(l.Path, in.Source), rowPrefix)
			if err != nil {
				return nil, err
			}
			for j := range es {
				es[j].Visibility = in.Visibility
				es[j].LayerIndex = i
			}
			entries = append(entries, es...)
		}
	}
	return entries, nil
}

// scanSkillDir lists sourceDir and, for each child directory that ships a
// SKILL.md, parses its frontmatter description and emits an Entry named
// "<prefix><dir>". An absent sourceDir is no skills (not an error — a layer
// need not ship every category, like the hook's `[[ -d ]] || return 0`). A dir
// lacking a SKILL.md is skipped (it isn't a skill).
func scanSkillDir(fs weavefs.FS, sourceDir, prefix string) ([]skill.Entry, error) {
	dirents, err := fs.ReadDir(sourceDir)
	if err != nil {
		return nil, nil // absent / unreadable source dir ⇒ no skills here
	}
	var out []skill.Entry
	for _, de := range dirents {
		if !de.IsDir() {
			continue
		}
		bodyPath := filepath.Join(sourceDir, de.Name(), "SKILL.md")
		data, rerr := fs.ReadFile(bodyPath)
		if rerr != nil {
			continue // a dir with no SKILL.md is not a skill (hook scans `*/`)
		}
		out = append(out, skill.Entry{
			Name:        prefix + de.Name(),
			Description: frontmatterDescription(string(data)),
			BodyPath:    bodyPath,
		})
	}
	return out, nil
}

// skillPrefix resolves a layer's skill-name prefix (#104 M2): its
// construct/config.json `localPrefix` when set, ELSE the layer's REPO NAME +
// "-" (the layer-root dir basename). So each layer prefixes its OWN skills by
// its repo name — `nous-`, `brain-`, `pair-` — and ariadne keeps `xx-` by
// setting it in its own config.json. (This is behavior-preserving while every
// derivative's config.json is still symlinked to ariadne's `xx-`; the repo-name
// default activates when #104 M3 un-symlinks those config.json files.)
func skillPrefix(fs weavefs.FS, layerRoot string) string {
	if data, err := fs.ReadFile(filepath.Join(layerRoot, "construct", "config.json")); err == nil {
		var cfg struct {
			LocalPrefix string `json:"localPrefix"`
		}
		if json.Unmarshal(data, &cfg) == nil && cfg.LocalPrefix != "" {
			return cfg.LocalPrefix
		}
	}
	return filepath.Base(layerRoot) + "-"
}

// frontmatterDescription extracts the `description:` field from a SKILL.md's
// leading YAML frontmatter block (the `---` … `---` fence). It reads the one
// field weave's menu needs; a full YAML parse is overkill (frontmatter here is
// flat key: value). A surrounding pair of quotes on the value is stripped
// (some skills quote the description). Returns "" if no description is present.
func frontmatterDescription(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "" // no frontmatter fence
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break // end of frontmatter
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "description" {
			continue
		}
		return unquote(strings.TrimSpace(val))
	}
	return ""
}

// unquote strips one symmetric pair of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

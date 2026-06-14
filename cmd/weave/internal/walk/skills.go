package walk

import (
	"encoding/json"
	"path/filepath"
	"strings"

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

// localSkillDir and adaptedSkillDir are the two skill source dirs each layer
// ships, relative to the layer root (ported from sync-local-skills.sh's
// LOCAL_DIR / ADAPTED_DIR).
const (
	localSkillRel   = "construct/local"
	adaptedSkillRel = "construct/adapted"
	defaultPrefix   = "xx-" // sync-local-skills.sh:19 fallback
)

// GatherSkills walks the resolved layers foundation-first and collects every
// skill each one ships into []skill.Entry — local skills prefixed, adapted bare
// — in the cascade order skill.Build expects (foundation layer's skills first,
// the consuming repo's last, so a downstream re-declaration overrides). Within
// a layer: local skills before adapted, each alphabetical (ReadDir is sorted).
// IO confined to this seam (ARCH-PURE).
func GatherSkills(fs weavefs.FS, layers []layer.Layer) ([]skill.Entry, error) {
	var entries []skill.Entry
	for _, l := range layers {
		prefix := localPrefix(fs, l.Path)
		// Local skills get the prefix; adapted keep their bare dir name.
		local, err := scanSkillDir(fs, filepath.Join(l.Path, localSkillRel), prefix)
		if err != nil {
			return nil, err
		}
		adapted, err := scanSkillDir(fs, filepath.Join(l.Path, adaptedSkillRel), "")
		if err != nil {
			return nil, err
		}
		entries = append(entries, local...)
		entries = append(entries, adapted...)
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

// localPrefix reads a layer's construct/config.json `localPrefix`, falling back
// to "xx-" when the file is absent/unparseable or the field is empty — ported
// from sync-local-skills.sh's `PREFIX="${PREFIX:-xx-}"`.
func localPrefix(fs weavefs.FS, layerRoot string) string {
	data, err := fs.ReadFile(filepath.Join(layerRoot, "construct", "config.json"))
	if err != nil {
		return defaultPrefix
	}
	var cfg struct {
		LocalPrefix string `json:"localPrefix"`
	}
	if json.Unmarshal(data, &cfg) != nil || cfg.LocalPrefix == "" {
		return defaultPrefix
	}
	return cfg.LocalPrefix
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

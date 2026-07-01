package retro

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/pkg/frontmatter"
)

// catalogCategories is the COMPLETE injected set — judge.AllCategories() omits
// EstimateQuality (scoped to standalone `sdlc judge` validity), but that prompt
// IS injected at change-code (changecode.go), so the manual must list it
// (ARCH-PURPOSE: deliver every injection point, not the easy subset).
func catalogCategories() []judge.Category {
	return append(append([]judge.Category{}, judge.AllCategories()...), judge.EstimateQuality)
}

// categoryBody is what the manual shows for a judge category. BuildPrompt is a
// pure function, so an empty PromptInput yields the deterministic prompt
// skeleton — EXCEPT `lessons`, which BuildPrompt renders as "" because it is a
// reminder ping, not a full agent prompt. For lessons the injected content IS
// judge.LessonsReminder. (This judge `lessons` category is a DIFFERENT injection
// point from the workshop/lessons.md file, which fileSources emits under Kind
// `lessons`.)
func categoryBody(c judge.Category) string {
	if body := judge.BuildPrompt(c, judge.PromptInput{}); strings.TrimSpace(body) != "" {
		return body
	}
	if c == judge.Lessons {
		return judge.LessonsReminder
	}
	return ""
}

// whenForCategory maps each category to its injection-trigger prose (pure).
func whenForCategory(c judge.Category) string {
	switch c {
	case judge.PlanQuality, judge.EstimateQuality:
		return "plan-quality gate at `sdlc change-code`"
	case judge.MilestoneReview:
		return "boundary review at `sdlc close` / `sdlc milestone-close`"
	case judge.Lessons:
		return "lessons reminder emitted at the review boundary"
	default:
		return "`sdlc judge " + string(c) + "`"
	}
}

// judgeSources is pure (judge.BuildPrompt + a constant do no IO).
func judgeSources() []InjectionSource {
	var out []InjectionSource
	for _, c := range catalogCategories() {
		out = append(out, InjectionSource{
			Kind:  KindSDLCPrompt,
			Title: string(c),
			When:  whenForCategory(c),
			Link:  "cmd/sdlc/internal/judge/prompts.go",
			Body:  categoryBody(c),
		})
	}
	return out
}

// helptextSources enumerates the embedded sdlc help-text files (the process
// manual baked into the binary). Injected with an fs.FS (production passes
// helptext.FS()) so it tests against a fake FS. Single enumeration path — the
// helptext package exposes only FS(), no separate Names() (ARCH-DRY).
func helptextSources(fsys fs.FS) []InjectionSource {
	ents, _ := fs.ReadDir(fsys, ".")
	var out []InjectionSource
	for _, e := range ents {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		stem := strings.TrimSuffix(name, ".md")
		content, _ := fs.ReadFile(fsys, name)
		out = append(out, InjectionSource{
			Kind:  KindHelpText,
			Title: stem,
			When:  helptextWhen(stem),
			Link:  "cmd/sdlc/helptext/" + name,
			Body:  firstParagraph(string(content)),
		})
	}
	return out
}

// helptextWhen keeps the trigger prose truthful. `root.md` is emitted by bare
// `sdlc --help`, not `sdlc root --help`; other stems map to a verb or sub-verb,
// so we stay generic-but-accurate rather than assert a `sdlc <stem>` that may
// not exist (e.g. set-status/fetch are sub-verbs).
func helptextWhen(stem string) string {
	if stem == "root" {
		return "printed by bare `sdlc --help` (the workflow contract)"
	}
	return "embedded help; printed by the matching `sdlc … --help` / on verb error"
}

// firstParagraph returns the first blank-line-delimited paragraph, trimmed.
func firstParagraph(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "\n\n"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

// skillSources enumerates <skillsDir>/*/SKILL.md — the on-demand agent skills.
// The trigger (`When`) is the frontmatter `description:` (parsed via the shared
// pkg/frontmatter, ARCH-DRY). Entries under .claude/skills are symlinks into
// construct/…; os.ReadFile follows them, and Link is resolved through the
// symlink + made repo-root-relative so the manual points at the real SKILL.md.
func skillSources(skillsDir, repoRoot string) []InjectionSource {
	entries, _ := os.ReadDir(skillsDir)
	var out []InjectionSource
	for _, e := range entries {
		skillPath := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			continue // not a skill dir (or no SKILL.md)
		}
		link := skillPath
		if resolved, rerr := filepath.EvalSymlinks(skillPath); rerr == nil {
			link = resolved
		}
		if rel, rerr := filepath.Rel(repoRoot, link); rerr == nil {
			link = rel
		}
		_, body, _ := frontmatter.Split(string(content))
		out = append(out, InjectionSource{
			Kind:  KindSkill,
			Title: e.Name(),
			When:  frontmatter.Description(string(content)),
			Link:  link,
			Body:  firstParagraph(body),
		})
	}
	return out
}

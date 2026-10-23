package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// copySkillTree copies src to dst, recreating symlinks as symlinks.
func copySkillTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dst, rel)
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(target, out)
		case info.IsDir():
			return os.MkdirAll(out, 0o755)
		default:
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return os.WriteFile(out, b, 0o644)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

// #399: a repository derived from ariadne discovers xx-universal-review in
// both harness skill directories, and the skill's grammar link reaches the one
// authored review-convention.md through each of them.
func TestUniversalReviewSkillDiscoveredInDerivedRepo(t *testing.T) {
	ariadne, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived := filepath.Join(parent, "derived")
	for _, root := range []string{base, derived} {
		mkfile(t, filepath.Join(root, "construct", "config.json"), `{"localPrefix": "xx-"}`+"\n")
		mkfile(t, filepath.Join(root, "construct", "base.manifest"), "prose AGENTS.local.md\nskill construct/local\n")
		mkfile(t, filepath.Join(root, "AGENTS.local.md"), "PROSE")
	}
	for _, skill := range []string{"universal-review", "fix"} {
		copySkillTree(t, filepath.Join(ariadne, "construct", "local", skill), filepath.Join(base, "construct", "local", skill))
	}
	mkfile(t, filepath.Join(derived, "construct", "deps"), "substrate ../base\n")

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	canon, err := os.ReadFile(filepath.Join(ariadne, "construct", "local", "fix", "review-convention.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{".claude/skills", ".agents/skills"} {
		skill := filepath.Join(derived, dir, "xx-universal-review")
		body, err := os.ReadFile(filepath.Join(skill, "SKILL.md"))
		if err != nil || !strings.Contains(string(body), "name: xx-universal-review") {
			t.Errorf("%s: xx-universal-review not discovered (err=%v)", dir, err)
			continue
		}
		grammar, err := os.ReadFile(filepath.Join(skill, "review-convention.md"))
		if err != nil || !bytes.Equal(grammar, canon) {
			t.Errorf("%s: the grammar link does not reach the canonical review-convention.md (err=%v)", dir, err)
		}
		for _, f := range []string{"round.md", "adapters/file-edit.md", "briefs/copy-editing.md", "export.json"} {
			if _, err := os.Stat(filepath.Join(skill, f)); err != nil {
				t.Errorf("%s: %s: %v", dir, f, err)
			}
		}
	}
}

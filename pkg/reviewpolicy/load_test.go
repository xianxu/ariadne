package reviewpolicy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/agentenv"
)

func TestLoad(t *testing.T) {
	root := t.TempDir()
	if p, err := Load(root); err != nil || p.EffectiveMode() != DefaultMode {
		t.Fatalf("absent config: %+v %v", p, err)
	}
	write := func(body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(ConfigPath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"boundary_reviewer": "other", "reviewer_order": {"bounded": ["gemini", "codex"]}, "reviewer_models": {"codex": "gpt-6-astra"}}`)
	p, err := Load(root)
	if err != nil || p.Mode != Other || p.Models[Codex] != "gpt-6-astra" || len(p.Orders[Bounded]) != 2 {
		t.Fatalf("loaded %+v %v", p, err)
	}
	for _, bad := range []string{`{"boundary_reviewer": "others"}`, `{"boundary_reviewr": "other"}`, `{"boundary_reviewer": `, `{"reviewer_order": {"bounded": ["claude"]}}`} {
		write(bad)
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), ConfigPath) {
			t.Errorf("%s: %v, want an error naming %s", bad, err, ConfigPath)
		}
	}
}

func TestAuthorFamilies(t *testing.T) {
	commits := []agentenv.Commit{
		{SHA: "a", Agents: []string{"codex/gpt-6-astra/medium"}},
		{SHA: "m", Merge: true, Agents: []string{"gemini/g/high"}},
		{SHA: "b", Agents: []string{"qoder/q/low", "gemini"}},
	}
	got := AuthorFamilies(Anthropic, commits)
	var s []string
	for _, f := range got {
		s = append(s, string(f))
	}
	if strings.Join(s, ",") != "anthropic,openai,google" {
		t.Errorf("families %v: the session first, a merge and an unknown agent ignored", s)
	}
}

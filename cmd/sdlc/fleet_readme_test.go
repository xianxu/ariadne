package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestREADME_DocumentsFleetQueries(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(repoRootForTest(t), "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	const heading = "## Fleet queries"
	start := strings.Index(text, heading)
	if start < 0 {
		t.Fatal("README missing fleet queries section")
	}
	section := text[start:]
	if next := strings.Index(section[len(heading):], "\n## "); next >= 0 {
		section = section[:len(heading)+next]
	}
	normalized := strings.Join(strings.Fields(section), " ")
	for _, want := range []string{
		"sdlc fleet inventory --path .",
		"sdlc fleet inventory --path /path/to/nested/checkout --json",
		"sdlc fleet policy --path /path/to/prospective/actor",
		"sdlc fleet policy --path /path/to/prospective/actor --json",
		"typed diagnostic to stdout",
		"exits nonzero",
		"never infers policy from a repository name",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("README fleet queries section missing %q:\n%s", want, section)
		}
	}
}

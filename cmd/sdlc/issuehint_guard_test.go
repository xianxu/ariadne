package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// TestIssueHintsRenderDecimalIDs guards the class behind #252 BR-5: a
// canonical ID interpolated into a `--issue` hint prints "000253", which the
// flag library parses as octal. Every `--issue %s` / `"--issue "+` hint must
// render its value through issue.CLIRef or a variable known to hold a decimal.
func TestIssueHintsRenderDecimalIDs(t *testing.T) {
	hint := regexp.MustCompile(`--issue (%s|"\s*\+)`)
	decimal := regexp.MustCompile(`CLIRef\(|\bissueStr\b|\bnum\)|unpadID\(`)
	err := filepath.Walk(".", func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			if !hint.MatchString(line) {
				continue
			}
			window := strings.Join(lines[i:min(i+4, len(lines))], "\n")
			if !decimal.MatchString(window) {
				t.Errorf("%s:%d: --issue hint without issue.CLIRef: %s", p, i+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"000253", "000258", "000008", "001000"} {
		n, err := strconv.ParseInt(issue.CLIRef(id), 0, 64) // base 0: what a padded value would become
		want, _ := strconv.Atoi(id)
		if err != nil || int(n) != want {
			t.Errorf("CLIRef(%s) = %s parses as %d, want %d", id, issue.CLIRef(id), n, want)
		}
	}
}

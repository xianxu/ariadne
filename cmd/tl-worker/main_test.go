package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/tlrules"
)

func helpOutput(t *testing.T, args ...string) string {
	t.Helper()
	root := buildRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("tl-worker %v: %v", args, err)
	}
	return out.String()
}

// workerRules reads testdata/rules.tsv: one row per rule the help must carry.
func workerRules(t *testing.T) []tlrules.Rule {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "rules.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := tlrules.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 10 {
		t.Fatalf("rules.tsv has %d rows; the table did not load", len(rows))
	}
	return rows
}

func TestWorkerHelpCarriesEveryRule(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {}} {
		help := helpOutput(t, args...)
		for _, r := range workerRules(t) {
			if !strings.Contains(help, r.Phrase) {
				t.Errorf("tl-worker %v: rule %s: help lacks %q", args, r.Name, r.Phrase)
			}
		}
	}
}

// The worker prose cites AGENTS §15, §5 and couch --skill instead of restating
// them; a restatement would drift from its source.
func TestWorkerHelpDoesNotRestateSection15(t *testing.T) {
	help := helpOutput(t, "--help")
	for _, cite := range []string{"AGENTS §15", "AGENTS §5", "`couch --skill`"} {
		if !strings.Contains(help, cite) {
			t.Errorf("help does not cite %s", cite)
		}
	}
	for _, restated := range []string{"sandbox exemption", "substitution", "$(", "in-pane question prompt", "WF_TEST_SUITES", "publish-description"} {
		if strings.Contains(help, restated) {
			t.Errorf("help restates %q, which its source already owns", restated)
		}
	}
}

func TestSkillStubsPointAtBinaries(t *testing.T) {
	for _, bin := range []string{"tl-worker", "tl-lead"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "construct", "local", bin, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		stub := string(raw)
		if !strings.Contains(stub, "`"+bin+" --help`") {
			t.Errorf("%s stub does not point at `%s --help`", bin, bin)
		}
		if n := strings.Count(stub, "\n"); n > 30 {
			t.Errorf("%s stub has %d lines; a pointer carries no rules", bin, n)
		}
	}
}

// AGENTS §15 is the deterministic load: every Couch slot reads the
// constitution, and §15 sends it to the right binary.
func TestSection15NamesTheBinaries(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "AGENTS.base.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	start := strings.Index(s, "### 15. Couch Slots")
	if start < 0 {
		t.Fatal("AGENTS.base.md has no §15")
	}
	sec := s[start:]
	if end := strings.Index(sec[1:], "\n## "); end >= 0 {
		sec = sec[:end+1]
	}
	for _, want := range []string{"`tl-worker --help`", "`tl-lead --help`"} {
		if !strings.Contains(sec, want) {
			t.Errorf("§15 does not name %s", want)
		}
	}
}

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
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
		t.Fatalf("tl-lead %v: %v", args, err)
	}
	return out.String()
}

// Each row names a topic the lead's procedure must cover and a phrase that
// only that topic's text carries.
func TestLeadHelpCoversPlaybook(t *testing.T) {
	topics := []struct{ topic, phrase string }{
		{"the eight-step playbook", "RUNNING A PROJECT"},
		{"playbook step 1, frame", "one end-to-end done_when against a real"},
		{"playbook step 2, contracts first", "Contracts first."},
		{"playbook step 8, MVP run and closing", "MVP run, then `closing`, then `closed`"},
		{"talking to the operator", "TALKING TO THE OPERATOR"},
		{"status file sections", "`## Needs you`"},
		{"session-sync: the operator's lock", "The operator's lock is the only lock"},
		{"dispatch judgment", "The dispatch message carries only"},
		{"product stewardship", "STEWARDSHIP"},
		{"design rule 1", "reduce to a few visible states"},
		{"design rule 2", "educated, probabilistic guess with the top alternatives"},
		{"the smoke loop via sdlc stage", "THE SMOKE LOOP VIA `sdlc stage`"},
		{"stage verification", "vcs.modified=false"},
		{"attention and load", "ATTENTION AND LOAD"},
		{"the project lifecycle", "set the project\n     `executing`"},
		{"grooming trains", "GROOMING THE BACKLOG INTO TRAINS"},
		{"reading a slot's state", "Idle is not free"},
		{"fleet traps", "FLEET TRAPS"},
		{"allocation across operators", "ALLOCATION ACROSS OPERATORS"},
		{"the project directory", "THE PROJECT DIRECTORY"},
	}
	for _, args := range [][]string{{"--help"}, {}} {
		help := helpOutput(t, args...)
		for _, tc := range topics {
			if !strings.Contains(help, tc.phrase) {
				t.Errorf("tl-lead %v: topic %q: help lacks %q", args, tc.topic, tc.phrase)
			}
		}
	}
}

// The prose is current rules and procedure only (#350 Revisions): no incident
// numbers, no dated stories, and every reference names its repository.
func TestLeadHelpHasNoNarratives(t *testing.T) {
	help := helpOutput(t, "--help")
	for _, tc := range []struct {
		what string
		re   *regexp.Regexp
	}{
		{"a kink number", regexp.MustCompile(`(?i)kink \d+`)},
		{"a learnings reference", regexp.MustCompile(`(?i)learnings \d+`)},
		{"a dated story", regexp.MustCompile(`\(\d\d-\d\d\b|\b10-1\d\b`)},
		{"a bare issue reference", regexp.MustCompile(`(^|[^\w.])#\d+`)},
		{"an until-it-lands interim rule", regexp.MustCompile(`(?i)until \S+#\d+ lands`)},
	} {
		if m := tc.re.FindString(help); m != "" {
			t.Errorf("help carries %s: %q", tc.what, m)
		}
	}
}

// The lead cites tl-worker for the worker's rules and does not restate them.
func TestLeadHelpDoesNotRestateWorker(t *testing.T) {
	help := helpOutput(t, "--help")
	if !strings.Contains(help, "`tl-worker --help`") {
		t.Error("help does not cite `tl-worker --help`")
	}
	f, err := os.Open(filepath.Join("..", "tl-worker", "testdata", "rules.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rules, err := tlrules.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	owned := 0
	for _, r := range rules {
		if r.Scope != tlrules.Own {
			continue
		}
		owned++
		if strings.Contains(help, r.Phrase) {
			t.Errorf("help restates tl-worker rule %s (%q)", r.Name, r.Phrase)
		}
	}
	if owned < 8 {
		t.Fatalf("only %d worker-owned rules loaded; the table did not load", owned)
	}
}

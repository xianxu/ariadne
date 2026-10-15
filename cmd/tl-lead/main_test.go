package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/pkg/tlrules"
)

func helpOutput(t *testing.T, args ...string) string {
	t.Helper()
	root := buildRoot(deps{})
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
		{"the quota section", "`## Quota`"},
		{"threads as ### subsections", "`### Project: <slug>`"},
		{"the mutual-focus rule", "source of truth for mutual focus"},
		{"notify only for what needs the operator now", "`couch --notify [--title T] MESSAGE`"},
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
		// The rules #377 ported from the ops draft, one phrase each.
		{"G1 Operator instructions is section 0", "0. `## Operator instructions`"},
		{"G2 the TL only links answers", "Your only write is appending"},
		{"G2 every bullet gets an answer", "every bullet gets one, questions and"},
		// #399: other documents' submits are review requests.
		{"review requests use the shared skill", "review request: answer it with the\n    `xx-universal-review` skill"},
		// #388: a deleted bullet closes its QA file.
		{"a deleted bullet closes its QA file", "A deleted bullet closes its QA file"},
		{"follow-ups never reopen a closed file", "never into the closed file"},
		{"G3 QA file path", "teams/<team>/qa/<operator>/YYYY-MM-DD-NN-<slug>.md"},
		{"G3 QA frontmatter", "`process-instruction: \"[free-form] answer the operator's 🤖 comments in place\"`"},
		{"G3 backlink without anchor", "with no `#anchor`"},
		{"G3 answers only the question", "answers only the question asked"},
		{"G4 Needs you holds only asks", "only asks, a decision or an action"},
		{"G4 remove when done", "Remove an\n       item the moment it is done"},
		{"G5 CI glance on main", "Glance at CI on main every pass"},
		{"G6 render from binaries", "Render sections from binaries wherever possible"},
		{"G6 a retyped fact is the next lift", "next piece to lift into a binary"},
		{"G7 quiet pane", "Keep the agent pane quiet"},
		{"G8 a lock with no submit", "A lock with no submit stays"},
		{"frontmatter key", "`couch-owner: <hub slot>` (couch still reads the older `owner:`)"},
		{"G10 verify slots after a disruption", "Verify the slots after a disruption"},
		{"G10 a send is not evidence", "is not evidence the slot resumed"},
		{"G11 filing publishes then resets", "`sdlc issue publish --issue N`, then reset"},
		// From the fresh comparison of the draft against this help.
		{"a dispatch names the repo only", "the repo\n    prefix is enough"},
		{"update on every thread change", "Update it whenever a\n    thread changes state"},
		{"Couch encodes no one repo's usage", "Don't ask Couch to encode one repo's usage"},
		{"stage, never move", "never `sdlc move` (which transfers ownership)"},
		{"waves are trains", "waves are fine"},
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

// The procedure names every command in the tree, and no retired script.
func TestLeadHelpNamesSubcommands(t *testing.T) {
	help := helpOutput(t, "--help")
	var walk func(c *cobra.Command, path string)
	n := 0
	walk = func(c *cobra.Command, path string) {
		for _, sub := range c.Commands() {
			if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			p := path + " " + sub.Name()
			if sub.HasSubCommands() {
				walk(sub, p)
				continue
			}
			n++
			if !strings.Contains(help, p+" ") {
				t.Errorf("help does not name %q", p)
			}
		}
	}
	walk(buildRoot(deps{}), "tl-lead")
	if n < 10 {
		t.Fatalf("walked only %d commands; the tree did not load", n)
	}
	for _, retired := range []string{"ss.py", "tl-status.py", "render-trains.py", "opinstr.py", "scripts/"} {
		if strings.Contains(help, retired) {
			t.Errorf("help names the retired %q", retired)
		}
	}
}

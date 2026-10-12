package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The brainstorming and writing-plans skills hand a design to `sdlc
// change-code`, whose plan-quality gate is the one mandatory automated review
// of a plan (#228, ARCH-DRY: one owner per review boundary, one findings
// ledger). These tests pin that neither skill grows back its own mandatory
// approval loop, and that both name the chain the design travels.

const reviewChain = "brainstorming/spec → durable plan → `change-code` plan-quality gate → code → milestone/close review"

func readAdapted(t *testing.T, rel ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{"..", "..", "construct", "adapted"}, rel...)...)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func TestBrainstormingSpecReviewIsAdvisory(t *testing.T) {
	skill := readAdapted(t, "superpowers-brainstorming", "SKILL.md")
	for _, banned := range []string{
		"re-dispatch until approved",
		"repeat until Approved",
		"Spec review passed?",
		"re-run the spec review loop",
	} {
		if strings.Contains(skill, banned) {
			t.Errorf("brainstorming SKILL.md carries %q: the spec review is a mandatory approval loop again (#228)", banned)
		}
	}
	for _, want := range []string{"Optional Spec Review", "advisory", "agreement is the design approval", reviewChain} {
		if !strings.Contains(skill, want) {
			t.Errorf("brainstorming SKILL.md lost %q (#228)", want)
		}
	}
}

func TestWritingPlansDefersToChangeCode(t *testing.T) {
	skill := readAdapted(t, "superpowers-writing-plans", "SKILL.md")
	for _, banned := range []string{"Repeat until ✅ Approved", "## Plan Review Loop"} {
		if strings.Contains(skill, banned) {
			t.Errorf("writing-plans SKILL.md carries %q: its plan reviewer is a second mandatory approval again (#228)", banned)
		}
	}
	for _, want := range []string{
		"The mandatory review of this plan is `sdlc change-code`'s plan-quality gate",
		"Optional editing aid",
		"run `sdlc change-code`",
		reviewChain,
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("writing-plans SKILL.md lost %q (#228)", want)
		}
	}
}

func TestDocumentReviewersDoNotApprove(t *testing.T) {
	for _, rel := range [][]string{
		{"superpowers-brainstorming", "spec-document-reviewer-prompt.md"},
		{"superpowers-writing-plans", "plan-document-reviewer-prompt.md"},
	} {
		prompt := readAdapted(t, rel...)
		if strings.Contains(prompt, "Approved") {
			t.Errorf("%s reports \"Approved\": an advisory reviewer must not present itself as an approval gate (#228)", filepath.Join(rel...))
		}
		if !strings.Contains(prompt, "Advisory, not a gate") {
			t.Errorf("%s no longer says it is advisory (#228)", filepath.Join(rel...))
		}
	}
}

// The atlas names the same chain in the same words, so the map and the skills
// cannot drift apart.
func TestLifecycleAtlasNamesReviewChain(t *testing.T) {
	path := filepath.Join("..", "..", "atlas", "workflow", "issue-lifecycle.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(body), reviewChain) {
		t.Errorf("%s no longer states the review chain %q (#228)", path, reviewChain)
	}
}

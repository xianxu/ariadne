package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestInstructionsLinkMonthAnswers(t *testing.T) {
	for _, c := range []struct{ target, parent string }{
		{"tl-status-op.side/2026-10/2026-10-23-01-answer.md", "../../tl-status-op.md"},
		{"tl-status-op.side/2026-10-23-01-answer.md", "../tl-status-op.md"},
	} {
		t.Run(c.target, func(t *testing.T) {
			f, file := indexFake(t)
			name := filepath.Join(filepath.Dir(file), c.target)
			if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
				t.Fatal(err)
			}
			original := "---\ntype: side-doc\nparent: " + c.parent + "\npresentation: propose\nprocess-instruction: \"[copy-edit] Keep names.\"\n---\n\n# Answer\n\nKept bytes.\n"
			if err := os.WriteFile(name, []byte(original), 0o644); err != nil {
				t.Fatal(err)
			}
			if out, err := link(t, f, file, "1", c.target); err != nil {
				t.Fatalf("instructions link: %v\n%s", err, out)
			}
			if !strings.Contains(f.sections[instructionsHeading], "1. what does S4 cost? → [answer]("+c.target+")") || !slices.Equal(entries(t, f), []string{c.target}) {
				t.Fatalf("link/index not published: %s\n%v", f.sections[instructionsHeading], entries(t, f))
			}
			got, err := os.ReadFile(name)
			if err != nil || string(got) != original {
				t.Fatalf("answer document changed: %v\n%s", err, got)
			}
		})
	}
}

func TestAnswerTargetMonthBounds(t *testing.T) {
	const side = "session.side/"
	for _, c := range []struct {
		name, target, parent string
		ok                   bool
	}{
		{"month", side + "2026-10/2026-10-23-01-answer.md", "../../session.md", true},
		{"leap day", side + "2024-02/2024-02-29-123-answer.md", "../../session.md", true},
		{"flat compatibility", side + "2026-10-23-01-answer.md", "../session.md", true},
		{"month mismatch", side + "2026-09/2026-10-23-01-answer.md", "../../session.md", false},
		{"invalid month", side + "2026-13/2026-13-23-01-answer.md", "../../session.md", false},
		{"invalid day", side + "2026-02/2026-02-29-01-answer.md", "../../session.md", false},
		{"extra depth", side + "2026-10/extra/2026-10-23-01-answer.md", "../../../session.md", false},
		{"undated", side + "2026-10/01-answer.md", "../../session.md", false},
		{"other root", "other.side/2026-10/2026-10-23-01-answer.md", "../../session.md", false},
		{"wrong parent depth", side + "2026-10/2026-10-23-01-answer.md", "../session.md", false},
		{"wrong parent file", side + "2026-10/2026-10-23-01-answer.md", "../../other.md", false},
		{"missing parent", side + "2026-10/2026-10-23-01-answer.md", "", false},
		{"absolute target", "/" + side + "2026-10/2026-10-23-01-answer.md", "../../session.md", false},
		{"traversal", "other/../" + side + "2026-10/2026-10-23-01-answer.md", "../../session.md", false},
		{"flat traversal", "other/../" + side + "2026-10-23-01-answer.md", "../session.md", false},
		{"nested traversal", side + "2026-09/../2026-10/2026-10-23-01-answer.md", "../../session.md", false},
		{"backslash", side + `2026-10\2026-10-23-01-answer.md`, "../../session.md", false},
		{"URL", "https://example.test/" + side + "2026-10/2026-10-23-01-answer.md", "../../session.md", false},
		{"fragment", side + "2026-10/2026-10-23-01-answer.md#answer", "../../session.md", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := answerTarget("session.md", c.target, func(string) ([]byte, error) {
				return []byte("---\ntype: side-doc\nparent: " + c.parent + "\n---\n"), nil
			})
			if (err == nil) != c.ok {
				t.Fatalf("answerTarget(%q): %v; want accepted=%v", c.target, err, c.ok)
			}
		})
	}
}

func TestInstructionsLinkMonthRefusesWithoutWrites(t *testing.T) {
	for _, c := range []struct{ name, target, parent string }{
		{"wrong month", "2026-09/2026-10-23-01-answer.md", "../../tl-status-op.md"},
		{"invalid date", "2026-02/2026-02-29-01-answer.md", "../../tl-status-op.md"},
		{"extra depth", "2026-10/extra/2026-10-23-01-answer.md", "../../../tl-status-op.md"},
		{"wrong parent", "2026-10/2026-10-23-01-answer.md", "../tl-status-op.md"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f, file := indexFake(t)
			target := "tl-status-op.side/" + c.target
			name := filepath.Join(filepath.Dir(file), target)
			if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, []byte("---\ntype: side-doc\nparent: "+c.parent+"\n---\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := link(t, f, file, "1", target); err == nil || f.writes != 0 || f.frontWrites != 0 || f.sections[instructionsHeading] != indexOps {
				t.Fatalf("refusal must leave both operator section and index unchanged: err=%v writes=%d/%d", err, f.writes, f.frontWrites)
			}
		})
	}
}

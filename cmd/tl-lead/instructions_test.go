package main

import (
	"github.com/xianxu/ariadne/pkg/sidedoc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The section as the operator types it: a preamble, repeated numbers,
// multi-line bullets with blank lines inside, image lines, and both answered
// forms.
const opSection = `Numbered; dd a bullet when satisfied.

1. Why did ariadne:3 stall?
   It was quiet from 14:00.

   Second paragraph of the same bullet.

1. Rename the train
2. Answered already → [answer](qa/op/2026-10-14-01-x.md)
3. Look at this screenshot
   ![](assets/tl-status-op/shot.png)

4. Done by hand → done: merged
`

func TestInstructionBullets(t *testing.T) {
	bs := instructionBullets(opSection)
	type row struct {
		pos, start, lines int
		first             string
		answered          bool
	}
	want := []row{
		{1, 2, 4, "1. Why did ariadne:3 stall?", false},
		{2, 7, 1, "1. Rename the train", false},
		{3, 8, 1, "2. Answered already → [answer](qa/op/2026-10-14-01-x.md)", true},
		{4, 9, 2, "3. Look at this screenshot", false},
		{5, 12, 1, "4. Done by hand → done: merged", true},
	}
	if len(bs) != len(want) {
		t.Fatalf("got %d bullets, want %d: %+v", len(bs), len(want), bs)
	}
	for i, w := range want {
		b := bs[i]
		if b.Pos != w.pos || b.Start != w.start || len(b.Lines) != w.lines || b.Lines[0] != w.first || b.Answered != w.answered {
			t.Errorf("bullet %d = {%d %d %d %q %v}, want %+v", i, b.Pos, b.Start, len(b.Lines), b.Lines[0], b.Answered, w)
		}
	}
	// A blank line inside a bullet stays; trailing ones go.
	if got := bs[0].Lines[2]; got != "" {
		t.Errorf("inner blank line lost: %q", got)
	}
	if len(instructionBullets("")) != 0 || len(instructionBullets("no bullets here\n")) != 0 {
		t.Error("a section with no item line has no bullets")
	}
}

// #412: a bullet is chat mode's item: -, * and N. lines at the section's
// shallowest list indent, sub-items inside it, and prose after a blank line
// belonging to no bullet.
func TestInstructionBulletsAreChatItems(t *testing.T) {
	const sec = `1. first
   - a sub-item
- check CI
* and the trains
  continued here

A note, not a bullet.
2. last`
	var got []string
	for _, b := range instructionBullets(sec) {
		got = append(got, strings.Join(b.Lines, "|"))
	}
	want := []string{"1. first|   - a sub-item", "- check CI", "* and the trains|  continued here", "2. last"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("bullets\n got %q\nwant %q", got, want)
	}
	// Linking bullet 2 lands on its own line, never on the line before it.
	if out, err := linkBullet(sec, 2, "x.side/2026-10-16-01-ci.md", true); err != nil || !strings.Contains(out, "- check CI → [answer](x.side/2026-10-16-01-ci.md)") || !strings.Contains(out, "   - a sub-item\n") {
		t.Errorf("link 2: %v\n%s", err, out)
	}
	// Fenced code starts no bullet and a tab indents. A link ending any line
	// answers a bullet, so a follow-up added after it keeps it answered
	// (row 14); a link inside a fence is code, not an answer. Code follows
	// CommonMark (#576): this fence belongs to item 1, so its contents are
	// indented under it; an unindented line would end item 1 and its fence.
	const edge = "1. q\n   ```\n   - not an item\n   x → [answer](z.md)\n   ```\n\t- tabbed sub-item\n2. r → [answer](x.md)\n   - and by repo?\n3. s → [answer](y.md)"
	bs := instructionBullets(edge)
	if len(bs) != 3 || len(bs[0].Lines) != 6 || bs[0].Answered || !bs[1].Answered || !bs[2].Answered {
		t.Errorf("edges: %+v", bs)
	}
	// Indented items only: the shallowest indent present is the item level.
	if bs := instructionBullets("  - a\n    - b\n  - c"); len(bs) != 2 || len(bs[0].Lines) != 2 {
		t.Errorf("indented items: %+v", bs)
	}
}

func TestRenderOpen(t *testing.T) {
	got := renderOpen(instructionBullets(opSection), "abc123")
	want := "1\t1. Why did ariadne:3 stall?\n" +
		"\t   It was quiet from 14:00.\n\t\n\t   Second paragraph of the same bullet.\n" +
		"2\t1. Rename the train\n" +
		"4\t3. Look at this screenshot\n\t   ![](assets/tl-status-op/shot.png)\n" +
		"hash: abc123\n"
	if got != want {
		t.Errorf("renderOpen:\n%s\nwant:\n%s", got, want)
	}
}

func TestLinkBullet(t *testing.T) {
	const target = "qa/op/2026-10-14-02-stall.md"
	// The link goes on the bullet's last non-blank line, even an image line,
	// and nothing else changes.
	for _, c := range []struct {
		pos      int
		lineWant string
	}{
		{1, "   Second paragraph of the same bullet. → [answer](" + target + ")"},
		{2, "1. Rename the train → [answer](" + target + ")"},
		{4, "   ![](assets/tl-status-op/shot.png) → [answer](" + target + ")"},
	} {
		got, err := linkBullet(opSection, c.pos, target, true)
		if err != nil {
			t.Fatalf("pos %d: %v", c.pos, err)
		}
		gl, wl := strings.Split(got, "\n"), strings.Split(opSection, "\n")
		changed := 0
		for i := range wl {
			if gl[i] != wl[i] {
				changed++
				if gl[i] != c.lineWant {
					t.Errorf("pos %d: line %d = %q, want %q", c.pos, i, gl[i], c.lineWant)
				}
			}
		}
		if changed != 1 || len(gl) != len(wl) {
			t.Errorf("pos %d: %d lines changed (of %d→%d), want exactly 1", c.pos, changed, len(wl), len(gl))
		}
	}
	if got, err := linkBullet(opSection, 3, target, true); err != errAlreadyLinked || got != opSection {
		t.Errorf("an answered bullet: err %v, changed %v; want errAlreadyLinked and no change", err, got != opSection)
	}
	for _, bad := range []string{"", "qa/a b.md", "qa/a.md\n9. fake", "qa/(x).md"} {
		if _, err := linkBullet(opSection, 1, bad, true); err == nil {
			t.Errorf("target %q: want a refusal", bad)
		}
	}
	// #412: chat mode's answer docs live in the status file's .side/ folder.
	const side = "tl-status-op.side/2026-10-16-01-s4-cost.md"
	if got, err := linkBullet("1. what does S4 cost?", 1, side, true); err != nil || got != "1. what does S4 cost? → [answer]("+side+")" {
		t.Errorf("a .side/ target: %q, %v", got, err)
	}
	// CRLF: the carriage return is trimmed before the link, as rstrip() does.
	if got, _ := linkBullet("1. a\r", 1, target, true); got != "1. a → [answer]("+target+")" {
		t.Errorf("CRLF line: %q", got)
	}
	for _, pos := range []int{0, 6} {
		if _, err := linkBullet(opSection, pos, target, true); err == nil || !strings.Contains(err.Error(), "no bullet") {
			t.Errorf("pos %d: err %v, want a no-bullet refusal", pos, err)
		}
	}
}

func TestInstructionsListAndLink(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	file, _ := statusFile(t, "")
	out, err := run(t, deps{ss: f}, "status", "instructions", "list", file)
	if err != nil {
		t.Fatal(err)
	}
	h := fakeHash(opSection)
	if !strings.HasSuffix(out, "hash: "+h+":legacy\n") || !strings.HasPrefix(out, "1\t1. Why") {
		t.Fatalf("list output:\n%s", out)
	}
	a, b, c := answerDoc(t, file, "2026-10-18-01-a.md"), answerDoc(t, file, "2026-10-18-02-b.md"), answerDoc(t, file, "2026-10-18-03-c.md")
	out, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "2", a, "--hash", h)
	if err != nil || f.writes != 1 || !strings.Contains(f.sections[instructionsHeading], "1. Rename the train → [answer]("+a+")") {
		t.Fatalf("link: err %v, writes %d, out %q", err, f.writes, out)
	}
	// The section moved (this write changed it): the old hash now refuses.
	if _, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "1", b, "--hash", h); exitCode(err) != 6 || f.writes != 1 {
		t.Errorf("a stale hash: exit %d, writes %d; want exit 6 and no write", exitCode(err), f.writes)
	}
	// Already answered: no write.
	h = fakeHash(f.sections[instructionsHeading])
	if out, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "2", c, "--hash", h); err != nil || f.writes != 1 || !strings.Contains(out, "already answered") {
		t.Errorf("an answered bullet: err %v, writes %d, out %q", err, f.writes, out)
	}
}

// The operator edits between the read and the write: link refuses rather than
// retrying, since the position came from the older read.
func TestInstructionsLinkNeverRetries(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	f.editBefore[1] = map[string]string{instructionsHeading: "1. A new first bullet\n" + opSection}
	file, _ := statusFile(t, "")
	_, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", answerDoc(t, file, "2026-10-18-01-a.md"), "--hash", fakeHash(opSection))
	if exitCode(err) != 6 || f.writeCalls != 1 || f.writes != 0 {
		t.Errorf("exit %d, write calls %d, writes %d; want exit 6 after one attempt and no write", exitCode(err), f.writeCalls, f.writes)
	}
}

func TestInstructionsLinkRefusals(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	file, _ := statusFile(t, "")
	h := fakeHash(opSection)
	a := answerDoc(t, file, "2026-10-18-01-a.md")
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", a); err == nil {
		t.Error("link without --hash must refuse")
	}
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "two", a, "--hash", h); err == nil {
		t.Error("a non-numeric POS must refuse")
	}
	f.holder = "operator"
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", a, "--hash", h); exitCode(err) != 3 || f.writes != 0 {
		t.Errorf("under the operator's lock: exit %d, writes %d; want exit 3, no write", exitCode(err), f.writes)
	}
}

// The parser reads the skill's own chat example as the skill means it: in
// before/, items 1 and 2 are open and the rest answered (item 3 keeps its
// link under a follow-up sub-item); in after/, every item is answered. This
// ties tl-lead to the reference fixture so the two can't drift (#412).
func TestInstructionBulletsReadTheChatExample(t *testing.T) {
	for tree, want := range map[string]string{"before": "1. 2.", "after": ""} {
		b, err := os.ReadFile(filepath.Join("..", "..", "construct", "local", "universal-review", "examples", "chat", tree, "status.md"))
		if err != nil {
			t.Fatal(err)
		}
		_, sec, ok := strings.Cut(string(b), "### Operator instructions (open)\n")
		sec, _, _ = strings.Cut(sec, "\n#")
		if !ok {
			t.Fatalf("%s: no Operator instructions section", tree)
		}
		bs := instructionBullets(sec)
		var open []string
		for _, bl := range bs {
			if !bl.Answered {
				open = append(open, strings.Fields(bl.Lines[0])[0])
			}
		}
		if len(bs) != 5 || strings.Join(open, " ") != want {
			t.Errorf("%s: %d bullets, open %q, want 5 and %q", tree, len(bs), open, want)
		}
	}
}

// answerDoc writes a valid answer doc of the status file (#455) and returns
// its link target.
func answerDoc(t *testing.T, file, name string) string {
	t.Helper()
	side := strings.TrimSuffix(filepath.Base(file), ".md") + ".side"
	if err := os.MkdirAll(filepath.Join(filepath.Dir(file), side), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: side-doc\nparent: ../" + filepath.Base(file) + "\nmarker: \"ask\"\ncollaboration-mode: discussion\n---\n\n# Ask\n"
	if err := os.WriteFile(filepath.Join(filepath.Dir(file), side, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return side + "/" + name
}

// link writes only side-doc answers of the session file (#455): each refusal
// names the expected shape and writes nothing, the index included.
func TestInstructionsLinkOnlySideDocAnswers(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	file, _ := statusFile(t, "")
	dir := filepath.Dir(file)
	write := func(rel, text string) string {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		return rel
	}
	fm := func(extra string) string { return "---\ntype: side-doc\n" + extra + "---\n# x\n" }
	for _, c := range []struct{ target, want string }{
		{write("qa/op/2026-10-18-01-a.md", fm("parent: ../../tl-status-op.md\n")), "not an answer doc of this session file"},
		{write("tl-status-op.side/01-why.md", fm("parent: ../tl-status-op.md\n")), "not an answer doc of this session file"},
		{write("tl-status-op.side/2026-10-18-01-a.side/2026-10-18-02-b.md", fm("parent: ../2026-10-18-01-a.md\n")), "not an answer doc of this session file"},
		{"tl-status-op.side/2026-10-18-09-missing.md", "write the answer doc first"},
		{write("tl-status-op.side/2026-10-18-03-untyped.md", "---\nparent: ../tl-status-op.md\n---\n"), "no `type: side-doc`"},
		{write("tl-status-op.side/2026-10-18-04-plain.md", "# no frontmatter\n"), "no `type: side-doc`"},
		{write("tl-status-op.side/2026-10-18-05-other.md", fm("parent: ../other.md\n")), "must name the session file"},
		{write("tl-status-op.side/2026-10-18-06-orphan.md", fm("")), "must name the session file"},
	} {
		_, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "1", c.target, "--hash", fakeHash(opSection))
		if err == nil || !strings.Contains(err.Error(), c.want) || f.writes != 0 || f.frontWrites != 0 {
			t.Errorf("%s: err %v, writes %d/%d; want a refusal naming %q and no write", c.target, err, f.writes, f.frontWrites, c.want)
		}
	}
	// A migrated answer keeps its own process-instruction (#455 decision 3),
	// and a quoted parent reads the same.
	ok := write("tl-status-op.side/2026-10-18-07-kept.md", fm("parent: \"../tl-status-op.md\"\nprocess-instruction: \"[free-form] x\"\n"))
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "1", ok, "--hash", fakeHash(opSection)); err != nil {
		t.Errorf("a valid answer refused: %v", err)
	}
}

// #576: a @@chat@@ tag alone gives the numbered grammar; any legacy
// declaration, or none, keeps `-`, `*` and `N.` items.
func TestLegacyGrammar(t *testing.T) {
	for _, tc := range []struct {
		name, doc string
		want      bool
	}{
		{"tag alone", "# S\n\n@@chat@@\n## Operator instructions\n1. a\n", false},
		{"tag and legacy comment", "# S\n\n@@chat@@\n## Operator instructions\n<!-- collaboration-mode: chat -->\n- a\n", true},
		{"legacy comment", "# S\n\n## Operator instructions\n<!-- collaboration-mode: chat -->\n- a\n", true},
		{"legacy map", "---\ncollaboration-mode: {chat: [Operator instructions]}\n---\n# S\n\n## Operator instructions\n- a\n", true},
		{"undeclared", "# S\n\n## Operator instructions\n- a\n", true},
		{"an enclosing tagged section governs", "@@chat@@\n# Parent\n## Operator instructions\n1. a\n", false},
		// The heading comes from the parse (#576 round 3): these are all tagged.
		{"a copy of the heading inside a fence comes first", "```\n## Operator instructions\n```\n@@chat@@\n## Operator instructions\n1. a\n", false},
		{"a closing ## on the heading", "@@chat@@\n## Operator instructions ##\n1. a\n", false},
		{"a heading indented by one space", "@@chat@@\n ## Operator instructions\n1. a\n", false},
	} {
		if got := legacyGrammar(tc.doc); got != tc.want {
			t.Errorf("%s: legacyGrammar %v, want %v", tc.name, got, tc.want)
		}
	}
	if bs := sidedoc.ChatItems("1. ask\n- a note, not an ask\n2. another", legacyGrammar("@@chat@@\n## Operator instructions\n")); len(bs) != 2 {
		t.Errorf("a tagged section takes numbered items only: %d bullets", len(bs))
	}
}

// #576 through the session-sync fake: list counts bullets in the grammar the
// file declares and prints it in its token; link refuses when the grammar
// changed outside the hashed section since that list, writing nothing.
func TestInstructionsGrammarThroughTheFake(t *testing.T) {
	t.Run("tagged file: numbered items only", func(t *testing.T) {
		f := newFake(map[string]string{instructionsHeading: "1. a\n- a note, not an ask\n2. b\n"})
		f.pre = "# S\n\n@@chat@@"
		file, _ := statusFile(t, "")
		out, err := run(t, deps{ss: f}, "status", "instructions", "list", file)
		if err != nil || !strings.Contains(out, "1\t1. a") || !strings.Contains(out, "2\t2. b") || !strings.HasSuffix(out, ":numbered\n") {
			t.Fatalf("list: err %v\n%s", err, out)
		}
		tok := strings.TrimSpace(strings.TrimPrefix(out[strings.LastIndex(out, "hash: "):], "hash: "))
		target := answerDoc(t, file, "2026-10-23-01-b.md")
		if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", target, "--hash", tok); err != nil || !strings.Contains(f.sections[instructionsHeading], "2. b → [answer]("+target+")") {
			t.Fatalf("link 2: err %v\n%s", err, f.sections[instructionsHeading])
		}
	})
	t.Run("heading-comment file: legacy items", func(t *testing.T) {
		f := newFake(map[string]string{instructionsHeading: "<!-- collaboration-mode: chat -->\n- a\n1. b\n"})
		file, _ := statusFile(t, "")
		out, err := run(t, deps{ss: f}, "status", "instructions", "list", file)
		if err != nil || !strings.Contains(out, "1\t- a") || !strings.Contains(out, "2\t1. b") || !strings.HasSuffix(out, ":legacy\n") {
			t.Fatalf("list: err %v\n%s", err, out)
		}
	})
	t.Run("a tag added after list: link refuses, no write", func(t *testing.T) {
		section := "- note\n1. a\n2. b\n"
		f := newFake(map[string]string{instructionsHeading: section})
		file, _ := statusFile(t, "")
		out, err := run(t, deps{ss: f}, "status", "instructions", "list", file)
		if err != nil {
			t.Fatal(err)
		}
		tok := strings.TrimSpace(strings.TrimPrefix(out[strings.LastIndex(out, "hash: "):], "hash: "))
		f.pre = "# S\n\n@@chat@@" // outside the hashed section
		target := answerDoc(t, file, "2026-10-23-01-a.md")
		if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", target, "--hash", tok); exitCode(err) != 6 || f.writes != 0 {
			t.Errorf("exit %d, writes %d; want exit 6 and no write (position 2 was \"1. a\" when listed)", exitCode(err), f.writes)
		}
		if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", target, "--hash", fakeHash(section)); exitCode(err) != 6 || f.writes != 0 {
			t.Errorf("a bare pre-#576 hash on a tagged file: exit %d, writes %d; want exit 6", exitCode(err), f.writes)
		}
	})
}

// #581: writer and reader share one target rule, so every target link
// writes reads back as answered and every one it refuses would read open.
func TestLinkBulletAgreesWithTheReader(t *testing.T) {
	for _, target := range []string{"a.md", "s.side/2026-10/2026-10-23-01-x.md", "", "a b.md", "a(b).md", "a\tb.md", "a b.md"} {
		section := "1. why?\n"
		body, err := linkBullet(section, 1, target, false)
		written := err == nil
		if written != sidedoc.ValidAnswerTarget(target) {
			t.Errorf("%q: linkBullet wrote %v, ValidAnswerTarget %v", target, written, !written)
		}
		if !written {
			body = strings.TrimSuffix(section, "\n") + " → [answer](" + target + ")\n"
		}
		if got := sidedoc.ChatItems(body, false)[0].Answered; got != written {
			t.Errorf("%q: reads answered %v, written %v", target, got, written)
		}
	}
}

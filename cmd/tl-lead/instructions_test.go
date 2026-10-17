package main

import (
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
	if out, err := linkBullet(sec, 2, "x.side/2026-10-16-01-ci.md"); err != nil || !strings.Contains(out, "- check CI → [answer](x.side/2026-10-16-01-ci.md)") || !strings.Contains(out, "   - a sub-item\n") {
		t.Errorf("link 2: %v\n%s", err, out)
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
		got, err := linkBullet(opSection, c.pos, target)
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
	if got, err := linkBullet(opSection, 3, target); err != errAlreadyLinked || got != opSection {
		t.Errorf("an answered bullet: err %v, changed %v; want errAlreadyLinked and no change", err, got != opSection)
	}
	for _, bad := range []string{"", "qa/a b.md", "qa/a.md\n9. fake", "qa/(x).md"} {
		if _, err := linkBullet(opSection, 1, bad); err == nil {
			t.Errorf("target %q: want a refusal", bad)
		}
	}
	// #412: chat mode's answer docs live in the status file's .side/ folder.
	const side = "tl-status-op.side/2026-10-16-01-s4-cost.md"
	if got, err := linkBullet("1. what does S4 cost?", 1, side); err != nil || got != "1. what does S4 cost? → [answer]("+side+")" {
		t.Errorf("a .side/ target: %q, %v", got, err)
	}
	// CRLF: the carriage return is trimmed before the link, as rstrip() does.
	if got, _ := linkBullet("1. a\r", 1, target); got != "1. a → [answer]("+target+")" {
		t.Errorf("CRLF line: %q", got)
	}
	for _, pos := range []int{0, 6} {
		if _, err := linkBullet(opSection, pos, target); err == nil || !strings.Contains(err.Error(), "no bullet") {
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
	if !strings.HasSuffix(out, "hash: "+h+"\n") || !strings.HasPrefix(out, "1\t1. Why") {
		t.Fatalf("list output:\n%s", out)
	}
	out, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "2", "qa/op/a.md", "--hash", h)
	if err != nil || f.writes != 1 || !strings.Contains(f.sections[instructionsHeading], "1. Rename the train → [answer](qa/op/a.md)") {
		t.Fatalf("link: err %v, writes %d, out %q", err, f.writes, out)
	}
	// The section moved (this write changed it): the old hash now refuses.
	if _, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "1", "qa/op/b.md", "--hash", h); exitCode(err) != 6 || f.writes != 1 {
		t.Errorf("a stale hash: exit %d, writes %d; want exit 6 and no write", exitCode(err), f.writes)
	}
	// Already answered: no write.
	h = fakeHash(f.sections[instructionsHeading])
	if out, err = run(t, deps{ss: f}, "status", "instructions", "link", file, "2", "qa/op/c.md", "--hash", h); err != nil || f.writes != 1 || !strings.Contains(out, "already answered") {
		t.Errorf("an answered bullet: err %v, writes %d, out %q", err, f.writes, out)
	}
}

// The operator edits between the read and the write: link refuses rather than
// retrying, since the position came from the older read.
func TestInstructionsLinkNeverRetries(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	f.editBefore[1] = map[string]string{instructionsHeading: "1. A new first bullet\n" + opSection}
	file, _ := statusFile(t, "")
	_, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", "qa/op/a.md", "--hash", fakeHash(opSection))
	if exitCode(err) != 6 || f.writeCalls != 1 || f.writes != 0 {
		t.Errorf("exit %d, write calls %d, writes %d; want exit 6 after one attempt and no write", exitCode(err), f.writeCalls, f.writes)
	}
}

func TestInstructionsLinkRefusals(t *testing.T) {
	f := newFake(map[string]string{instructionsHeading: opSection})
	file, _ := statusFile(t, "")
	h := fakeHash(opSection)
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", "qa/op/a.md"); err == nil {
		t.Error("link without --hash must refuse")
	}
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "two", "qa/op/a.md", "--hash", h); err == nil {
		t.Error("a non-numeric POS must refuse")
	}
	f.holder = "operator"
	if _, err := run(t, deps{ss: f}, "status", "instructions", "link", file, "2", "qa/op/a.md", "--hash", h); exitCode(err) != 3 || f.writes != 0 {
		t.Errorf("under the operator's lock: exit %d, writes %d; want exit 3, no write", exitCode(err), f.writes)
	}
}

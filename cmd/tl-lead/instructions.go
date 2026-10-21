package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

// instructionsHeading is the operator's section of the status file: their
// numbered questions and instructions, which the TL only links answers to.
const instructionsHeading = "Operator instructions"

var (
	itemStartRE = regexp.MustCompile(`^(?:[-*]|\d+\.)\s`)
	answeredRE  = regexp.MustCompile(`→\s*(\[answer\]\([^)]*\)\s*$|done:)`)
)

// Bullet is one operator bullet. Pos is its 1-based position in the section:
// the operator reuses numbers, so the number never identifies a bullet.
type Bullet struct {
	Pos      int
	Start    int // index of its first line in the section
	Lines    []string
	Answered bool
}

// instructionBullets splits the section body into bullets, by the chat-mode
// item of xx-universal-review's collaboration-modes.md §3 (#412): a `-`, `*`
// or `N.` line at the section's shallowest list indent starts one, and it
// runs over its sub-items and continuation lines until the next such line, or
// until a blank line followed by prose indented no deeper than that indent
// (prose belongs to no bullet). Fenced code starts no bullet; a tab indents to
// the next multiple of four. Trailing blank lines are dropped. A bullet is
// answered by a link ending any of its lines outside fences, so a follow-up
// added after the link keeps it answered; `→ done:` is the older form.
func instructionBullets(section string) []Bullet {
	lines := strings.Split(section, "\n")
	fenced := make([]bool, len(lines))
	in := false
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fenced[i], in = true, !in
			continue
		}
		fenced[i] = in
	}
	start := func(i int) (int, bool) {
		if fenced[i] || !itemStartRE.MatchString(strings.TrimLeft(lines[i], " \t")) {
			return 0, false
		}
		return indentOf(lines[i]), true
	}
	shallow := -1
	for i := range lines {
		if n, ok := start(i); ok && (shallow < 0 || n < shallow) {
			shallow = n
		}
	}
	var out []Bullet
	open, blank := false, false
	for i, l := range lines {
		n, isItem := start(i)
		switch {
		case strings.TrimSpace(l) == "" && !fenced[i]:
			blank = true
			if open {
				b := &out[len(out)-1]
				b.Lines = append(b.Lines, l)
			}
			continue
		case isItem && n == shallow:
			out = append(out, Bullet{Pos: len(out) + 1, Start: i, Lines: []string{l}})
			open = true
		case open && blank && !fenced[i] && indentOf(l) <= shallow:
			open = false
		case open:
			b := &out[len(out)-1]
			b.Lines = append(b.Lines, l)
		}
		blank = false
	}
	for i := range out {
		b := &out[i]
		for len(b.Lines) > 1 && strings.TrimSpace(b.Lines[len(b.Lines)-1]) == "" {
			b.Lines = b.Lines[:len(b.Lines)-1]
		}
		for j, l := range b.Lines {
			if !fenced[b.Start+j] && answeredRE.MatchString(l) {
				b.Answered = true
			}
		}
	}
	return out
}

// indentOf is a line's leading indent in columns, a tab reaching the next
// multiple of four.
func indentOf(l string) int {
	n := 0
	for _, r := range l {
		switch r {
		case ' ':
			n++
		case '\t':
			n += 4 - n%4
		default:
			return n
		}
	}
	return n
}

// errAlreadyLinked: the bullet already carries an answer; nothing to write.
var errAlreadyLinked = errors.New("already answered")

// linkBullet appends ` → [answer](target)` to bullet pos's last non-blank
// line and returns the new section body.
func linkBullet(section string, pos int, target string) (string, error) {
	// The target lands inside the operator's section as a link: whitespace
	// could start a new line (even a fake bullet) and a paren would end the
	// link early.
	if target == "" || strings.ContainsAny(target, "()") || strings.IndexFunc(target, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("refused: TARGET %q must be a relative path with no whitespace or parentheses", target)
	}
	bs := instructionBullets(section)
	if pos < 1 || pos > len(bs) {
		return "", fmt.Errorf("refused: no bullet %d in ## %s (%d bullets)", pos, instructionsHeading, len(bs))
	}
	b := bs[pos-1]
	if b.Answered {
		return section, errAlreadyLinked
	}
	lines := strings.Split(section, "\n")
	last := b.Start + len(b.Lines) - 1
	lines[last] = strings.TrimRightFunc(lines[last], unicode.IsSpace) + " → [answer](" + target + ")"
	return strings.Join(lines, "\n"), nil
}

// renderOpen lists the open bullets as `POS<TAB>first line`, continuation
// lines tab-indented, then the section hash that a later link must name.
func renderOpen(bs []Bullet, hash string) string {
	var sb strings.Builder
	for _, b := range bs {
		if b.Answered {
			continue
		}
		fmt.Fprintf(&sb, "%d\t%s\n", b.Pos, b.Lines[0])
		for _, l := range b.Lines[1:] {
			fmt.Fprintf(&sb, "\t%s\n", l)
		}
	}
	fmt.Fprintf(&sb, "hash: %s\n", hash)
	return sb.String()
}

// linkInstruction writes one answer link, pinned to the read `list` reported.
// Unlike writeSection it never retries on stale: positions come from that
// read, and recomputing against a newer one could link a different bullet.
// The answer doc's index entry is written first and the link second (#424):
// a split leaves an entry without a link, which rerunning link completes and
// a reconcile otherwise drops, never a link without its entry, which would
// read as the operator's "reviewed".
func linkInstruction(d deps, out io.Writer, file string, pos int, target, hash string) error {
	ss := d.ss
	cur, h, err := ss.ReadSection(file, instructionsHeading)
	if err != nil {
		return refusal(file, instructionsHeading, err)
	}
	moved := exitError{6, fmt.Errorf("refused: ## %s changed since list; list again", instructionsHeading)}
	if h != hash {
		return moved
	}
	body, err := linkBullet(cur, pos, target)
	if errors.Is(err, errAlreadyLinked) {
		fmt.Fprintf(out, "bullet %d is already answered; nothing written\n", pos)
		return nil
	}
	if err != nil {
		return err
	}
	if err := answerTarget(file, target, func(rel string) ([]byte, error) {
		return os.ReadFile(filepath.Join(filepath.Dir(file), rel))
	}); err != nil {
		return err
	}
	if err := syncIndex(d, out, file, []string{target}); err != nil {
		return err
	}
	err = writeBody(body, func(path string) error { return ss.WriteSection(file, instructionsHeading, path, hash, nil) })
	if errors.Is(err, ErrStale) {
		return moved
	}
	if err != nil {
		return refusal(file, instructionsHeading, err)
	}
	fmt.Fprintf(out, "linked bullet %d → %s\n", pos, target)
	return nil
}

func newInstructionsCmd(d deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "instructions",
		Short: "Read ## Operator instructions and link answers to its bullets (by position; numbers repeat)",
	}
	list := &cobra.Command{
		Use:   "list [FILE]",
		Short: "Print the open (unanswered) bullets by position, then the section hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, hash, err := d.ss.ReadSection(args[0], instructionsHeading)
			if err != nil {
				return refusal(args[0], instructionsHeading, err)
			}
			fmt.Fprint(cmd.OutOrStdout(), renderOpen(instructionBullets(body), hash))
			return nil
		},
	}
	var hash string
	link := &cobra.Command{
		Use:   "link [FILE] POS TARGET --hash H",
		Short: "List TARGET in the discussions: index, then append ' → [answer](TARGET)' to bullet POS, if the section still has the hash list printed",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			pos, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("refused: POS %q is not a number", args[1])
			}
			return linkInstruction(d, cmd.OutOrStdout(), args[0], pos, args[2], hash)
		},
	}
	link.Flags().StringVar(&hash, "hash", "", "the section hash `list` printed (required)")
	_ = link.MarkFlagRequired("hash")
	c.AddCommand(withDefaultStatusFile(list, 1), withDefaultStatusFile(link, 3))
	return c
}

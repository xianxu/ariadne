package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// instructionsHeading is the operator's section of the status file: their
// numbered questions and instructions, which the TL only links answers to.
const instructionsHeading = "Operator instructions"

// Bullet is one operator bullet: a chat item of the shared grammar
// (pkg/sidedoc ChatItems, collaboration-modes.md §3). Pos is its 1-based
// position in the section: the operator reuses numbers, so the number never
// identifies a bullet.
type Bullet = sidedoc.ChatItem

// instructionBullets splits the section by the legacy item grammar (`-`, `*`
// and `N.` items), which older session files and migrate use.
func instructionBullets(section string) []Bullet {
	return sidedoc.ChatItems(section, true)
}

// legacyGrammar reports whether the operator's section takes the legacy item
// grammar: every form but a `@@chat@@` tag alone does, which takes numbered
// items only (#576). A session file that declares nothing keeps the legacy
// grammar it always had. doc is the whole file, peeked without moving
// Couch's diff baseline.
func legacyGrammar(doc string) bool {
	scan := sidedoc.ChatSections(doc)
	// The heading comes from the parse, so an indented heading or a closing
	// `##` is found and a copy inside a fence is not.
	h, ok := scan.Heading(2, instructionsHeading)
	if !ok {
		return true
	}
	// The section itself governs when it is chat; otherwise the innermost
	// enclosing chat section, tagged or legacy, does.
	for _, sec := range scan.Sections {
		if sec.Heading == h.Line {
			return sec.Source != "tag" || sec.Legacy
		}
	}
	if sec, ok := scan.Innermost(h.Line); ok {
		return sec.Source != "tag" || sec.Legacy
	}
	return true
}

// sectionBullets splits the operator's section by the grammar its file
// declares.
func sectionBullets(d deps, file, section string) ([]Bullet, bool, error) {
	doc, err := d.ss.Peek(file)
	if err != nil {
		return nil, false, err
	}
	legacy := legacyGrammar(doc)
	return sidedoc.ChatItems(section, legacy), legacy, nil
}

// grammarName names the item grammar in the token `list` prints, so `link`
// numbers bullets the way that list did (#576): the section hash covers only
// the section, and a tag or legacy declaration outside it can change the
// grammar between the two.
func grammarName(legacy bool) string {
	if legacy {
		return "legacy"
	}
	return "numbered"
}

// tokenName and tokenHelp are the one description of the token `list`
// prints and `link --hash` takes back, so every surface (Use, Short, the flag
// help, help.md's wording) names the same thing.
const (
	tokenName = "TOKEN"
	tokenHelp = "the HASH:GRAMMAR token `list` printed: the section hash and the item grammar its positions were counted in"
)

// listToken is what `list` prints after `hash:` and `link --hash` takes back:
// the section hash and the grammar its positions were counted in.
func listToken(hash string, legacy bool) string { return hash + ":" + grammarName(legacy) }

// parseListToken splits a `--hash` token. A bare hash, from a list before
// #576, counted positions in the legacy grammar.
func parseListToken(tok string) (hash, grammar string) {
	if h, g, ok := strings.Cut(tok, ":"); ok {
		return h, g
	}
	return tok, grammarName(true)
}

// errAlreadyLinked: the bullet already carries an answer; nothing to write.
var errAlreadyLinked = errors.New("already answered")

// linkBullet appends ` → [answer](target)` to bullet pos's last non-blank
// line, splitting the section by the given grammar, and returns the new
// section body. The ask's bytes stay a prefix of the new line, trailing
// whitespace included, and a line's `\r` stays last: couch's chat guard
// (pair#838) takes only an append at the item's end and refuses any other
// edit to the ask.
func linkBullet(section string, pos int, target string, legacy bool) (string, error) {
	// The target lands inside the operator's section as a link: whitespace
	// could start a new line (even a fake bullet) and a paren would end the
	// link early. The reader answers an item only by such a target (#581).
	if !sidedoc.ValidAnswerTarget(target) {
		return "", fmt.Errorf("refused: TARGET %q must be a relative path with no whitespace or parentheses", target)
	}
	bs := sidedoc.ChatItems(section, legacy)
	if pos < 1 || pos > len(bs) {
		return "", fmt.Errorf("refused: no bullet %d in ## %s (%d bullets)", pos, instructionsHeading, len(bs))
	}
	b := bs[pos-1]
	if b.Answered {
		return section, errAlreadyLinked
	}
	lines := strings.Split(section, "\n")
	last := b.Start + len(b.Lines) - 1
	ask := strings.TrimRight(lines[last], "\r")
	lines[last] = ask + " → [answer](" + target + ")" + lines[last][len(ask):]
	return strings.Join(lines, "\n"), nil
}

// renderOpen lists the open bullets as `POS<TAB>first line`, continuation
// lines tab-indented, then the list token (tokenHelp) a later link must name.
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
// a link without its entry would read as the operator's "reviewed". When the
// link's write is refused (couch's chat guard, a stale section, the lock),
// unlist takes the entry back out: the frontmatter's old bytes, or a
// reconcile when the frontmatter moved meanwhile. Each write is scoped: a whole-file apply would advance couch's
// diff baseline over the operator's unread edits elsewhere (#426).
func linkInstruction(d deps, out io.Writer, file string, pos int, target, hash string) error {
	ss := d.ss
	cur, h, err := ss.ReadSection(file, instructionsHeading)
	if err != nil {
		return refusal(file, instructionsHeading, err)
	}
	moved := exitError{6, fmt.Errorf("refused: ## %s changed since list; list again", instructionsHeading)}
	hash, listed := parseListToken(hash)
	if h != hash {
		return moved
	}
	doc, err := ss.Peek(file)
	if err != nil {
		return refusal(file, instructionsHeading, err)
	}
	legacy := legacyGrammar(doc)
	if grammarName(legacy) != listed {
		return exitError{6, fmt.Errorf("refused: ## %s now counts bullets in the %s grammar, list used %s (a @@chat@@ tag or legacy declaration changed); list again", instructionsHeading, grammarName(legacy), listed)}
	}
	body, err := linkBullet(cur, pos, target, legacy)
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
		err = moved
	} else if err != nil {
		err = refusal(file, instructionsHeading, err)
	}
	if err != nil {
		if rerr := unlist(d, out, file, doc, target); rerr != nil {
			fmt.Fprintf(out, "index: %s stays listed without its link (%v); the next reconcile drops it, or rerun link\n", target, rerr)
		}
		return err
	}
	fmt.Fprintf(out, "linked bullet %d → %s\n", pos, target)
	return nil
}

// unlist undoes link's index write after its link was refused. While the
// frontmatter is still what that write made of doc's, it goes back to doc's
// block byte for byte; after any other edit to it, a reconcile drops the
// entry that has no link.
func unlist(d deps, out io.Writer, file, doc, target string) error {
	before, _, _ := sidedoc.FrontBlock(doc)
	block, hash, err := d.ss.ReadFrontmatter(file)
	if err != nil {
		return err
	}
	if block == before {
		return nil
	}
	if listed, _, err := sidedoc.ReconcileIndex(doc, []string{target}); err == nil {
		if front, _, _ := sidedoc.FrontBlock(listed); block == front {
			err := writeBody(before, func(path string) error { return d.ss.WriteFrontmatter(file, path, hash) })
			if err == nil {
				fmt.Fprintf(out, "index: unlisted %s (its link was refused)\n", target)
				return nil
			}
			if !errors.Is(err, ErrStale) {
				return err
			}
		}
	}
	return syncIndex(d, out, file, nil)
}

func newInstructionsCmd(d deps) *cobra.Command {
	c := &cobra.Command{
		Use:   "instructions",
		Short: "Read ## Operator instructions and link answers to its bullets (by position; numbers repeat)",
	}
	list := &cobra.Command{
		Use:   "list [FILE]",
		Short: "Print the open (unanswered) bullets by position, then the " + tokenName + " link needs (" + tokenHelp + ")",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, hash, err := d.ss.ReadSection(args[0], instructionsHeading)
			if err != nil {
				return refusal(args[0], instructionsHeading, err)
			}
			bs, legacy, err := sectionBullets(d, args[0], body)
			if err != nil {
				return refusal(args[0], instructionsHeading, err)
			}
			fmt.Fprint(cmd.OutOrStdout(), renderOpen(bs, listToken(hash, legacy)))
			return nil
		},
	}
	var hash string
	link := &cobra.Command{
		Use:   "link [FILE] POS TARGET --hash " + tokenName,
		Short: "List TARGET in the discussions: index, then append ' → [answer](TARGET)' to bullet POS (a refused append drops the entry again), if the section and its grammar still match the " + tokenName + " list printed",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			pos, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("refused: POS %q is not a number", args[1])
			}
			return linkInstruction(d, cmd.OutOrStdout(), args[0], pos, args[2], hash)
		},
	}
	link.Flags().StringVar(&hash, "hash", "", tokenHelp+" (required)")
	_ = link.MarkFlagRequired("hash")
	c.AddCommand(withDefaultStatusFile(list, 1), withDefaultStatusFile(link, 3))
	return c
}

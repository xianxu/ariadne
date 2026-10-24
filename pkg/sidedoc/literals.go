package sidedoc

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmtext "github.com/yuin/goldmark/text"
)

// CodeLiteralMask uses actual CommonMark code nodes, so unmatched inline
// delimiters cannot hide markers across paragraph, heading or container bounds.
// Goldmark is already the repository's Markdown parser; do not duplicate its
// block/inline grammar here. Only literal code content is masked.
// The returned work count covers parser input bytes, AST visits and mask writes;
// it does not count operations inside Goldmark. Callers bound their input size.
func CodeLiteralMask(text string) ([]bool, int) {
	literal := make([]bool, len(text))
	source := []byte(text)
	root := goldmark.DefaultParser().Parse(gmtext.NewReader(source))
	work := len(source) // one parser input; internal parser operations are external
	mark := func(start, end int) {
		for i := start; i < end; i++ {
			literal[i] = true
			work++
		}
	}
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		work++
		switch node := n.(type) {
		case *ast.CodeSpan:
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				work++
				segment := child.(*ast.Text).Segment
				mark(segment.Start, segment.Stop)
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeBlock, *ast.FencedCodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				work++
				segment := lines.At(i)
				mark(segment.Start, segment.Stop)
			}
			if fenced, ok := n.(*ast.FencedCodeBlock); ok && fenced.Info != nil {
				mark(fenced.Info.Segment.Start, fenced.Info.Segment.Stop)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return literal, work
}

// CodeLines marks each line of text that belongs to a code block, fenced or
// indented, fence lines included, by CommonMark's rules: a fence closes only
// on its own character and at least its length, which a toggle on any
// fence-like line gets wrong (pair#838). Lines are text's "\n"-split lines.
func CodeLines(text string) []bool {
	code, _ := codeLines(text)
	return code
}

// codeLines is CodeLines with its work count: the lines split, the bytes the
// line cursor advances over and the lines marked; goldmark's parse is outside
// the count, as in CodeLiteralMask.
func codeLines(text string) ([]bool, int) {
	lines := strings.Split(text, "\n")
	work := len(lines)
	code := make([]bool, len(lines))
	source := []byte(text)
	// Blocks arrive in source order, so one cursor counts each newline once.
	cursor, cursorLine := 0, 0
	lineAt := func(off int) int {
		cursorLine += strings.Count(text[cursor:off], "\n")
		work += off - cursor
		cursor = off
		return cursorLine
	}
	isFence := func(i int) bool {
		t := strings.TrimSpace(lines[i])
		return strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
	}
	root := goldmark.DefaultParser().Parse(gmtext.NewReader(source))
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			segs := n.Lines()
			_, fenced := n.(*ast.FencedCodeBlock)
			if segs.Len() == 0 {
				return ast.WalkSkipChildren, nil
			}
			first, last := lineAt(segs.At(0).Start), lineAt(segs.At(segs.Len()-1).Start)
			for i := first; i <= last; i++ {
				code[i] = true
				work++
			}
			// The fence lines sit just outside the content.
			if fenced && first > 0 && isFence(first-1) {
				code[first-1] = true
			}
			if fenced && last+1 < len(lines) && isFence(last+1) {
				code[last+1] = true
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return code, work
}

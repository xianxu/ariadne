package sidedoc

import (
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

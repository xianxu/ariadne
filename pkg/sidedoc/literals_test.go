package sidedoc

import (
	"strings"
	"testing"
)

func TestCodeLiteralMaskCountedTraversal(t *testing.T) {
	for _, unit := range []string{
		"`multiline\n[s](path.md)\nend`\n\n",
		"    [s](path.md)\n\n",
		"- ```\n  [s](path.md)\n  ```\n\n",
		"> ```\n> [s](path.md)\n> ```\n\n",
		"`unmatched\n\n[s](path.md)\n\ncloser`\n\n",
	} {
		one := strings.Repeat(unit, 128)
		two := strings.Repeat(unit, 256)
		mask, n := CodeLiteralMask(one)
		maskTwice, twice := CodeLiteralMask(two)
		if len(mask) != len(one) || len(maskTwice) != len(two) {
			t.Fatal("mask must use source byte coordinates")
		}
		// Counts are our parser input/AST traversal/mask writes, never a
		// claim about operations internal to the Goldmark dependency.
		if twice > 2*n+16 || twice > 4*len(two)+16 {
			t.Fatalf("nonlinear owned traversal: N=%d 2N=%d bytes=%d", n, twice, len(two))
		}
	}
}

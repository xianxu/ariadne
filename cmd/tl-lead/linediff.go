package main

import "strings"

// lineDiff renders a minimal line diff of a against b: "-" for a line only in
// a, "+" for one only in b, " " for a shared line. Inputs are status files (about
// a screen), so the quadratic LCS table is fine.
func lineDiff(a, b string) string {
	x, y := strings.Split(a, "\n"), strings.Split(b, "\n")
	lcs := make([][]int, len(x)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(y)+1)
	}
	for i := len(x) - 1; i >= 0; i-- {
		for j := len(y) - 1; j >= 0; j-- {
			if x[i] == y[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var sb strings.Builder
	i, j := 0, 0
	for i < len(x) || j < len(y) {
		switch {
		case i < len(x) && j < len(y) && x[i] == y[j]:
			sb.WriteString(" " + x[i] + "\n")
			i, j = i+1, j+1
		case j < len(y) && (i == len(x) || lcs[i][j+1] > lcs[i+1][j]):
			sb.WriteString("+" + y[j] + "\n")
			j++
		default:
			sb.WriteString("-" + x[i] + "\n")
			i++
		}
	}
	return sb.String()
}

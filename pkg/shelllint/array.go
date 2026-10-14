package shelllint

import (
	"regexp"
	"strings"
)

const reasonArray = `array expansion under set -u: macOS /bin/bash 3.2 treats an empty array as ` +
	`unbound and exits; use ${a[@]+"${a[@]}"} (or a default, ${a[@]:-…}) (#353)`

// nounset is a set command that turns on -u.
var nounset = regexp.MustCompile(`(^|[\s;&|])set\s+(-[a-zA-Z]*u[a-zA-Z]*|-o\s+nounset)(\s|$)`)

// ArrayFindings reports, in a script that enables nounset anywhere, each
// ${name[@]} or ${name[*]} expansion that bash 3.2 rejects for an empty array:
// the bare form, and the bash-4 case operators (^ ,). Measured on macOS
// /bin/bash 3.2.57 (#353): substring (:1), pattern removal (# %), replacement
// (/) and the default/alternate operators do not fail. It does not judge
// whether an array can be empty: guard it, ${a[@]+"${a[@]}"}. A count
// (${#a[@]}) is safe. Single-quoted text and comments are not expansions.
func ArrayFindings(src string) []Finding {
	lines := strings.Split(src, "\n")
	parsed := make([]lexed, len(lines))
	strict := false
	for i, raw := range lines {
		parsed[i] = lex(raw)
		if nounset.MatchString(parsed[i].masked()) {
			strict = true
		}
	}
	if !strict {
		return nil
	}
	var out []Finding
	for i, l := range parsed {
		for range unguardedArrays(l.expanding()) {
			out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(lines[i]), Reason: reasonArray})
		}
	}
	return out
}

// unguardedArrays returns the offsets of unsafe array expansions in text. A
// guarded expansion is skipped whole, including the array it guards.
func unguardedArrays(text string) []int {
	var at []int
	for i := 0; i < len(text); {
		j := strings.Index(text[i:], "${")
		if j < 0 {
			break
		}
		j += i
		k := j + 2
		m := k
		for m < len(text) && (text[m] == '_' || isAlpha(text[m]) || (m > k && isDigit(text[m]))) {
			m++
		}
		if m == k || !(strings.HasPrefix(text[m:], "[@]") || strings.HasPrefix(text[m:], "[*]")) {
			i = k // a scalar, a count (${#…}) or not an expansion
			continue
		}
		op := text[m+3:]
		switch {
		case strings.HasPrefix(op, "+") || strings.HasPrefix(op, ":+"):
			i = closingBrace(text, j) // a guard: skip it whole
		case hasAnyPrefix(op, "}", "^", ","):
			at = append(at, j)
			i = m + 3
		default:
			i = m + 3
		}
	}
	return at
}

// closingBrace is the offset just past the } closing the ${ at open.
func closingBrace(text string, open int) int {
	depth := 0
	for i := open; i < len(text); i++ {
		switch {
		case strings.HasPrefix(text[i:], "${"):
			depth++
			i++
		case text[i] == '}':
			if depth--; depth == 0 {
				return i + 1
			}
		}
	}
	return len(text)
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func isAlpha(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

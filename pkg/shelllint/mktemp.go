// Package shelllint holds pure checks over shell source (scripts and Makefile
// recipes) that agents and tooling run. It reads text only; the repository
// test that applies it lives beside it.
package shelllint

import (
	"regexp"
	"strings"
)

// Finding is one unsafe construct on a line (1-based).
type Finding struct {
	Line   int
	Text   string
	Reason string
}

const (
	reasonTemplate = `mktemp without a TMPDIR-rooted template: macOS ignores TMPDIR without one ` +
		`and writes under /var/folders, which the sandbox denies; use mktemp -d "${TMPDIR:-/tmp}/name.XXXXXX" (#342)`
	reasonCd = `cd or pushd into a mktemp substitution: when mktemp fails, cd "" succeeds and stays in the ` +
		`current directory; assign it, check it (|| exit 1), then cd "${d:?}" (#342)`
)

// cdIntoMktemp is cd or pushd whose argument is a $(...) substitution running
// mktemp.
var cdIntoMktemp = regexp.MustCompile(`(^|[\s;(&|])(cd|pushd)\s+"?\$\([^)]*\bmktemp\b`)

// MktempFindings reports each mktemp call without a template argument that
// contains XXX and is rooted at a variable (${TMPDIR:-/tmp}/…, or a directory
// already made from it), and each cd into a mktemp substitution. Makefile `$$`
// escapes are read as `$`; comments and words inside strings are not code.
func MktempFindings(src string) []Finding {
	var out []Finding
	for i, line := range strings.Split(src, "\n") {
		code := strings.ReplaceAll(stripComment(line), "$$", "$")
		for _, at := range commandWords(code, "mktemp") {
			if !templated(arguments(code[at+len("mktemp"):])) {
				out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(line), Reason: reasonTemplate})
			}
		}
		if cdIntoMktemp.MatchString(code) {
			out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(line), Reason: reasonCd})
		}
	}
	return out
}

// stripComment drops a # comment: a # at the start of a word, outside quotes.
func stripComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else if c == '\\' && quote == '"' {
				i++
			}
		case c == '\'' || c == '"':
			quote = c
		case c == '\\':
			i++
		case c == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t' || line[i-1] == ';'):
			return line[:i]
		}
	}
	return line
}

// commandWords returns the offsets of word where it runs as a command. The
// rule is inverted rather than enumerating what may precede a command (if,
// sudo -u x, time -p, VAR=x, …): every standalone, unquoted occurrence in code
// counts, including inside $(…) within double quotes. Literal quoted text, a
// part of a larger word, a $word or ${word} expansion, and a lookup that only
// names the command (command -v, which, type, hash) do not.
func commandWords(code, word string) []int {
	masked := codeOnly(code)
	var at []int
	for i := 0; ; {
		j := strings.Index(masked[i:], word)
		if j < 0 {
			return at
		}
		j += i
		end := j + len(word)
		standalone := (j == 0 || !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$.{/-", rune(masked[j-1]))) &&
			(end == len(masked) || strings.ContainsRune(" \t);|&`", rune(masked[end])))
		if standalone && !lookup(strings.Fields(masked[:j])) {
			at = append(at, j)
		}
		i = end
	}
}

// lookup reports whether the words before an occurrence make it an argument
// that names a command rather than a call.
func lookup(before []string) bool {
	n := len(before)
	if n == 0 {
		return false
	}
	switch word(before[n-1]) {
	case "which", "type", "hash":
		return true
	case "-v", "-V":
		return n >= 2 && word(before[n-2]) == "command"
	}
	return false
}

// word is a field without what opened it: the text after its last ( or `.
func word(field string) string {
	return field[strings.LastIndexAny(field, "(`")+1:]
}

// codeOnly blanks quoted literal text, keeping command substitutions (which
// run code even inside double quotes), so offsets match the input.
func codeOnly(line string) string {
	out := []byte(line)
	type ctx struct {
		kind  byte // 0 code, '"' or '\''
		depth int  // parentheses inside a $( substitution
	}
	stack := []ctx{{}}
	for i := 0; i < len(line); i++ {
		top := &stack[len(stack)-1]
		c := line[i]
		switch top.kind {
		case '\'':
			if c == '\'' {
				stack = stack[:len(stack)-1]
			} else {
				out[i] = ' '
			}
		case '"':
			switch {
			case c == '"':
				stack = stack[:len(stack)-1]
			case c == '\\':
				out[i] = ' '
				if i+1 < len(line) {
					i++
					out[i] = ' '
				}
			case c == '$' && i+1 < len(line) && line[i+1] == '(':
				stack = append(stack, ctx{depth: 1})
				i++
			default:
				out[i] = ' '
			}
		default:
			switch {
			case c == '\'' || c == '"':
				stack = append(stack, ctx{kind: c})
			case c == '\\':
				i++
			case c == '(':
				top.depth++
			case c == ')':
				if top.depth > 0 {
					top.depth--
					if top.depth == 0 && len(stack) > 1 {
						stack = stack[:len(stack)-1]
					}
				}
			}
		}
	}
	return string(out)
}

// arguments splits what follows a command word up to the end of that command
// (an unquoted ) ; | & or backquote), removing quotes.
func arguments(rest string) []string {
	var args []string
	var cur strings.Builder
	var quote byte
	flush := func() {
		if cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '\'' || c == '"':
			quote = c
		case strings.IndexByte(");|&`", c) >= 0:
			flush()
			return args
		case c == ' ' || c == '\t':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return args
}

// templated reports whether an argument is a template rooted at a variable.
// An option's value (-t prefix) does not count: macOS ignores TMPDIR for it.
func templated(args []string) bool {
	for i, a := range args {
		if strings.HasPrefix(a, "-") || (i > 0 && args[i-1] == "-t") {
			continue
		}
		if strings.Contains(a, "XXX") && strings.HasPrefix(a, "$") {
			return true
		}
	}
	return false
}

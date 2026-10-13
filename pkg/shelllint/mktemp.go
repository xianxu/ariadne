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
// already made from it), and each cd or pushd into a mktemp substitution.
// Makefile `$$` escapes are read as `$`; comments and quoted text are not code.
func MktempFindings(src string) []Finding {
	var out []Finding
	for i, raw := range strings.Split(src, "\n") {
		l := lex(strings.ReplaceAll(raw, "$$", "$"))
		for _, at := range commandWords(l, "mktemp") {
			if !templated(l.arguments(at + len("mktemp"))) {
				out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(raw), Reason: reasonTemplate})
			}
		}
		if cdIntoMktemp.MatchString(l.text) {
			out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(raw), Reason: reasonCd})
		}
	}
	return out
}

// class is what one byte of a line is to the shell.
type class byte

const (
	code   class = iota // runs: commands, operators, substitutions
	quoted              // literal text inside quotes
	delim               // a quote character opening or closing quoted text
)

// lexed is one line, cut at its comment, with each byte classified. It is the
// one place quoting is understood: masking, comment and argument splitting all
// read it.
type lexed struct {
	text string
	cls  []class
}

// lex classifies line's bytes and drops its comment (a # at the start of a
// word, in code). Command substitutions, $(…) and `…`, are code even inside
// double quotes.
func lex(line string) lexed {
	cls := make([]class, len(line))
	type ctx struct {
		kind  byte // 0 top-level code, '(' a $( substitution, '`', '"', '\''
		depth int  // open parentheses inside a $( substitution
	}
	stack := []ctx{{}}
	for i := 0; i < len(line); i++ {
		top := &stack[len(stack)-1]
		c := line[i]
		switch top.kind {
		case '\'':
			cls[i] = quoted
			if c == '\'' {
				cls[i] = delim
				stack = stack[:len(stack)-1]
			}
		case '"':
			switch {
			case c == '"':
				cls[i] = delim
				stack = stack[:len(stack)-1]
			case c == '$' && i+1 < len(line) && line[i+1] == '(':
				stack = append(stack, ctx{kind: '(', depth: 1})
				i++ // both bytes are code
			case c == '`':
				stack = append(stack, ctx{kind: '`'})
			case c == '\\' && i+1 < len(line):
				cls[i], cls[i+1] = quoted, quoted
				i++
			default:
				cls[i] = quoted
			}
		default: // code: top level or inside a substitution
			switch {
			case c == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t' || line[i-1] == ';'):
				return lexed{text: line[:i], cls: cls[:i]}
			case c == '\'' || c == '"':
				cls[i] = delim
				stack = append(stack, ctx{kind: c})
			case c == '\\':
				i++ // the escaped byte stays code
			case c == '`' && top.kind == '`':
				stack = stack[:len(stack)-1]
			case c == '`':
				stack = append(stack, ctx{kind: '`'})
			case c == '$' && i+1 < len(line) && line[i+1] == '(':
				stack = append(stack, ctx{kind: '(', depth: 1})
				i++
			case c == '(' && top.kind == '(':
				top.depth++
			case c == ')' && top.kind == '(':
				if top.depth--; top.depth == 0 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}
	return lexed{text: line, cls: cls}
}

// masked is the code with quoted text and quote delimiters blanked, so
// offsets match text.
func (l lexed) masked() string {
	b := []byte(l.text)
	for i, c := range l.cls {
		if c != code {
			b[i] = ' '
		}
	}
	return string(b)
}

// arguments splits what follows offset up to the end of that command (an
// unquoted ) ; | & or backquote), with quotes removed.
func (l lexed) arguments(from int) []string {
	var args []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
		}
	}
	for i := from; i < len(l.text); i++ {
		c := l.text[i]
		switch l.cls[i] {
		case delim:
		case quoted:
			cur.WriteByte(c)
		default:
			switch {
			case strings.IndexByte(");|&`", c) >= 0:
				flush()
				return args
			case c == ' ' || c == '\t':
				flush()
			default:
				cur.WriteByte(c)
			}
		}
	}
	flush()
	return args
}

// commandWords returns the offsets of word where it runs as a command. The
// rule is inverted rather than enumerating what may precede a command (if,
// sudo -u x, time -p, VAR=x, …): every standalone occurrence in code counts,
// including inside a substitution within double quotes. Quoted text, a part
// of a larger word, a $word or ${word} expansion, and a lookup that only
// names the command (command -v, which, type, hash) do not.
func commandWords(l lexed, word string) []int {
	masked := l.masked()
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

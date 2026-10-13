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

// assignment is a NAME=value word, which may prefix a command.
var assignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

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

// commandWords returns the offsets where word starts a command: after the
// start of the line, an opening substitution or subshell, a separator, a brace
// group, a keyword or prefix command (sudo, time, …), or an env assignment. A
// word inside a string, an identifier or a ${…} expansion is not one.
func commandWords(code, word string) []int {
	var at []int
	for i := 0; ; {
		j := strings.Index(code[i:], word)
		if j < 0 {
			return at
		}
		j += i
		end := j + len(word)
		if (end == len(code) || strings.ContainsRune(" \t);|&", rune(code[end]))) && commandPosition(code[:j]) {
			at = append(at, j)
		}
		i = end
	}
}

func commandPosition(before string) bool {
	trimmed := strings.TrimRight(before, " \t")
	if trimmed == "" {
		return true
	}
	if strings.ContainsRune("(`;|&", rune(trimmed[len(trimmed)-1])) {
		return true
	}
	fields := strings.Fields(trimmed)
	last := fields[len(fields)-1]
	switch last {
	case "{", "then", "do", "else", "!", "exec", "command", "sudo", "time", "env", "nohup", "nice":
		return true
	}
	return assignment.MatchString(last)
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

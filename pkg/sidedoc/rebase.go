package sidedoc

import (
	"path"
	"regexp"
	"strings"
)

var (
	// An inline link or image target: `](target)` or `](target "title")`.
	inlineRE = regexp.MustCompile(`\]\(([^()\s]+)(\s+"[^"\n]*")?\)`)
	// A reference definition: `[label]: target`.
	refDefRE = regexp.MustCompile(`^(\s{0,3}\[[^\]\n]+\]:\s*)(\S+)`)
	// A dated answer doc (collaboration-modes.md §4): `YYYY-MM-DD-NN-slug.md`,
	// the number two digits or wider, the slug lowercase words of at most 40.
	answerRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{2,}-([a-z0-9]+(?:-[a-z0-9]+)*)\.md$`)
)

// AnswerName reports whether a file name is a dated answer doc's.
func AnswerName(name string) bool {
	m := answerRE.FindStringSubmatch(name)
	return m != nil && len(m[1]) <= 40
}

// MoveUnder maps a slash path under fromRoot to its place under toRoot. ok is
// false for a path outside fromRoot.
func MoveUnder(p, fromRoot, toRoot string) (string, bool) {
	rel, ok := strings.CutPrefix(p, fromRoot+"/")
	if !ok || rel == "" {
		return "", false
	}
	return toRoot + "/" + rel, true
}

// RebaseLinks keeps a document's relative links resolving when it moves from
// `from` to `to`, or when what it links moves (collaboration-modes.md §5).
// Paths are slash paths relative to one root. moved maps a root-relative
// target to its new place; exists reports whether a root-relative path is
// there. It rewrites inline link and image targets, reference definitions
// and the frontmatter's `parent:`, keeping any #fragment. It leaves alone
// URLs, absolute paths, targets that leave the root, fenced code and code
// spans. A target found under neither its old nor its new name is returned in
// dangling, as written, and left unchanged. Nothing changes when nothing moved.
func RebaseLinks(text, from, to string, moved func(string) (string, bool), exists func(string) bool) (out string, dangling []string) {
	rebase := func(target string) string {
		p, frag, hasFrag := strings.Cut(target, "#")
		if hasFrag {
			frag = "#" + frag
		}
		if p == "" || strings.HasPrefix(p, "/") || strings.Contains(p, ":") {
			return target // a fragment, an absolute path, a URL or scheme
		}
		at := path.Join(path.Dir(from), p)
		if at == ".." || strings.HasPrefix(at, "../") {
			return target
		}
		next, isMoved := moved(at)
		if !isMoved {
			next = at
		}
		if !exists(at) && !exists(next) {
			dangling = append(dangling, target)
			return target
		}
		if !isMoved && from == to {
			return target
		}
		rel := RelPath(path.Dir(to), next)
		if strings.HasSuffix(p, "/") {
			rel += "/"
		}
		if rel == path.Clean(p) || rel+"/" == p {
			return target // the same place, spelled as the author spelled it
		}
		return rel + frag
	}
	literal, _ := CodeLiteralMask(text)
	lines := strings.Split(text, "\n")
	front := frontLines(lines)
	offset := 0
	for i, line := range lines {
		if i > 0 && i < front {
			lines[i] = rebaseParent(line, rebase)
		} else if i >= front {
			lines[i] = rebaseLine(line, literal[offset:offset+len(line)], rebase)
		}
		offset += len(line) + 1
	}
	return strings.Join(lines, "\n"), dangling
}

// frontLines is the number of lines the leading frontmatter block spans,
// both `---` lines included; 0 when there is none.
func frontLines(lines []string) int {
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			return i + 1
		}
	}
	return 0
}

// rebaseParent rewrites a `parent:` frontmatter line's value, plain or
// double-quoted.
func rebaseParent(line string, rebase func(string) string) string {
	v, ok := strings.CutPrefix(line, "parent:")
	if !ok {
		return line
	}
	body := strings.TrimRight(v, "\r")
	cr := v[len(body):]
	val := strings.TrimSpace(body)
	quoted := len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"'
	if quoted {
		val = val[1 : len(val)-1]
	}
	if val == "" || strings.ContainsAny(val, `"\ `) {
		return line
	}
	next := rebase(val)
	if next == val {
		return line
	}
	if quoted {
		next = `"` + next + `"`
	}
	return "parent: " + next + cr
}

// rebaseLine rewrites link targets outside the shared document code mask.
func rebaseLine(line string, literal []bool, rebase func(string) string) string {
	inCode := func(at int) bool { return literal[at] }
	var b strings.Builder
	last := 0
	replace := func(start, end int) {
		if start < last {
			return // overlaps a target already rewritten: one link, read once
		}
		target := line[start:end]
		if next := rebase(target); next != target {
			b.WriteString(line[last:start])
			b.WriteString(next)
			last = end
		}
	}
	if m := refDefRE.FindStringSubmatchIndex(line); m != nil && !inCode(m[4]) {
		replace(m[4], m[5])
	}
	for _, m := range inlineRE.FindAllStringSubmatchIndex(line, -1) {
		if !inCode(m[0]) {
			replace(m[2], m[3])
		}
	}
	if last == 0 {
		return line
	}
	b.WriteString(line[last:])
	return b.String()
}

// RelPath is the slash path from directory dir to target, both relative to
// the same root.
func RelPath(dir, target string) string {
	split := func(p string) []string {
		if p == "." || p == "" {
			return nil
		}
		return strings.Split(p, "/")
	}
	d, t := split(path.Clean(dir)), split(path.Clean(target))
	i := 0
	for i < len(d) && i < len(t) && d[i] == t[i] {
		i++
	}
	parts := make([]string, 0, len(d)-i+len(t)-i)
	for range d[i:] {
		parts = append(parts, "..")
	}
	parts = append(parts, t[i:]...)
	if len(parts) == 0 {
		return "."
	}
	return strings.Join(parts, "/")
}

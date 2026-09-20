package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

const ignoreBegin = "# BEGIN weave generated"
const ignoreEnd = "# END weave generated"

// managedIgnoreText puts the owned block before authored rules, allowing local
// negations to override it. Only exact entries from the old fixed list migrate.
func managedIgnoreText(current string, entries []string) (string, error) {
	legacy := map[string]bool{}
	for _, e := range GeneratedRuntimeGitignoreEntries {
		legacy[e] = true
	}
	var kept strings.Builder
	inside, seen := false, false
	for _, line := range strings.SplitAfter(current, "\n") {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		switch trimmed {
		case ignoreBegin:
			if inside || seen {
				return "", fmt.Errorf("malformed weave gitignore block")
			}
			inside = true
			seen = true
		case ignoreEnd:
			if !inside {
				return "", fmt.Errorf("malformed weave gitignore block")
			}
			inside = false
		default:
			if !inside && !legacy[trimmed] {
				kept.WriteString(line)
			}
		}
	}
	if inside {
		return "", fmt.Errorf("unterminated weave gitignore block")
	}
	set := map[string]bool{}
	for _, e := range entries {
		set[e] = true
	}
	sorted := make([]string, 0, len(set))
	for e := range set {
		sorted = append(sorted, e)
	}
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return kept.String(), nil
	}
	return ignoreBegin + "\n" + strings.Join(sorted, "\n") + "\n" + ignoreEnd + "\n" + kept.String(), nil
}

// escapeIgnore quotes gitignore metacharacters so an output pathname cannot
// turn into a rule covering unrelated authored paths.
func escapeIgnore(path string) string {
	r := strings.NewReplacer("\\", "\\\\", "*", "\\*", "?", "\\?", "[", "\\[", "]", "\\]", " ", "\\ ")
	return "/" + r.Replace(filepath.ToSlash(path))
}
func managedIgnore(fs weavefs.FS, root string, ids []outputIdentity) (string, error) {
	p := filepath.Join(root, ".gitignore")
	if fi, e := fs.Lstat(p); e == nil && !fi.Mode().IsRegular() {
		return "", fmt.Errorf("gitignore is not a regular file: %s", p)
	} else if e != nil && !os.IsNotExist(e) {
		return "", e
	}
	b, e := fs.ReadFile(p)
	if e != nil && !os.IsNotExist(e) {
		return "", e
	}
	entries := []string{escapeIgnore(filepath.Dir(InventoryPath)) + "/"}
	for _, id := range ids {
		entries = append(entries, escapeIgnore(id.Path))
	}
	return managedIgnoreText(string(b), entries)
}
func writeManagedIgnore(fs weavefs.FS, root, next string) error {
	p := filepath.Join(root, ".gitignore")
	current, e := fs.ReadFile(p)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if string(current) == next {
		return nil
	}
	return fs.WriteFileAtomic(p, []byte(next))
}

// GeneratedGitignoreEntries derives exact output paths for dry-run/planning.
// Scaffolds, touches and seeds remain authored entrypoints, never broad ignores.
func GeneratedGitignoreEntries(actions []Action) []string {
	set := map[string]bool{escapeIgnore(filepath.Dir(InventoryPath)) + "/": true}
	for _, a := range actions {
		var path string
		switch a := a.(type) {
		case Symlink:
			path = a.Dst
		case WriteFile:
			path = a.Path
		case MergeSettings:
			path = a.Target
		default:
			continue
		}
		set[escapeIgnore(path)] = true
	}
	out := make([]string, 0, len(set))
	for path := range set {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

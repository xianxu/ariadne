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

// splitIgnore separates the weave block's entries from the authored rules
// around it. With migrate, exact entries from the old fixed list, which lived
// outside the block, are dropped; only the artifacts pass, which records their
// replacement, migrates (#264).
func splitIgnore(current string, migrate bool) (string, []string, error) {
	legacy := map[string]bool{}
	if migrate {
		for _, e := range GeneratedRuntimeGitignoreEntries {
			legacy[e] = true
		}
	}
	var kept strings.Builder
	var block []string
	inside, seen := false, false
	for _, line := range strings.SplitAfter(current, "\n") {
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		switch trimmed {
		case ignoreBegin:
			if inside || seen {
				return "", nil, fmt.Errorf("malformed weave generated block")
			}
			inside = true
			seen = true
		case ignoreEnd:
			if !inside {
				return "", nil, fmt.Errorf("malformed weave generated block")
			}
			inside = false
		default:
			if inside {
				if trimmed != "" {
					block = append(block, trimmed)
				}
			} else if !legacy[trimmed] {
				kept.WriteString(line)
			}
		}
	}
	if inside {
		return "", nil, fmt.Errorf("unterminated weave generated block")
	}
	return kept.String(), block, nil
}

// managedIgnoreText puts the owned block before authored rules, allowing local
// negations to override it.
func managedIgnoreText(current string, entries []string, migrate bool) (string, error) {
	kept, _, err := splitIgnore(current, migrate)
	if err != nil {
		return "", err
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
		return kept, nil
	}
	return ignoreBegin + "\n" + strings.Join(sorted, "\n") + "\n" + ignoreEnd + "\n" + kept, nil
}

// escapeIgnore quotes gitignore metacharacters so an output pathname cannot
// turn into a rule covering unrelated authored paths.
func escapeIgnore(path string) string {
	r := strings.NewReplacer("\\", "\\\\", "*", "\\*", "?", "\\?", "[", "\\[", "]", "\\]", " ", "\\ ")
	return "/" + r.Replace(filepath.ToSlash(path))
}

// managedIgnore derives the block from the next inventory. An existing entry is
// dropped only when weave owned that path (it appears in the old inventory);
// any other entry was committed by another checkout's compile and survives
// (#263) — a dependency never compiled locally has an empty inventory, not an
// empty set of generated outputs.
func managedIgnore(fs weavefs.FS, root string, old, ids []outputIdentity, migrate bool) (string, error) {
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
	owned := map[string]bool{}
	for _, id := range old {
		owned[escapeIgnore(id.Path)] = true
	}
	_, block, e := splitIgnore(string(b), migrate)
	if e != nil {
		return "", e
	}
	for _, entry := range block {
		if !owned[entry] {
			entries = append(entries, entry)
		}
	}
	for _, id := range ids {
		entries = append(entries, escapeIgnore(id.Path))
	}
	return managedIgnoreText(string(b), entries, migrate)
}
func writeManagedIgnore(fs weavefs.FS, root, next string) error {
	return writeManagedFile(fs, root, ".gitignore", next)
}

// writeManagedFile publishes a managed-block file at the repository root, skipping
// the write when its content already matches.
func writeManagedFile(fs weavefs.FS, root, name, next string) error {
	p := filepath.Join(root, name)
	current, e := fs.ReadFile(p)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if string(current) == next {
		return nil
	}
	return weavefs.Publish(fs, root, p, []byte(next), nil)
}

// GeneratedGitignoreEntries derives exact output paths for dry-run/planning.
// Scaffolds, touches and seeds remain authored entrypoints, never broad ignores.
func GeneratedGitignoreEntries(actions []Action) []string {
	set := map[string]bool{escapeIgnore(filepath.Dir(InventoryPath)) + "/": true}
	for _, a := range actions {
		var path string
		switch a := outputAction(a).(type) {
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

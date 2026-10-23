// Package reviewexport builds the universal review skill's export manifest
// (#399): the versioned list of the files parley snapshots for its standalone
// copy of the review protocol. The skill directory is the one authored source;
// the manifest names each member with its size and hash, so a consumer can
// tell a current snapshot from a stale one. Building is pure over an fs.FS.
package reviewexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

// Schema and Version identify the manifest format. A consumer refuses any
// other pair, so a change to the member shape bumps Version. Version 4
// replaces the mode role and its order/scope/frontier flags with flagless
// briefs (#575: one collaborative mode; scope comes from the instruction).
// Version 3 retired the mode-level deletions field (#437). Version 2 added
// the fixture role (#424).
const (
	Schema  = "ariadne/universal-review-export"
	Version = 4
)

// GrammarSource is the repository path of the one authored marker grammar,
// which the skill directory links as review-convention.md.
const GrammarSource = "construct/local/fix/review-convention.md"

// Manifest is export.json.
type Manifest struct {
	Schema  string   `json:"schema"`
	Version int      `json:"version"`
	Members []Member `json:"members"`
}

// Member is one exported file. Path is relative to the skill directory. Name
// is set for adapters, briefs and fixtures; Source only for the grammar.
type Member struct {
	Role   string `json:"role"`
	Name   string `json:"name,omitempty"`
	Path   string `json:"path"`
	Source string `json:"source,omitempty"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

var roleRank = map[string]int{"round": 0, "grammar": 1, "collaboration-modes": 2, "adapter": 3, "brief": 4, "fixture": 5}

// exportedCases are the examples, besides the move cases, that the export
// carries as fixtures for pair and parley to pin (#575): the one list, read
// by Build and by TestExportCurrent.
var exportedCases = []string{
	"chat-tag", "chat-tag-misplaced",
	"ask-question", "ask-command", "ask-command-unclear", "ask-unsure",
	"scope-paragraph", "scope-section", "scope-document",
	"legacy-chat-scalar", "legacy-chat-map", "legacy-chat-comment", "legacy-keys-ignored",
}

// Brief is a parsed briefs/<name>.md: a named brief the agent may load when
// an instruction calls for it (#575). Its frontmatter is the name alone.
type Brief struct {
	Name string
}

// retiredFlags are v3 mode keys a brief must not carry: leaving one in would
// suggest a setting the round no longer reads.
var retiredFlags = []string{"order", "scope", "frontier", "deletions"}

// ParseBrief reads a brief's frontmatter by parley's rules: a leading `---\n`
// fence (so CRLF files are refused, as parley refuses them) and flat
// `key: value` lines with values taken literally (quotes included). The only
// key is `name`, which must equal the file's basename; a retired mode flag or
// any other key is refused.
func ParseBrief(content, basename string) (Brief, error) {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return Brief{}, fmt.Errorf("brief %s: missing frontmatter", basename)
	}
	// Searching from the opening fence's newline, as parley does, lets an
	// empty block (`---\n---`) parse, and fail later for want of a name.
	fm, _, ok := strings.Cut("\n"+rest, "\n---")
	if !ok {
		return Brief{}, fmt.Errorf("brief %s: unterminated frontmatter", basename)
	}
	b := Brief{}
	for _, line := range strings.Split(fm, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch k = strings.TrimSpace(k); {
		case k == "name":
			b.Name = strings.TrimSpace(v)
		case slices.Contains(retiredFlags, k):
			return Brief{}, fmt.Errorf("brief %s: %s is a retired mode flag (#575: the instruction sets the scope)", basename, k)
		default:
			return Brief{}, fmt.Errorf("brief %s: unknown frontmatter key %q; a brief has only name", basename, k)
		}
	}
	if b.Name != basename {
		return Brief{}, fmt.Errorf("brief %s: name %q must equal the file's basename", basename, b.Name)
	}
	return b, nil
}

// Build reads the skill directory and returns its manifest: round.md, the
// grammar link, collaboration-modes.md (#412), every adapters/*.md, every
// briefs/*.md, and every file of each move case and each exportedCases case
// under examples/ (role fixture, named by its case). The other examples are
// not exported. Sizes and hashes cover each file's resolved contents (the
// grammar is a symlink). The order is total: role, then name, then path.
func Build(skill fs.FS) (Manifest, error) {
	m := Manifest{Schema: Schema, Version: Version}
	add := func(mem Member) error {
		b, err := fs.ReadFile(skill, mem.Path)
		if err != nil {
			return fmt.Errorf("export member %s: %w", mem.Path, err)
		}
		sum := sha256.Sum256(b)
		mem.Bytes, mem.SHA256 = len(b), hex.EncodeToString(sum[:])
		m.Members = append(m.Members, mem)
		return nil
	}
	if err := add(Member{Role: "round", Path: "round.md"}); err != nil {
		return Manifest{}, err
	}
	if err := add(Member{Role: "grammar", Path: "review-convention.md", Source: GrammarSource}); err != nil {
		return Manifest{}, err
	}
	if err := add(Member{Role: "collaboration-modes", Path: "collaboration-modes.md"}); err != nil {
		return Manifest{}, err
	}
	adapters, err := fs.Glob(skill, "adapters/*.md")
	if err != nil {
		return Manifest{}, err
	}
	if len(adapters) == 0 {
		return Manifest{}, fmt.Errorf("export: no adapters/*.md")
	}
	for _, p := range adapters {
		if err := add(Member{Role: "adapter", Name: strings.TrimSuffix(path.Base(p), ".md"), Path: p}); err != nil {
			return Manifest{}, err
		}
	}
	briefs, err := fs.Glob(skill, "briefs/*.md")
	if err != nil {
		return Manifest{}, err
	}
	if len(briefs) == 0 {
		return Manifest{}, fmt.Errorf("export: no briefs/*.md")
	}
	for _, p := range briefs {
		b, err := fs.ReadFile(skill, p)
		if err != nil {
			return Manifest{}, err
		}
		br, err := ParseBrief(string(b), strings.TrimSuffix(path.Base(p), ".md"))
		if err != nil {
			return Manifest{}, err
		}
		if err := add(Member{Role: "brief", Name: br.Name, Path: p}); err != nil {
			return Manifest{}, err
		}
	}
	cases, err := fs.Glob(skill, "examples/*/case.md")
	if err != nil {
		return Manifest{}, err
	}
	for _, c := range cases {
		b, err := fs.ReadFile(skill, c)
		if err != nil {
			return Manifest{}, err
		}
		dir := path.Dir(c)
		if _, _, move := ParseCase(string(b)); move == nil && !slices.Contains(exportedCases, path.Base(dir)) {
			continue
		}
		err = fs.WalkDir(skill, dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			return add(Member{Role: "fixture", Name: path.Base(dir), Path: p})
		})
		if err != nil {
			return Manifest{}, err
		}
	}
	slices.SortFunc(m.Members, func(a, b Member) int {
		if d := roleRank[a.Role] - roleRank[b.Role]; d != 0 {
			return d
		}
		if d := strings.Compare(a.Name, b.Name); d != 0 {
			return d
		}
		return strings.Compare(a.Path, b.Path)
	})
	return m, nil
}

// Marshal is the manifest's one serialization: indented, newline-terminated,
// fields in declaration order, so the same tree always yields the same bytes.
func Marshal(m Manifest) ([]byte, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// ParseCase reads an example's case.md: its `submitted:` and `outcome:`
// lines, and its `move: OLD NEW` fields when the case is a move
// (collaboration-modes.md §5).
func ParseCase(text string) (submitted, outcome string, move []string) {
	for _, l := range strings.Split(text, "\n") {
		if v, ok := strings.CutPrefix(l, "move: "); ok {
			move = strings.Fields(v)
		}
		if v, ok := strings.CutPrefix(l, "submitted: "); ok {
			submitted = v
		}
		if v, ok := strings.CutPrefix(l, "outcome: "); ok {
			outcome = v
		}
	}
	return submitted, outcome, move
}

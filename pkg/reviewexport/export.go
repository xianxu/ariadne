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
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
)

// Schema and Version identify the manifest format. A consumer refuses any
// other pair, so a change to the member shape bumps Version.
const (
	Schema  = "ariadne/universal-review-export"
	Version = 1
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

// Member is one exported file. Path is relative to the skill directory. Mode
// members carry their flags as a dense record; Name is set for adapters and
// modes; Source only for the grammar.
type Member struct {
	Role      string `json:"role"`
	Name      string `json:"name,omitempty"`
	Path      string `json:"path"`
	Source    string `json:"source,omitempty"`
	Order     int    `json:"order,omitempty"`
	Scope     string `json:"scope,omitempty"`
	Deletions string `json:"deletions,omitempty"`
	Frontier  string `json:"frontier,omitempty"`
	Bytes     int    `json:"bytes"`
	SHA256    string `json:"sha256"`
}

var roleRank = map[string]int{"round": 0, "grammar": 1, "collaboration-modes": 2, "adapter": 3, "mode": 4}

// defaults are parley's values for an omitted flag; the export records the
// resolved value, so its mode records are dense.
var defaults = map[string]string{"scope": "markers-only", "deletions": "propose-strike", "frontier": "on"}

var valid = map[string][]string{
	"scope":     {"whole-doc", "markers-only"},
	"deletions": {"apply-with-gutter-why", "propose-strike", "apply"},
	"frontier":  {"on", "off"},
}

// Mode is a parsed modes/<name>.md frontmatter. Order 0 means the mode has
// none, and sorts after every ordered mode.
type Mode struct {
	Name, Scope, Deletions, Frontier string
	Order                            int
}

// ParseMode reads a mode file's frontmatter by parley's rules: a leading
// `---\n` fence (so CRLF files are refused, as parley refuses them), flat
// `key: value` lines with values taken literally (quotes included), an
// omitted flag taking its default, an invalid one refused, `order` optional. The name must equal the file's
// basename (parley drops such a file; the authored source refuses it).
func ParseMode(content, basename string) (Mode, error) {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return Mode{}, fmt.Errorf("mode %s: missing frontmatter", basename)
	}
	// Searching from the opening fence's newline, as parley does, lets an
	// empty block (`---\n---`) parse, and fail later for want of a name.
	fm, _, ok := strings.Cut("\n"+rest, "\n---")
	if !ok {
		return Mode{}, fmt.Errorf("mode %s: unterminated frontmatter", basename)
	}
	flags := map[string]string{}
	for _, line := range strings.Split(fm, "\n") {
		if k, v, ok := strings.Cut(line, ":"); ok {
			flags[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	m := Mode{Name: flags["name"]}
	if m.Name != basename {
		return Mode{}, fmt.Errorf("mode %s: name %q must equal the file's basename", basename, m.Name)
	}
	for key, vals := range valid {
		v := flags[key]
		if v == "" {
			v = defaults[key]
		}
		if !slices.Contains(vals, v) {
			return Mode{}, fmt.Errorf("mode %s: %s=%q is not one of %s", basename, key, v, strings.Join(vals, ", "))
		}
		switch key {
		case "scope":
			m.Scope = v
		case "deletions":
			m.Deletions = v
		case "frontier":
			m.Frontier = v
		}
	}
	if o, ok := flags["order"]; ok && o != "" {
		order, err := strconv.Atoi(o)
		if err != nil || order < 1 {
			return Mode{}, fmt.Errorf("mode %s: order %q must be a positive integer", basename, o)
		}
		m.Order = order
	}
	return m, nil
}

// Build reads the skill directory and returns its manifest: round.md, the
// grammar link, collaboration-modes.md (#412), every adapters/*.md and every
// modes/*.md. The examples/ fixtures are not exported. Sizes and hashes
// cover each file's resolved contents (the grammar is a symlink). The order is
// total: role, then mode order (unordered last), then name, then path, as
// parley lists its modes.
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
	modes, err := fs.Glob(skill, "modes/*.md")
	if err != nil {
		return Manifest{}, err
	}
	if len(modes) == 0 {
		return Manifest{}, fmt.Errorf("export: no modes/*.md")
	}
	for _, p := range modes {
		b, err := fs.ReadFile(skill, p)
		if err != nil {
			return Manifest{}, err
		}
		md, err := ParseMode(string(b), strings.TrimSuffix(path.Base(p), ".md"))
		if err != nil {
			return Manifest{}, err
		}
		if err := add(Member{Role: "mode", Name: md.Name, Path: p, Order: md.Order, Scope: md.Scope, Deletions: md.Deletions, Frontier: md.Frontier}); err != nil {
			return Manifest{}, err
		}
	}
	slices.SortFunc(m.Members, func(a, b Member) int {
		if d := roleRank[a.Role] - roleRank[b.Role]; d != 0 {
			return d
		}
		if d := orderKey(a.Order) - orderKey(b.Order); d != 0 {
			return d
		}
		if d := strings.Compare(a.Name, b.Name); d != 0 {
			return d
		}
		return strings.Compare(a.Path, b.Path)
	})
	return m, nil
}

func orderKey(o int) int {
	if o == 0 {
		return math.MaxInt
	}
	return o
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

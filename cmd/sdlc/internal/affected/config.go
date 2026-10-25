// Package affected plans the affected suite (#473): which Go packages and declared test groups a change can break.
// It does no IO; LoadGoGraph (golist.go) is its one seam.
package affected

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// Group is a declared non-Go test group, selected by changed paths or by its Go roots' dependencies.
type Group struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths,omitempty"`
	Go    []string `json:"go,omitempty"` // package dirs relative to the module root, e.g. "./cmd/pair"
}

// Config is the parsed .sdlc/affected.json.
type Config struct {
	Version        int           `json:"version"`
	Command        []string      `json:"command"`
	Weight         int           `json:"weight,omitempty"`
	TimeoutMinutes int           `json:"timeout_minutes,omitempty"`
	Full           []FullRule    `json:"full,omitempty"`
	Groups         []Group       `json:"groups,omitempty"`
	Owners         []OwnerDecl   `json:"owners,omitempty"`
	Sidecars       []SidecarDecl `json:"sidecars,omitempty"`
	Traceability   []string      `json:"traceability,omitempty"` // traceability files (#616), exact repository paths
	GraphPrepare   *Preparation  `json:"graph_prepare,omitempty"`
}

// FullRule is a `full` entry (#598): a glob, or {"glob", "except"}, where a path the except globs match falls through
// to ordinary classification instead (pair keeps cmd/couch/** full but lets its tests select their package).
type FullRule struct {
	Glob   string   `json:"glob"`
	Except []string `json:"except,omitempty"`
}

// UnmarshalJSON takes either form, the object strictly.
func (r *FullRule) UnmarshalJSON(data []byte) error {
	var glob string
	if err := json.Unmarshal(data, &glob); err == nil {
		*r = FullRule{Glob: glob}
		return nil
	}
	type plain FullRule
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p plain
	if err := dec.Decode(&p); err != nil {
		return fmt.Errorf("a full entry is a glob or {\"glob\", \"except\"}: %w", err)
	}
	*r = FullRule(p)
	return nil
}

// Matches reports whether path is full by this rule: the glob matches and no except glob does.
func (r FullRule) Matches(p string) (bool, error) {
	ok, err := Match(r.Glob, p)
	if err != nil || !ok {
		return false, err
	}
	for _, e := range r.Except {
		if ok, err := Match(e, p); err != nil || ok {
			return false, err
		}
	}
	return true, nil
}

// OwnerDecl is an inline `owners` rule (#598): a change to a path its globs match also counts as a change to the
// declared Go roots, to the declared target paths, or to the one target the strip derives (prefix + target + suffix).
// Targets are classified as changed paths are, one level deep. Exactly one of go, as and strip is set.
type OwnerDecl struct {
	Paths []string `json:"paths"`
	Go    []string `json:"go,omitempty"`
	As    []string `json:"as,omitempty"`
	Strip *Strip   `json:"strip,omitempty"`
}

// Strip derives a target path from a matched one.
type Strip struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

// SidecarDecl declares a sidecar convention (#598): a file at prefix + target + suffix is about target. A change to
// the sidecar counts as a change to target, and the base copy's optional `reads` globs count as target's inputs.
type SidecarDecl struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

// Preparation establishes generated graph inputs in the head's private checkout.
type Preparation struct {
	Command        []string `json:"command"`
	TimeoutMinutes int      `json:"timeout_minutes,omitempty"`
}

// groupNameRE is the local-check name rule, so group names need no quoting in make.
var groupNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)

// ParseConfig parses affected.json strictly: a misread selection would skip tests, so anything unexpected refuses.
func ParseConfig(data []byte) (Config, error) {
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return Config{}, fmt.Errorf("affected.json: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("affected.json: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Config{}, errors.New("affected.json: trailing data after the object")
	}
	if cfg.Version != 1 {
		return Config{}, fmt.Errorf("affected.json: version %d, want 1", cfg.Version)
	}
	if len(cfg.Command) == 0 {
		return Config{}, errors.New("affected.json: command is empty")
	}
	for _, a := range cfg.Command {
		if a == "" {
			return Config{}, errors.New("affected.json: command has an empty argument")
		}
	}
	if p := cfg.GraphPrepare; p != nil {
		if len(p.Command) == 0 {
			return Config{}, errors.New("affected.json: graph_prepare command is empty")
		}
		for _, a := range p.Command {
			if a == "" {
				return Config{}, errors.New("affected.json: graph_prepare command has an empty argument")
			}
		}
		if p.TimeoutMinutes == 0 {
			p.TimeoutMinutes = 2
		}
		if p.TimeoutMinutes < 1 || p.TimeoutMinutes > 30 {
			return Config{}, errors.New("affected.json: graph_prepare timeout_minutes outside 1..30")
		}
	}
	if cfg.Weight == 0 {
		cfg.Weight = 1
	}
	if cfg.Weight < 1 || cfg.Weight > 8 {
		return Config{}, fmt.Errorf("affected.json: weight %d outside 1..8", cfg.Weight)
	}
	if cfg.TimeoutMinutes == 0 {
		cfg.TimeoutMinutes = 30
	}
	if cfg.TimeoutMinutes < 1 || cfg.TimeoutMinutes > 120 {
		return Config{}, fmt.Errorf("affected.json: timeout_minutes %d outside 1..120", cfg.TimeoutMinutes)
	}
	for _, r := range cfg.Full {
		if !validGlob(r.Glob) {
			return Config{}, fmt.Errorf("affected.json: full: bad glob %q", r.Glob)
		}
		for _, e := range r.Except {
			if !validGlob(e) {
				return Config{}, fmt.Errorf("affected.json: full %q: bad except glob %q", r.Glob, e)
			}
		}
	}
	for i, o := range cfg.Owners {
		if err := validOwner(o); err != nil {
			return Config{}, fmt.Errorf("affected.json: owners[%d]: %w", i, err)
		}
	}
	for i, sc := range cfg.Sidecars {
		if err := validSidecar(sc.Prefix, sc.Suffix); err != nil {
			return Config{}, fmt.Errorf("affected.json: sidecars[%d]: %w", i, err)
		}
	}
	files := map[string]bool{}
	for _, f := range cfg.Traceability {
		if !cleanRelPath(f) || hasGlobMeta(f) {
			return Config{}, fmt.Errorf("affected.json: traceability: %q must be an exact, clean repository-relative path", f)
		}
		if files[f] {
			return Config{}, fmt.Errorf("affected.json: traceability: duplicate %q", f)
		}
		files[f] = true
	}
	seen := map[string]bool{}
	for _, gr := range cfg.Groups {
		if !groupNameRE.MatchString(gr.Name) {
			return Config{}, fmt.Errorf("affected.json: group name %q must match %s", gr.Name, groupNameRE)
		}
		if seen[gr.Name] {
			return Config{}, fmt.Errorf("affected.json: duplicate group %q", gr.Name)
		}
		seen[gr.Name] = true
		if len(gr.Paths) == 0 && len(gr.Go) == 0 {
			return Config{}, fmt.Errorf("affected.json: group %q needs paths or go", gr.Name)
		}
		for _, g := range gr.Paths {
			if !validGlob(g) {
				return Config{}, fmt.Errorf("affected.json: group %q: bad glob %q", gr.Name, g)
			}
		}
		for _, r := range gr.Go {
			if !validGoRoot(r) {
				return Config{}, fmt.Errorf("affected.json: group %q: go root %q must be \".\" or \"./dir\"", gr.Name, r)
			}
		}
	}
	return cfg, nil
}

func validOwner(o OwnerDecl) error {
	if len(o.Paths) == 0 {
		return errors.New("paths is empty")
	}
	for _, g := range o.Paths {
		if !validGlob(g) {
			return fmt.Errorf("bad glob %q", g)
		}
	}
	set := 0
	for _, b := range []bool{len(o.Go) > 0, len(o.As) > 0, o.Strip != nil} {
		if b {
			set++
		}
	}
	if set != 1 {
		return errors.New("needs exactly one of go, as and strip")
	}
	for _, r := range o.Go {
		if !validGoRoot(r) {
			return fmt.Errorf("go root %q must be \".\" or \"./dir\"", r)
		}
	}
	for _, t := range o.As {
		if !cleanRelPath(t) {
			return fmt.Errorf("as target %q must be a clean repository-relative path", t)
		}
	}
	if o.Strip != nil && o.Strip.Prefix == "" && o.Strip.Suffix == "" {
		return errors.New("strip needs a prefix or a suffix")
	}
	return nil
}

// validSidecar: the prefix is a clean directory ending in "/", the suffix a non-empty file-name tail.
func validSidecar(prefix, suffix string) error {
	dir, ok := strings.CutSuffix(prefix, "/")
	if !ok || !cleanRelPath(dir) {
		return fmt.Errorf("prefix %q must be a clean repository-relative directory ending in /", prefix)
	}
	if suffix == "" || strings.ContainsAny(suffix, "/\\\x00") {
		return fmt.Errorf("suffix %q must be a non-empty file-name tail", suffix)
	}
	return nil
}

// cleanRelPath reports a canonical repository-relative path: no traversal, root, empty segment or backslash.
func cleanRelPath(p string) bool {
	return p != "" && p != "." && !path.IsAbs(p) && path.Clean(p) == p && p != ".." && !strings.HasPrefix(p, "../") &&
		!strings.ContainsAny(p, "\\\x00")
}

func validGoRoot(r string) bool {
	if r == "." {
		return true
	}
	rest, ok := strings.CutPrefix(r, "./")
	return ok && rest != "" && path.Clean(rest) == rest && !strings.HasPrefix(rest, "..")
}

// goRootDir maps a validated go root to a Graph.Dir value ("./a/b" → "a/b", "." → ".").
func goRootDir(r string) string {
	if rest := strings.TrimPrefix(r, "./"); rest != "" && rest != "." {
		return rest
	}
	return "."
}

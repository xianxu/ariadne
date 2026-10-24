package affected

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// Mapping is the compiled declared ownership (#598): every declaration source — affected.json `owners`, sidecar
// conventions and sidecar `reads` — as one rule list, which Plan consults in its claim step. It is built from bytes
// read at main's pinned rev; a branch never declares its own selection.
type Mapping struct {
	rules   []ownerRule
	exact   map[string][]int // a glob-free rule path → the rules naming it
	globbed []globRef        // globbed rule paths, scanned behind their literal prefix
	strips  []int            // sidecar conventions, matched by prefix and suffix
}

// globRef is one globbed rule path and the literal leading segments every path it matches must start with.
type globRef struct {
	rule    int
	glob    string
	literal string
}

// ownerRule is one declaration. Exactly one of go, as and strip is set; source names it in a refusal.
type ownerRule struct {
	source string
	paths  []string
	goRoot []string
	as     []string
	strip  *Strip
}

// SidecarFile is one sidecar's base copy: its repository path and content.
type SidecarFile struct {
	Path    string
	Content []byte
}

// Compile builds the Mapping from the parsed config and the sidecars read under its conventions. A sidecar that does
// not parse, or whose reads are not clean globs, refuses: its test's inputs are unknown, so nothing may be narrowed.
func Compile(cfg Config, sidecars []SidecarFile) (Mapping, error) {
	var m Mapping
	for i, o := range cfg.Owners {
		m.rules = append(m.rules, ownerRule{source: fmt.Sprintf("affected.json owners[%d]", i), paths: o.Paths, goRoot: o.Go, as: o.As, strip: o.Strip})
	}
	for i, sc := range cfg.Sidecars {
		m.rules = append(m.rules, ownerRule{source: fmt.Sprintf("affected.json sidecars[%d]", i), strip: &Strip{Prefix: sc.Prefix, Suffix: sc.Suffix}})
	}
	for _, f := range sidecars {
		target, ok := sidecarTarget(cfg.Sidecars, f.Path)
		if !ok {
			return Mapping{}, fmt.Errorf("sidecar %s is under no declared convention", f.Path)
		}
		reads, err := ParseSidecar(f.Content)
		if err != nil {
			return Mapping{}, fmt.Errorf("sidecar %s: %w", f.Path, err)
		}
		if len(reads) > 0 {
			m.rules = append(m.rules, ownerRule{source: "sidecar " + f.Path, paths: reads, as: []string{target}})
		}
	}
	m.exact = map[string][]int{}
	for i, r := range m.rules {
		if r.paths == nil {
			m.strips = append(m.strips, i)
			continue
		}
		for _, p := range r.paths {
			if strings.ContainsAny(p, "*?[\\") {
				m.globbed = append(m.globbed, globRef{rule: i, glob: p, literal: literalPrefix(p)})
				continue
			}
			m.exact[p] = append(m.exact[p], i)
		}
	}
	return m, nil
}

// ParseSidecar returns a sidecar's declared reads. The sidecar is a JSON object whose every member but `reads` is
// opaque to ariadne; an absent `reads` declares nothing. Duplicate keys, trailing data and a reads value that is not
// an array of clean globs refuse.
func ParseSidecar(data []byte) ([]string, error) {
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	var obj map[string]json.RawMessage
	if err := dec.Decode(&obj); err != nil {
		return nil, fmt.Errorf("not a JSON object: %w", err)
	}
	if obj == nil {
		return nil, errors.New("not a JSON object")
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing data after the object")
	}
	raw, ok := obj["reads"]
	if !ok {
		return nil, nil
	}
	var reads []string
	if err := json.Unmarshal(raw, &reads); err != nil || reads == nil {
		return nil, errors.New("reads must be an array of globs")
	}
	for _, g := range reads {
		if !validGlob(g) || !cleanRelPath(g) {
			return nil, fmt.Errorf("reads: %q must be a clean repository-relative glob", g)
		}
	}
	return reads, nil
}

// sidecarTarget is the path a sidecar file is about, under the first convention it fits.
func sidecarTarget(decls []SidecarDecl, file string) (string, bool) {
	for _, sc := range decls {
		if t, ok := stripTarget(Strip{Prefix: sc.Prefix, Suffix: sc.Suffix}, file); ok {
			return t, true
		}
	}
	return "", false
}

// stripTarget derives prefix + target + suffix → target, refusing a target that is not a clean relative path (a
// traversal the strip would otherwise produce, or nothing left).
func stripTarget(s Strip, p string) (string, bool) {
	rest, ok := strings.CutPrefix(p, s.Prefix)
	if !ok {
		return "", false
	}
	t, ok := strings.CutSuffix(rest, s.Suffix)
	if !ok || !cleanRelPath(t) {
		return "", false
	}
	return t, true
}

// literalPrefix is a glob's leading glob-free segments with their slashes ("a/b/*.go" → "a/b/"): a cheap filter
// before Match, never a substitute for it.
func literalPrefix(glob string) string {
	segs := strings.Split(glob, "/")
	n := 0
	for _, seg := range segs[:len(segs)-1] {
		if strings.ContainsAny(seg, "*?[\\") {
			break
		}
		n += len(seg) + 1
	}
	return glob[:n]
}

// matching returns the rules whose paths match p, in declaration order. A malformed glob is an error.
func (m Mapping) matching(p string) ([]int, error) {
	hit := map[int]bool{}
	for _, i := range m.exact[p] {
		hit[i] = true
	}
	for _, gr := range m.globbed {
		if hit[gr.rule] || !strings.HasPrefix(p, gr.literal) {
			continue
		}
		ok, err := Match(gr.glob, p)
		if err != nil {
			return nil, fmt.Errorf("%s: bad glob %q: %v", m.rules[gr.rule].source, gr.glob, err)
		}
		if ok {
			hit[gr.rule] = true
		}
	}
	for _, i := range m.strips {
		s := m.rules[i].strip
		if strings.HasPrefix(p, s.Prefix) && strings.HasSuffix(p, s.Suffix) {
			hit[i] = true
		}
	}
	out := make([]int, 0, len(hit))
	for i := range m.rules {
		if hit[i] {
			out = append(out, i)
		}
	}
	return out, nil
}

// claim is what one changed path selects: packages and groups, or the reason it makes the selection full.
type claim struct {
	pkgs   map[string]bool
	groups map[string]bool
	full   string
}

func (c claim) claimed() bool { return len(c.pkgs) > 0 || len(c.groups) > 0 }

// claimPath classifies p, `full` aside: the graph's owners, the groups whose paths match, and — for a changed path,
// not a declared target (one level) — every declared rule matching it. A declared root that is not a package, a
// target the strip cannot derive, a target matching `full`, and a target that claims nothing make it full.
func claimPath(p string, g Graph, byDir map[string]string, cfg Config, m Mapping, declared bool) claim {
	c := claim{pkgs: map[string]bool{}, groups: map[string]bool{}}
	for _, pkg := range g.owners(byDir, p) {
		c.pkgs[pkg] = true
	}
	for _, gr := range cfg.Groups {
		for _, pat := range gr.Paths {
			ok, err := Match(pat, p)
			if err != nil {
				c.full = fmt.Sprintf("group %s: bad glob %q: %v", gr.Name, pat, err)
				return c
			}
			if ok {
				c.groups[gr.Name] = true
				break
			}
		}
	}
	if !declared {
		return c
	}
	rules, err := m.matching(p)
	if err != nil {
		c.full = err.Error()
		return c
	}
	for _, i := range rules {
		r := m.rules[i]
		for _, root := range r.goRoot {
			pkg, ok := byDir[goRootDir(root)]
			if !ok {
				c.full = fmt.Sprintf("%s: %s is not a package", r.source, root)
				return c
			}
			c.pkgs[pkg] = true
		}
		targets := r.as
		if r.strip != nil {
			t, ok := stripTarget(*r.strip, p)
			if !ok {
				c.full = fmt.Sprintf("%s: no target derives from %s", r.source, p)
				return c
			}
			targets = []string{t}
		}
		for _, t := range targets {
			isFull, err := fullMatch(cfg, t)
			if err != nil {
				c.full = err.Error()
				return c
			}
			if isFull {
				c.full = fmt.Sprintf("%s (declared by %s for %s)", t, r.source, p)
				return c
			}
			tc := claimPath(t, g, byDir, cfg, m, false)
			if tc.full != "" {
				c.full = tc.full
				return c
			}
			if !tc.claimed() {
				c.full = fmt.Sprintf("%s: target %s of %s claims nothing", r.source, t, p)
				return c
			}
			for k := range tc.pkgs {
				c.pkgs[k] = true
			}
			for k := range tc.groups {
				c.groups[k] = true
			}
		}
	}
	return c
}

// fullMatch reports whether a `full` rule selects p; a malformed glob is an error.
func fullMatch(cfg Config, p string) (bool, error) {
	for _, r := range cfg.Full {
		ok, err := r.Matches(p)
		if err != nil {
			return false, fmt.Errorf("bad glob %q: %v", r.Glob, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// Owned returns every package a change to p can break, declared ownership included (#598): what Plan's claim step
// selects for p, sorted. splitImpact's uncertain fallback uses it, so landing attribution sees what selection sees.
// A path whose declaration makes the selection full still returns the packages it claimed before that.
func Owned(p string, g Graph, cfg Config, m Mapping) []string {
	return sortedKeys(claimPath(p, g, g.byDir(), cfg, m, true).pkgs)
}

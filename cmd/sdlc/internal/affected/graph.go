package affected

import (
	"path"
	"sort"
)

// Graph is the module's package graph (#473).
type Graph struct {
	Module string
	Dir    map[string]string          // import path → module-relative dir ("." for root)
	Deps   map[string]map[string]bool // import path → transitive in-module deps, test imports included
	Embeds map[string]string          // module-relative embedded file → the package embedding it (go:embed)
}

func (g Graph) byDir() map[string]string {
	idx := make(map[string]string, len(g.Dir))
	for p, d := range g.Dir {
		idx[d] = p
	}
	return idx
}

// PackageOf returns the package that owns file, exactly: a file directly in a package dir, a file under that
// package's testdata/, or a file the package embeds. Any other file, including one in a plain subdirectory of a
// package (the module root's above all), is unowned and so unclassified unless a group claims it (#473 M1 BR-1).
func (g Graph) PackageOf(file string) (string, bool) {
	return g.packageIn(g.byDir(), file)
}

func (g Graph) packageIn(byDir map[string]string, file string) (string, bool) {
	if p, ok := byDir[path.Dir(file)]; ok {
		return p, true
	}
	for d := path.Dir(file); d != "." && d != "/"; {
		parent := path.Dir(d)
		if path.Base(d) == "testdata" {
			if p, ok := byDir[parent]; ok {
				return p, true
			}
		}
		d = parent
	}
	p, ok := g.Embeds[file]
	return p, ok
}

// Dependents returns changed plus every package depending on one of them, sorted.
func (g Graph) Dependents(changed map[string]bool) []string {
	out := map[string]bool{}
	for p := range changed {
		out[p] = true
	}
	for p, deps := range g.Deps {
		for d := range deps {
			if changed[d] {
				out[p] = true
				break
			}
		}
	}
	return sortedKeys(out)
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

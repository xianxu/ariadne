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
	Embeds map[string][]string        // module-relative embedded file → every package embedding it (go:embed)
}

func (g Graph) byDir() map[string]string {
	idx := make(map[string]string, len(g.Dir))
	for p, d := range g.Dir {
		idx[d] = p
	}
	return idx
}

// Owners returns every package a change to file can break, sorted: the package whose dir holds it, each package
// whose testdata/ holds it, and each package that embeds it. A file can have several (a root package may embed a
// file inside a sub-package's dir), and the union is the rule; none means unowned, so the file is unclassified
// unless a group claims it. A plain subdirectory file, under the root package above all, has no owner (#473 M1
// BR-1, BR-7).
func (g Graph) Owners(file string) []string {
	return g.owners(g.byDir(), file)
}

func (g Graph) owners(byDir map[string]string, file string) []string {
	set := map[string]bool{}
	if p, ok := byDir[path.Dir(file)]; ok {
		set[p] = true
	}
	for d := path.Dir(file); d != "." && d != "/"; d = path.Dir(d) {
		if path.Base(d) == "testdata" {
			if p, ok := byDir[path.Dir(d)]; ok {
				set[p] = true
			}
		}
	}
	for _, p := range g.Embeds[file] {
		set[p] = true
	}
	return sortedKeys(set)
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

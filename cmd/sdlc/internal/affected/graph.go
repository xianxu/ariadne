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
}

func (g Graph) byDir() map[string]string {
	idx := make(map[string]string, len(g.Dir))
	for p, d := range g.Dir {
		idx[d] = p
	}
	return idx
}

// PackageOf returns the package owning file: the deepest package dir containing it, so testdata and embedded
// assets belong to their package.
func (g Graph) PackageOf(file string) (string, bool) {
	return packageIn(g.byDir(), file)
}

func packageIn(byDir map[string]string, file string) (string, bool) {
	for d := path.Dir(file); ; d = path.Dir(d) {
		if p, ok := byDir[d]; ok {
			return p, true
		}
		if d == "." || d == "/" {
			return "", false
		}
	}
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

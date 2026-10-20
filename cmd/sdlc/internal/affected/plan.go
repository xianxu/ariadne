package affected

import "fmt"

// Selection is what the affected suite runs: Full, or the Go packages and groups (both sorted).
type Selection struct {
	Full   bool
	Reason string
	GoPkgs []string
	Groups []string
}

// Empty reports a selection that runs nothing (a docs-only change).
func (s Selection) Empty() bool {
	return !s.Full && len(s.GoPkgs) == 0 && len(s.Groups) == 0
}

func full(reason string) Selection { return Selection{Full: true, Reason: reason} }

// Plan classifies each changed path by the plan's affected.Plan rules (#473): a `full` match selects everything;
// otherwise package ownership (Owners) and group `paths` are additive, never first-match (PQ-1). A path neither
// claims is ignored when it is a doc (isDoc, the caller's docs rule) and Full otherwise: a doc a package owns, an
// embedded or testdata .md, still selects that package (close review). Changed packages then expand to their
// dependents, and a group whose `go` root is among them is selected.
func Plan(paths []string, g Graph, cfg Config, isDoc func(string) bool) Selection {
	byDir := g.byDir()
	changed := map[string]bool{}
	groups := map[string]bool{}
	for _, p := range paths {
		for _, pat := range cfg.Full {
			ok, err := Match(pat, p)
			if err != nil {
				return full(fmt.Sprintf("bad glob %q: %v", pat, err))
			}
			if ok {
				return full(p)
			}
		}
		claimed := false
		for _, pkg := range g.owners(byDir, p) {
			changed[pkg] = true
			claimed = true
		}
		for _, gr := range cfg.Groups {
			for _, pat := range gr.Paths {
				ok, err := Match(pat, p)
				if err != nil {
					return full(fmt.Sprintf("group %s: bad glob %q: %v", gr.Name, pat, err))
				}
				if ok {
					groups[gr.Name] = true
					claimed = true
					break
				}
			}
		}
		if !claimed && (isDoc == nil || !isDoc(p)) {
			return full("unclassified: " + p)
		}
	}
	if len(changed) == 0 && len(groups) == 0 {
		return Selection{} // docs no package or group claims
	}
	pkgs := g.Dependents(changed)
	selected := make(map[string]bool, len(pkgs))
	for _, p := range pkgs {
		selected[p] = true
	}
	// A stale go root must not silently skip its group, so every root is checked, not only until one hits.
	for _, gr := range cfg.Groups {
		for _, r := range gr.Go {
			pkg, ok := byDir[goRootDir(r)]
			if !ok {
				return full(fmt.Sprintf("group %s: %s is not a package", gr.Name, r))
			}
			if selected[pkg] {
				groups[gr.Name] = true
			}
		}
	}
	return Selection{GoPkgs: pkgs, Groups: sortedKeys(groups)}
}

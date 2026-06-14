// Package layer resolves the layer dependency graph: it turns a set of
// per-layer dependency edges into a foundation-first, deduped application
// order. Pure — no IO (ARCH-PURE); edges are read at the boundary and passed
// in. Mirrors the ordering of setup.sh's discover_ancestors (ARCH-DRY).
package layer

import "fmt"

// node visit states for cycle detection.
const (
	unvisited = iota
	active    // on the current DFS path
	done      // fully resolved + emitted
)

// Resolve returns root and its transitive dependencies in foundation-first
// topological order — every node appears after all of its dependencies —
// with each layer emitted exactly once (a diamond collapses to one
// application). deps maps a layer to the layers it depends on; a layer
// absent from deps (or with an empty list) is a foundation. Returns an error
// if the graph contains a dependency cycle.
func Resolve(root string, deps map[string][]string) ([]string, error) {
	state := make(map[string]int)
	var order []string

	var visit func(n string, path []string) error
	visit = func(n string, path []string) error {
		switch state[n] {
		case done:
			return nil // already emitted — dedup (handles diamonds)
		case active:
			return fmt.Errorf("dependency cycle: %v -> %s", path, n)
		}
		state[n] = active
		for _, dep := range deps[n] {
			if err := visit(dep, append(path, n)); err != nil {
				return err
			}
		}
		state[n] = done
		order = append(order, n) // post-order ⇒ deps before dependents
		return nil
	}

	if err := visit(root, nil); err != nil {
		return nil, err
	}
	return order, nil
}

package layergraph

import "strings"

// ParseDeps extracts the substrate dependency edges from the text content of a
// construct/deps file, in file order — the sole source of layer-dependency
// edges (the map Resolve consumes). Pure: it takes the file's content, never
// the file (reading it is an IO-seam concern, ARCH-PURE).
//
// The grammar is ported verbatim from
// construct/scripts/lib-deps.sh:deps_substrate_targets (ARCH-DRY — weave must
// parse construct/deps identically to the shell that parses it today):
//
//   - Each line is truncated at the first '#' (so whole-line and trailing
//     comments drop), then whitespace-split into positional columns.
//   - A row needs ≥2 columns (kind + target); blank, comment-only, or
//     otherwise short rows are skipped silently — matching the shell's
//     `[[ $# -ge 2 ]] || continue`, which never errors on a malformed row.
//   - Only `substrate` rows contribute an edge; `data` (and any other kind)
//     rows are ignored.
//
// The returned slice is the per-layer relpath edge list; resolving each
// relpath to an on-disk sibling is the transitive-walk concern (Walk).
// ParseDeps returns no error today — lib-deps.sh rejects nothing — but keeps
// the error in its signature so a future strict mode is a non-breaking change.
func ParseDeps(content string) ([]string, error) {
	var edges []string
	for _, line := range strings.Split(content, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i] // strip trailing/whole-line comment (lib-deps: ${line%%#*})
		}
		fields := strings.Fields(line) // whitespace word-split (lib-deps: set -- $line)
		if len(fields) < 2 {
			continue // blank / comment-only / malformed (lib-deps: $# -ge 2)
		}
		if fields[0] != "substrate" {
			continue // ignore data + any other kind
		}
		edges = append(edges, fields[1])
	}
	return edges, nil
}

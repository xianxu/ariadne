package layergraph

import (
	"fmt"
	"strings"
)

// Dependency preserves the source and mount information from a construct/deps
// row. Substrate paths remain the only edges in the layer graph.
type Dependency struct {
	Kind   string
	Path   string
	Source string
	Mount  string
}

// ParseRows parses the existing whitespace/comment grammar, retaining recognized
// substrate and data rows. Malformed declarations fail with their line number.
func ParseRows(content string) ([]Dependency, error) {
	var rows []Dependency
	for lineNumber, line := range strings.Split(content, "\n") {
		line, _, _ = strings.Cut(line, "#")
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if (f[0] == "substrate" && (len(f) < 2 || len(f) > 3)) || (f[0] == "data" && len(f) != 3) || (f[0] != "substrate" && f[0] != "data") {
			return nil, fmt.Errorf("construct/deps line %d: expected substrate <path> [source] or data <source> <mount>", lineNumber+1)
		}
		switch f[0] {
		case "substrate":
			row := Dependency{Kind: f[0], Path: f[1]}
			if len(f) > 2 {
				row.Source = f[2]
			}
			rows = append(rows, row)
		case "data":
			if len(f) >= 3 {
				rows = append(rows, Dependency{Kind: f[0], Source: f[1], Mount: f[2]})
			}
		}
	}
	return rows, nil
}

// ParseDeps projects substrate paths from the shared parser for existing graph
// consumers. Source metadata and data mounts do not introduce topology edges.
func ParseDeps(content string) ([]string, error) {
	rows, err := ParseRows(content)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, row := range rows {
		if row.Kind == "substrate" {
			paths = append(paths, row.Path)
		}
	}
	return paths, nil
}

package refresh

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// declarationLabel never includes a source. Paths and mounts are untrusted too:
// omit URL/scp-like labels entirely instead of trying to redact their secrets.
func declarationLabel(row layergraph.Dependency) string {
	label := row.Path
	if row.Kind == "data" {
		label = row.Mount
	}
	if label == "" || strings.ContainsAny(label, ":@?%") || strings.HasPrefix(label, "//") || strings.IndexFunc(label, unicode.IsControl) >= 0 {
		return row.Kind
	}
	return fmt.Sprintf("%s %q", row.Kind, label)
}

// declarationChanges compares declaration ordinals, not physical file lines.
// Keeping the ordering distinguishes duplicate rows and reordered declarations.
func declarationChanges(before, after []layergraph.Dependency) string {
	var changes []string
	for i := 0; i < max(len(before), len(after)); i++ {
		prefix := fmt.Sprintf("row %d: ", i+1)
		switch {
		case i >= len(before):
			changes = append(changes, prefix+declarationLabel(after[i])+" added")
		case i >= len(after):
			changes = append(changes, prefix+declarationLabel(before[i])+" removed")
		default:
			a, b := before[i], after[i]
			if a == b {
				continue
			}
			var fields []string
			for _, f := range []struct{ name, a, b string }{
				{"kind", a.Kind, b.Kind}, {"path", a.Path, b.Path}, {"source", a.Source, b.Source}, {"mount", a.Mount, b.Mount},
			} {
				if f.a != f.b {
					fields = append(fields, f.name)
				}
			}
			label := declarationLabel(a)
			if other := declarationLabel(b); other != label {
				label += " -> " + other
			}
			changes = append(changes, prefix+label+" ("+strings.Join(fields, ", ")+" changed)")
		}
	}
	return strings.Join(changes, "; ")
}

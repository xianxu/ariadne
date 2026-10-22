package main

import (
	"fmt"
	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// projectWorkingdoc is shared by settlement and drift detection. It never
// resolves pending review items: the editing surface owns their acceptance.
func projectWorkingdoc(text, from, to string, exists func(string) bool) (string, []string, error) {
	m, err := parseWorkingdoc(text)
	if err != nil {
		return "", nil, err
	}
	// Use the migration's existing parent-coordinate contract. A YAML form
	// the shared rebaser cannot safely rewrite must not silently change meaning
	// when the copy gets a different directory.
	if _, err := moveMetadataFields(text); err != nil {
		return "", nil, err
	}
	if markers := workingdocMarkers(text); len(markers) > 0 {
		return "", nil, fmt.Errorf("unresolved review markers: %+v", markers)
	}
	if m.Pointer != nil {
		text = text[:m.pointerStart] + text[m.pointerEnd:]
	}
	moved := func(string) (string, bool) { return "", false }
	projected, warnings := sidedoc.RebaseLinks(text, from, to, moved, exists)
	projected, err = sidedoc.RebaseIndex(projected, from, to, moved)
	return projected, warnings, err
}

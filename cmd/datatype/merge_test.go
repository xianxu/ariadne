package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// writeProto writes a prototype file with a frontmatter `description:` so the
// merge can read it via pkg/frontmatter.Description.
func writeProto(t *testing.T, path, desc string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: type\ndescription: " + desc + "\n---\n\nbody of " + filepath.Base(path) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestMergeTypes_DAGUnionLocalWinsByFilename is the merge-policy gate. It builds
// a two-layer DAG (ariadne foundation, nous leaf) plus a nous-project-local
// datatype/ dir and asserts:
//   - the union spans every layer's construct/datatype/*.md PLUS the leaf's
//     project-local datatype/*.md,
//   - keying is by FILENAME (product.md → name "product", though its `type:`
//     frontmatter is "type" — the naming trap, plan-review N10),
//   - a downstream layer's event.md SHADOWS the ancestor's (leaf wins),
//   - sorted by name, with the winning layer's BodyPath + Description.
func TestMergeTypes_DAGUnionLocalWinsByFilename(t *testing.T) {
	parent := t.TempDir()
	ariadne := filepath.Join(parent, "ariadne")
	nous := filepath.Join(parent, "nous")

	// ariadne (foundation): continuation, event, product.
	writeProto(t, filepath.Join(ariadne, "construct", "datatype", "continuation.md"), "ariadne continuation")
	writeProto(t, filepath.Join(ariadne, "construct", "datatype", "event.md"), "ariadne event")
	// product.md: filename is the type name even though `type:` says "type".
	writeProto(t, filepath.Join(ariadne, "construct", "datatype", "product.md"), "ariadne product")

	// nous (leaf, shared dir): event shadows ariadne's.
	writeProto(t, filepath.Join(nous, "construct", "datatype", "event.md"), "nous event (wins)")

	// nous project-local datatype/ override: a leaf-only type.
	leafLocal := filepath.Join(nous, "datatype")
	writeProto(t, filepath.Join(leafLocal, "trip.md"), "nous-local trip")

	// Foundation-first roots: ariadne then nous.
	roots := []string{ariadne, nous}

	got, err := mergeTypes(layergraph.OSFS{}, roots, leafLocal)
	if err != nil {
		t.Fatalf("mergeTypes: %v", err)
	}

	// Expected: continuation, event(nous), product, trip — sorted by name.
	type expect struct {
		name     string
		desc     string
		bodyPath string
	}
	want := []expect{
		{"continuation", "ariadne continuation", filepath.Join(ariadne, "construct", "datatype", "continuation.md")},
		{"event", "nous event (wins)", filepath.Join(nous, "construct", "datatype", "event.md")},
		{"product", "ariadne product", filepath.Join(ariadne, "construct", "datatype", "product.md")},
		{"trip", "nous-local trip", filepath.Join(leafLocal, "trip.md")},
	}

	if len(got) != len(want) {
		names := make([]string, len(got))
		for i, p := range got {
			names[i] = p.Name
		}
		t.Fatalf("mergeTypes returned %d protos %v, want %d", len(got), names, len(want))
	}
	for i, w := range want {
		if got[i].Name != w.name {
			t.Fatalf("proto[%d].Name = %q, want %q (sorted by name)", i, got[i].Name, w.name)
		}
		if got[i].Description != w.desc {
			t.Fatalf("proto[%d] %q Description = %q, want %q", i, w.name, got[i].Description, w.desc)
		}
		if got[i].BodyPath != w.bodyPath {
			t.Fatalf("proto[%d] %q BodyPath = %q, want %q", i, w.name, got[i].BodyPath, w.bodyPath)
		}
	}
}

// TestMergeTypes_LeafLocalShadowsSharedAndAncestor: the leaf project-local
// datatype/ overlay is applied LAST, so it wins over both an ancestor's and the
// leaf's own construct/datatype/ entry of the same filename.
func TestMergeTypes_LeafLocalShadowsSharedAndAncestor(t *testing.T) {
	parent := t.TempDir()
	ariadne := filepath.Join(parent, "ariadne")
	nous := filepath.Join(parent, "nous")

	writeProto(t, filepath.Join(ariadne, "construct", "datatype", "event.md"), "ancestor event")
	writeProto(t, filepath.Join(nous, "construct", "datatype", "event.md"), "leaf shared event")
	leafLocal := filepath.Join(nous, "datatype")
	writeProto(t, filepath.Join(leafLocal, "event.md"), "leaf-local event (wins)")

	got, err := mergeTypes(layergraph.OSFS{}, []string{ariadne, nous}, leafLocal)
	if err != nil {
		t.Fatalf("mergeTypes: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 proto (event), got %d", len(got))
	}
	if got[0].Name != "event" || got[0].Description != "leaf-local event (wins)" {
		t.Fatalf("leaf-local did not win: got %+v", got[0])
	}
	if got[0].BodyPath != filepath.Join(leafLocal, "event.md") {
		t.Fatalf("BodyPath = %q, want leaf-local", got[0].BodyPath)
	}
}

// TestMergeTypes_MissingDirsTolerated: a root without construct/datatype/, and
// an absent leaf-local dir, are silently skipped (not errors) — a layer need not
// define any prototypes.
func TestMergeTypes_MissingDirsTolerated(t *testing.T) {
	parent := t.TempDir()
	ariadne := filepath.Join(parent, "ariadne")
	bare := filepath.Join(parent, "bare") // no construct/datatype/

	writeProto(t, filepath.Join(ariadne, "construct", "datatype", "continuation.md"), "c")

	// leaf-local points at a non-existent dir.
	got, err := mergeTypes(layergraph.OSFS{}, []string{ariadne, bare}, filepath.Join(bare, "datatype"))
	if err != nil {
		t.Fatalf("mergeTypes should tolerate missing dirs, got err: %v", err)
	}
	if len(got) != 1 || got[0].Name != "continuation" {
		t.Fatalf("want only continuation, got %+v", got)
	}
}

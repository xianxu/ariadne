package reviewexport

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

const monthSide = "d.side/2026-10/2026-10-22-01-why.md"

func monthCase() fstest.MapFS {
	m := goodCase()
	old := "after/d.side/01-why.md"
	body := strings.ReplaceAll(string(m[old].Data), "../d.md", "../../d.md")
	body = strings.Replace(body, "collaboration-mode: discussion", "collaboration-mode: discussion\npresentation-mode: highlight\nprocessing-mode: free-form", 1)
	delete(m, old)
	m["after/"+monthSide] = &fstest.MapFile{Data: []byte(body)}
	m["after/d.md"] = &fstest.MapFile{Data: []byte("# D\n\nText. → [discussion](" + monthSide + ")\n")}
	return m
}

func TestMonthSideDocChecker(t *testing.T) {
	m := monthCase()
	after, _ := fs.Sub(m, "after")
	sides, errs := checkTree(after)
	if _, ok := sides[monthSide]; !ok {
		t.Errorf("month side doc not inspected: %v", sides)
	}
	if len(errs) != 0 {
		t.Fatalf("valid month tree: %v", errs)
	}
	if errs := checkCase(m); len(errs) != 0 {
		t.Fatalf("valid month round: %v", errs)
	}
	if errs := checkCase(goodCase()); len(errs) != 0 {
		t.Fatalf("legacy flat round: %v", errs)
	}
	mutate := func(m fstest.MapFS, p, old, new string) {
		m[p] = &fstest.MapFile{Data: []byte(strings.Replace(string(m[p].Data), old, new, 1))}
	}
	for _, c := range []struct {
		name string
		edit func(fstest.MapFS)
		want string
	}{
		{"empty invalid directory", func(m fstest.MapFS) { m["after/d.side/archive"] = &fstest.MapFile{Mode: fs.ModeDir} }, "month"},
		{"empty deeper directory", func(m fstest.MapFS) { m["after/d.side/2026-10/deep"] = &fstest.MapFile{Mode: fs.ModeDir} }, "nested"},
		{"undated month name", func(m fstest.MapFS) { renameMonth(m, "d.side/2026-10/01-why.md") }, "dated"},
		{"mismatched month", func(m fstest.MapFS) { renameMonth(m, "d.side/2026-09/2026-10-22-01-why.md") }, "month"},
		{"invalid month", func(m fstest.MapFS) { renameMonth(m, "d.side/2026-13/2026-13-22-01-why.md") }, "month"},
		{"invalid date", func(m fstest.MapFS) { renameMonth(m, "d.side/2026-10/2026-10-32-01-why.md") }, "date"},
		{"deeper directory", func(m fstest.MapFS) { renameMonth(m, "d.side/2026-10/deep/2026-10-22-01-why.md") }, "nested"},
		{"arbitrary directory", func(m fstest.MapFS) { renameMonth(m, "d.side/archive/2026-10-22-01-why.md") }, "month"},
		{"nested side root", func(m fstest.MapFS) { renameMonth(m, "d.side/nested.side/2026-10-22-01-why.md") }, ".side/ inside"},
		{"missing presentation default", func(m fstest.MapFS) { mutate(m, "after/"+monthSide, "presentation-mode: highlight\n", "") }, "presentation-mode"},
		{"missing processing default", func(m fstest.MapFS) { mutate(m, "after/"+monthSide, "processing-mode: free-form\n", "") }, "processing-mode"},
		{"missing collaboration mode", func(m fstest.MapFS) { mutate(m, "after/"+monthSide, "collaboration-mode: discussion\n", "") }, "collaboration-mode"},
		{"wrong parent", func(m fstest.MapFS) {
			mutate(m, "after/"+monthSide, "parent: ../../d.md", "parent: ../../e.md")
			m["after/e.md"] = &fstest.MapFile{Data: []byte("# E\n")}
		}, "whose parent"},
		{"wrong backlink", func(m fstest.MapFS) { mutate(m, "after/"+monthSide, "[D](../../d.md)", "[D](../../e.md)") }, "backlink"},
		{"missing backlink", func(m fstest.MapFS) { mutate(m, "after/"+monthSide, "_From", "From") }, "no backlink"},
		{"duplicate links", func(m fstest.MapFS) { mutate(m, "after/d.md", "# D", "# D\n\n→ [discussion]("+monthSide+")") }, "two links point"},
		{"wrong index parent", func(m fstest.MapFS) {
			m["after/e.md"] = &fstest.MapFile{Data: []byte("---\ndiscussions:\n  - doc: " + monthSide + "\n---\n# E\n")}
		}, "names a side doc whose parent"},
		{"legacy unlinked reuse", func(m fstest.MapFS) {
			legacy := goodCase()["after/d.side/01-why.md"]
			m["before/d.side/01-why.md"] = legacy
			m["after/d.side/01-why.md"] = legacy
		}, "should have been relinked"},
		{"flat dated collision", func(m fstest.MapFS) {
			legacy := goodCase()["after/d.side/01-why.md"]
			m["before/d.side/2026-10-22-01-old.md"] = legacy
			m["after/d.side/2026-10-22-01-old.md"] = legacy
		}, "named 2026-10-22-02-why.md"},
		{"month dated collision", func(m fstest.MapFS) {
			m["before/d.side/2026-10/2026-10-22-01-old.md"] = m["after/"+monthSide]
			m["after/d.side/2026-10/2026-10-22-01-old.md"] = m["after/"+monthSide]
		}, "named 2026-10-22-02-why.md"},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := monthCase()
			c.edit(m)
			errs := strings.Join(checkCase(m), "\n")
			if !strings.Contains(errs, c.want) {
				t.Errorf("errors %q, want %q", errs, c.want)
			}
		})
	}
}

func renameMonth(m fstest.MapFS, newPath string) {
	m["after/"+newPath] = m["after/"+monthSide]
	delete(m, "after/"+monthSide)
	m["after/d.md"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(m["after/d.md"].Data), monthSide, newPath))}
}

func TestMonthSideDocMixedLayoutControls(t *testing.T) {
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	t.Run("legacy dated flat remains valid without modern modes", func(t *testing.T) {
		m := goodCase()
		old := "after/d.side/01-why.md"
		m["after/d.side/2026-10-22-01-why.md"] = m[old]
		delete(m, old)
		m["after/d.md"] = f(strings.ReplaceAll(string(m["after/d.md"].Data), "01-why.md", "2026-10-22-01-why.md"))
		if errs := checkCase(m); len(errs) != 0 {
			t.Fatal(errs)
		}
	})
	for _, difference := range []string{"marker", "section", "root", "already linked"} {
		t.Run(difference, func(t *testing.T) {
			m := monthCase()
			legacy := string(goodCase()["after/d.side/01-why.md"].Data)
			legacyPath := "d.side/01-why.md"
			switch difference {
			case "marker":
				legacy = strings.ReplaceAll(legacy, `marker: "why?"`, `marker: "other?"`)
			case "section":
				legacy = strings.ReplaceAll(legacy, `section "D"`, `section "Other"`)
			case "root":
				legacyPath = "other.side/01-why.md"
			case "already linked":
				link := "\n→ [discussion](d.side/01-why.md)\n"
				m["before/d.md"] = f(string(m["before/d.md"].Data) + link)
				m["after/d.md"] = f(string(m["after/d.md"].Data) + link)
			}
			m["before/"+legacyPath] = f(legacy)
			m["after/"+legacyPath] = f(legacy)
			if errs := checkCase(m); len(errs) != 0 {
				t.Fatal(errs)
			}
		})
	}
	t.Run("flat taken date advances month allocation", func(t *testing.T) {
		m := monthCase()
		legacy := strings.ReplaceAll(string(goodCase()["after/d.side/01-why.md"].Data), `marker: "why?"`, `marker: "other?"`)
		m["before/d.side/2026-10-22-01-old.md"] = f(legacy)
		m["after/d.side/2026-10-22-01-old.md"] = f(legacy)
		renameMonth(m, "d.side/2026-10/2026-10-22-02-why.md")
		if errs := checkCase(m); len(errs) != 0 {
			t.Fatal(errs)
		}
	})
	t.Run("month taken date prevents flat collision", func(t *testing.T) {
		m := goodCase()
		modern := monthCase()["after/"+monthSide]
		m["before/"+monthSide] = modern
		m["after/"+monthSide] = modern
		old := "after/d.side/01-why.md"
		m["after/d.side/2026-10-22-01-why.md"] = m[old]
		delete(m, old)
		m["after/d.md"] = f(strings.ReplaceAll(string(m["after/d.md"].Data), "01-why.md", "2026-10-22-01-why.md"))
		errs := strings.Join(checkCase(m), "\n")
		for _, want := range []string{"named 2026-10-22-02-why.md", "should have been relinked"} {
			if !strings.Contains(errs, want) {
				t.Errorf("errors %q, want %q", errs, want)
			}
		}
	})
}

func TestMonthSideDocCanonicalDefaults(t *testing.T) {
	legacyParent := func() fstest.MapFS {
		m := monthCase()
		for _, p := range []string{"before/d.md", "after/d.md"} {
			m[p] = &fstest.MapFile{Data: []byte("---\nprocess-instruction: \"[review:direct]\"\n---\n" + string(m[p].Data))}
		}
		return m
	}
	t.Run("legacy parent does not pass down composite alias", func(t *testing.T) {
		if errs := checkCase(legacyParent()); len(errs) != 0 {
			t.Fatal(errs)
		}
	})
	t.Run("existing month alias stays readable", func(t *testing.T) {
		m := legacyParent()
		body := strings.Replace(string(m["after/"+monthSide].Data), "type: side-doc", "type: side-doc\nprocess-instruction: \"[review:direct]\"", 1)
		m["before/"+monthSide] = &fstest.MapFile{Data: []byte(body)}
		m["after/"+monthSide] = m["before/"+monthSide]
		if errs := checkCase(m); len(errs) != 0 {
			t.Fatal(errs)
		}
	})
	t.Run("copied composite alias rejected", func(t *testing.T) {
		m := legacyParent()
		m["after/"+monthSide] = &fstest.MapFile{Data: []byte(strings.Replace(string(m["after/"+monthSide].Data), "type: side-doc", "type: side-doc\nprocess-instruction: \"[review:direct]\"", 1))}
		if errs := strings.Join(checkCase(m), "\n"); !strings.Contains(errs, "legacy process-instruction") {
			t.Fatalf("errors %q, want legacy process-instruction rejection", errs)
		}
	})
	t.Run("alias cannot replace canonical fields", func(t *testing.T) {
		m := legacyParent()
		body := string(m["after/"+monthSide].Data)
		body = strings.Replace(body, "presentation-mode: highlight\nprocessing-mode: free-form\n", "process-instruction: \"[review:direct]\"\n", 1)
		m["after/"+monthSide] = &fstest.MapFile{Data: []byte(body)}
		errs := strings.Join(checkCase(m), "\n")
		for _, want := range []string{"presentation-mode", "processing-mode"} {
			if !strings.Contains(errs, want) {
				t.Errorf("errors %q, want explicit %s rejection", errs, want)
			}
		}
	})
	t.Run("new side doc starts with default modes", func(t *testing.T) {
		m := monthCase()
		m["after/"+monthSide] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(m["after/"+monthSide].Data), "presentation-mode: highlight", "presentation-mode: direct"))}
		if errs := strings.Join(checkCase(m), "\n"); !strings.Contains(errs, "new side doc presentation-mode") {
			t.Fatalf("errors %q, want default rejection", errs)
		}
	})
	t.Run("explicit processing instruction is allowed", func(t *testing.T) {
		m := legacyParent()
		m["after/"+monthSide] = &fstest.MapFile{Data: []byte(strings.Replace(string(m["after/"+monthSide].Data), "type: side-doc", "type: side-doc\nprocessing-instruction: \"Compare the alternatives\"", 1))}
		if errs := checkCase(m); len(errs) != 0 {
			t.Fatal(errs)
		}
	})
}

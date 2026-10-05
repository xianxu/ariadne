package layergraph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// layer makes dir a layer (base.manifest) declaring deps ("" for none).
func layer(t *testing.T, dir, deps string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "construct", "base.manifest"), "prose AGENTS.local.md\n")
	if deps != "" {
		writeFile(t, filepath.Join(dir, "construct", "deps"), deps)
	}
}

func summary(ds []DeclaredSubstrate) string {
	var out []string
	for _, d := range ds {
		out = append(out, filepath.Base(d.Path)+"<"+filepath.Base(d.Owner)+" present="+map[bool]string{true: "y", false: "n"}[d.Present]+" src="+d.Source)
	}
	return strings.Join(out, "; ")
}

// #294: every declared substrate, present or absent, transitively over present
// layers, deduplicated by path in discovery order, with its owner and source.
func TestDeclaredSubstrates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(p string)
		want  string
	}{
		{"transitive present chain", func(p string) {
			layer(t, filepath.Join(p, "d"), "substrate ../m\n")
			layer(t, filepath.Join(p, "m"), "substrate ../b https://x/b.git\n")
			layer(t, filepath.Join(p, "b"), "")
		}, "m<d present=y src=; b<m present=y src=https://x/b.git"},
		{"absent substrate reported, not descended", func(p string) {
			layer(t, filepath.Join(p, "d"), "substrate ../gone https://x/gone.git\n")
		}, "gone<d present=n src=https://x/gone.git"},
		{"unresolvable parent skipped", func(p string) {
			layer(t, filepath.Join(p, "d"), "substrate ../nowhere/x\nsubstrate ../b\n")
			layer(t, filepath.Join(p, "b"), "")
		}, "b<d present=y src="},
		{"dedup across owners, first owner kept", func(p string) {
			layer(t, filepath.Join(p, "d"), "substrate ../a\nsubstrate ../b\n")
			layer(t, filepath.Join(p, "a"), "substrate ../c\n")
			layer(t, filepath.Join(p, "b"), "substrate ../c\n")
			layer(t, filepath.Join(p, "c"), "")
		}, "a<d present=y src=; b<d present=y src=; c<a present=y src="},
		{"the root is never its own substrate", func(p string) {
			layer(t, filepath.Join(p, "d"), "substrate ../a\n")
			layer(t, filepath.Join(p, "a"), "substrate ../d\n")
		}, "a<d present=y src="},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := canon(t, t.TempDir())
			tc.build(p)
			got, err := DeclaredSubstrates(OSFS{}, filepath.Join(p, "d"))
			if err != nil {
				t.Fatal(err)
			}
			if summary(got) != tc.want {
				t.Fatalf("got  %s\nwant %s", summary(got), tc.want)
			}
			for _, d := range got {
				if !filepath.IsAbs(d.Path) || !filepath.IsAbs(d.Owner) {
					t.Fatalf("paths must be absolute: %+v", d)
				}
			}
		})
	}
}

// #294: the errors Walk raises are DeclaredSubstrates' too.
func TestDeclaredSubstratesErrors(t *testing.T) {
	p := canon(t, t.TempDir())
	layer(t, filepath.Join(p, "d"), "substrate\n") // malformed row
	if _, err := DeclaredSubstrates(OSFS{}, filepath.Join(p, "d")); err == nil || !strings.Contains(err.Error(), "construct/deps line 1") {
		t.Fatalf("malformed: %v", err)
	}
	q := canon(t, t.TempDir())
	layer(t, filepath.Join(q, "d"), "substrate ../notalayer\n")
	writeFile(t, filepath.Join(q, "notalayer", "README.md"), "present, no manifest")
	if _, err := DeclaredSubstrates(OSFS{}, filepath.Join(q, "d")); err == nil || !strings.Contains(err.Error(), "base.manifest") {
		t.Fatalf("present without manifest: %v", err)
	}
}

// #294: Walk and DeclaredSubstrates are one traversal: Walk's layers (minus the
// root itself) are exactly the present substrates DeclaredSubstrates reports.
func TestWalkIsThePresentSubsetOfDeclaredSubstrates(t *testing.T) {
	p := canon(t, t.TempDir())
	layer(t, filepath.Join(p, "d"), "substrate ../a\nsubstrate ../b\nsubstrate ../gone\n")
	layer(t, filepath.Join(p, "a"), "substrate ../c\n")
	layer(t, filepath.Join(p, "b"), "substrate ../c\nsubstrate ../also-gone\n")
	layer(t, filepath.Join(p, "c"), "")
	root := filepath.Join(p, "d")
	roots, err := Walk(OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	declared, err := DeclaredSubstrates(OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	walked := map[string]bool{}
	for _, r := range roots {
		if r != root {
			walked[r] = true
		}
	}
	present := map[string]bool{}
	for _, d := range declared {
		if d.Present {
			present[d.Path] = true
		}
	}
	if len(walked) != len(present) || len(present) != 3 {
		t.Fatalf("walk %v vs present declared %v", walked, present)
	}
	for r := range walked {
		if !present[r] {
			t.Fatalf("%s walked but not declared present", r)
		}
	}
}

// plainFS offers only the FS seam (no safe declaration reader), like weave's
// fs and test fakes.
type plainFS struct{ OSFS }

func (plainFS) ReadDeclaration() {} // shadows OSFS's: not a DeclarationReader

// #294: construct/deps reads are bounded whichever FS reads them.
func TestDeclarationReadsAreBounded(t *testing.T) {
	p := canon(t, t.TempDir())
	layer(t, filepath.Join(p, "d"), "# "+strings.Repeat("x", int(DeclarationLimit))+"\n")
	for name, fs := range map[string]FS{"OSFS": OSFS{}, "plain FS": plainFS{}} {
		if _, err := DeclaredSubstrates(fs, filepath.Join(p, "d")); err == nil || !strings.Contains(err.Error(), "limit") {
			t.Errorf("%s: oversized declaration: %v", name, err)
		}
	}
	// OSFS refuses a symlinked declaration (the shared reader's rule).
	q := canon(t, t.TempDir())
	layer(t, filepath.Join(q, "d"), "")
	writeFile(t, filepath.Join(q, "real-deps"), "substrate ../x\n")
	if err := os.Symlink(filepath.Join(q, "real-deps"), filepath.Join(q, "d", "construct", "deps")); err != nil {
		t.Fatal(err)
	}
	if _, err := DeclaredSubstrates(OSFS{}, filepath.Join(q, "d")); err == nil || !strings.Contains(err.Error(), "ordinary file") {
		t.Errorf("symlinked declaration: %v", err)
	}
}

// #294: under OSFS a construct/deps that exists but cannot be read is an
// error for Walk and DeclaredSubstrates alike, never "no dependencies"; and a
// relative root still yields absolute paths.
func TestUnreadableDeclarationIsLoudAndRootsAreAbsolute(t *testing.T) {
	p := canon(t, t.TempDir())
	layer(t, filepath.Join(p, "d"), "")
	if err := os.MkdirAll(filepath.Join(p, "d", "construct", "deps"), 0o755); err != nil { // a directory, not a file
		t.Fatal(err)
	}
	if _, err := Walk(OSFS{}, filepath.Join(p, "d")); err == nil {
		t.Fatal("Walk read an unreadable construct/deps as none")
	}
	if _, err := DeclaredSubstrates(OSFS{}, filepath.Join(p, "d")); err == nil {
		t.Fatal("DeclaredSubstrates read an unreadable construct/deps as none")
	}

	q := canon(t, t.TempDir())
	layer(t, filepath.Join(q, "d"), "substrate ../b\n")
	layer(t, filepath.Join(q, "b"), "")
	t.Chdir(q)
	got, err := DeclaredSubstrates(OSFS{}, "d")
	if err != nil || len(got) != 1 || !filepath.IsAbs(got[0].Owner) || !filepath.IsAbs(got[0].Path) {
		t.Fatalf("relative root: %+v %v", got, err)
	}
}

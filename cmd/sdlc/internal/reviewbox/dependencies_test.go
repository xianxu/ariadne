package reviewbox

import (
	"fmt"
	"strings"
	"testing"
)

func TestDependencyManifest(t *testing.T) {
	if got := DependencyManifest(nil, true); got != "" {
		t.Fatalf("empty manifest = %q", got)
	}
	deps := []Dependency{
		{Kind: "substrate", Dir: "/review/quote\" path", Head: strings.Repeat("a", 40), FetchAge: "2h3m"},
		{Kind: "data", Dir: "/review/data", Head: strings.Repeat("b", 64)},
	}
	for _, fenced := range []bool{true, false} {
		got := DependencyManifest(deps, fenced)
		if !strings.Contains(got, fmt.Sprintf("\nSDLC_DEP_SUBSTRATE=%q\n", deps[0].Dir)) {
			t.Errorf("binding is not a standalone quoted assignment: %s", got)
		}
		for _, want := range []string{"SDLC_DEP_SUBSTRATE", "SDLC_DEP_DATA", fmt.Sprintf("%q", deps[0].Dir), deps[0].Head, deps[1].Head, "2h3m", "unknown"} {
			if !strings.Contains(got, want) {
				t.Errorf("manifest missing %q: %s", want, got)
			}
		}
		if fenced && !strings.Contains(got, "read-only") {
			t.Errorf("fenced mode absent: %s", got)
		}
		if !fenced && (!strings.Contains(got, "writable disposable") || !strings.Contains(got, "discarded")) {
			t.Errorf("disposable mode absent: %s", got)
		}
		reversed := []Dependency{deps[1], deps[0]}
		if got != DependencyManifest(reversed, fenced) {
			t.Error("manifest depends on input order")
		}
		if deps[0].Kind != "substrate" {
			t.Fatal("manifest mutated caller input")
		}
	}
}

func dependencyTreeEntry(mode, kind, oid, size, path string) string {
	return fmt.Sprintf("%s %s %s %7s\t%s\x00", mode, kind, oid, size, path)
}

func TestDependencyTreeBudgetAcceptsGitBlobModesAndRawPaths(t *testing.T) {
	var raw string
	for i, mode := range []string{"100644", "100755", "120000"} {
		raw += dependencyTreeEntry(mode, "blob", strings.Repeat("a", 40), fmt.Sprint(i), "a\tline\n"+mode)
	}
	raw += dependencyTreeEntry("100644", "blob", strings.Repeat("b", 64), "536870909", "large")
	entries, size, err := DependencyTreeBudget([]byte(raw))
	if err != nil || entries != 4 || size != 512<<20 {
		t.Fatalf("budget = %d, %d, %v", entries, size, err)
	}
	entries, size, err = DependencyTreeBudget(nil)
	if err != nil || entries != 0 || size != 0 {
		t.Fatalf("empty tree = %d, %d, %v", entries, size, err)
	}
}

func TestDependencyTreeBudgetRefusesMalformedOrUnboundedTrees(t *testing.T) {
	oid := strings.Repeat("a", 40)
	valid := dependencyTreeEntry("100644", "blob", oid, "1", "file")
	cases := map[string]string{
		"unterminated":     strings.TrimSuffix(valid, "\x00"),
		"missing metadata": "file\x00",
		"missing path":     dependencyTreeEntry("100644", "blob", oid, "1", ""),
		"gitlink":          dependencyTreeEntry("160000", "commit", oid, "-", "submodule"),
		"tree":             dependencyTreeEntry("040000", "tree", oid, "-", "directory"),
		"wrong mode":       dependencyTreeEntry("100664", "blob", oid, "1", "file"),
		"wrong kind":       dependencyTreeEntry("100644", "commit", oid, "1", "file"),
		"short oid":        dependencyTreeEntry("100644", "blob", "abc", "1", "file"),
		"nonhex oid":       dependencyTreeEntry("100644", "blob", strings.Repeat("z", 40), "1", "file"),
		"negative size":    dependencyTreeEntry("100644", "blob", oid, "-1", "file"),
		"signed size":      dependencyTreeEntry("100644", "blob", oid, "+1", "file"),
		"missing size":     dependencyTreeEntry("100644", "blob", oid, "-", "file"),
		"overflow size":    dependencyTreeEntry("100644", "blob", oid, "18446744073709551616", "file"),
		"oversize blob":    dependencyTreeEntry("100644", "blob", oid, "536870913", "file"),
		"aggregate size":   dependencyTreeEntry("100644", "blob", oid, "536870912", "file") + valid,
		"excess entries":   strings.Repeat(valid, 100001),
		"excess metadata":  strings.Repeat("x", (32<<20)+1),
		"empty record":     "\x00",
		"extra metadata":   "100644 blob " + oid + " 1 extra\tfile\x00",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := DependencyTreeBudget([]byte(raw)); err == nil {
				t.Fatal("accepted invalid dependency tree")
			}
		})
	}
}

func TestDependencyTreeBudgetAcceptsEntryBound(t *testing.T) {
	raw := strings.Repeat(dependencyTreeEntry("100644", "blob", strings.Repeat("c", 40), "0", "file"), 100000)
	entries, size, err := DependencyTreeBudget([]byte(raw))
	if err != nil || entries != 100000 || size != 0 {
		t.Fatalf("at entry limit = %d, %d, %v", entries, size, err)
	}
}

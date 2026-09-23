package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddressGrammar(t *testing.T) {
	for _, s := range []string{"repo", "repo:0", "repo:12", ":1", "a-b.c_d:2", "repo name:2", "仓库:3", "-repo:4"} {
		a, err := ParseAddress(s)
		if err != nil {
			t.Fatal(s, err)
		}
		b, err := ParseAddress(a.String())
		if err != nil || a != b {
			t.Fatalf("roundtrip %q: %+v %v", s, b, err)
		}
	}
	for _, s := range []string{"", ":", "../x:1", "a/b:1", "a:01", "a:-1", "a:+1", "a:99999999999999999999999999999", "a:1:2", "a\nb:1", ".:1", "..:1", "a\\b:1"} {
		if _, err := ParseAddress(s); err == nil {
			t.Errorf("accepted %q", s)
		}
	}
}
func FuzzAddress(f *testing.F) {
	for _, s := range []string{"repo", ":0", "a:01", "../x:1", "r:42"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		a, e := ParseAddress(s)
		if e != nil {
			return
		}
		b, e := ParseAddress(a.String())
		if e != nil || a != b {
			t.Fatal(a, b, e)
		}
	})
}
func TestSlotPathContainment(t *testing.T) {
	seen := map[string]bool{}
	for n := 1; n < 1000; n++ {
		p, e := SlotPath("/fleet", "repo", n)
		if e != nil || filepath.Dir(p) != "/fleet/worktree" || seen[p] {
			t.Fatal(p, e)
		}
		seen[p] = true
	}
	if _, e := SlotPath("/fleet", "../escape", 1); e == nil {
		t.Fatal("escape accepted")
	}
}
func TestClassify(t *testing.T) {
	v := Vantage{RepoIdentity: "/fleet/repo/.git", PrimaryRoot: "/fleet/repo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/worktree/repo-slot1"}
	ws := []Worktree{{Path: v.PrimaryRoot, Branch: "main", HEAD: strings.Repeat("a", 40)}, {Path: v.WorktreeRoot, Branch: "issue", HEAD: strings.Repeat("b", 40)}}
	id, e := Classify(v, ws, map[string]string{"main-slot1": strings.Repeat("a", 40)})
	if e != nil || id.Kind != "slot" || *id.Address != "repo:1" || *id.RestingBranch != "main-slot1" {
		t.Fatal(id, e)
	}
	ws = append(ws, Worktree{Path: "/elsewhere", Branch: "other", HEAD: strings.Repeat("a", 40)})
	if _, e := Classify(v, ws, map[string]string{"main-slot1": strings.Repeat("a", 40)}); e != nil {
		t.Fatal(e)
	}
	ws[2].Branch = "main-slot1"
	if _, e := Classify(v, ws, map[string]string{"main-slot1": strings.Repeat("a", 40)}); e == nil {
		t.Fatal("occupied baseline accepted")
	}
	ws = ws[:2]
	ws[1].Branch = "main"
	if _, e := Classify(v, ws, map[string]string{"main-slot1": strings.Repeat("a", 40)}); e == nil {
		t.Fatal("wrong reserved branch accepted")
	}
	ws[1].Branch = "issue"
	if _, e := Classify(v, ws, nil); e == nil {
		t.Fatal("missing baseline accepted")
	}
	ws = append(ws, ws[1])
	if _, e := Classify(v, ws, map[string]string{"main-slot1": strings.Repeat("a", 40)}); e == nil {
		t.Fatal("duplicate membership accepted")
	}
}

func TestClassifyTopologyPermutations(t *testing.T) {
	v := Vantage{RepoIdentity: "/fleet/repo/.git", PrimaryRoot: "/fleet/repo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/worktree/repo-slot1"}
	trees := []Worktree{{Path: v.PrimaryRoot, Branch: "main", HEAD: strings.Repeat("a", 40)}, {Path: v.WorktreeRoot, Branch: "issue", HEAD: strings.Repeat("b", 40)}, {Path: "/ordinary", Branch: "other", HEAD: strings.Repeat("a", 40)}}
	refs := map[string]string{"main-slot1": strings.Repeat("a", 40)}
	want, e := Classify(v, trees, refs)
	if e != nil {
		t.Fatal(e)
	}
	for i := range trees {
		for j := range trees {
			trees[i], trees[j] = trees[j], trees[i]
			got, e := Classify(v, trees, refs)
			if e != nil || got.RepoIdentity != want.RepoIdentity || *got.Address != *want.Address || *got.Branch != *want.Branch {
				t.Fatal(got, e)
			}
			trees[i], trees[j] = trees[j], trees[i]
		}
	}
}

func TestClassifyRejectsUnaddressablePrimary(t *testing.T) {
	v := Vantage{RepoIdentity: "/fleet/bad\nrepo/.git", PrimaryRoot: "/fleet/bad\nrepo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/bad\nrepo"}
	if _, e := Classify(v, []Worktree{{Path: v.PrimaryRoot, Branch: "main", HEAD: strings.Repeat("a", 40)}}, nil); e == nil {
		t.Fatal("unaddressable primary accepted")
	}
}

func TestClassifyRejectsMalformedOID(t *testing.T) {
	malformed := []string{"", "0", strings.Repeat("0", 39), strings.Repeat("0", 41), strings.Repeat("0", 63), strings.Repeat("0", 65), "abc", strings.Repeat("a", 39), strings.Repeat("a", 65), strings.Repeat("g", 40), strings.Repeat("A", 40), strings.Repeat("a", 39) + "\n"}
	for _, field := range []string{"primary HEAD", "ordinary HEAD", "slot HEAD", "unrelated HEAD", "resting ref"} {
		for _, oid := range malformed {
			t.Run(fmt.Sprintf("%s/%q", field, oid), func(t *testing.T) {
				v := Vantage{RepoIdentity: "/fleet/repo/.git", PrimaryRoot: "/fleet/repo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/repo"}
				trees := []Worktree{{Path: v.PrimaryRoot, Branch: "main", HEAD: strings.Repeat("a", 40)}, {Path: "/fleet/worktree/repo-slot1", Branch: "issue", HEAD: strings.Repeat("b", 40)}, {Path: "/ordinary", Branch: "other", HEAD: strings.Repeat("c", 40)}}
				refs := map[string]string{"main-slot1": strings.Repeat("a", 40)}
				switch field {
				case "primary HEAD":
					trees[0].HEAD = oid
				case "ordinary HEAD":
					v.WorktreeRoot = trees[2].Path
					trees[2].HEAD = oid
				case "slot HEAD":
					v.WorktreeRoot = trees[1].Path
					trees[1].HEAD = oid
				case "unrelated HEAD":
					trees[2].HEAD = oid
				case "resting ref":
					v.WorktreeRoot = trees[1].Path
					refs["main-slot1"] = oid
				}
				if id, e := Classify(v, trees, refs); e == nil {
					t.Fatalf("accepted malformed %s %q: %+v", field, oid, id)
				}
			})
		}
	}
}
func TestClassifyFullOIDForms(t *testing.T) {
	for _, length := range []int{40, 64} {
		for _, zero := range []bool{false, true} {
			t.Run(fmt.Sprintf("length=%d/zero=%v", length, zero), func(t *testing.T) {
				oid := strings.Repeat("a", length)
				if zero {
					oid = strings.Repeat("0", length)
				}
				v := Vantage{RepoIdentity: "/fleet/repo/.git", PrimaryRoot: "/fleet/repo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/repo"}
				id, e := Classify(v, []Worktree{{Path: v.PrimaryRoot, Branch: "main", HEAD: oid}}, nil)
				if e != nil || (id.Head == nil) != zero {
					t.Fatal(id, e)
				}
			})
		}
	}
}
func TestClassifyRejectsZeroRestingCommit(t *testing.T) {
	for _, length := range []int{40, 64} {
		v := Vantage{RepoIdentity: "/fleet/repo/.git", PrimaryRoot: "/fleet/repo", FleetRoot: "/fleet", WorktreeRoot: "/fleet/worktree/repo-slot1"}
		if _, e := Classify(v, []Worktree{{Path: v.WorktreeRoot, Branch: "issue", HEAD: strings.Repeat("a", length)}}, map[string]string{"main-slot1": strings.Repeat("0", length)}); e == nil {
			t.Fatal("zero resting commit accepted")
		}
	}
}

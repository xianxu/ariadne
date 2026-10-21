package main

import (
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"reflect"
	"testing"
)

func TestLandingCheckSelection(t *testing.T) {
	d := localCheckDecl{Version: 2, Checks: []localCheck{{Name: "vm", RunScope: "landing", Resource: "runner", Required: true, PathsFile: "paths.json"}, {Name: "disabled", RunScope: "landing"}, {Name: "suite", Required: true}}}
	policies := map[string][]string{"paths.json": {"src/**", "!src/docs/**", "src/docs/critical.md"}}
	for _, tc := range []struct {
		paths [][]string
		want  []string
	}{
		{[][]string{{"src/a.go"}, {"README.md"}}, []string{"vm"}},
		{[][]string{{"README.md"}}, nil},
		{[][]string{{"src/docs/a.md"}}, nil},
		{[][]string{{"src/docs/critical.md"}}, []string{"vm"}},
	} {
		got, err := selectLandingChecks(d, policies, tc.paths)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, c := range got {
			names = append(names, c.Name)
		}
		if !reflect.DeepEqual(names, tc.want) {
			t.Fatalf("%v => %v want %v", tc.paths, names, tc.want)
		}
	}
	if _, err := selectLandingChecks(d, nil, [][]string{{"src/a.go"}}); err == nil {
		t.Fatal("missing policy admitted")
	}
}

func TestLandingCheckPaths(t *testing.T) {
	for _, raw := range []string{`{"version":1,"paths":[]}`, `{"version":1,"paths":["!src/**"]}`, `{"version":1,"paths":["["]}`, `{"version":1,"paths":["../x"]}`, `{"version":1,"paths":null}`, `{"version":1,"paths":["a"],"PATHS":["b"]}`} {
		if _, err := parseLandingPaths([]byte(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	p, err := parseLandingPaths([]byte(`{"version":1,"paths":["src/**","!src/docs/**"],"future":1e9999}`))
	if err != nil || len(p.Paths) != 2 {
		t.Fatalf("%+v %v", p, err)
	}
}

func TestLandingCheckProofBinding(t *testing.T) {
	p := LandingCheckProof{Head: "head", Policy: "policy", Cars: "cars", Run: localRunRecord{Name: "vm", SHA: "head", Exit: 0, LogPath: "retained", LogSHA256: "digest"}}
	if !landingProofMatches(p, "vm", "head", "policy", "cars") {
		t.Fatal("bound pass rejected")
	}
	for _, args := range [][4]string{{"other", "head", "policy", "cars"}, {"vm", "new-head", "policy", "cars"}, {"vm", "head", "new-policy", "cars"}, {"vm", "head", "policy", "new-cars"}} {
		if landingProofMatches(p, args[0], args[1], args[2], args[3]) {
			t.Fatalf("accepted wrong binding %v", args)
		}
	}
	p.Run.Exit = 37
	if landingProofMatches(p, "vm", "head", "policy", "cars") {
		t.Fatal("failed proof accepted")
	}
}

func TestLandingCheckMatchingCounted(t *testing.T) {
	d := localCheckDecl{Checks: []localCheck{{Name: "vm", Required: true, RunScope: "landing", PathsFile: "paths"}}}
	for _, n := range []int{8, 16} {
		calls := 0
		paths := []string{}
		for i := 0; i < n; i++ {
			paths = append(paths, fmt.Sprintf("docs/%d.md", i))
		}
		_, err := selectLandingChecksWithMatch(d, map[string][]string{"paths": {"src/**", "lib/**"}}, [][]string{paths, paths}, func(pattern, name string) (bool, error) { calls++; return affected.Match(pattern, name) })
		if err != nil || calls != 2*n {
			t.Fatalf("paths%d calls%d err%v", n, calls, err)
		}
	}
}

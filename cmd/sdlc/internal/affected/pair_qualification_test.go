package affected

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestPairQualification (#598): pair's real affected.json (testdata/pair-affected.json, pair 93a57a380) with the
// edits pair#853 makes — tests/** dropped, cmd/couch/** keeping its production code full but excepting its tests, the
// reader-certificate convention declared — over a pair-shaped graph, with pair's real request_lifetime certificate
// (pair#825's) as the sidecar.
func TestPairQualification(t *testing.T) {
	raw, err := os.ReadFile("testdata/pair-affected.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseConfig(raw); err != nil {
		t.Fatalf("pair's affected.json today: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	var fulls []any
	for _, f := range doc["full"].([]any) {
		switch f {
		case "tests/**":
		case "cmd/couch/**":
			fulls = append(fulls, map[string]any{"glob": "cmd/couch/**", "except": []string{"cmd/couch/*_test.go"}})
		default:
			fulls = append(fulls, f)
		}
	}
	doc["full"] = fulls
	doc["sidecars"] = []map[string]string{{"prefix": "tests/affected-readers/", "suffix": ".json"}}
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfig(edited)
	if err != nil {
		t.Fatalf("pair's edited affected.json: %v", err)
	}
	cert, err := os.ReadFile("testdata/pair-request_lifetime_test.go.json")
	if err != nil {
		t.Fatal(err)
	}
	const certPath = "tests/affected-readers/cmd/internal/testcouch/agentfixture/request_lifetime_test.go.json"
	mp, err := Compile(cfg, Sources{Sidecars: []SidecarFile{{Path: certPath, Content: cert}}})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	const m = "github.com/xianxu/pair/"
	dirs := []string{"cmd/internal/testcouch/agentfixture", "cmd/internal/testcouch", "cmd/couch", "cmd/internal/other"}
	for _, gr := range cfg.Groups {
		dirs = append(dirs, gr.Go...)
	}
	g := Graph{Module: "github.com/xianxu/pair", Dir: map[string]string{}, Deps: map[string]map[string]bool{}}
	for _, d := range dirs {
		d = goRootDir(d)
		g.Dir[m+d] = d
		g.Deps[m+d] = map[string]bool{}
	}
	// testcouch builds the agent fixture; couch's acceptance tests use testcouch.
	g.Deps[m+"cmd/internal/testcouch"] = map[string]bool{m + "cmd/internal/testcouch/agentfixture": true}
	g.Deps[m+"cmd/couch"] = map[string]bool{m + "cmd/internal/testcouch": true, m + "cmd/internal/testcouch/agentfixture": true}
	agentUp := []string{m + "cmd/couch", m + "cmd/internal/testcouch", m + "cmd/internal/testcouch/agentfixture"}
	cases := []struct {
		name  string
		paths []string
		want  Selection
	}{
		{"pair#825's certificate selects its test's package and dependents", []string{certPath}, Selection{GoPkgs: agentUp, Groups: []string{}}},
		{"an arbitrary new certificate selects its own test's package", []string{"tests/affected-readers/cmd/internal/other/x_test.go.json"}, Selection{GoPkgs: []string{m + "cmd/internal/other"}, Groups: []string{}}},
		{"a certificate for a package that is not there is full, never nothing", []string{"tests/affected-readers/cmd/gone/x_test.go.json"},
			full("affected.json sidecars[0]: target cmd/gone/x_test.go of tests/affected-readers/cmd/gone/x_test.go.json claims nothing")},
		{"a couch test edit selects couch", []string{"cmd/couch/editorial_acceptance_test.go"}, Selection{GoPkgs: []string{m + "cmd/couch"}, Groups: []string{}}},
		{"couch production code stays full", []string{"cmd/couch/main.go"}, full("cmd/couch/main.go")},
		{"couch's own testdata stays full", []string{"cmd/couch/testdata/slash-journal/a.json"}, full("cmd/couch/testdata/slash-journal/a.json")},
		{"the selector stays full", []string{"scripts/test-affected.py"}, full("scripts/test-affected.py")},
		{"the policy stays full", []string{".sdlc/affected.json"}, full(".sdlc/affected.json")},
		{"the shared harness stays full", []string{"tests/lib/with-pair-test-binary.sh"}, full("tests/lib/with-pair-test-binary.sh")},
		{"the inventory checker stays full", []string{"tests/affected-config-test.py"}, full("tests/affected-config-test.py")},
		{"an undeclared tests/ path is unclassified, so full", []string{"tests/new-fixture.txt"}, full("unclassified: tests/new-fixture.txt")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Plan(c.paths, g, cfg, mp, isDocTest); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Plan(%q)\n got  %+v\n want %+v", c.paths, got, c.want)
			}
		})
	}
}

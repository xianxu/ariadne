package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

// #598: declared ownership through the real affected run. The sidecars are read from main's tree, never the head's.

const declaredFixtureConfig = `{"version":1,"command":["make","test-affected"],
	"full":["go.mod","Makefile*"],"groups":[{"name":"lua","paths":["**/*.lua"]}],
	"sidecars":[{"prefix":"tests/readers/","suffix":".json"}]}`

// declaredFixture is affectedFixture with sidecars committed on main (and merged into the branch).
func declaredFixture(t *testing.T, sidecars map[string]string) (localRunDeps, func(map[string]string) string, string, string, affected.Config) {
	t.Helper()
	d, _, change, out, marker := affectedFixture(t)
	must := func(args ...string) {
		t.Helper()
		if o, err := d.Git(args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, o)
		}
	}
	must("switch", "-q", "main")
	for rel, body := range sidecars {
		p := filepath.Join(d.Root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("add", "-A")
	must("commit", "-q", "-m", "sidecars")
	must("switch", "-q", "br")
	must("merge", "-q", "--ff-only", "main")
	cfg, err := affected.ParseConfig([]byte(declaredFixtureConfig))
	if err != nil {
		t.Fatal(err)
	}
	return d, change, out, marker, cfg
}

func runDeclaredFixture(t *testing.T, d localRunDeps, cfg affected.Config, sha string) localRunRecord {
	t.Helper()
	rec, err := runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha, affectedOrdinary)
	if err != nil {
		t.Fatalf("%v\n%s", err, d.Stderr)
	}
	return rec
}

func TestAffectedDeclaredSidecarsFromMain(t *testing.T) {
	cases := []struct {
		name    string
		changed map[string]string
		want    string // the consumer's env; "" means the suite ran instead
	}{
		{"a declared read selects its reader", map[string]string{"fixtures/mid/a.txt": "x\n"}, "example.com/m/mid\n\n"},
		{"an edited sidecar selects its test's package", map[string]string{"tests/readers/mid/mid_test.go.json": `{"reads":["fixtures/mid/**"],"note":1}`}, "example.com/m/mid\n\n"},
		{"a branch-added sidecar is path-only", map[string]string{"tests/readers/tonly/tonly_test.go.json": `{"reads":["fixtures/tonly/**"]}`}, "example.com/m/tonly\n\n"},
		// The head's own declaration is not trusted: main declares nothing for this input, so it is unclassified.
		{"a read only the head declares does not count", map[string]string{
			"tests/readers/tonly/tonly_test.go.json": `{"reads":["fixtures/tonly/**"]}`, "fixtures/tonly/a.txt": "x\n"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, change, out, marker, cfg := declaredFixture(t, map[string]string{
				"tests/readers/mid/mid_test.go.json": `{"version":1,"disposition":"local_or_literal_fixture","reads":["fixtures/mid/**"]}`,
			})
			sha := change(c.changed)
			rec := runDeclaredFixture(t, d, cfg, sha)
			if c.want == "" {
				if rec.Name != suiteCheckName || !exists(marker) || exists(out) {
					t.Fatalf("want the suite: %+v\n%s", rec, d.Stderr)
				}
				return
			}
			got, err := os.ReadFile(out)
			if err != nil || string(got) != c.want || exists(marker) {
				t.Fatalf("env %q (%v), want %q; suite ran %v\n%s", got, err, c.want, exists(marker), d.Stderr)
			}
		})
	}
}

// An unreadable sidecar on main narrows nothing: the selection is full, naming it, and the suite runs.
func TestAffectedDeclaredMalformedSidecarRunsTheSuite(t *testing.T) {
	d, change, out, marker, cfg := declaredFixture(t, map[string]string{"tests/readers/mid/mid_test.go.json": `{"reads": "fixtures"}`})
	sha := change(map[string]string{"leaf/leaf.go": "package leaf\n\nfunc F() int { return 3 }\n"})
	if rec := runDeclaredFixture(t, d, cfg, sha); rec.Name != suiteCheckName || !exists(marker) || exists(out) {
		t.Fatalf("%+v\n%s", rec, d.Stderr)
	}
	if !strings.Contains(fmt.Sprint(d.Stderr), "declared ownership unavailable: sidecar tests/readers/mid/mid_test.go.json") {
		t.Fatalf("the fallback does not name the sidecar:\n%s", d.Stderr)
	}
}

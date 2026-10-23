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

// affectedFixture is localRunRepo plus, on main, a module (leaf ← mid, leaf
// ← tonly's tests, other) and a Makefile whose test-affected target is a fake
// consumer of the env contract: it writes both variables to $OUT. Its test
// target, the suite, touches $SUITE_MARKER. change commits files on a branch
// and returns the pushed head.
func affectedFixture(t *testing.T) (d localRunDeps, gh *ciFakeGH, change func(files map[string]string) string, out, marker string) {
	t.Helper()
	d, _, gh, _ = localRunRepo(t, "acquired")
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(d.Root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must := func(args ...string) string {
		t.Helper()
		o, err := d.Git(args...)
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, o)
		}
		return strings.TrimSpace(o)
	}
	write("go.mod", "module example.com/m\n\ngo 1.22\n")
	write("leaf/leaf.go", "package leaf\n\nfunc F() int { return 1 }\n")
	write("mid/mid.go", "package mid\n\nimport \"example.com/m/leaf\"\n\nfunc G() int { return leaf.F() }\n")
	write("tonly/tonly.go", "package tonly\n")
	write("tonly/tonly_test.go", "package tonly\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/leaf\"\n)\n\nfunc TestT(t *testing.T) { _ = leaf.F() }\n")
	write("other/other.go", "package other\n")
	write("x.lua", "return 1\n")
	write("Makefile", "test-affected:\n\tprintf '%s\\n%s\\n' \"$$SDLC_AFFECTED_GO_PKGS\" \"$$SDLC_AFFECTED_GROUPS\" > \"$$OUT\"\n\ntest:\n\ttouch \"$$SUITE_MARKER\"\n")
	must("add", "-A")
	must("commit", "-q", "-m", "fixture")
	must("switch", "-q", "-c", "br")
	out, marker = filepath.Join(t.TempDir(), "out"), filepath.Join(t.TempDir(), "suite-ran")
	t.Setenv("OUT", out)
	t.Setenv("SUITE_MARKER", marker)
	change = func(files map[string]string) string {
		t.Helper()
		for rel, body := range files {
			write(rel, body)
		}
		must("add", "-A")
		must("commit", "-q", "-m", "change")
		sha := must("rev-parse", "HEAD")
		gh.setHead(9, sha)
		return sha
	}
	return d, gh, change, out, marker
}

func affectedFixtureConfig(t *testing.T) affected.Config {
	t.Helper()
	cfg, err := affected.ParseConfig([]byte(`{"version":1,"command":["make","test-affected"],
		"full":["go.mod","Makefile*"],"groups":[{"name":"lua","paths":["**/*.lua"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

var affectedFixtureSuite = localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, Required: true, TimeoutMinutes: 30}

func runAffectedFixture(t *testing.T, d localRunDeps, sha string) localRunRecord {
	t.Helper()
	rec, err := runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", affectedFixtureConfig(t)), affectedFixtureConfig(t), affectedFixtureSuite, sha, affectedOrdinary)
	if err != nil {
		t.Fatalf("%v\n%s", err, d.Stderr)
	}
	return rec
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// #473: the env contract, through a fake consumer: a leaf change and a Lua
// change select the leaf, its dependents (a test-only one included) and the
// lua group, and nothing else; the record names the selection.
func TestAffectedContractFixture(t *testing.T) {
	d, gh, change, out, marker := affectedFixture(t)
	sha := change(map[string]string{"leaf/leaf.go": "package leaf\n\nfunc F() int { return 2 }\n", "x.lua": "return 2\n"})
	if rec := runAffectedFixture(t, d, sha); !rec.passed() {
		t.Fatalf("%+v\n%s", rec, d.Stderr)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("the consumer never ran: %v\n%s", err, d.Stderr)
	}
	if want := "example.com/m/leaf example.com/m/mid example.com/m/tonly\nlua\n"; string(got) != want {
		t.Fatalf("contract env %q, want %q", got, want)
	}
	if gh.statuses[sha]["sdlc-local/affected"] != "success" || !strings.HasPrefix(gh.descs[sha]["sdlc-local/affected"], "3 packages, groups lua · ") {
		t.Fatalf("statuses %v descs %v", gh.statuses[sha], gh.descs[sha])
	}
	if _, ok := gh.statuses[sha]["sdlc-local/suite"]; ok || exists(marker) {
		t.Fatal("a subset selection ran the suite")
	}
}

// #473 D5: a full selection runs and records suite, never affected.
func TestAffectedFullSelectionRunsTheSuite(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"go.mod changed":           {"go.mod": "module example.com/m\n\ngo 1.23\n"},
		"unclassified path":        {"tools/x.py": "print(1)\n"},
		"package graph unloadable": {"leaf/leaf.go": "not go\n"},
	} {
		t.Run(name, func(t *testing.T) {
			d, gh, change, out, marker := affectedFixture(t)
			sha := change(files)
			if rec := runAffectedFixture(t, d, sha); rec.Name != suiteCheckName || !rec.passed() {
				t.Fatalf("%+v\n%s", rec, d.Stderr)
			}
			if !exists(marker) || exists(out) || gh.statuses[sha]["sdlc-local/suite"] != "success" {
				t.Fatalf("suite ran %v, consumer ran %v, statuses %v\n%s", exists(marker), exists(out), gh.statuses[sha], d.Stderr)
			}
			if _, ok := gh.statuses[sha]["sdlc-local/affected"]; ok {
				t.Fatal("a full selection was recorded as affected")
			}
		})
	}
}

// #473: a docs-only change records affected without running anything.
func TestAffectedDocsOnlyRecordsWithoutRunning(t *testing.T) {
	d, gh, change, out, marker := affectedFixture(t)
	sha := change(map[string]string{"README.md": "# m\n"})
	if rec := runAffectedFixture(t, d, sha); !rec.passed() {
		t.Fatalf("%+v", rec)
	}
	if exists(out) || exists(marker) {
		t.Fatal("a docs-only change ran a command")
	}
	if gh.statuses[sha]["sdlc-local/affected"] != "success" || !strings.HasPrefix(gh.descs[sha]["sdlc-local/affected"], "nothing affected (docs-only)") {
		t.Fatalf("statuses %v descs %v", gh.statuses[sha], gh.descs[sha])
	}
}

// #473 close review: a doc a package owns is not docs-only: the run loads
// the graph and tests that package and its dependents.
func TestAffectedOwnedDocRunsItsPackage(t *testing.T) {
	d, gh, change, out, _ := affectedFixture(t)
	sha := change(map[string]string{"leaf/NOTES.md": "# notes\n"})
	if rec := runAffectedFixture(t, d, sha); !rec.passed() || gh.statuses[sha]["sdlc-local/affected"] != "success" {
		t.Fatalf("%+v %v", rec, gh.statuses[sha])
	}
	if got, err := os.ReadFile(out); err != nil || string(got) != "example.com/m/leaf example.com/m/mid example.com/m/tonly\n\n" {
		t.Fatalf("contract env %q %v", got, err)
	}
}

// #473 M1 BR-2: a head with no change against main is not docs-only; it
// refuses rather than recording a pass nothing ran for.
func TestAffectedNoChangeRefuses(t *testing.T) {
	d, gh, _, out, marker := affectedFixture(t)
	sha, err := d.Git("rev-parse", "main")
	if err != nil {
		t.Fatal(err)
	}
	sha = strings.TrimSpace(sha)
	gh.setHead(9, sha)
	_, err = runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", affectedFixtureConfig(t)), affectedFixtureConfig(t), affectedFixtureSuite, sha, affectedOrdinary)
	if err == nil || !strings.Contains(err.Error(), "no changes") {
		t.Fatalf("err %v", err)
	}
	if len(gh.statuses[sha]) != 0 || exists(out) || exists(marker) {
		t.Fatalf("statuses %v, consumer %v, suite %v", gh.statuses[sha], exists(out), exists(marker))
	}
}

// #473 D1: affected.json's absence is the repo's off switch; a bad one refuses.
func TestReadAffectedConfig(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "acquired")
	if _, ok, err := readAffectedConfig(d.Git, "main"); ok || err != nil {
		t.Fatalf("absent: ok %v err %v", ok, err)
	}
	commit := func(body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(d.Root, ".sdlc"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d.Root, affectedConfigPath), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "cfg"}} {
			if out, err := d.Git(args...); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
	}
	commit(`{"version":1,"command":["make","test-affected"]}`)
	if cfg, ok, err := readAffectedConfig(d.Git, "main"); !ok || err != nil || cfg.Command[0] != "make" {
		t.Fatalf("present: %+v ok %v err %v", cfg, ok, err)
	}
	commit(`{"version":1,"command":["make"],"surprise":true}`)
	if _, _, err := readAffectedConfig(d.Git, "main"); err == nil || !strings.Contains(err.Error(), affectedConfigPath) {
		t.Fatalf("bad file: err %v", err)
	}
}

// #473: the selection as a local check: env(1) carries the contract, the
// suite's flake policy rides along, and the scope names the selection.
func TestAffectedLocalCheck(t *testing.T) {
	cfg := affected.Config{Command: []string{"make", "test-affected"}, Weight: 1, TimeoutMinutes: 30}
	sel := affected.Selection{GoPkgs: []string{"a", "b"}, Groups: []string{"lua"}}
	c := affectedLocalCheck(cfg, sel, localCheck{KnownFlakes: "accept"})
	want := []string{"env", "SDLC_AFFECTED_GO_PKGS=a b", "SDLC_AFFECTED_GROUPS=lua", "make", "test-affected"}
	if strings.Join(c.Command, "|") != strings.Join(want, "|") || c.Name != affectedCheckName || c.KnownFlakes != "accept" || c.Scope != "2 packages, groups lua" {
		t.Fatalf("%+v", c)
	}
	if got := affectedScope(affected.Selection{Groups: []string{"lua"}}); got != "groups lua" {
		t.Fatal(got)
	}
}

// #473 M1 BR-4: affected is split from the declared names, and --all with any
// --name, affected included, refuses as it did before.
func TestSplitAffectedName(t *testing.T) {
	if _, _, err := splitAffectedName([]string{affectedCheckName}, true); err == nil {
		t.Fatal("--all with --name affected was accepted")
	}
	got, aff, err := splitAffectedName([]string{"suite", affectedCheckName, "e2e"}, false)
	if err != nil || !aff || strings.Join(got, ",") != "suite,e2e" {
		t.Fatalf("%v %v %v", got, aff, err)
	}
	if got, aff, err := splitAffectedName(nil, true); err != nil || aff || len(got) != 0 {
		t.Fatalf("--all alone: %v %v %v", got, aff, err)
	}
}

// #473 M2 BR-9: the affected name resolves for a decision that needs origin/
// main's declaration (republishing a saved known-flake pass): affected.json's
// command with the suite's policy. Without the file it refuses.
func TestDeclaredCheckResolvesAffected(t *testing.T) {
	files := map[string]string{
		localChecksPath:    `{"version":1,"checks":[{"name":"suite","command":["make","test"],"required":true,"known_flakes":"accept"}]}`,
		affectedConfigPath: `{"version":1,"command":["make","test-affected"],"weight":2}`,
	}
	git := func(args ...string) (string, error) {
		switch args[0] {
		case "fetch":
			return "", nil
		case "rev-parse":
			return strings.Repeat("a", 40), nil
		case "ls-tree":
			if _, ok := files[args[len(args)-1]]; ok {
				return args[len(args)-1] + "\n", nil
			}
			return "", nil
		case "show":
			return files[strings.TrimPrefix(args[1], strings.Repeat("a", 40)+":")], nil
		}
		return "", fmt.Errorf("unexpected git %v", args)
	}
	c, err := declaredCheck(context.Background(), git, affectedCheckName, strings.Repeat("b", 40))
	if err != nil || c.Name != affectedCheckName || strings.Join(c.Command, " ") != "make test-affected" || c.Weight != 2 || c.KnownFlakes != "accept" {
		t.Fatalf("%+v %v", c, err)
	}
	delete(files, affectedConfigPath)
	if _, err := declaredCheck(context.Background(), git, affectedCheckName, strings.Repeat("b", 40)); err == nil || !strings.Contains(err.Error(), affectedConfigPath) {
		t.Fatalf("no affected.json: %v", err)
	}
}

// #473 M2: an affected.json main cannot parse degrades a car to requiring the
// suite; it never stops the train.
func TestCarWithBrokenAffectedConfigNeedsSuite(t *testing.T) {
	root, gh, commit, vouch, _ := carGateFixture(t, false)
	procedureWrite(t, root, affectedConfigPath, `{"version":1,"command":["make"],"surprise":1}`+"\n")
	base := commit("go.sum")
	git(t, root, "update-ref", "refs/remotes/origin/main", base)
	head := commit("feature.go")
	postStatus(t, gh, head, affectedCheckName)
	if err := vouch(head); err == nil || !strings.Contains(err.Error(), "--name "+suiteCheckName) {
		t.Fatalf("a broken affected.json: %v", err)
	}
	postStatus(t, gh, head, suiteCheckName)
	if err := vouch(head); err != nil {
		t.Fatalf("suite still vouches: %v", err)
	}
}

// #561: boarding mode never escalates a full selection to the suite: it
// refuses before any command runs or any status is posted, naming the
// selection's reason, the repair, and the full-proof alternative.
func TestAffectedBoardingNeverEscalatesToSuite(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"go.mod changed":           {"go.mod": "module example.com/m\n\ngo 1.23\n"},
		"unclassified path":        {"tools/x.py": "print(1)\n"},
		"package graph unloadable": {"leaf/leaf.go": "not go\n"},
	} {
		t.Run(name, func(t *testing.T) {
			d, gh, change, out, marker := affectedFixture(t)
			sha := change(files)
			cfg := affectedFixtureConfig(t)
			_, err := runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha, affectedBoarding)
			if err == nil {
				t.Fatalf("boarding accepted a full selection\n%s", d.Stderr)
			}
			for _, want := range []string{"boarding", "full selection", affectedConfigPath,
				"sdlc local-check run --name suite --rev " + sha, "sdlc merge --yes", "train_interval_minutes"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal lacks %q: %v", want, err)
				}
			}
			if exists(marker) || exists(out) || len(gh.statuses[sha]) != 0 {
				t.Fatalf("suite ran %v, consumer ran %v, statuses %v\n%s", exists(marker), exists(out), gh.statuses[sha], d.Stderr)
			}
		})
	}
	// A partial and a docs-only selection board exactly as an ordinary run does.
	for name, files := range map[string]map[string]string{
		"partial":   {"leaf/leaf.go": "package leaf\n\nfunc F() int { return 2 }\n"},
		"docs-only": {"README.md": "# m\n"},
	} {
		t.Run(name, func(t *testing.T) {
			d, gh, change, _, marker := affectedFixture(t)
			sha := change(files)
			cfg := affectedFixtureConfig(t)
			rec, err := runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha, affectedBoarding)
			if err != nil || !rec.passed() || rec.Name != affectedCheckName || gh.statuses[sha]["sdlc-local/affected"] != "success" || exists(marker) {
				t.Fatalf("rec %+v err %v statuses %v suite %v", rec, err, gh.statuses[sha], exists(marker))
			}
		})
	}
}

// #561: --boarding qualifies affected alone; with suite, mixed names or
// --all it refuses rather than being silently ignored.
func TestLocalCheckBoardingFlag(t *testing.T) {
	for _, tc := range []struct {
		names    []string
		all, ok  bool
		boarding bool
	}{
		{names: []string{affectedCheckName}, boarding: true, ok: true},
		{names: []string{affectedCheckName}, ok: true},
		{names: []string{suiteCheckName}, ok: true},
		{names: []string{suiteCheckName}, boarding: true},
		{names: []string{affectedCheckName, suiteCheckName}, boarding: true},
		{names: []string{affectedCheckName, "e2e"}, boarding: true},
		{all: true, boarding: true},
	} {
		mode, err := affectedModeFor(tc.names, tc.all, tc.boarding)
		if (err == nil) != tc.ok {
			t.Errorf("%v all=%v boarding=%v: err %v", tc.names, tc.all, tc.boarding, err)
			continue
		}
		if err == nil && (mode == affectedBoarding) != tc.boarding {
			t.Errorf("%v boarding=%v: mode %v", tc.names, tc.boarding, mode)
		}
		if err != nil && !strings.Contains(err.Error(), "--boarding") {
			t.Errorf("refusal does not name the flag: %v", err)
		}
	}
	if run, _, err := NewLocalCheckCmd().Find([]string{"run"}); err != nil || run.Flags().Lookup("boarding") == nil {
		t.Fatal("local-check run has no --boarding flag")
	}
}

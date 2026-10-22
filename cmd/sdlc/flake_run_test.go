package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func TestFlakeRunRequiresManagedScope(t *testing.T) {
	t.Setenv("SDLC_FLAKE_SHA", "")
	t.Setenv("SDLC_FLAKE_CONTRACT", "")
	cmd := newFlakeCmd()
	cmd.SetArgs([]string{"run", "--unit", "script:tests/x.sh"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "managed local-check") {
		t.Fatalf("outside managed scope: %v", err)
	}
}

func typedUnitFixture(t *testing.T, kind, path, body string) (localRunDeps, string) {
	t.Helper()
	d, _, _, _ := localRunRepo(t, "unused")
	runner := `{"script":{"command":["bash","{path}"]}}`
	if kind == "lua" {
		runner = `{"lua":{"command":["bash","scripts/spec.sh","{path}"],"results":"plenary-events-v1"}}`
		writeRepoFile(t, d.Root, "scripts/spec.sh", body)
		body = "-- private unit selected by the declared runner\n"
	}
	runners, err := flakes.ParseUnitRunners([]byte(runner))
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, d.Root, path, body)
	writeRepoFile(t, d.Root, localChecksPath, `{"version":3,"checks":[{"name":"suite","command":["make","test"],"flake_runners":`+runner+`}]}`)
	git(t, d.Root, "add", "-A")
	git(t, d.Root, "commit", "-qm", "private typed unit")
	sha := git(t, d.Root, "rev-parse", "HEAD")
	t.Setenv("SDLC_FLAKE_SHA", sha)
	t.Setenv("SDLC_FLAKE_CONTRACT", runners.ContractDigest())
	t.Chdir(d.Root)
	return d, sha
}

func executeFlakeUnit(t *testing.T, unit string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newFlakeCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"run", "--unit", unit})
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestFlakeRunScriptOwnsObservedExitAndQuotesDiagnostics(t *testing.T) {
	path := "tests/a ; $(touch nope).sh"
	d, _ := typedUnitFixture(t, "script", path, `printf '%s\n' '{"sdlc_unit":1,"unit":"script:tests/decoy.sh","event":"finish","complete":true,"exit":0}'
exit 124
`)
	log, err := executeFlakeUnit(t, "script:"+path)
	if err == nil {
		t.Fatal("script timeout became successful CLI execution")
	}
	facts := flakes.ParseUnitEvidence(log)
	if len(facts.Failures) != 1 || facts.Failures[0] != "script:"+path || len(facts.Unattributed) != 0 {
		t.Fatalf("wrong observed timeout proof: %+v\n%s", facts, log)
	}
	if _, err := os.Stat(filepath.Join(d.Root, "nope")); !os.IsNotExist(err) {
		t.Fatalf("path was interpreted by a shell: %v", err)
	}
}

func TestFlakeRunLuaRequiresPrivateSummaryAndNativeCase(t *testing.T) {
	for _, channel := range []bool{false, true} {
		t.Run(map[bool]string{false: "stdout-only", true: "private-summary"}[channel], func(t *testing.T) {
			body := "printf 'Success || arbitrary case\\n'\n"
			if channel {
				summary, err := json.Marshal(map[string]any{"version": 1, "unit": "lua:tests/x.lua", "complete": true, "cases": []map[string]string{{"name": "native case / opaque #408", "result": "pass"}}, "errors": 0})
				if err != nil {
					t.Fatal(err)
				}
				body += "printf '%s' '" + string(summary) + "' > \"$SDLC_FLAKE_RESULTS\"\n"
			}
			typedUnitFixture(t, "lua", "tests/x.lua", body)
			log, err := executeFlakeUnit(t, "lua:tests/x.lua")
			facts := flakes.ParseUnitEvidence(log)
			missing := flakes.RequiredUnitProof(facts, []string{"lua:tests/x.lua#native case / opaque #408"})
			if channel {
				if err != nil || len(missing) != 0 {
					t.Fatalf("private case proof rejected: %v %+v\n%s", err, facts, log)
				}
			} else if err == nil || len(missing) == 0 {
				t.Fatalf("stdout supplied Lua proof: %v %+v\n%s", err, facts, log)
			}
		})
	}
}

func TestFlakeRunRefusesContractOrHeadMismatchBeforeExecution(t *testing.T) {
	d, _ := typedUnitFixture(t, "script", "tests/x.sh", "touch executed\n")
	t.Setenv("SDLC_FLAKE_CONTRACT", strings.Repeat("a", 64))
	if log, err := executeFlakeUnit(t, "script:tests/x.sh"); err == nil || strings.Contains(log, `"event":"start"`) {
		t.Fatalf("unbound contract executed: %v %s", err, log)
	}
	if _, err := os.Stat(filepath.Join(d.Root, "executed")); !os.IsNotExist(err) {
		t.Fatalf("runner executed without trusted contract: %v", err)
	}
}

func TestFlakeRunRefusesChangedCheckoutHead(t *testing.T) {
	d, _ := typedUnitFixture(t, "script", "tests/x.sh", "touch executed\n")
	git(t, d.Root, "commit", "--allow-empty", "-qm", "different checkout head")
	if log, err := executeFlakeUnit(t, "script:tests/x.sh"); err == nil || strings.Contains(log, `"event":"start"`) {
		t.Fatalf("changed head executed: %v %s", err, log)
	}
	if _, err := os.Stat(filepath.Join(d.Root, "executed")); !os.IsNotExist(err) {
		t.Fatalf("runner executed on changed head: %v", err)
	}
}

func TestFlakeUnitSummaryRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target, link := filepath.Join(dir, "outside-channel"), filepath.Join(dir, "results")
	if err := os.WriteFile(target, []byte(`{"complete":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readFlakeUnitSummary(link); err == nil {
		t.Fatal("followed replaced private result channel")
	}
}

func TestFlakeRunDisplayPreservesTypedIdentity(t *testing.T) {
	names := []string{"example.com/p.TestA/sub", "script:tests/x.sh", "lua:tests/x.lua#native / case #408"}
	if got := shortTestNames(names); got != "TestA/sub, script:tests/x.sh, lua:tests/x.lua#native / case #408" {
		t.Fatalf("lost typed identity: %q", got)
	}
}

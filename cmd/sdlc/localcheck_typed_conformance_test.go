//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const typedConformanceGo = "example.com/private-typed/go_case.TestNative"
const typedConformanceScript = "script:tests/nvim-route-poke-admission-test.sh"
const typedConformanceLuaUnit = "lua:tests/integration/session_sync_native_spec.lua"
const typedConformanceLuaCase = "session-sync native edits retries refused admission with native o entry (#408 retry)"

// Reuse the private resume harness's production-main subprocess, real Git
// transport, immutable-store reader and status sink. The actual unit adapter
// is the built production binary, selected through the fixture's private PATH.
func newPrivateTypedFixture(t *testing.T, binary, library, nvim string) *privateResumeFixture {
	t.Helper()
	producer, err := os.ReadFile("../../scripts/sdlc-plenary-events.lua")
	if err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"HOME", "TMPDIR", "GOTMPDIR", "WF_TEST_SCRATCH", "WF_SDLC_QUEUE_ROOT", "XDG_STATE_HOME", "XDG_CACHE_HOME"} {
		t.Setenv(key, t.TempDir())
	}
	for key, val := range map[string]string{"GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.DevNull, "WF_TEST_LOCK_FILE": filepath.Join(t.TempDir(), "suite.lock"), "WF_TEST_SUITES": "1", "WF_TEST_LOCK": "", "WF_TEST_LOCK_HELD_BY": "", "WF_TEST_LOCK_TIMEOUT": "30", "WF_ISSUES_DIR": "workshop/issues", "COUCH_TREE": "", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOFLAGS": ""} {
		t.Setenv(key, val)
	}
	d, _, _, _ := localRunRepo(t, "unused")
	f := &privateResumeFixture{t: t, root: d.Root, marks: t.TempDir(), statuses: filepath.Join(t.TempDir(), "statuses.jsonl")}
	f.store = filepath.Join(f.root, ".git", "sdlc", "test-stats")
	t.Setenv("PRIVATE_TYPED_MARKS", f.marks)
	t.Setenv("PRIVATE_RESUME_MARKS", f.marks)
	t.Setenv("SDLC_PLENARY_DIR", library)
	write := func(rel, body string) { writeRepoFile(t, f.root, rel, body) }
	write("go.mod", "module example.com/private-typed\n\ngo 1.22\n")
	write("go_case/typed_test.go", `package go_case
import ("os";"path/filepath";"strings";"testing")
func TestNative(t *testing.T) {
 root:=os.Getenv("PRIVATE_TYPED_MARKS")
 f,err:=os.OpenFile(filepath.Join(root,"go-count"),os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600);if err!=nil {t.Fatal(err)};if _,err=f.WriteString("go\n");err!=nil {t.Fatal(err)};if err=f.Close();err!=nil {t.Fatal(err)}
 raw,err:=os.ReadFile(filepath.Join(root,"mode"));if err!=nil {t.Fatal(err)}
 if strings.TrimSpace(string(raw))=="fail" {t.Fatal("private same-head native Go failure")}
}
func TestDecoy(t *testing.T) { if err:=os.WriteFile(filepath.Join(os.Getenv("PRIVATE_TYPED_MARKS"),"go-decoy"),nil,0600);err!=nil {t.Fatal(err)} }
`)
	write("scripts/test-lock.py", string(helper))
	write("scripts/sdlc-plenary-events.lua", string(producer))
	write("tests/nvim-route-poke-admission-test.sh", `printf 'script\n' >> "$PRIVATE_TYPED_MARKS/script-count"
mode=$(cat "$PRIVATE_TYPED_MARKS/mode")
if [ "$mode" = fail ]; then
 python3 -c 'import subprocess,sys
try: subprocess.run(["sleep","5"],timeout=0.05,check=True)
except subprocess.TimeoutExpired:
 print("run_headless: TIMEOUT — child did not exit.");sys.exit(124)'
 exit $?
fi
exit 0
`)
	write("tests/integration/session_sync_native_spec.lua", `local modefile=assert(io.open(vim.env.PRIVATE_TYPED_MARKS..'/mode','r'))
local mode=modefile:read('*l');modefile:close()
describe('session-sync native edits',function()
 before_each(function() local f=assert(io.open(vim.env.PRIVATE_TYPED_MARKS..'/hooks','a'));f:write('B');f:close() end)
 after_each(function() local f=assert(io.open(vim.env.PRIVATE_TYPED_MARKS..'/hooks','a'));f:write('A');f:close() end)
 if mode~='omit' then
  it('retries refused admission with native o entry (#408 retry)',function() assert.equals('pass',mode) end)
 end
 it('keeps a sibling case in the whole spec',function()
  local f=assert(io.open(vim.env.PRIVATE_TYPED_MARKS..'/sibling-count','a'));f:write('sibling\n');f:close()
  assert.equals(1,1)
 end)
end)
`)
	write("tests/typed-child.lua", `vim.opt.runtimepath:append(vim.env.SDLC_PLENARY_DIR)
local file=vim.fn.argv()[1]
local f=assert(io.open(vim.env.PRIVATE_TYPED_MARKS..'/lua-count','a'));f:write('lua\n');f:close()
dofile('scripts/sdlc-plenary-events.lua').install(file)
require('plenary.busted').run(file)
`)
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private typed inputs")
	wireLocalCheckSignalCLI(t, f.root)
	if err := os.Symlink(binary, filepath.Join(os.Getenv("SDLC_QUEUE_BIN"), "sdlc")); err != nil {
		t.Fatal(err)
	}
	runners, err := flakes.ParseUnitRunners([]byte(fmt.Sprintf(`{"script":{"command":["bash","{path}"]},"lua":{"command":[%q,"--headless","-u","NONE","-i","NONE","-n","-c","lua dofile('tests/typed-child.lua')","--","{path}"],"results":"plenary-events-v1"}}`, nvim)))
	if err != nil {
		t.Fatal(err)
	}
	decl := localCheckDecl{Version: 3, Checks: []localCheck{
		{Name: "lint", Command: []string{"true"}, TimeoutMinutes: 1, Weight: 1},
		{Name: suiteCheckName, Command: []string{"sh", "-c", `printf 'suite\n' >> "$PRIVATE_TYPED_MARKS/suite-count"; sdlc flake run --unit ` + typedConformanceScript + `; a=$?; sdlc flake run --unit ` + typedConformanceLuaUnit + `; b=$?; go test -count=1 -json -run '^TestNative$' ./go_case; c=$?; test "$a" = 0 && test "$b" = 0 && test "$c" = 0`}, TimeoutMinutes: 2, Weight: 1, FlakeRunners: runners},
	}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	write(localChecksPath, string(raw))
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private typed declaration")
	git(t, f.root, "push", "-qu", "origin", "main")
	trunk, err := gitx.NewTrunkFileContext(context.Background(), f.root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	_, err = trunk.Bootstrap(map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes(), "workshop/issue-cards/000001-flakes.md": []byte("---\nid: 000001\nstatus: open\n---\n\n# Private typed flakes\n\n## Problem\nPrivate typed evidence.\n")}, "private empty registry", func(gitx.BootstrapResult) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	trackerRoot := strings.TrimSpace(git(t, f.root, "rev-parse", "refs/remotes/origin/issue-tracker"))
	write(tracker.CutoverMarkerPath, string(tracker.CutoverMarkerBytes(trackerRoot)))
	git(t, f.root, "add", "-A")
	git(t, f.root, "commit", "-qm", "private tracker marker")
	git(t, f.root, "push", "-q", "origin", "main")
	f.sha = strings.TrimSpace(git(t, f.root, "rev-parse", "HEAD"))
	t.Setenv("SDLC_PRIVATE_GH_STATE", f.statuses)
	gh := `#!/usr/bin/env python3
import json,os,sys
a=sys.argv[1:]
if a[:2]==['pr','list']: sys.exit(0)
if a[:3]==['api','-X','POST'] and a[3].startswith('repos/fixture/queue/statuses/'):
 fields=dict(x.split('=',1) for x in a[4:] if x!='-f');fields['endpoint']=a[3]
 with open(os.environ['SDLC_PRIVATE_GH_STATE'],'a') as f: f.write(json.dumps(fields)+'\n')
 print('{}');sys.exit(0)
sys.exit('unexpected private gh command: '+repr(a))
`
	if err := os.WriteFile(filepath.Join(os.Getenv("SDLC_QUEUE_BIN"), "gh"), []byte(gh), 0700); err != nil {
		t.Fatal(err)
	}
	f.cleanupRetainedFixture()
	return f
}

func typedFixtureMode(t *testing.T, f *privateResumeFixture, mode string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.marks, "mode"), []byte(mode+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}
func typedFixtureCounts(f *privateResumeFixture) string {
	return string(f.read(filepath.Join(f.marks, "suite-count"))) + "/" + string(f.read(filepath.Join(f.marks, "script-count"))) + "/" + string(f.read(filepath.Join(f.marks, "lua-count"))) + "/" + string(f.read(filepath.Join(f.marks, "sibling-count"))) + "/" + string(f.read(filepath.Join(f.marks, "go-count")))
}

// Explicit acceptance: SDLC_PRIVATE_TYPED_CONFORMANCE=1 SDLC_PLENARY_DIR=...
// go test ./cmd/sdlc -run '^TestLocalCheckTypedConformance$' -count=1 -v.
func TestLocalCheckTypedConformance(t *testing.T) {
	if os.Getenv("SDLC_PRIVATE_TYPED_CONFORMANCE") != "1" {
		t.Skip("opt in with SDLC_PRIVATE_TYPED_CONFORMANCE=1 and SDLC_PLENARY_DIR")
	}
	library := os.Getenv("SDLC_PLENARY_DIR")
	if library == "" {
		t.Fatal("native conformance requires installed SDLC_PLENARY_DIR")
	}
	if _, err := os.Stat(filepath.Join(library, "lua", "plenary", "busted.lua")); err != nil {
		t.Fatal(err)
	}
	library, err := filepath.Abs(library)
	if err != nil {
		t.Fatal(err)
	}
	nvim, err := exec.LookPath("nvim")
	if err != nil {
		t.Fatal(err)
	}
	binary := testfix.BuiltBinary(t, ".", ".")
	// Keep Go's normal cache when the private fixture replaces HOME.
	cache, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOCACHE", strings.TrimSpace(string(cache)))
	for _, mode := range []string{"pass", "omit", "crash"} {
		t.Run(mode, func(t *testing.T) {
			f := newPrivateTypedFixture(t, binary, library, nvim)
			typedFixtureMode(t, f, "fail")
			f.cli(false, "local-check", "run", "--name", "suite")
			initial := f.snapshot()
			if len(initial.Attempts) != 1 {
				t.Fatalf("initial %+v", initial)
			}
			f.parent = initial.Attempts[0]
			f.parentRaw = f.read(filepath.Join(f.store, "attempts", f.parent.ID+".json"))
			f.parentLog = f.read(suiteAttemptLogFile(f.store, f.parent))
			facts := flakes.ParseUnitEvidence(string(f.parentLog))
			if f.parent.Version != 4 || f.parent.RunnerContractSHA256 == "" || len(facts.Unattributed) != 0 || len(facts.Failures) != 2 || len(f.parent.Failures) != 3 || !slices.Contains(f.parent.Failures, typedConformanceGo) || !strings.Contains(string(f.parentLog), `"exit":124`) {
				t.Fatalf("native typed failure: parent=%+v facts=%+v log=%s", f.parent, facts, f.parentLog)
			}
			typedFixtureMode(t, f, "pass")
			// The real suite passes, but publication must still refuse unexplained history.
			out := f.cli(false, "local-check", "run", "--name", "suite")
			if !strings.Contains(out, "unexplained") {
				t.Fatalf("missing same-head guard: %s", out)
			}
			f.cli(true, "flake", "register", "--issue", "1", "--test", typedConformanceGo, "--test", typedConformanceScript, "--test", typedConformanceLuaUnit+"#"+typedConformanceLuaCase)
			before := typedFixtureCounts(f)
			if before != "suite\nsuite\n/script\nscript\n/lua\nlua\n/sibling\nsibling\n/go\ngo\n" {
				t.Fatalf("pre-resume actual command inventory %q", before)
			}
			if mode == "crash" {
				typedFixtureMode(t, f, "pass")
				typedFixtureInterruptedResume(t, f, before)
				return
			}
			typedFixtureMode(t, f, mode)
			f.cli(mode == "pass", f.resumeArgs()...)
			after := f.snapshot()
			if len(after.Resumes) != 1 {
				t.Fatalf("resume graph %+v", after)
			}
			var child *SuiteAttempt
			for i := range after.Attempts {
				if after.Attempts[i].ID == after.Resumes[0].ChildID {
					child = &after.Attempts[i]
				}
			}
			if child == nil || child.Version != 4 || child.RerunOf != f.parent.ID || child.RunnerContractSHA256 != f.parent.RunnerContractSHA256 {
				t.Fatalf("typed child %+v", child)
			}
			childRaw := f.read(suiteAttemptLogFile(f.store, *child))
			missing := flakes.RerunMissingProof(string(childRaw), []string{typedConformanceGo, typedConformanceScript, typedConformanceLuaUnit + "#" + typedConformanceLuaCase})
			if (len(missing) == 0) != (mode == "pass") {
				t.Fatalf("child positive proof missing=%v log=%s", missing, childRaw)
			}
			counts := typedFixtureCounts(f)
			if counts != "suite\nsuite\n/script\nscript\nscript\n/lua\nlua\nlua\n/sibling\nsibling\nsibling\n/go\ngo\ngo\n" {
				t.Fatalf("resume broadened or omitted command inventory %q", counts)
			}
			f.cli(mode == "pass", "local-check", "republish", "--name", "suite", "--attempt", f.parent.ID, "--rev", f.sha)
			f.cli(mode == "pass", f.resumeArgs()...)
			if typedFixtureCounts(f) != counts {
				t.Fatal("saved resume/republish launched another unit")
			}
			f.unchanged()
			if _, err := os.Stat(filepath.Join(f.marks, "go-decoy")); !os.IsNotExist(err) {
				t.Fatalf("unselected Go test executed: %v", err)
			}
			statuses := f.read(f.statuses)
			hasSuccess := bytes.Contains(statuses, []byte(`"state": "success"`))
			if hasSuccess != (mode == "pass") {
				t.Fatalf("guard publication acceptance=%v mode=%s statuses=%s", hasSuccess, mode, statuses)
			}
			hooks := string(f.read(filepath.Join(f.marks, "hooks")))
			hookCalls := 5
			if mode == "pass" {
				hookCalls++
			}
			if hooks != strings.Repeat("BA", hookCalls) {
				t.Fatalf("native hooks changed: %q", hooks)
			}
			t.Logf("private native typed mode=%s parent=%s child=%s exact-file/whole-script proof; retained original bytes", mode, f.parent.ID, child.ID)
		})
	}
}

// Stop after the existing worktree-lock effect, before any reserved child
// starts. Reconciliation may add only its immutable interrupted audit, while
// the original typed/Go proof and every protected resource remain unchanged.
func typedFixtureInterruptedResume(t *testing.T, f *privateResumeFixture, counts string) {
	t.Helper()
	f.installBarrier("afterlock")
	f.killAtBarrier(f.resumeArgs()...)
	pending := f.snapshot()
	if len(pending.Resumes) != 1 || len(pending.Attempts) != 2 || len(pending.Reconciliations) != 0 {
		t.Fatalf("typed pending fence: %+v", pending)
	}
	if typedFixtureCounts(f) != counts {
		t.Fatal("crash barrier launched a reserved test child")
	}
	inputs := f.retainedInputs()
	statuses := f.read(f.statuses)
	args := f.reconcileArgs()
	f.cli(true, args...)
	terminal := f.snapshot()
	if len(terminal.Attempts) != 2 || len(terminal.Reconciliations) != 1 || terminal.Reconciliations[0].Outcome != "interrupted" {
		t.Fatalf("typed interruption audit: %+v", terminal)
	}
	receiptPath := filepath.Join(f.store, "attempts", terminal.Reconciliations[0].ID+".json")
	receipt := f.read(receiptPath)
	f.assertRetained(inputs)
	f.cli(true, args...)
	f.cli(false, f.resumeArgs()...)
	f.cli(false, "local-check", "republish", "--name", "suite", "--attempt", f.parent.ID, "--rev", f.sha)
	if typedFixtureCounts(f) != counts || !bytes.Equal(statuses, f.read(f.statuses)) || !bytes.Equal(receipt, f.read(receiptPath)) {
		t.Fatal("interrupted recovery executed tests, posted success or rewrote its receipt")
	}
	f.assertRetained(inputs)
	// The independent full-suite pass is ordinary GC input, outside the pinned
	// interrupted graph. Do not mistake its normal retirement for lost proof.
	for _, a := range terminal.Attempts {
		if a.ID != f.parent.ID {
			delete(inputs, filepath.Join(f.store, "attempts", a.ID+".json"))
			delete(inputs, suiteAttemptLogFile(f.store, a))
		}
	}
	f.assertPinnedAfterPrune(inputs)
	f.unchanged()
}

package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLandingCheckProductionTrainAndSingle(t *testing.T) {
	for _, tc := range []struct {
		name                                                string
		multi, fail, disabled, unrelated, noCI, off, prefix bool
	}{
		{name: "train", multi: true}, {name: "single"}, {name: "failure", multi: true, fail: true}, {name: "disabled", multi: true, disabled: true}, {name: "unrelated", multi: true, unrelated: true}, {name: "no-ci", noCI: true}, {name: "trains-disabled", off: true}, {name: "good-prefix", multi: true, prefix: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			multi := tc.multi
			s := newTrainSetup(t, "train")
			s.noCI = tc.noCI
			trace := filepath.Join(t.TempDir(), "order")
			decl := localCheckDecl{Version: 2, Checks: []localCheck{
				{Name: "suite", Required: true, Command: []string{"sh", "-c", `echo suite >> "$1"`, "fixture", trace}},
				{Name: "vm", RunScope: "landing", Resource: "runner", Required: true, Command: []string{"sh", "-c", `git rev-parse HEAD >> "$1"`, "fixture", trace}},
			}}
			if tc.off {
				procedureWrite(t, s.server, landingConfigPath, `{"train_interval_minutes":0}`)
			}
			if tc.prefix {
				decl.Checks[0].Command[2] += "; test ! -e second.go"
			}
			if tc.disabled {
				decl.Checks[1].Required = false
			}
			if tc.fail {
				decl.Checks[1].Command[2] += "; exit 37"
			}
			if tc.unrelated {
				decl.Checks[1].PathsFile = "paths.json"
				procedureWrite(t, s.server, "paths.json", `{"version":1,"paths":["unrelated/**"]}`)
			}
			raw, err := json.Marshal(decl)
			if err != nil {
				t.Fatal(err)
			}
			procedureWrite(t, s.server, localChecksPath, string(raw))
			git(t, s.server, "add", ".")
			git(t, s.server, "commit", "-qm", "landing gate")
			git(t, s.server, "push", "-q", "origin", "main")
			if multi {
				s.pushCar(t, "000102-second", "second", 102, 2)
			}
			result := s.merge(s.first)
			waitQueued(t, s.line, 1, []chan error{result})
			if !tc.off {
				waitVouched(t, s.line, 1, []chan error{result})
			}
			if multi {
				s.queuedCar(t, "000102-second", 102, 2)
				waitVouched(t, s.line, 2, []chan error{result})
			}
			s.blocker.Release()
			mergeErr := finished(t, result)
			if tc.fail {
				if mergeErr == nil || !strings.Contains(mergeErr.Error(), "landing checks failed") {
					t.Fatalf("runner failure %v", mergeErr)
				}
				if s.gh.merges != 0 {
					t.Fatal("merged after VM failure")
				}
			} else if mergeErr != nil {
				t.Fatal(mergeErr)
			}
			raw, err = os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Fields(string(raw))
			want := 2
			if tc.prefix {
				want = 4
			}
			if tc.disabled || tc.unrelated {
				want = 1
			}
			if len(lines) != want || lines[0] != "suite" || (want >= 2 && len(lines[len(lines)-1]) != 40) {
				t.Fatalf("suite then exactly one final-head runner: %q", raw)
			}
			if want >= 2 && !tc.fail {
				if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", lines[len(lines)-1], "main"); err != nil {
					t.Fatalf("runner head did not land: %v", err)
				}
			}
			if multi && len(s.gh.created) != 1 {
				t.Fatalf("not a train: %v", s.gh.created)
			}
		})
	}
}

func TestLandingCheckProductionLegacy(t *testing.T) {
	for _, tc := range []struct {
		name       string
		noCI, fail bool
	}{{name: "normal"}, {name: "no-ci", noCI: true}, {name: "failure", fail: true}} {
		t.Run(tc.name, func(t *testing.T) {
			trace := filepath.Join(t.TempDir(), "order")
			script := `echo vm >> "$1"`
			if tc.fail {
				script += "; exit 37"
			}
			decl := localCheckDecl{Version: 2, Checks: []localCheck{
				{Name: "suite", Required: true, Command: []string{"sh", "-c", `echo suite >> "$1"`, "fixture", trace}},
				{Name: "vm", RunScope: "landing", Resource: "runner", Required: true, Command: []string{"sh", "-c", script, "fixture", trace}},
			}}
			raw, err := json.Marshal(decl)
			if err != nil {
				t.Fatal(err)
			}
			root := tempRepoWith(t, map[string]string{localChecksPath: string(raw), "Makefile.workflow": "# fake\n", "scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n")})
			head := git(t, root, "rev-parse", "HEAD")
			ci := newCIFakeGH()
			ci.setHead(42, head)
			gh := &e2eGH{openPR: "42", ciSeamGH: ciSeamGH{ci: ci}}
			swapMergeDeps(t, gh, nil)
			f := &mergeFlags{Yes: true, NoJudge: true, NoCI: tc.noCI, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
			if tc.fail {
				msg, died := expectDie(t, func() { _ = runMerge(io.Discard, io.Discard, f) })
				if !died || !strings.Contains(msg, "landing check vm") {
					t.Fatalf("failure %v %s", died, msg)
				}
				if gh.prMergeCalls != 0 {
					t.Fatal("legacy merged VM failure")
				}
			} else if err := runMerge(io.Discard, io.Discard, f); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(trace)
			if err != nil || string(got) != "suite\nvm\n" {
				t.Fatalf("legacy order/retry %q %v", got, err)
			}
		})
	}
}

func TestLandingCheckManualCLIUsesLandedMain(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "manual")
	decl := localCheckDecl{Version: 2, Checks: []localCheck{{Name: "vm", RunScope: "landing", Resource: "runner", Command: []string{"sh", "-c", `echo ran > "$1"`, "fixture", marker}}}}
	raw, err := json.Marshal(decl)
	if err != nil {
		t.Fatal(err)
	}
	root := tempRepoWith(t, map[string]string{localChecksPath: string(raw)})
	head, main := git(t, root, "rev-parse", "HEAD"), git(t, root, "rev-parse", "origin/main")
	ci := newCIFakeGH()
	ci.setHead(42, head)
	ci.setHead(43, main)
	gh := &e2eGH{openPR: "42", ciSeamGH: ciSeamGH{ci: ci}}
	swapMergeDeps(t, gh, nil)
	if err := runLocalCheckCmd(context.Background(), io.Discard, localCheckFlags{Names: []string{"vm"}, Rev: "HEAD"}); err == nil || !strings.Contains(err.Error(), "pinned main") {
		t.Fatalf("off-main burn-in %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("off-main launched: %v", err)
	}
	if err := runLocalCheckCmd(context.Background(), io.Discard, localCheckFlags{All: true, Rev: "HEAD"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("--all launched VM: %v", err)
	}
	if err := runLocalCheckCmd(context.Background(), io.Discard, localCheckFlags{Names: []string{"vm"}, Rev: main}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal(err)
	}
}

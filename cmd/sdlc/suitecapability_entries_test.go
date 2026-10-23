package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func suiteCapabilityDenied() error {
	return &suiteEnvironmentError{Cause: syscall.EPERM, Diagnosis: suiteStartupDiagnosis{Resource: "private Unix listener"}}
}

func assertCapabilityRefusal(t *testing.T, err error, check, sha string) {
	t.Helper()
	var environment *suiteEnvironmentError
	if !errors.As(err, &environment) || !errors.Is(err, syscall.EPERM) {
		t.Fatalf("expected capability refusal, got %v", err)
	}
	if environment.Check != check || environment.Head != sha || environment.LogPath != "" {
		t.Fatalf("preflight must bind check/head without inventing a log: %+v", environment)
	}
	if !strings.Contains(environment.Error(), "could-not-run") || strings.Contains(environment.Error(), "; log ") {
		t.Fatalf("preflight diagnostic: %v", err)
	}
}

// Read-only observation of the fixture's execution/evidence directory. A
// preflight must not create even an empty inflight writer or queue marker.
func capabilityEvidenceFiles(t *testing.T, d localRunDeps) map[string]string {
	t.Helper()
	root := filepath.Join(d.Root, ".git", "sdlc")
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			files[path] = "directory"
			return nil
		}
		raw, err := os.ReadFile(path)
		if err == nil {
			files[path] = string(raw)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestSuiteCapabilityOrdinaryEntryRefusesBeforeEffects(t *testing.T) {
	for _, mode := range []string{"suite", "affected", "baseline", "typed"} {
		t.Run(mode, func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "acquired")
			c := suiteCheck("true")
			if mode == "affected" {
				c.Name = "affected"
			}
			d.SkipRecord = mode == "baseline"
			var raw string
			if mode == "typed" {
				c, raw = typedBindingFixture(t, "bash")
			}
			git := d.Git
			d.Git = func(args ...string) (string, error) {
				if args[0] == "worktree" {
					t.Fatalf("worktree IO before capability authorization: %v", args)
				}
				if raw != "" && args[0] == "ls-tree" {
					return localChecksPath, nil
				}
				if raw != "" && args[0] == "show" && args[1] == sha+":"+localChecksPath {
					return raw, nil
				}
				return git(args...)
			}
			// An absent helper must not mask the independent environment refusal.
			if err := os.Remove(filepath.Join(d.Root, "Makefile.workflow")); err != nil {
				t.Fatal(err)
			}
			calls := 0
			d.suiteCapabilities = func(context.Context) error { calls++; return suiteCapabilityDenied() }
			d.FlakeRegistry = func(context.Context) (flakes.Registry, error) {
				t.Fatal("capability refusal entered flake registry")
				return flakes.Registry{}, nil
			}
			d.normalTempDir = func() (string, error) {
				t.Fatal("execution scratch allocated before capability authorization")
				return "", nil
			}
			before := capabilityEvidenceFiles(t, d)
			rec, err := runLocalCheck(context.Background(), d, c, sha)
			assertCapabilityRefusal(t, err, c.Name, sha)
			if calls != 1 || rec.Exit != -1 || rec.LogPath != "" {
				t.Fatalf("calls=%d record=%+v", calls, rec)
			}
			if after := capabilityEvidenceFiles(t, d); !reflect.DeepEqual(before, after) {
				t.Fatalf("refusal allocated execution/evidence state: before=%v after=%v", before, after)
			}
			if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
				t.Fatalf("refusal published: %v %v", gh.statuses, gh.comments)
			}
		})
	}
}

func TestSuiteCapabilityRepairSameHeadAndNonTestBypass(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	calls := 0
	denied := true
	d.suiteCapabilities = func(context.Context) error {
		calls++
		if denied {
			return suiteCapabilityDenied()
		}
		return nil
	}
	c := suiteCheck("true")
	_, err := runLocalCheck(context.Background(), d, c, sha)
	assertCapabilityRefusal(t, err, suiteCheckName, sha)
	if rows := gateAttempts(t, d); len(rows) != 0 || len(gh.statuses[sha]) != 0 {
		t.Fatalf("refusal became a failed attempt: %+v %v", rows, gh.statuses)
	}
	denied = false
	if rec, err := runLocalCheck(context.Background(), d, c, sha); err != nil || !rec.passed() {
		t.Fatalf("same-head repair refused: %+v %v", rec, err)
	}
	if calls != 2 || len(gateAttempts(t, d)) != 1 || gh.statuses[sha][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("same-head repair: calls=%d attempts=%+v statuses=%v", calls, gateAttempts(t, d), gh.statuses)
	}
	denied = true
	c.Name = "ordinary-check"
	if rec, err := runLocalCheck(context.Background(), d, c, sha); err != nil || !rec.passed() || calls != 2 {
		t.Fatalf("non-test check entered capability probe: %+v %v calls=%d", rec, err, calls)
	}
}

func TestSuiteCapabilityAutomaticRerunKeepsGenuineParent(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	marker := filepath.Join(t.TempDir(), "child-ran")
	scriptRerun(t, `printf child > "`+marker+`"; `+gatePassJSON)
	calls := 0
	d.suiteCapabilities = func(context.Context) error {
		calls++
		if calls == 1 {
			return nil
		}
		return suiteCapabilityDenied()
	}
	_, err := runLocalCheck(context.Background(), d, suiteFailing(gateKnown), sha)
	assertCapabilityRefusal(t, err, suiteCheckName, sha)
	rows := gateAttempts(t, d)
	if calls != 2 || len(rows) != 1 || rows[0].Check != suiteCheckName || rows[0].Exit != 1 || rows[0].RerunOf != "" || rows[0].Outcome == string(flakes.KnownFlakePassed) {
		t.Fatalf("genuine failure lost or invalid child saved: calls=%d attempts=%+v", calls, rows)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("child executed despite refusal: %v", err)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("refused child published status: %v %v", gh.statuses, gh.comments)
	}
}

func TestSuiteCapabilityResumeEligibilityBeforeEffects(t *testing.T) {
	d, sha, store, link := resumeSavedFixture(t, false)
	if err := os.Remove(filepath.Join(store, "attempts", link.ID+".json")); err != nil {
		t.Fatal(err)
	}
	// Prove the setup is eligible rather than merely blocked by an old link.
	parentRaw, err := os.ReadFile(filepath.Join(store, "attempts", "aa.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := lockResumeParent(context.Background(), store, "aa")
	if err != nil {
		t.Fatal(err)
	}
	in, _, err := resumeInputs(context.Background(), d, store, lock.Attempt)
	lock.Close()
	if err != nil || resolveResumeProof(in).State != ResumeEligible {
		t.Fatalf("fixture not eligible: %+v %v", resolveResumeProof(in), err)
	}
	git := d.Git
	d.Git = func(args ...string) (string, error) {
		if args[0] == "worktree" {
			t.Fatalf("resume touched worktree before preflight: %v", args)
		}
		return git(args...)
	}
	if err := os.Remove(filepath.Join(d.Root, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	calls := 0
	d.suiteCapabilities = func(context.Context) error { calls++; return suiteCapabilityDenied() }
	before := capabilityEvidenceFiles(t, d)
	_, err = resumeSavedSuite(context.Background(), d, suiteCheck("true"), sha, "aa")
	assertCapabilityRefusal(t, err, suiteCheckName, sha)
	if calls != 1 || !strings.Contains(err.Error(), "sdlc local-check resume --name suite --attempt aa --rev "+sha) {
		t.Fatalf("resume retry identity lost: calls=%d err=%v", calls, err)
	}
	if after := capabilityEvidenceFiles(t, d); !reflect.DeepEqual(before, after) {
		t.Fatalf("resume refusal created child/link or changed parent: before=%v after=%v", before, after)
	}
	if after, _ := os.ReadFile(filepath.Join(store, "attempts", "aa.json")); string(after) != string(parentRaw) {
		t.Fatal("refusal rewrote genuine parent")
	}
	gh := d.GH.(*ciFakeGH)
	if len(gh.statuses[sha]) != 0 {
		t.Fatalf("resume refusal published: %v", gh.statuses)
	}
}

func TestSuiteCapabilitySavedResumeRepostBypassesProbe(t *testing.T) {
	d, sha, _, _ := resumeSavedFixture(t, true)
	d.suiteCapabilities = func(context.Context) error {
		t.Fatal("saved-pass repost probed execution capabilities")
		return suiteCapabilityDenied()
	}
	rec, err := resumeSavedSuite(context.Background(), d, suiteCheck("false"), sha, "aa")
	if err != nil || !rec.passed() || rec.ResumeParent != "aa" || rec.ResumeChild != "cc" {
		t.Fatalf("saved-pass repost failed: %+v %v", rec, err)
	}
	if gh := d.GH.(*ciFakeGH); gh.statuses[sha][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("saved-pass repost missing: %v", gh.statuses)
	}
}

func TestSuiteCapabilityAffectedGraphBeforeEffects(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	d.suiteCapabilities = func(context.Context) error { return suiteCapabilityDenied() }
	ad := affectedDeps{
		changed: func(string) ([]string, error) { return []string{"code.go"}, nil },
		graph: func(context.Context, string) (affected.Graph, error) {
			t.Fatal("affected graph launched before preflight")
			return affected.Graph{}, nil
		},
		run: func(context.Context, localCheck, string) (localRunRecord, error) {
			t.Fatal("affected execution before preflight")
			return localRunRecord{}, nil
		},
		post: func(context.Context, localRunRecord) error { t.Fatal("affected status before preflight"); return nil },
	}
	_, err := runAffectedCheck(context.Background(), d, ad, affected.Config{}, suiteCheck("true"), sha, affectedOrdinary)
	assertCapabilityRefusal(t, err, affectedCheckName, sha)
	if len(gateAttempts(t, d)) != 0 || len(gh.statuses[sha]) != 0 {
		t.Fatal("affected preflight created evidence")
	}
}

func TestSuiteCapabilityResumeRetryAfterRepair(t *testing.T) {
	d, sha, store, link := resumeSavedFixture(t, false)
	if err := os.Remove(filepath.Join(store, "attempts", link.ID+".json")); err != nil {
		t.Fatal(err)
	}
	d.suiteCapabilities = func(context.Context) error { return suiteCapabilityDenied() }
	_, err := resumeSavedSuite(context.Background(), d, suiteCheck("true"), sha, "aa")
	assertCapabilityRefusal(t, err, suiteCheckName, sha)
	d.suiteCapabilities = func(context.Context) error { return nil }
	scriptRerun(t, `printf '%s\n' '{"Action":"pass","Package":"example.org/p","Test":"TestOne"}' '{"Action":"pass","Package":"example.org/p","Elapsed":1}'`)
	rec, err := resumeSavedSuite(context.Background(), d, suiteCheck("true"), sha, "aa")
	if err != nil || !rec.passed() || rec.ResumeParent != "aa" {
		t.Fatalf("repair did not permit exact retry: %+v %v", rec, err)
	}
}

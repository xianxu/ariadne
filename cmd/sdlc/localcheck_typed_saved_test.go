package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func typedSavedFixture(t *testing.T) (localRunDeps, string, string, SuiteAttempt) {
	t.Helper()
	d, sha, _, _ := localRunRepo(t, "unused")
	raw := `{"version":3,"checks":[{"name":"suite","command":["make","test"],"flake_runners":{"lua":{"command":["scripts/spec.sh","{path}"],"results":"plenary-events-v1"}}}]}`
	decl, err := parseLocalChecks([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	original := d.Git
	d.Git = func(args ...string) (string, error) {
		if args[0] == "fetch" {
			return "", nil
		}
		if args[0] == "rev-parse" && len(args) > 2 && args[2] == "origin/main^{commit}" {
			return sha, nil
		}
		if args[0] == "ls-tree" && args[len(args)-1] == localChecksPath {
			return localChecksPath, nil
		}
		if args[0] == "show" && args[1] == sha+":"+localChecksPath {
			return raw, nil
		}
		return original(args...)
	}
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyRerun, nil }
	d.FlakeRegistry = gateRegistry(t, "lua:tests/native_spec.lua#native retry")
	logdir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(logdir)
	a := storedAttemptFixture("aa")
	a.Version = 4
	a.Repo = d.Repo
	a.SHA = sha
	a.Exit = 0
	a.Failures = nil
	a.Outcome = string(flakes.Passed)
	a.LogPath = "failed/aa.log"
	wall := 1.0
	a.WallSeconds = &wall
	a.RunnerContractSHA256 = decl.Checks[0].FlakeRunners.ContractDigest()
	log := []byte(typedPassingLog(t, "lua:tests/native_spec.lua", "native retry"))
	a.LogSHA256 = resumeRecordDigest(log)
	resumeSavedWrite(t, store, a, log)
	return d, sha, store, a
}

func TestTypedSavedSelectionRequiresCurrentAuthority(t *testing.T) {
	d, sha, store, a := typedSavedFixture(t)
	if _, _, err := loadSavedAttempt(store, d.Repo, sha, suiteCheckName, a.ID, "h"); err == nil {
		t.Fatal("legacy loader reused typed authority")
	}
	rec, newest, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, a.ID, "h")
	if err != nil || !newest || !rec.passed() {
		t.Fatalf("valid typed pass %+v newest=%v err=%v", rec, newest, err)
	}
	original := d.Git
	d.Git = func(args ...string) (string, error) {
		out, err := original(args...)
		if args[0] == "show" && strings.Contains(args[1], localChecksPath) {
			out = strings.Replace(out, "scripts/spec.sh", "scripts/other.sh", 1)
		}
		return out, err
	}
	if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, a.ID, "h"); err == nil {
		t.Fatal("changed runner contract reused")
	}
}

func TestTypedSavedSelectionReprovesRawLogFacts(t *testing.T) {
	for _, tc := range []struct {
		name, log string
		failures  []string
	}{
		{name: "malformed", log: `{"sdlc_unit":1`},
		{name: "failure despite outer zero", log: "failure"},
		{name: "record mismatch", failures: []string{"lua:tests/native_spec.lua#native retry"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, sha, store, a := typedSavedFixture(t)
			log := tc.log
			if tc.name == "failure despite outer zero" {
				log = typedLogFrames(t, "lua:tests/native_spec.lua", "native retry", 1)
				a.Failures = []string{"lua:tests/native_spec.lua#native retry"}
			}
			if tc.name == "record mismatch" {
				log = typedPassingLog(t, "lua:tests/native_spec.lua", "native retry")
				a.Failures = tc.failures
			}
			a.LogSHA256 = resumeRecordDigest([]byte(log))
			resumeSavedWrite(t, store, a, []byte(log))
			if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, a.ID, "h"); err == nil {
				t.Fatal("unproved v4 pass reused")
			}
			if err := guardSuitePass(context.Background(), d, suiteCheckName, sha); err == nil {
				t.Fatal("guard accepted unproved v4 pass")
			}
		})
	}
}

func typedSavedAutomatic(t *testing.T) (localRunDeps, string, string, SuiteAttempt, SuiteAttempt) {
	t.Helper()
	d, sha, store, p := typedSavedFixture(t)
	fail := []byte(typedLogFrames(t, "lua:tests/native_spec.lua", "native retry", 1))
	p.Exit = 1
	p.Outcome = string(flakes.KnownFlakePassed)
	p.Failures = []string{"lua:tests/native_spec.lua#native retry"}
	p.KnownFlakes = []flakes.Entry{{Test: p.Failures[0], Issue: "000999"}}
	p.LogSHA256 = resumeRecordDigest(fail)
	resumeSavedWrite(t, store, p, fail)
	c := p
	c.ID = "bb"
	c.Check = suiteRerunCheck
	c.RerunOf = p.ID
	c.Exit = 0
	c.Failures = nil
	c.LogPath = "failed/bb.log"
	c.StartedAt = p.FinishedAt.Add(time.Second)
	c.FinishedAt = c.StartedAt.Add(time.Second)
	pass := []byte(typedPassingLog(t, "lua:tests/native_spec.lua", "native retry"))
	c.LogSHA256 = resumeRecordDigest(pass)
	resumeSavedWrite(t, store, c, pass)
	return d, sha, store, p, c
}

func TestTypedSavedAutomaticRerunReprovesExactCase(t *testing.T) {
	d, sha, store, p, c := typedSavedAutomatic(t)
	rec, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, p.ID, "h")
	if err != nil || !rec.passed() {
		t.Fatalf("valid automatic rerun %+v %v", rec, err)
	}
	if err := guardSuitePass(context.Background(), d, suiteCheckName, sha); err != nil {
		t.Fatal(err)
	}
	log := []byte(typedPassingLog(t, "lua:tests/native_spec.lua", "another case"))
	c.LogSHA256 = resumeRecordDigest(log)
	resumeSavedWrite(t, store, c, log)
	if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, p.ID, "h"); err == nil {
		t.Fatal("unrelated case promoted automatic parent")
	}
	if err := guardSuitePass(context.Background(), d, suiteCheckName, sha); err == nil {
		t.Fatal("guard trusted automatic claim without case")
	}
}

func TestTypedSavedAutomaticChildBounds(t *testing.T) {
	for _, variant := range []string{"absent", "duplicate", "contract", "legacy", "rejected"} {
		t.Run(variant, func(t *testing.T) {
			d, sha, store, p, c := typedSavedAutomatic(t)
			switch variant {
			case "absent":
				if err := os.Remove(filepath.Join(store, "attempts", c.ID+".json")); err != nil {
					t.Fatal(err)
				}
			case "duplicate":
				duplicate := c
				duplicate.ID = "cc"
				duplicate.LogPath = "failed/cc.log"
				resumeSavedWrite(t, store, duplicate, []byte(typedPassingLog(t, "lua:tests/native_spec.lua", "native retry")))
			case "contract":
				c.RunnerContractSHA256 = strings.Repeat("d", 64)
			case "legacy":
				c.Version = 2
				c.RunnerContractSHA256 = ""
				c.KnownFlakes = nil
			case "rejected":
				c.Outcome = string(flakes.RerunInconclusive)
			}
			if variant != "absent" && variant != "duplicate" {
				resumeSavedWriteJSON(t, store, c.ID, c)
			}
			if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, p.ID, "h"); err == nil {
				t.Fatal("unproved child accepted")
			}
		})
	}
}

func TestTypedSavedSnapshotUncertaintyRefuses(t *testing.T) {
	d, sha, store, a := typedSavedFixture(t)
	if err := os.WriteFile(filepath.Join(store, "attempts", "unreadable.json"), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, a.ID, "h"); err == nil {
		t.Fatal("incomplete snapshot supplied typed proof")
	}
}

func TestTypedSavedAcceptRequiresCurrentPolicy(t *testing.T) {
	d, sha, store, p, c := typedSavedAutomatic(t)
	if err := os.Remove(filepath.Join(store, "attempts", c.ID+".json")); err != nil {
		t.Fatal(err)
	}
	p.Outcome = string(flakes.KnownFlakeAccept)
	resumeSavedWriteJSON(t, store, p.ID, p)
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyAccept, nil }
	rec, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, p.ID, "h")
	if err != nil || !rec.passed() {
		t.Fatalf("current accept policy: %+v %v", rec, err)
	}
	d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyRerun, nil }
	if _, _, err := loadSavedAttemptCurrent(context.Background(), d, store, sha, suiteCheckName, p.ID, "h"); err == nil {
		t.Fatal("accept survived policy revoke")
	}
}

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func typedBindingFixture(t *testing.T, command string) (localCheck, string) {
	t.Helper()
	raw := `{"version":3,"checks":[{"name":"suite","command":["make","test"],"flake_runners":{"script":{"command":["` + command + `","{path}"]}}}]}`
	decl, err := parseLocalChecks([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return decl.Checks[0], raw
}

func TestTypedCheckBindingPinsTrustedContract(t *testing.T) {
	c, raw := typedBindingFixture(t, "bash")
	sha := strings.Repeat("a", 40)
	var calls [][]string
	d := localRunDeps{Git: func(args ...string) (string, error) {
		calls = append(calls, args)
		if args[0] == "ls-tree" {
			return localChecksPath, nil
		}
		return raw, nil
	}}
	bound, err := bindTypedCheck(d, c, sha)
	if err != nil {
		t.Fatal(err)
	}
	if !bound.StrictGroupCleanup || bound.runnerLifetime == nil || !bound.runnerLifetime.mayRetire() {
		t.Fatalf("typed lifetime not shared: %+v", bound)
	}
	if !reflect.DeepEqual(calls, [][]string{{"ls-tree", "--full-tree", "--name-only", sha, localChecksPath}, {"show", sha + ":" + localChecksPath}}) {
		t.Fatalf("not pinned: %v", calls)
	}
	env := bound.typedUnitEnvironment([]string{"KEEP=yes", flakeUnitSHAEnv + "=old", flakeUnitContractEnv + "=old", flakeUnitNameEnv + "=old", flakeUnitResultsEnv + "=old"})
	want := []string{"KEEP=yes", flakeUnitSHAEnv + "=" + sha, flakeUnitContractEnv + "=" + c.FlakeRunners.ContractDigest()}
	if !reflect.DeepEqual(env, want) {
		t.Fatalf("env %v; want %v", env, want)
	}
	for _, check := range []string{suiteCheckName, affectedCheckName, suiteRerunCheck} {
		a := SuiteAttempt{Version: suiteAttemptVersion, Check: check}
		bound.bindSuiteAttempt(&a)
		if a.Version != 4 || a.RunnerContractSHA256 != c.FlakeRunners.ContractDigest() {
			t.Fatalf("attempt not bound: %+v", a)
		}
	}
}

func TestTypedCheckBindingRefusesBeforeExecution(t *testing.T) {
	c, _ := typedBindingFixture(t, "bash")
	_, different := typedBindingFixture(t, "sh")
	for _, tc := range []struct {
		name, raw string
		readErr   error
	}{
		{name: "different", raw: different}, {name: "absent"}, {name: "unreadable", readErr: errors.New("read failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := localRunDeps{Git: func(args ...string) (string, error) {
				if tc.readErr != nil {
					return "", tc.readErr
				}
				if args[0] == "ls-tree" && tc.raw != "" {
					return localChecksPath, nil
				}
				return tc.raw, nil
			}}
			if _, err := bindTypedCheck(d, c, strings.Repeat("a", 40)); err == nil {
				t.Fatal("accepted unverified typed authority")
			}
		})
	}
}

func TestTypedCheckBindingLegacyStripsEnvelopeWithoutGit(t *testing.T) {
	d := localRunDeps{Git: func(...string) (string, error) { t.Fatal("legacy check read typed contract"); return "", nil }}
	d, err := bindTypedCheck(d, localCheck{Name: suiteCheckName}, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if d.StrictGroupCleanup || d.runnerLifetime != nil {
		t.Fatal("legacy lifetime changed")
	}
	if got := d.typedUnitEnvironment([]string{"KEEP=yes", flakeUnitSHAEnv + "=stale", flakeUnitContractEnv + "=stale", flakeUnitNameEnv + "=stale", flakeUnitResultsEnv + "=stale"}); !reflect.DeepEqual(got, []string{"KEEP=yes"}) {
		t.Fatalf("stale authority: %v", got)
	}
	a := SuiteAttempt{Version: suiteAttemptVersion, Check: suiteCheckName}
	d.bindSuiteAttempt(&a)
	if a.Version != suiteAttemptVersion || a.RunnerContractSHA256 != "" {
		t.Fatalf("legacy writer changed: %+v", a)
	}
}

func TestTypedCheckBindingRetainsSharedLifetime(t *testing.T) {
	c, raw := typedBindingFixture(t, "bash")
	lifetime := &runnerArtifactLifetime{retirable: false}
	d := localRunDeps{runnerLifetime: lifetime, Git: func(args ...string) (string, error) {
		if args[0] == "ls-tree" {
			return localChecksPath, nil
		}
		return raw, nil
	}}
	bound, err := bindTypedCheck(d, c, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if bound.runnerLifetime != lifetime || bound.runnerLifetime.mayRetire() {
		t.Fatal("binding reset unconfirmed retirement")
	}
}

func TestTypedCheckBindingRejectsExitZeroWithInvalidFacts(t *testing.T) {
	for _, facts := range []SuiteLogFacts{{Failed: true}, {Unattributed: []string{"malformed unit frame"}}, {Warnings: []string{"truncated line"}}} {
		rec := localRunRecord{Exit: 0}
		localRunDeps{flakeContract: strings.Repeat("a", 64)}.applyTypedRunFacts(&rec, facts)
		if rec.Exit != 0 || rec.Outcome == "" || rec.passed() {
			t.Fatalf("invalid typed pass: %+v", rec)
		}
	}
	rec := localRunRecord{Exit: 0}
	localRunDeps{}.applyTypedRunFacts(&rec, SuiteLogFacts{Failed: true})
	if rec.Outcome != "" {
		t.Fatal("legacy success changed")
	}
}

func TestTypedCheckBindingDoesNotGrantResumeOwnerRelease(t *testing.T) {
	c, raw := typedBindingFixture(t, "bash")
	d := localRunDeps{Git: func(args ...string) (string, error) {
		if args[0] == "ls-tree" {
			return localChecksPath, nil
		}
		return raw, nil
	}}
	bound, err := bindTypedCheck(d, c, strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(t.TempDir(), "raw.log")
	marker := filepath.Join(filepath.Dir(logPath), runnerArtifactOwnerFile)
	if err := os.WriteFile(marker, []byte("manual-resume-owner"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := bound.releaseTypedAttemptOwner(context.Background(), logPath); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(marker); err != nil || string(raw) != "manual-resume-owner" {
		t.Fatalf("binding removed resume marker: %q %v", raw, err)
	}
	bound.ordinaryTypedOwner = true
	if err := bound.releaseTypedAttemptOwner(context.Background(), logPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("ordinary owner not released: %v", err)
	}
}

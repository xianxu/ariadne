package main

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Host conformance exercises production preflight under actual kernel denial.
// It runs a tiny command fixture, never a nested suite or queue waiter.
func TestSuiteCapabilityNativeSandbox(t *testing.T) {
	const child = "ARIADNE_572_NATIVE_CHILD"
	if mode := os.Getenv(child); mode != "" {
		d, sha, gh, _ := localRunRepo(t, "skipped")
		before := capabilityEvidenceFiles(t, d)
		d.Git = func(args ...string) (string, error) {
			t.Fatalf("preflight denial reached Git/execution preparation: %v", args)
			return "", nil
		}
		_, err := runLocalCheck(context.Background(), d, suiteCheck("true"), sha)
		assertCapabilityRefusal(t, err, suiteCheckName, sha)
		if !strings.Contains(err.Error(), mode) || !strings.Contains(err.Error(), "whole Bash call") {
			t.Fatalf("missing capability/guidance: %v", err)
		}
		if !reflect.DeepEqual(before, capabilityEvidenceFiles(t, d)) || len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
			t.Fatal("native refusal created queue/evidence/status effects")
		}
		t.Logf("native %s denial: %v; queue/attempt/status unchanged", mode, err)
		return
	}
	if runtime.GOOS != "darwin" || os.Getenv("ARIADNE_NATIVE_SANDBOX") != "1" {
		t.Skip("host conformance: ARIADNE_NATIVE_SANDBOX=1 on macOS")
	}
	if err := suiteHostCapabilities(context.Background()); err != nil {
		t.Fatalf("host capability control unavailable: %v", err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, rule string }{
		{"boot identity", `(deny sysctl-read (sysctl-name "kern.boottime"))`},
		{"Unix listener", "(deny network-bind)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "sandbox-exec", "-p", "(version 1)(allow default)"+tc.rule,
				executable, "-test.run=^TestSuiteCapabilityNativeSandbox$", "-test.v", "-test.timeout=45s")
			cmd.Env = append(os.Environ(), child+"="+tc.name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native %s: %v\n%s", tc.name, err, out)
			}
			if !strings.Contains(string(out), "queue/attempt/status unchanged") {
				t.Fatalf("missing child proof: %s", out)
			}
			t.Logf("%s", out)
		})
	}
	// A fresh real listener proves the child denial did not mutate host state.
	if err := suiteHostCapabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
}

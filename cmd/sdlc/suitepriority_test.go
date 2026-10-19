package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestSuitePrioritySubject(t *testing.T) {
	repo := landingRepo(t)
	git(t, repo, "-c", "user.name=T", "-c", "user.email=t@e.com", "commit", "--allow-empty", "-qm", "initial")
	git(t, repo, "checkout", "-qb", "000452-suite-priority")
	sha := git(t, repo, "rev-parse", "HEAD")
	for _, tc := range []struct {
		sha  string
		want int
	}{{"", 452}, {strings.TrimSpace(sha), 452}, {"different", 0}} {
		got, err := suiteSubject(repo, tc.sha)
		if err != nil || got.Issue != tc.want || !filepath.IsAbs(got.CommonDir) {
			t.Fatalf("subject %+v, %v; want %d", got, err, tc.want)
		}
	}
	git(t, repo, "checkout", "--detach", "-q")
	got, err := suiteSubject(repo, "")
	if err != nil || got.Issue != 0 {
		t.Fatalf("detached %+v %v", got, err)
	}
}

func TestSuitePrioritySnapshotLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	common := filepath.Join(t.TempDir(), "repo.git")
	subjects := []suitePrioritySubject{{CommonDir: common, Issue: 452}, {CommonDir: common, Issue: 453}}
	// A read of absent logs must not create the landing directory.
	got := suiteGrantSnapshot(subjects, now)
	if len(got) != 1 || len(got[0].Deadlines) != 0 || got[0].Error != "" {
		t.Fatalf("absent %+v", got)
	}
	if _, err := os.Stat(common); !os.IsNotExist(err) {
		t.Fatalf("read created files: %v", err)
	}
	dir := filepath.Join(common, "sdlc", "landing")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	grant := func(n int) landingline.Grant {
		return landingline.Grant{Issue: n, Grantor: landingline.Grantor{Address: "ops:0"}, Reason: "fleet blocked", Granted: now.Add(-time.Minute), Expires: now.Add(time.Hour)}
	}
	active, future, unauthorized := grant(452), grant(453), grant(454)
	future.Granted = now.Add(time.Minute)
	slot := 3
	unauthorized.Grantor.Slot = &slot
	writeSuiteGrants(t, dir, active, future, unauthorized)
	got = suiteGrantSnapshot(append(subjects, suitePrioritySubject{CommonDir: common, Issue: 454}), now)
	if len(got) != 1 || len(got[0].Deadlines) != 1 || got[0].Deadlines["452"] != float64(active.Expires.Unix()) || got[0].RefreshAt != float64(future.Granted.Unix()) {
		t.Fatalf("active %+v", got)
	}
	got = suiteGrantSnapshot(subjects, now.Add(2*time.Minute))
	if len(got[0].Deadlines) != 2 {
		t.Fatalf("future activation %+v", got)
	}
	got = suiteGrantSnapshot(subjects, now.Add(time.Hour))
	if len(got[0].Deadlines) != 0 {
		t.Fatalf("expired %+v", got)
	}
	revoke := active
	revoke.Revoked = true
	writeSuiteGrants(t, dir, active, revoke)
	if got = suiteGrantSnapshot(subjects, now); len(got[0].Deadlines) != 0 {
		t.Fatalf("revoked %+v", got)
	}
	// A torn revoke could conceal the latest decision: never confer priority.
	f, err := os.OpenFile(filepath.Join(dir, landingline.GrantsFile), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, "{torn")
	f.Close()
	if got = suiteGrantSnapshot(subjects, now); got[0].Error == "" || len(got[0].Deadlines) != 0 {
		t.Fatalf("malformed %+v", got)
	}
}

func writeSuiteGrants(t *testing.T, dir string, grants ...landingline.Grant) {
	t.Helper()
	var buf bytes.Buffer
	for _, g := range grants {
		if err := json.NewEncoder(&buf).Encode(g); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, landingline.GrantsFile), buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSuitePriorityProtocol(t *testing.T) {
	now := time.Now()
	for _, input := range []string{`{"version":2,"subjects":[]}`, `{"version":1,"subjects":[],"unknown":1}`, `{"version":1,"subjects":[{"common_dir":"relative","issue":1}]}`, `{"version":1,"subjects":[{"common_dir":"/tmp/r","issue":-1}]}`, `{"version":1,"subjects":[]} {}`, strings.Repeat(" ", 65537)} {
		var out bytes.Buffer
		if err := runSuitePriority(&out, strings.NewReader(input), false, ".", now); err == nil {
			t.Fatalf("accepted %.100s", input)
		}
	}
	var out bytes.Buffer
	if err := runSuitePriority(&out, strings.NewReader(`{"version":1,"subjects":[]}`), false, ".", now); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != `{"version":1,"sources":[]}` {
		t.Fatalf("response %s", out.String())
	}
}

// A subprocess launcher enters the actual protocol, not a fake priority policy.
// SDLC_SUITE_PRIORITY_PROCESS=1 <test binary> -test.run '^TestSuitePriorityProviderProcess$' -- landing suite-priority ...
func TestSuitePriorityProviderProcess(t *testing.T) {
	if os.Getenv("SDLC_SUITE_PRIORITY_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) > 0 && args[0] == "landing" {
		args = args[1:]
	}
	cmd := newLandingCmd(defaultLandingDeps())
	cmd.SetArgs(args)
	cmd.SetIn(os.Stdin)
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestSuiteQueuePriorityRealHelper(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "provider")
	// A literal shell-quoted executable path, not caller-supplied command text.
	quoted := "'" + strings.ReplaceAll(binary, "'", "'\"'\"'") + "'"
	script := "#!/bin/sh\nexport SDLC_SUITE_PRIORITY_PROCESS=1\nexec " + quoted + " -test.run '^TestSuitePriorityProviderProcess$' -- \"$@\"\n"
	if err := os.WriteFile(launcher, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", "../../scripts/test/test-lock-priority.test.py")
	cmd.Env = append(os.Environ(), "SUITE_PRIORITY_TEST_PROVIDER="+launcher)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("real helper: %v\n%s", err, out)
	}
}

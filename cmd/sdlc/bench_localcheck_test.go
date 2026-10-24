package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBenchSuitePrivateRepoIdentity(t *testing.T) {
	for _, repo := range []string{"owner/repo", "file:/bench/public/origin.git"} {
		a := SuiteAttempt{Version: 1, ID: "a1", Repo: repo, SHA: strings.Repeat("a", 40), Check: "suite", StartedAt: time.Now().Add(-time.Second).UTC(), FinishedAt: time.Now().UTC(), LogSHA256: strings.Repeat("b", 64)}
		raw, _ := json.Marshal(a)
		if _, err := decodeSuiteAttempt(raw); err != nil {
			t.Fatalf("identity %s: %v", repo, err)
		}
		future := strings.Replace(string(futureProofFixture("d3")), "x/ariadne", repo, 1)
		if _, err := decodeUnsupportedProof([]byte(future), "d3.json"); err != nil {
			t.Fatalf("future private identity %s: %v", repo, err)
		}
	}
	for _, repo := range []string{"file:relative", "file:/a/../b", "file://host/path", "file:/", "file:/bad\npath", "file:/bad\x00path", "https://example.com/repo", "owner/repo/extra"} {
		if validSuiteRepo(repo) {
			t.Fatalf("invalid repo accepted %q", repo)
		}
	}
}

func TestBenchReceiptPreservesFutureFlakeEntryFields(t *testing.T) {
	old := []byte(`{"record":{"KnownFlakes":[{"test":"TestX","issue":"42","future":{"retained":true}}]}}`)
	current := []byte(`{"record":{"KnownFlakes":[{"test":"TestX","issue":"42"}]}}`)
	got, err := mergeBenchStatusExtensions(old, current)
	if err != nil || !strings.Contains(string(got), `"future":{"retained":true}`) {
		t.Fatalf("nested array fields lost: %s %v", got, err)
	}
	changed := []byte(`{"record":{"KnownFlakes":[{"test":"TestY","issue":"43"}]}}`)
	got, err = mergeBenchStatusExtensions(old, changed)
	if err != nil || strings.Contains(string(got), `"future"`) {
		t.Fatalf("extension moved to another flake: %s %v", got, err)
	}
}

func benchLocalCheckFixture(t *testing.T) (benchPublicContext, string) {
	t.Helper()
	b := benchContextFixture(t)
	prepareTestGit(t, b.WorkRoot, "config", "user.name", "Bench Test")
	prepareTestGit(t, b.WorkRoot, "config", "user.email", "bench@example.invalid")
	prepareTestWrite(t, b.WorkRoot, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["sh","-c","echo bench-proof"],"timeout_minutes":1,"weight":1}]}`, 0600)
	prepareTestWrite(t, b.WorkRoot, "Makefile.workflow", "# fixture\n", 0600)
	prepareTestWrite(t, b.WorkRoot, "scripts/test-lock.py", fakeSuiteLockHelper("raise Exception('historical admission must not run')\n"), 0755)
	prepareTestGit(t, b.WorkRoot, "add", localChecksPath, "Makefile.workflow", "scripts/test-lock.py")
	prepareTestGit(t, b.WorkRoot, "commit", "-qm", "suite fixture")
	prepareTestGit(t, b.WorkRoot, "push", "-q", "-u", "origin", "main")
	return b, prepareTestGit(t, b.WorkRoot, "rev-parse", "HEAD")
}

func TestBenchLocalCheckCLIPrivateReceiptNoGitHub(t *testing.T) {
	// Opt in after building once outside the suite; ordinary tests exercise
	// the same Cobra command without rebuilding the full executable per test.
	cliBinary := os.Getenv("ARIADNE_BENCH_TEST_SDLC_BINARY")
	if cliBinary != "" {
		if !filepath.IsAbs(cliBinary) {
			t.Fatal("ARIADNE_BENCH_TEST_SDLC_BINARY must be absolute")
		}
		info, err := os.Stat(cliBinary)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
			t.Fatalf("built bench CLI unavailable: %s: %v", cliBinary, err)
		}
		t.Logf("production CLI conformance: %s", cliBinary)
	}
	b, head := benchLocalCheckFixture(t)
	sentinel := filepath.Join(t.TempDir(), "gh-called")
	bin := t.TempDir()
	prepareTestWrite(t, bin, "gh", "#!/bin/sh\nprintf 'invoked\\n' >> \"$BENCH_GH_SENTINEL\"\nexit 89\n", 0755)
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestBenchPrepareGateHelper$", "--"}, args...)...)
		if cliBinary != "" {
			cmd = exec.Command(cliBinary, args...)
		}
		cmd.Dir = b.WorkRoot
		cmd.Env = append(benchPrepareEnv(), "ARIADNE_BENCH_PREPARE_TEST_GATE=1", "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "BENCH_GH_SENTINEL="+sentinel)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	run("local-check", "run", "--name", "suite")
	first, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite")
	if err != nil || !first.Record.passed() || first.Record.LogSHA256 == "" {
		t.Fatalf("run receipt %+v %v", first, err)
	}
	run("local-check", "republish", "--name", "suite")
	got, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite")
	if err != nil || got.Record.LogSHA256 != first.Record.LogSHA256 {
		t.Fatalf("republish receipt %+v %v", got, err)
	}
	// A newer writer's nested extension survives a later status publication.
	path, err := benchLocalCheckReceiptPath(&b, head, "suite")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(object["record"], &record); err != nil {
		t.Fatal(err)
	}
	record["future_measurement"] = json.RawMessage(`{"nested":{"kept":1e400}}`)
	object["record"], _ = json.Marshal(record)
	object["future_status"] = json.RawMessage(`{"kept":[true,null]}`)
	raw, _ = json.Marshal(object)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	run("local-check", "republish", "--name", "suite")
	raw, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"future_measurement":{"nested":{"kept":1e400}}`) || !strings.Contains(string(raw), `"future_status":{"kept":[true,null]}`) {
		t.Fatalf("future fields lost: %s", raw)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatalf("GH sentinel invoked or unreadable: %v", err)
	}
	if err := os.WriteFile(first.Record.LogPath, []byte("changed proof bytes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite"); err == nil {
		t.Fatal("stale receipt admitted after log tampering")
	}
}

func TestBenchLocalCheckReceiptRefusesSymlinkAndContextLoss(t *testing.T) {
	b, head := benchLocalCheckFixture(t)
	rec := localRunRecord{Name: "suite", SHA: head, LogSHA256: strings.Repeat("b", 64)}
	if err := writeBenchLocalCheckReceipt(context.Background(), b.WorkRoot, rec); err != nil {
		t.Fatal(err)
	}
	path, err := benchLocalCheckReceiptPath(&b, head, "suite")
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.Rename(path, outside); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite"); err == nil {
		t.Fatal("symlink receipt read")
	}
	if err := writeBenchLocalCheckReceipt(context.Background(), b.WorkRoot, rec); err == nil {
		t.Fatal("symlink receipt overwritten")
	}
	if err := os.Remove(filepath.Join(b.WorkRoot, benchPublicContextPath)); err != nil {
		t.Fatal(err)
	}
	if err := writeBenchLocalCheckReceipt(context.Background(), b.WorkRoot, rec); err == nil {
		t.Fatal("missing context treated as valid bench")
	}
}

func TestBenchLocalCheckReceiptRejectsLaterUnrecordedFailure(t *testing.T) {
	b, head := benchLocalCheckFixture(t)
	query := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", b.WorkRoot}, args...)...).CombinedOutput()
		return string(out), err
	}
	d, err := localCheckPublicationDeps(context.Background(), b.WorkRoot, query, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runLocalCheck(context.Background(), d, suiteCheck("true"), head); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite"); err != nil {
		t.Fatal(err)
	}
	d.RecordStatus = func(context.Context, localRunRecord) error { return errors.New("receipt device unavailable") }
	if _, err := runLocalCheck(context.Background(), d, suiteCheck("false"), head); err == nil {
		t.Fatal("failed receipt write hidden")
	}
	if _, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite"); err == nil {
		t.Fatal("old success receipt hid a newer saved failure")
	}
}

func TestBenchLocalCheckReceiptRevalidatesKnownFlakeAllowance(t *testing.T) {
	b, head := benchLocalCheckFixture(t)
	query := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", b.WorkRoot}, args...)...).CombinedOutput()
		return string(out), err
	}
	d, err := localCheckPublicationDeps(context.Background(), b.WorkRoot, query, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	// The captured run knew an open flake. The reader must consult the current
	// real registry/policy; that old classification alone cannot grant a pass.
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	c := suiteFailing(gateKnown)
	c.KnownFlakes = "accept"
	rec, err := runLocalCheck(context.Background(), d, c, head)
	if err != nil || !rec.passed() {
		t.Fatalf("known-flake run: %+v %v", rec, err)
	}
	if _, err := readBenchLocalCheckReceipt(context.Background(), b.WorkRoot, head, "suite"); err == nil {
		t.Fatal("old known-flake allowance accepted without current registry/policy")
	}
}

func TestBenchLocalCheckPublicationDepsRejectDriftWithoutGitHub(t *testing.T) {
	b, _ := benchLocalCheckFixture(t)
	oldDetect := detectRepo
	detectRepo = func() (string, error) { t.Fatal("bench path attempted production discovery"); return "", nil }
	t.Cleanup(func() { detectRepo = oldDetect })
	query := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", b.WorkRoot}, args...)...).CombinedOutput()
		return string(out), err
	}
	d, err := localCheckPublicationDeps(context.Background(), b.WorkRoot, query, io.Discard)
	if err != nil || d.GH != nil || d.PR != 0 || d.Repo != b.OriginIdentity || d.RecordStatus == nil {
		t.Fatalf("private deps %+v %v", d, err)
	}
	if _, err := localCheckPublicationDeps(context.Background(), b.HubRoot, query, io.Discard); err == nil {
		t.Fatal("hub publication allowed")
	}
	prepareTestGit(t, b.WorkRoot, "remote", "set-url", "origin", "https://github.com/example/production.git")
	if _, err := localCheckPublicationDeps(context.Background(), b.WorkRoot, query, io.Discard); err == nil {
		t.Fatal("moved origin allowed")
	}
	if err := d.RecordStatus(context.Background(), localRunRecord{}); err == nil {
		t.Fatal("cached recorder ignored moved origin")
	}
}

func TestBenchLocalCheckReceiptAfterSameHeadGuard(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "skipped")
	d.GH = nil
	calls := 0
	d.RecordStatus = func(context.Context, localRunRecord) error { calls++; return nil }
	if _, err := runLocalCheck(context.Background(), d, suiteCheck("false"), sha); err != nil {
		t.Fatalf("failure records honestly: %v %s", err, stderr)
	}
	if calls != 1 {
		t.Fatalf("failure callback calls %d", calls)
	}
	var unexplained *unexplainedFailureError
	if _, err := runLocalCheck(context.Background(), d, suiteCheck("true"), sha); !errors.As(err, &unexplained) {
		t.Fatalf("passing rerun must refuse: %v %s", err, stderr)
	}
	if calls != 1 {
		t.Fatal("receipt callback bypassed native same-head failure guard")
	}
}

func TestBenchLocalCheckReceiptCodec(t *testing.T) {
	b := &benchPublicContext{Version: 1, RunID: "bench-run", WorkRoot: "/bench/public/work", OriginIdentity: "file:/bench/public/origin.git"}
	head := strings.Repeat("a", 40)
	rec := localRunRecord{Name: "suite", SHA: head, LogSHA256: strings.Repeat("b", 64), Host: "test"}
	receipt := benchLocalCheckReceipt{Version: 1, RunID: b.RunID, WorkRoot: b.WorkRoot, OriginIdentity: b.OriginIdentity, Head: head, Check: "suite", Record: rec}
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeBenchLocalCheckReceipt(raw, b, head, "suite"); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*benchLocalCheckReceipt){
		func(r *benchLocalCheckReceipt) { r.Version = 2 }, func(r *benchLocalCheckReceipt) { r.RunID = "other" },
		func(r *benchLocalCheckReceipt) { r.Record.SHA = strings.Repeat("c", 40) },
		func(r *benchLocalCheckReceipt) { r.Record.Name = "other" }, func(r *benchLocalCheckReceipt) { r.OriginIdentity = "file:/production" },
	} {
		changed := receipt
		mutate(&changed)
		body, _ := json.Marshal(changed)
		if _, err := decodeBenchLocalCheckReceipt(body, b, head, "suite"); err == nil {
			t.Fatalf("accepted %+v", changed)
		}
	}
	duplicate := strings.Replace(string(raw), `"version":1`, `"version":1,"version":1`, 1)
	if _, err := decodeBenchLocalCheckReceipt([]byte(duplicate), b, head, "suite"); err == nil {
		t.Fatal("duplicate version accepted")
	}
	if _, err := decodeBenchLocalCheckReceipt([]byte(`{"version":1}`), b, head, "suite"); err == nil {
		t.Fatal("missing required receipt accepted")
	}
	for _, field := range []string{"Exit", "TimedOut", "Outcome", "Duration"} {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		var record map[string]json.RawMessage
		if err := json.Unmarshal(object["record"], &record); err != nil {
			t.Fatal(err)
		}
		delete(record, field)
		object["record"], _ = json.Marshal(record)
		body, _ := json.Marshal(object)
		if _, err := decodeBenchLocalCheckReceipt(body, b, head, "suite"); err == nil {
			t.Fatalf("missing %s fabricated a pass", field)
		}
	}
	for _, check := range []string{"../suite", "a/b", "", ".."} {
		if _, err := benchLocalCheckReceiptPath(b, head, check); err == nil {
			t.Fatalf("unsafe check %q", check)
		}
	}
}

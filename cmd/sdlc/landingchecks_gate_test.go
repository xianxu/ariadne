package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The same stateful execution seam is used for suite and external checks.
// Each launch produces real retained evidence and records exact order/head.
type landingCheckFake struct {
	t       *testing.T
	calls   []string
	fail    string
	badHead bool
}

func (f *landingCheckFake) run(_ context.Context, c localCheck, head string) (localRunRecord, error) {
	f.calls = append(f.calls, c.Name+":"+head)
	p := filepath.Join(f.t.TempDir(), "output.log")
	if err := os.WriteFile(p, []byte(c.Name+" evidence"), 0600); err != nil {
		f.t.Fatal(err)
	}
	sum, _, err := hashAndTail(p, 1)
	if err != nil {
		f.t.Fatal(err)
	}
	r := localRunRecord{Name: c.Name, SHA: head, LogPath: p, LogSHA256: sum}
	if f.fail == c.Name {
		r.Exit = 37
	}
	if f.badHead {
		r.SHA = "stale"
	}
	return r, nil
}
func landingGateRepo(t *testing.T) (localRunDeps, string, string) {
	t.Helper()
	d, _, _, _ := localRunRepo(t, "unused")
	write := func(path, body string) {
		t.Helper()
		p := filepath.Join(d.Root, path)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(localChecksPath, `{"version":2,"checks":[{"name":"suite","command":["true"]},{"name":"vm","scope":"landing","resource":"runner","required":true,"command":["true"],"paths_file":"paths.json"}]}`)
	write("paths.json", `{"version":1,"paths":["src/**","watched/**"]}`)
	commit := func() string {
		t.Helper()
		if _, e := d.Git("add", "-A"); e != nil {
			t.Fatal(e)
		}
		if _, e := d.Git("commit", "-qm", "fixture"); e != nil {
			t.Fatal(e)
		}
		h, e := d.Git("rev-parse", "HEAD")
		if e != nil {
			t.Fatal(e)
		}
		return strings.TrimSpace(h)
	}
	base := commit()
	write("src/a.go", "code")
	head := commit()
	return d, base, head
}
func TestLandingCheckGateOrderAndOwnerEvidence(t *testing.T) {
	d, base, head := landingGateRepo(t)
	fake := &landingCheckFake{t: t}
	owner := landingCheckOwner{}
	for range 2 {
		if err := owner.run(context.Background(), d.Git, base, head, []string{head}, false, fake.run, fake.run); err != nil {
			t.Fatal(err)
		}
	}
	if got := strings.Join(fake.calls, ","); got != "suite:"+head+",vm:"+head {
		t.Fatalf("order/duplicate %s", got)
	}
	// Bare saved/manual status cannot satisfy a fresh owner's gate.
	fresh := landingCheckOwner{}
	if err := fresh.run(context.Background(), d.Git, base, head, []string{head}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if len(fake.calls) != 3 || fake.calls[2] != "vm:"+head {
		t.Fatalf("fresh owner accepted unbound status: %v", fake.calls)
	}
}
func TestLandingCheckGateFailureNeverRetries(t *testing.T) {
	for _, bad := range []bool{false, true} {
		t.Run(fmt.Sprint(bad), func(t *testing.T) {
			d, base, head := landingGateRepo(t)
			fake := &landingCheckFake{t: t, fail: "vm", badHead: bad}
			owner := landingCheckOwner{}
			for range 2 {
				if err := owner.run(context.Background(), d.Git, base, head, []string{head}, true, fake.run, fake.run); err == nil {
					t.Fatal("failed/stale runner passed")
				}
			}
			if len(fake.calls) != 1 {
				t.Fatalf("runner retried %v", fake.calls)
			}
		})
	}
}

func TestLandingRequirementsNeverBecomeCarChecks(t *testing.T) {
	decl := localCheckDecl{Version: 2, Checks: []localCheck{{Name: "vm", RunScope: "landing", Resource: "runner", Required: true}}}
	if v := decideCIGate(ciInputs{Local: decl}); !v.Reason.passes() {
		t.Fatalf("car demanded VM: %+v", v)
	}
	if names := unrecordedRequired(decl, nil, nil); len(names) != 0 {
		t.Fatalf("ancestor lookup demanded VM: %v", names)
	}
	if v := decideCIGate(ciInputs{Local: decl, Checks: []prCheck{{Name: localStatusContext("vm"), Bucket: "fail"}}}); v.Reason != ciRed {
		t.Fatalf("observed red VM hidden: %+v", v)
	}
}
func TestLandingManualBurnInTarget(t *testing.T) {
	checks := []localCheck{{Name: "vm", RunScope: "landing", Required: false}}
	if err := validateManualLandingTarget(checks, "branch", "main"); err == nil {
		t.Fatal("disabled check ran off main")
	}
	if err := validateManualLandingTarget(checks, "main", "main"); err != nil {
		t.Fatal(err)
	}
	checks[0].Required = true
	if err := validateManualLandingTarget(checks, "branch", "main"); err != nil {
		t.Fatal(err)
	}
}

func TestLandingCheckGateReintegrationEvidence(t *testing.T) {
	d, base, car := landingGateRepo(t)
	owner := landingCheckOwner{}
	fake := &landingCheckFake{t: t}
	if err := owner.run(context.Background(), d.Git, base, car, []string{car}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	change := func(path string) string {
		t.Helper()
		p := filepath.Join(d.Root, path)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("changed"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, e := d.Git("add", "-A"); e != nil {
			t.Fatal(e)
		}
		if _, e := d.Git("commit", "-qm", "move"); e != nil {
			t.Fatal(e)
		}
		h, e := d.Git("rev-parse", "HEAD")
		if e != nil {
			t.Fatal(e)
		}
		return strings.TrimSpace(h)
	}
	docs := change("notes.md")
	if err := owner.run(context.Background(), d.Git, base, docs, []string{car}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("unwatched docs reran VM: %v", fake.calls)
	}
	watched := change("watched/readme.md")
	if err := owner.run(context.Background(), d.Git, base, watched, []string{car}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if len(fake.calls) != 2 || fake.calls[1] != "vm:"+watched {
		t.Fatalf("watched docs waived: %v", fake.calls)
	}
	code := change("src/new.go")
	if err := owner.run(context.Background(), d.Git, base, code, []string{car}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if len(fake.calls) != 3 || fake.calls[2] != "vm:"+code {
		t.Fatalf("code reused old proof: %v", fake.calls)
	}
	// A successful record's original exact SHA is retained, never copied.
	if p := owner.proofs["vm"]; p.Run.SHA != code {
		t.Fatalf("proof %+v", p)
	}
	if err := owner.run(context.Background(), d.Git, base, code, []string{car, docs}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if len(fake.calls) != 4 {
		t.Fatalf("changed cars reused proof %v", fake.calls)
	}
}
func TestLandingCheckGateNeedsRealSuite(t *testing.T) {
	d, base, head := landingGateRepo(t)
	fake := &landingCheckFake{t: t, fail: "suite"}
	owner := landingCheckOwner{}
	if err := owner.run(context.Background(), d.Git, base, head, []string{head}, false, fake.run, fake.run); err == nil {
		t.Fatal("failed suite admitted runner")
	}
	if len(fake.calls) != 1 || !strings.HasPrefix(fake.calls[0], "suite:") {
		t.Fatalf("runner after failed suite: %v", fake.calls)
	}
	// Omit both an explicit suite and a Makefile test target on pinned base.
	raw := `{"version":2,"checks":[{"name":"vm","scope":"landing","resource":"runner","required":true,"command":["true"]}]}`
	if err := os.WriteFile(filepath.Join(d.Root, localChecksPath), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if _, e := d.Git("add", localChecksPath); e != nil {
		t.Fatal(e)
	}
	if _, e := d.Git("commit", "-qm", "no suite"); e != nil {
		t.Fatal(e)
	}
	pinned, e := d.Git("rev-parse", "HEAD")
	if e != nil {
		t.Fatal(e)
	}
	pinned = strings.TrimSpace(pinned)
	owner = landingCheckOwner{}
	fake = &landingCheckFake{t: t}
	if err := owner.run(context.Background(), d.Git, pinned, head, []string{head}, true, fake.run, fake.run); err == nil || !strings.Contains(err.Error(), "real declared/default suite") {
		t.Fatalf("synthetic suite accepted: %v", err)
	}
	if len(fake.calls) != 0 {
		t.Fatal(fake.calls)
	}
}

func TestLandingCheckGateMainMoveCannotWaiveWatchedDocs(t *testing.T) {
	d, base, head := landingGateRepo(t)
	owner := landingCheckOwner{}
	fake := &landingCheckFake{t: t}
	if err := owner.run(context.Background(), d.Git, base, head, []string{head}, true, fake.run, fake.run); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("checkout", "-qb", "moving-main", base); err != nil {
		t.Fatal(err)
	}
	change := func(path string) string {
		t.Helper()
		p := filepath.Join(d.Root, path)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("doc"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, e := d.Git("add", "-A"); e != nil {
			t.Fatal(e)
		}
		if _, e := d.Git("commit", "-qm", "main moved"); e != nil {
			t.Fatal(e)
		}
		h, e := d.Git("rev-parse", "HEAD")
		if e != nil {
			t.Fatal(e)
		}
		return strings.TrimSpace(h)
	}
	docs := change("notes.md")
	if !owner.allowsMainMove(d.Git, base, docs, []string{head}) {
		t.Fatal("unwatched docs cannot reuse proof")
	}
	watched := change("watched/notes.md")
	if owner.allowsMainMove(d.Git, base, watched, []string{head}) {
		t.Fatal("issue/docs-only readiness waived watched path")
	}
}

func TestLandingCheckGateCanceledOwnerLaunchesNothing(t *testing.T) {
	d, base, head := landingGateRepo(t)
	fake := &landingCheckFake{t: t}
	owner := landingCheckOwner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := owner.run(ctx, d.Git, base, head, []string{head}, false, fake.run, fake.run); err == nil {
		t.Fatal("canceled owner passed")
	}
	if len(fake.calls) != 0 {
		t.Fatalf("canceled owner launched %v", fake.calls)
	}
}

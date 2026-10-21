package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkingdocCommandSurface(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	gitInit(t, repo.root)
	os.WriteFile(filepath.Join(repo.root, ".hub"), []byte("product\n"), 0600)
	t.Chdir(repo.root)
	d.documents = repo
	root := buildRoot(d)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "settle", source})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "settled") || ss.stamps != 1 {
		t.Fatalf("output %s stamps %d", out.String(), ss.stamps)
	}
	snapshot, _ := os.ReadFile(source)
	out.Reset()
	root = buildRoot(d)
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(source)
	if !bytes.Equal(snapshot, after) || ss.stamps != 1 || !strings.Contains(out.String(), "settled") {
		t.Fatalf("status wrote or lost state: %s", out.String())
	}
}

func TestWorkingdocStampCannotExceedSourceLimit(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	header := "---\ntype: workingdoc\n---\n"
	os.WriteFile(source, []byte(header+strings.Repeat("x", maxWorkingdocBytes-len(header))), 0600)
	if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err == nil {
		t.Fatal("accepted source with no room for its pointer")
	}
	if repo.publishes != 0 || ss.stamps != 0 {
		t.Fatal("size refusal happened after publication")
	}
}

func newWorkingdocFixture(t *testing.T) (string, *workingdocSyncFake, *workingdocRepositoryFake, deps) {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{".git", "session.side"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0700); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(root, "session.side/2026-10-21-01-direction.md")
	if err := os.WriteFile(source, []byte("---\ntype: side-doc\nowner: ops:0\ntopic: Direction\n---\n# Direction\nAgreed.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ss := &workingdocSyncFake{root: root, registered: map[string]*syncRegistration{source: {}}, locks: map[string]bool{}}
	repo := &workingdocRepositoryFake{root: root, blobs: map[settlementPointer][]byte{}, current: map[string][]byte{}}
	return source, ss, repo, deps{ss: ss, now: func() time.Time { return time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC) }}
}

func TestWorkingdocSettleAndDrift(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	result, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Target != "workshop/workingdoc/2026-10-21-direction.md" || !result.Published || !result.Stamped {
		t.Fatalf("result %+v", result)
	}
	b, _ := os.ReadFile(source)
	m, err := parseWorkingdoc(string(b))
	if err != nil || m.Pointer == nil {
		t.Fatalf("source not stamped: %v", err)
	}
	if repo.publishes != 1 || ss.stamps != 1 {
		t.Fatalf("effects %d/%d", repo.publishes, ss.stamps)
	}
	if _, err = settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err != nil {
		t.Fatal(err)
	}
	if repo.publishes != 1 {
		t.Fatal("retry duplicated the publication")
	}
	b, _ = os.ReadFile(source)
	b = []byte(strings.Replace(string(b), "Agreed.", "Changed.", 1))
	os.WriteFile(source, b, 0600)
	if _, err = settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err != nil {
		t.Fatal(err)
	}
	if repo.publishes != 2 {
		t.Fatal("re-settle did not update copy")
	}
}

func TestWorkingdocSourceChangeCannotBeStamped(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	repo.afterPublish = func() {
		b, _ := os.ReadFile(source)
		os.WriteFile(source, append(b, []byte("New open work.\n")...), 0600)
	}
	result, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err == nil || !result.Published || result.Stamped || ss.stamps != 0 {
		t.Fatalf("partial outcome %+v %v", result, err)
	}
	b, _ := os.ReadFile(source)
	if strings.Contains(string(b), "settled-into:") || !strings.Contains(string(b), "New open work.") {
		t.Fatal("new work lost or falsely stamped")
	}
}

func TestWorkingdocRefusalsHaveNoEffects(t *testing.T) {
	for _, kind := range []string{"locked", "comment", "proposal", "unregistered", "collision"} {
		t.Run(kind, func(t *testing.T) {
			source, ss, repo, d := newWorkingdocFixture(t)
			switch kind {
			case "locked":
				ss.locks[source] = true
			case "comment", "proposal":
				b, _ := os.ReadFile(source)
				marker := "🤖[open question]"
				if kind == "proposal" {
					marker = "🤖{new text}"
				}
				os.WriteFile(source, append(b, []byte(marker+"\n")...), 0600)
			case "unregistered":
				delete(ss.registered, source)
			case "collision":
				p := filepath.Join(repo.root, "workshop/workingdoc/2026-10-21-direction.md")
				os.MkdirAll(filepath.Dir(p), 0700)
				os.WriteFile(p, []byte("Independent content"), 0600)
			}
			if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err == nil {
				t.Fatal("refusal accepted")
			}
			if repo.publishes != 0 || ss.stamps != 0 {
				t.Fatal("refusal performed effects")
			}
		})
	}
}

// These exercise the coordinator with persistent files and remote blobs: an
// interrupted response must never be mistaken for either an absent effect or
// permission to publish the same snapshot twice.
func TestWorkingdocStampFailureRecovery(t *testing.T) {
	for _, failure := range []string{"before-write", "acknowledgment-lost", "late-lock"} {
		t.Run(failure, func(t *testing.T) {
			source, ss, repo, d := newWorkingdocFixture(t)
			original := readWorkingdocTestFile(t, source)
			unreachable := errors.New("Couch response unavailable")
			switch failure {
			case "before-write":
				ss.beforeApply = func(string) error { return unreachable }
			case "acknowledgment-lost":
				ss.afterApply = func(string) error { return unreachable }
			case "late-lock":
				repo.afterPublish = func() { ss.locks[source] = true }
			}
			first, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
			if !first.Published || repo.publishes != 1 {
				t.Fatalf("publication was not confirmed: result=%+v err=%v publishes=%d", first, err, repo.publishes)
			}
			if failure == "acknowledgment-lost" {
				if err != nil || !first.Stamped || ss.stamps != 1 {
					t.Fatalf("readback did not confirm lost acknowledgment: result=%+v err=%v stamps=%d", first, err, ss.stamps)
				}
				assertWorkingdocTestPointer(t, source, original, first)
			} else {
				if err == nil || first.Stamped || ss.stamps != 0 || !strings.Contains(err.Error(), "source stamp failed or unconfirmed") {
					t.Fatalf("failed stamp reported success: result=%+v err=%v stamps=%d", first, err, ss.stamps)
				}
				if !bytes.Equal(readWorkingdocTestFile(t, source), original) {
					t.Fatal("failed stamp changed the source")
				}
			}
			ss.beforeApply, ss.afterApply, repo.afterPublish = nil, nil, nil
			delete(ss.locks, source)
			retry, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
			if err != nil || !retry.Published || !retry.Stamped || repo.publishes != 1 || ss.stamps != 1 {
				t.Fatalf("retry duplicated publication or failed to stamp: result=%+v err=%v publishes=%d stamps=%d", retry, err, repo.publishes, ss.stamps)
			}
			assertWorkingdocTestPointer(t, source, original, retry)
		})
	}
}

func TestWorkingdocUncertainStampReadbackRecovery(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	original := readWorkingdocTestFile(t, source)
	parked := source + ".temporarily-unavailable"
	ss.afterApply = func(file string) error {
		if err := os.Rename(file, parked); err != nil {
			t.Fatal(err)
		}
		return errors.New("Apply response lost while source unavailable")
	}
	first, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err == nil || !first.Published || first.Stamped || ss.stamps != 1 || repo.publishes != 1 {
		t.Fatalf("unknown readback falsely confirmed: result=%+v err=%v stamps=%d publishes=%d", first, err, ss.stamps, repo.publishes)
	}
	assertWorkingdocTestPointer(t, parked, original, first)
	if err := os.Rename(parked, source); err != nil {
		t.Fatal(err)
	}
	ss.afterApply = nil
	retry, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err != nil || !retry.Stamped || !retry.Published || ss.stamps != 1 || repo.publishes != 1 {
		t.Fatalf("recovered acknowledgment repeated an effect: result=%+v err=%v stamps=%d publishes=%d", retry, err, ss.stamps, repo.publishes)
	}
	assertWorkingdocTestPointer(t, source, original, retry)
}

func TestWorkingdocUncertainStampPreservesNewSourceWork(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	original := readWorkingdocTestFile(t, source)
	ss.afterApply = func(file string) error {
		stamped := readWorkingdocTestFile(t, file)
		if err := os.WriteFile(file, append(stamped, []byte("New operator work.\n")...), 0600); err != nil {
			t.Fatal(err)
		}
		return errors.New("Apply acknowledgment lost")
	}
	first, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err == nil || !first.Published || first.Stamped || ss.stamps != 1 || repo.publishes != 1 {
		t.Fatalf("divergent readback falsely confirmed: result=%+v err=%v stamps=%d publishes=%d", first, err, ss.stamps, repo.publishes)
	}
	newSource := readWorkingdocTestFile(t, source)
	if !strings.HasSuffix(string(newSource), "New operator work.\n") || strings.Contains(string(repo.current[first.Target]), "New operator work.") {
		t.Fatal("concurrent source work was lost or incorrectly published")
	}
	ss.afterApply = nil
	retry, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, "")
	if err != nil || !retry.Stamped || !retry.Published || ss.stamps != 2 || repo.publishes != 2 {
		t.Fatalf("new source snapshot was not settled exactly once: result=%+v err=%v stamps=%d publishes=%d", retry, err, ss.stamps, repo.publishes)
	}
	assertWorkingdocTestPointer(t, source, append(original, []byte("New operator work.\n")...), retry)
	if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, ""); err != nil || repo.publishes != 2 || ss.stamps != 2 {
		t.Fatalf("completed retry repeated an effect: err=%v publishes=%d stamps=%d", err, repo.publishes, ss.stamps)
	}
}

func readWorkingdocTestFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func assertWorkingdocTestPointer(t *testing.T, source string, original []byte, result settlementResult) {
	t.Helper()
	got := readWorkingdocTestFile(t, source)
	metadata, err := parseWorkingdoc(string(got))
	if err != nil || metadata.Pointer == nil || metadata.Pointer.Path != result.Target || metadata.Pointer.Commit != result.Commit {
		t.Fatalf("pointer is not the confirmed receipt: metadata=%+v result=%+v err=%v", metadata, result, err)
	}
	want, err := stampWorkingdoc(string(original), *metadata.Pointer)
	if err != nil || string(got) != want {
		t.Fatalf("source bytes beyond the managed pointer changed: err=%v", err)
	}
}

func TestWorkingdocUnsafePointerRefusesBeforePublication(t *testing.T) {
	source, ss, repo, d := newWorkingdocFixture(t)
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	target := "workshop/workingdoc/a #b.md"
	if _, err := settleWorkingdoc(context.Background(), d, repo, repo.root, source, target); err == nil {
		t.Fatal("accepted unrepresentable managed pointer")
	}
	after, _ := os.ReadFile(source)
	if repo.publishes != 0 || ss.stamps != 0 || !bytes.Equal(before, after) {
		t.Fatal("pointer refusal changed source or published")
	}
	if _, err := os.Stat(filepath.Join(repo.root, target)); !os.IsNotExist(err) {
		t.Fatalf("target created before pointer validation: %v", err)
	}
}

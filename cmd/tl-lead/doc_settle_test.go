package main

import (
	"bytes"
	"context"
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

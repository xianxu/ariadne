package reviewusage

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestStoreAdmissionReclaimsDeadHolder(t *testing.T) {
	store, intent := storeFixture(t)
	child := exec.Command("sh", "-c", "exit 0")
	if err := child.Run(); err != nil {
		t.Fatal(err)
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	common := strings.TrimSpace(testfix.Capture(t, store.root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(store.root, common)
	}
	lockDir := filepath.Join(common, "sdlc-review-usage-admission", "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		t.Fatal(err)
	}
	meta := repolock.Metadata{PID: child.Process.Pid, Hostname: host, StartedAt: time.Now(), Command: "dead review admission"}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), repolock.Encode(meta), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := store.SaveRequested(ctx, intent); err != nil {
		t.Fatalf("dead admission holder was not reclaimed: %v", err)
	}
	if _, err := os.Stat(lockDir); !os.IsNotExist(err) {
		t.Fatalf("admission lock retained: %v", err)
	}
}

func storeFixture(t *testing.T) (*Store, Intent) {
	t.Helper()
	root := testfix.Repo(t, testfix.InitialCommit())
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	intent := testIntent()
	intent.Root = root
	intent.ReviewedHEAD = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	return store, intent
}
func TestStoreCASAndAdmissionCap(t *testing.T) {
	ctx := context.Background()
	store, intent := storeFixture(t)
	store.limit = 2
	oid, err := store.SaveRequested(ctx, intent)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := store.Load(ctx, intent.ID)
	if err != nil || entry.OID != oid || entry.Record.Phase != Requested {
		t.Fatal("requested progress missing")
	}
	started, err := Step(entry.Record, Event{Kind: Start, Intent: intent, PID: 42, At: "2026-10-14T10:00:01Z"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(ctx, started, "wrong-generation"); err == nil {
		t.Fatal("invalid CAS admitted")
	}
	if err := store.Record(ctx, started); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(ctx, started, oid); !errors.Is(err, gitx.ErrRecoveryChanged) {
		t.Fatalf("stale CAS accepted: %v", err)
	}
	second := intent
	second.ID = "review-two"
	second.Attempt = 2
	if _, err := store.SaveRequested(ctx, second); err != nil {
		t.Fatal(err)
	}
	third := intent
	third.ID = "review-three"
	if _, err := store.SaveRequested(ctx, third); err == nil {
		t.Fatal("full store admitted launch")
	}
	entries, err := store.List(ctx)
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries %v: %v", entries, err)
	}
	if err := store.Record(ctx, started); err != nil {
		t.Fatal("exact update charged capacity")
	}
	rewritten := started
	rewritten.Intent.Repository = "foreign"
	if err := store.Record(ctx, rewritten); err == nil {
		t.Fatal("intent rewritten")
	}
}
func TestStoreConcurrentAdmissionAndCanceledTerminal(t *testing.T) {
	ctx := context.Background()
	store, intent := storeFixture(t)
	store.limit = 1
	result := make(chan error, 2)
	for _, id := range []string{"one", "two"} {
		go func(id string) { v := intent; v.ID = id; _, err := store.SaveRequested(ctx, v); result <- err }(id)
	}
	successes := 0
	for i := 0; i < 2; i++ {
		if err := <-result; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("cap race admitted %d", successes)
	}
	records, err := store.List(ctx)
	if err != nil || len(records) != 1 {
		t.Fatal(err)
	}
	r := records[0].Record
	s, _ := Step(r, Event{Kind: Start, Intent: r.Intent, PID: 42, At: "2026-10-14T10:00:01Z"})
	finish, _ := Step(s, Event{Kind: Finish, Intent: r.Intent, PID: 42, StartedAt: s.StartedAt, At: "2026-10-14T10:00:02Z", Outcome: "canceled"})
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := store.Record(canceled, finish); err == nil {
		t.Fatal("canceled persistence completed")
	}
	if err := store.Record(ctx, finish); err != nil {
		t.Fatal(err)
	}
	terminal, err := store.Load(ctx, r.Intent.ID)
	if err != nil || terminal.Record != finish {
		t.Fatal("fresh accounting context lost known start/finish")
	}
}
func TestStoreStrictDocument(t *testing.T) {
	r, _ := NewRequested(testIntent())
	raw, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := Parse(raw); err != nil || got != r {
		t.Fatal("record did not roundtrip")
	}
	for _, bad := range []string{strings.Replace(string(raw), `"version":1`, `"version":1,"version":1`, 1), strings.Replace(string(raw), `"phase":`, `"Phase":`, 1), strings.Replace(string(raw), `"phase":"requested"`, `"phase":"unknown"`, 1), string(raw) + "{}"} {
		if _, err := Parse([]byte(bad)); err == nil {
			t.Fatal("malformed document admitted")
		}
	}
}

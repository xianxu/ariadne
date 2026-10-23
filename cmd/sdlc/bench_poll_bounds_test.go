package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchObserverMetadataReadBounds(t *testing.T) {
	// Approved M1 polling budget, independent of future generic store changes.
	if benchEventLimit != 2000 || benchRecordLimit != 1<<20 {
		t.Fatal("observer metadata budget changed; revise the approved polling cost")
	}
	s, err := createBenchStore(filepath.Join(t.TempDir(), "run"), benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	r.Events = make([]json.RawMessage, 2001)
	for i := range r.Events {
		r.Events[i] = json.RawMessage(`null`)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) >= 1<<20 {
		t.Fatal("event-count control unintentionally exceeds byte budget")
	}
	if err := os.WriteFile(s.recordPath(), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.read(); err == nil || !strings.Contains(err.Error(), "event budget") {
		t.Fatal("reader replayed more than 2000 transitions", err)
	}
	c := benchContextFixture(t)
	p := benchAuditPath(&c)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(strings.Repeat(" ", (1<<20)+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchAudit(context.Background(), &c); err == nil {
		t.Fatal("proposal reader accepted more than 1 MiB")
	}
}

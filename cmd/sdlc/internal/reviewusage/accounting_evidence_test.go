package reviewusage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestAccountingEvidencePreservesInvalidIdentityWithoutAuthority(t *testing.T) {
	store, intent := storeFixture(t)
	ctx := context.Background()
	if _, err := store.SaveRequested(ctx, intent); err != nil {
		t.Fatal(err)
	}
	invalid := Record{Version: 99, Intent: intent, Phase: Requested}
	invalid.Intent.ID = "invalid"
	encoded, err := json.Marshal(invalid)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := store.raw(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Save(invalid.Intent.ID, "", gitx.RecoveryData{Document: encoded}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.List(ctx); err == nil {
		t.Fatal("strict List admitted invalid receipt")
	}
	entries, err := store.AccountingEvidence(ctx)
	if err == nil || len(entries) != 2 {
		t.Fatalf("partial evidence not retained: %d %v", len(entries), err)
	}
	validCount := 0
	invalidCount := 0
	for _, entry := range entries {
		if Validate(entry.Record) == nil {
			validCount++
		} else if entry.Record.Intent.ID == invalid.Intent.ID && entry.Record.Intent.Root == intent.Root {
			invalidCount++
		}
	}
	if validCount != 1 || invalidCount != 1 {
		t.Fatalf("usable identity lost or granted authority: %+v", entries)
	}
}

func TestParseAccountingEvidenceStrictFailuresRemainQuarantined(t *testing.T) {
	record, err := NewRequested(testIntent())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.TrimSuffix(string(encoded), "}") + `,"unknown":true}`,
		strings.Replace(string(encoded), `"version":1`, `"version":99,"version":1`, 1),
	} {
		got, err := ParseAccountingEvidence([]byte(raw))
		if err == nil || Validate(got) == nil || got.Intent != record.Intent {
			t.Fatalf("invalid shape acquired authority or lost identity: %+v %v", got, err)
		}
	}
}

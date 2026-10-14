package tracker

import (
	"strings"
	"testing"
)

func TestTransferAssetPinsRejectUnsafeAndBindGeneration(t *testing.T) {
	spec := operationSpec()
	// Decode the new optional wire field through the public receipt parser, so
	// the pre-feature run fails for unsupported asset pins rather than compile.
	tr, err := NewTransfer(spec)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := MarshalReceipt(tr.Receipt())
	if err != nil {
		t.Fatal(err)
	}
	asset := strings.Repeat("a", 40) + " workshop/issues/assets/000001/" + strings.Repeat("b", 64) + ".png"
	raw = []byte(strings.Replace(string(raw), `"source_blob":`, `"source_assets":"`+asset+`","source_blob":`, 1))
	parsed, err := ParseReceipt(raw, spec.Repository)
	if err != nil {
		t.Fatalf("valid asset pin rejected: %v", err)
	}
	wire, err := MarshalReceipt(parsed)
	if err != nil || !strings.Contains(string(wire), asset) {
		t.Fatalf("asset pin not preserved: %s %v", wire, err)
	}
	for _, bad := range []string{"../escape.png", "workshop/issues/assets/000002/" + strings.Repeat("b", 64) + ".png"} {
		altered := []byte(strings.Replace(string(raw), "workshop/issues/assets/000001/"+strings.Repeat("b", 64)+".png", bad, 1))
		if _, err := ParseReceipt(altered, spec.Repository); err == nil {
			t.Fatalf("unsafe asset pin accepted: %s", bad)
		}
	}
}

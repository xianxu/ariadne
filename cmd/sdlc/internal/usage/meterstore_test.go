package usage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMeterStorePreservesOnRefusalAndCancel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "usage-meter.json")
	ctx := context.Background()
	o := Observation{Agent: "codex", Unit: "credits", Remaining: 100, At: time.Now().UTC()}
	if _, e := RecordMeter(ctx, path, o); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal(info.Mode())
	}
	o.Remaining = 99
	if _, e = RecordMeter(ctx, path, o); e == nil {
		t.Fatal("conflict accepted")
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	o.At = o.At.Add(time.Hour)
	if _, e = RecordMeter(c, path, o); e == nil {
		t.Fatal("cancel accepted")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("old meter overwritten")
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".usage-meter-*"))
	if len(files) != 0 {
		t.Fatal(files)
	}
}
func TestMeterStoreBoundAndMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meter.json")
	m, e := LoadMeter(path)
	if e != nil || len(m.Observations) != 0 {
		t.Fatal(m, e)
	}
	if e = os.WriteFile(path, make([]byte, 65537), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = LoadMeter(path); e == nil {
		t.Fatal("oversize accepted")
	}
}

func TestMeterStoreRefusesUnknownSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meter.json")
	for _, raw := range []string{"null", "{}", "{\"version\":2}"} {
		os.WriteFile(path, []byte(raw), 0600)
		if _, e := LoadMeter(path); e == nil {
			t.Fatal("unsupported schema accepted", raw)
		}
	}
}

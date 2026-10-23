package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

func ledgerFixtureRow(t *testing.T, id string) bench.LedgerRow {
	t.Helper()
	p := bench.AgentProfile{Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}
	r := bench.LedgerRow{Version: 1, CorpusVersion: "v1", Repo: "fixture", TaskID: "000001", Type: "small-fix", Repeat: 1, RunID: id, ParentOID: strings.Repeat("a", 40), LandingOID: strings.Repeat("b", 40), OriginalIssueOID: strings.Repeat("c", 40), Outcome: "stopped", Worker: p, TL: p, Reviewer: p, OriginalHours: 0.1, Measurements: bench.Measurements{GrossNanos: 1, Autonomous: true, Worker: bench.RoleUsage{Coverage: "unknown"}, TL: bench.RoleUsage{Coverage: "unknown"}, Reviewer: bench.RoleUsage{Coverage: "unknown"}, Leaks: bench.LeakReport{Status: "unknown", Coverage: bench.LeakCoverage{Basis: "missing transcript"}}}, Acceptance: bench.Acceptance{Status: "unavailable", Reason: "stopped"}, RawDigest: strings.Repeat("d", 64), Synthetic: true}
	if err := bench.ValidateLedgerRow(r); err != nil {
		t.Fatal(err)
	}
	return r
}
func ledgerFixturePath(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(p, "ledger.jsonl")
}
func ledgerWrite(t *testing.T, p string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func ledgerBytes(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBenchLedgerAppendIdempotenceAndOriginalBytes(t *testing.T) {
	p := ledgerFixturePath(t)
	r := ledgerFixtureRow(t, "r1")
	r.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"large":1e400}`)}
	r.Measurements.TL.Unknown = map[string]json.RawMessage{"future_nested": json.RawMessage(`{"n":1}`)}
	raw, err := bench.EncodeLedgerRow(r)
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte(" "), raw...)
	original = append(original, '\n')
	ledgerWrite(t, p, original)
	if err := recordBenchRow(context.Background(), p, r); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, ledgerBytes(t, p)) {
		t.Fatal("idempotence rewrote original bytes")
	}
	second := ledgerFixtureRow(t, "r2")
	if err := recordBenchRow(context.Background(), p, second); err != nil {
		t.Fatal(err)
	}
	after := ledgerBytes(t, p)
	if !bytes.HasPrefix(after, original) {
		t.Fatal("append rewrote older row")
	}
	got, err := readBenchLedger(context.Background(), p)
	if err != nil || len(got) != 2 || got[0].Unknown["future"] == nil || got[0].Measurements.TL.Unknown["future_nested"] == nil {
		t.Fatalf("round trip: %+v %v", got, err)
	}
	r.RawDigest = strings.Repeat("e", 64)
	if err := recordBenchRow(context.Background(), p, r); err == nil {
		t.Fatal("conflicting digest accepted")
	}
	r.RawDigest = strings.Repeat("d", 64)
	r.OriginalHours = 0.2
	if err := recordBenchRow(context.Background(), p, r); err == nil {
		t.Fatal("same digest concealed changed content")
	}
	if !bytes.Equal(after, ledgerBytes(t, p)) {
		t.Fatal("refusal mutated ledger")
	}
}

func TestBenchLedgerCorruptionAndBoundsRefuseWithoutRepair(t *testing.T) {
	row := ledgerFixtureRow(t, "r1")
	raw, _ := bench.EncodeLedgerRow(row)
	duplicate := append(append(append([]byte{}, raw...), '\n'), raw...)
	duplicate = append(duplicate, '\n')
	cases := map[string][]byte{"partial": raw, "blank": []byte("\n"), "malformed": []byte("{broken}\n"), "duplicate": duplicate, "oversized line": append(bytes.Repeat([]byte(" "), bench.LedgerRowLimit+1), '\n')}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			p := ledgerFixturePath(t)
			ledgerWrite(t, p, content)
			for _, read := range []func() error{func() error { return preflightBenchLedger(context.Background(), p) }, func() error { _, err := readBenchLedger(context.Background(), p); return err }, func() error { return recordBenchRow(context.Background(), p, ledgerFixtureRow(t, "r2")) }} {
				if err := read(); err == nil {
					t.Fatal("corruption accepted")
				}
			}
			if !bytes.Equal(content, ledgerBytes(t, p)) {
				t.Fatal("corruption silently repaired")
			}
		})
	}
}

func TestBenchLedgerMissingPreflightAndPathRefusals(t *testing.T) {
	p := ledgerFixturePath(t)
	if err := preflightBenchLedger(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preflight created file: %v", err)
	}
	if rows, err := readBenchLedger(context.Background(), p); err != nil || len(rows) != 0 {
		t.Fatal(rows, err)
	}
	target := filepath.Join(filepath.Dir(p), "target")
	ledgerWrite(t, target, nil)
	if err := os.Symlink(target, p); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{p, "relative.jsonl", filepath.Dir(p), filepath.Dir(p) + "/../escape.jsonl", filepath.Join(filepath.Dir(p), "missing", "ledger.jsonl")} {
		if err := recordBenchRow(context.Background(), path, ledgerFixtureRow(t, "r1")); err == nil {
			t.Errorf("unsafe path accepted: %s", path)
		}
	}
	if len(ledgerBytes(t, target)) != 0 {
		t.Fatal("symlink target changed")
	}
	alias := filepath.Join(filepath.Dir(p), "alias")
	if err := os.Symlink(filepath.Dir(p), alias); err != nil {
		t.Fatal(err)
	}
	if err := preflightBenchLedger(context.Background(), filepath.Join(alias, "other.jsonl")); err == nil {
		t.Fatal("symlink ancestor accepted")
	}
}

func TestBenchLedgerContextAndConcurrentAppends(t *testing.T) {
	p := ledgerFixturePath(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := recordBenchRow(ctx, p, ledgerFixtureRow(t, "r1")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled append created ledger")
	}
	ledgerWrite(t, p, nil)
	held, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	if err := flock.Exclusive(held, false); err != nil {
		t.Fatal(err)
	}
	wait, cancelWait := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- preflightBenchLedger(wait, p) }()
	cancelWait()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	held.Close()
	row1, row2 := ledgerFixtureRow(t, "r1"), ledgerFixtureRow(t, "r2")
	errs := make(chan error, 4)
	var wg sync.WaitGroup
	for _, row := range []bench.LedgerRow{row1, row2, row1, row2} {
		wg.Add(1)
		go func(row bench.LedgerRow) { defer wg.Done(); errs <- recordBenchRow(context.Background(), p, row) }(row)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, err := readBenchLedger(context.Background(), p)
	if err != nil || len(rows) != 2 {
		t.Fatal(rows, err)
	}
}

func TestBenchLedgerScanVisitsEachRowOnce(t *testing.T) {
	for _, n := range []int{10, 20} {
		var content bytes.Buffer
		for i := 0; i < n; i++ {
			row := ledgerFixtureRow(t, strings.Repeat("r", i+1))
			raw, _ := bench.EncodeLedgerRow(row)
			content.Write(raw)
			content.WriteByte('\n')
		}
		visits := 0
		decode := func(raw []byte) (bench.LedgerRow, error) { visits++; return bench.DecodeLedgerRow(raw) }
		scan, err := scanBenchLedger(context.Background(), bytes.NewReader(content.Bytes()), decode)
		if err != nil || len(scan.rows) != n || visits != n {
			t.Fatalf("N=%d visits=%d rows=%d err=%v", n, visits, len(scan.rows), err)
		}
	}
}

func TestBenchLedgerCapacityAndReplacedInode(t *testing.T) {
	visits := 0
	decode := func([]byte) (bench.LedgerRow, error) {
		visits++
		return bench.LedgerRow{RunID: fmt.Sprint(visits)}, nil
	}
	scan, err := scanBenchLedger(context.Background(), strings.NewReader(strings.Repeat("{}\n", benchLedgerRowLimit+1)), decode)
	if err == nil || len(scan.rows) != benchLedgerRowLimit || visits != benchLedgerRowLimit {
		t.Fatalf("row budget not enforced before decoder: rows=%d visits=%d err=%v", len(scan.rows), visits, err)
	}
	p := ledgerFixturePath(t)
	f, _, err := openBenchLedger(context.Background(), p, true)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	parent, err := benchLedgerParent(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(p, p+".retained"); err != nil {
		t.Fatal(err)
	}
	ledgerWrite(t, p, []byte("replacement\n"))
	if err := checkBenchLedgerFile(p, f, parent); err == nil {
		t.Fatal("replaced inode accepted")
	}
	if !bytes.Equal(ledgerBytes(t, p), []byte("replacement\n")) || len(ledgerBytes(t, p+".retained")) != 0 {
		t.Fatal("identity check changed either inode")
	}
	large := ledgerFixturePath(t)
	oversized, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := oversized.Truncate(benchLedgerByteLimit + 1); err != nil {
		t.Fatal(err)
	}
	oversized.Close()
	if err := preflightBenchLedger(context.Background(), large); err == nil {
		t.Fatal("oversized ledger admitted")
	}
}

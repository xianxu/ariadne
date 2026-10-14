package estimate

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func correctionFixture(t *testing.T) ([]LedgerRow, LedgerRow) {
	t.Helper()
	base := LedgerRow{Issue: "repo#1", Estimate: 2, Actual: 1, WindowTrusted: true}
	rows, err := ParseRowsChecked(FormatRow(base))
	if err != nil || len(rows) != 1 {
		t.Fatalf("parse: %v %+v", err, rows)
	}
	replacement := base
	replacement.Actual = 3
	replacement.Correction = &CorrectionRecord{Version: 1, Operation: "op-1", Generation: "generation-1", Disposition: "replace", Targets: []string{rows[0].RawID}, InputDigest: strings.Repeat("a", 64)}
	return rows, replacement
}

func TestCorrectionRoundTripAndRawIdentity(t *testing.T) {
	rows, repair := correctionFixture(t)
	raw := FormatRow(repair)
	got, err := ParseRowsChecked(raw)
	if err != nil || len(got) != 1 || got[0].Correction == nil || got[0].Correction.Operation != "op-1" {
		t.Fatalf("round trip: %v %+v", err, got)
	}
	rawBase := FormatRow(rows[0]) + " "
	got, err = ParseRowsChecked(rawBase + "\n")
	if err != nil || got[0].RawID != fmt.Sprintf("%x", sha256.Sum256([]byte(rawBase))) {
		t.Fatalf("raw identity: %v %+v", err, got)
	}
	if len(strings.Split(FormatRow(rows[0]), "\t")) != 30 {
		t.Fatal("ordinary row encoding changed")
	}
}

func TestResolveCorrectionBeforeTrustFilter(t *testing.T) {
	rows, repair := correctionFixture(t)
	repair.Correction.Disposition = "exclude"
	repair.WindowTrusted = false
	parsed, err := ParseRowsChecked(FormatRow(repair))
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, parsed...)
	resolved, err := ResolveCorrections(rows)
	if err != nil {
		t.Fatal(err)
	}
	var eligible []LedgerRow
	for _, r := range resolved {
		if r.WindowTrusted {
			eligible = append(eligible, r)
		}
	}
	if got := NewestPerIssue(eligible); len(got) != 0 {
		t.Fatalf("excluded trusted target resurrected: %+v", got)
	}
}

func TestResolveCorrectionPreservesOrdinaryRecloseAndIdempotence(t *testing.T) {
	rows, repair := correctionFixture(t)
	next := rows[0]
	next.Actual = 4
	next.RawID = strings.Repeat("b", 64)
	rows = append(rows, next)
	parsed, err := ParseRowsChecked(FormatRow(repair))
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, parsed[0], parsed[0])
	got, err := ResolveCorrections(rows)
	if err != nil || len(got) != 2 || got[0].Actual != 4 || got[1].Actual != 3 {
		t.Fatalf("resolve: %v %+v", err, got)
	}
}

func TestParseCorrectionRejectsMalformed(t *testing.T) {
	rows, repair := correctionFixture(t)
	base := FormatRow(rows[0]) + "\t"
	valid := strings.Split(FormatRow(repair), "\t")[30]
	for _, payload := range []string{"null", "{}", valid + " {}", strings.Replace(valid, `"version":1`, `"Version":1`, 1), strings.Replace(valid, `"version":1`, `"version":1,"version":1`, 1), strings.Replace(valid, `"version":1`, `"version":2`, 1), strings.Replace(valid, `"replace"`, `"unknown"`, 1), strings.Replace(valid, `"version":1`, `"extra":1,"version":1`, 1), strings.Replace(valid, rows[0].RawID, "bad", 1), strings.Repeat(" ", 65537)} {
		if _, err := ParseRowsChecked(base + payload); err == nil {
			t.Errorf("accepted malformed correction %.100q", payload)
		}
	}
	// Corrupt sample numbers may not hide a correction record.
	bad := strings.Split(base, "\t")
	bad[4] = "N/A"
	if _, err := ParseRowsChecked(strings.Join(bad, "\t") + "{}"); err == nil {
		t.Fatal("invalid correction hidden by nonnumeric actual")
	}
}

func TestResolveCorrectionRejectsInvalidTargets(t *testing.T) {
	for _, kind := range []string{"missing", "cross-issue", "forward", "conflicting", "generation", "operation"} {
		t.Run(kind, func(t *testing.T) {
			rows, repair := correctionFixture(t)
			switch kind {
			case "missing":
				repair.Correction.Targets = []string{strings.Repeat("f", 64)}
			case "cross-issue":
				repair.Issue = "repo#2"
			}
			parsed, err := ParseRowsChecked(FormatRow(repair))
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "forward":
				rows = []LedgerRow{parsed[0], rows[0]}
			case "conflicting", "generation", "operation":
				second := repair
				c := *repair.Correction
				second.Correction = &c
				c.Operation = "op-2"
				if kind == "generation" {
					c.Generation = "generation-2"
					c.Targets = []string{parsed[0].RawID}
				}
				if kind == "operation" {
					c.Operation = "op-1"
					second.Actual = 99
				}
				more, e := ParseRowsChecked(FormatRow(second))
				if e != nil {
					t.Fatal(e)
				}
				rows = append(rows, parsed[0], more[0])
			default:
				rows = append(rows, parsed[0])
			}
			if _, err := ResolveCorrections(rows); err == nil {
				t.Fatal("accepted invalid correction")
			}
		})
	}
}

func FuzzCorrectionParser(f *testing.F) {
	f.Add("{}")
	f.Add(`{"version":1,"operation":"op","generation":"g","disposition":"exclude","targets":["` + strings.Repeat("a", 64) + `"],"input_digest":"` + strings.Repeat("b", 64) + `"}`)
	base := FormatRow(LedgerRow{Issue: "repo#1", Estimate: 1, Actual: 1}) + "\t"
	f.Fuzz(func(t *testing.T, payload string) {
		rows, err := ParseRowsChecked(base + payload)
		if err == nil {
			_, _ = ResolveCorrections(rows)
		}
	})
}

func TestCorrectionChainAndLegacyIdentity(t *testing.T) {
	rows, repair := correctionFixture(t)
	// CR belongs to the exact raw bytes; parsing compatibility does not normalize identity.
	legacy := "repo#2\t1\t0\t1\t2\t0.50\tm\t-\tyes\t2026-01-01\r"
	legacyRows, err := ParseRowsChecked(legacy + "\n")
	if err != nil || len(legacyRows) != 1 || legacyRows[0].RawID != fmt.Sprintf("%x", sha256.Sum256([]byte(legacy))) {
		t.Fatalf("legacy identity: %v %+v", err, legacyRows)
	}
	replacement, err := ParseRowsChecked(FormatRow(repair))
	if err != nil {
		t.Fatal(err)
	}
	exclusion := repair
	c := *repair.Correction
	exclusion.Correction = &c
	c.Operation = "op-2"
	c.Disposition = "exclude"
	c.Targets = []string{replacement[0].RawID}
	excludes, err := ParseRowsChecked(FormatRow(exclusion))
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, replacement[0], excludes[0])
	rows = append(rows, legacyRows...)
	got, err := ResolveCorrections(rows)
	if err != nil || len(got) != 1 || got[0].Issue != "repo#2" {
		t.Fatalf("chain failed: %v %+v", err, got)
	}
}

func FuzzCorrectionExactSuppression(f *testing.F) {
	f.Add(uint8(4), uint64(5))
	f.Add(uint8(1), uint64(1))
	f.Add(uint8(64), uint64(0xffffffffffffffff))
	f.Fuzz(func(t *testing.T, size uint8, mask uint64) {
		n := int(size%64) + 1
		var text strings.Builder
		for i := 0; i < n; i++ {
			text.WriteString(FormatRow(LedgerRow{Issue: "repo#1", Estimate: 2, Actual: float64(i + 1), WindowTrusted: i%2 == 0}) + "\n")
		}
		rows, err := ParseRowsChecked(text.String())
		if err != nil {
			t.Fatal(err)
		}
		var targets []string
		kept := map[string]bool{}
		for i, r := range rows {
			if mask&(uint64(1)<<i) != 0 {
				targets = append(targets, r.RawID)
			} else {
				kept[r.RawID] = true
			}
		}
		if len(targets) == 0 {
			return
		}
		exclusion := LedgerRow{Issue: "repo#1", Correction: &CorrectionRecord{Version: 1, Operation: "property", Generation: "g", Disposition: "exclude", Targets: targets, InputDigest: strings.Repeat("a", 64)}}
		extra, err := ParseRowsChecked(FormatRow(exclusion))
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, extra...)
		got, err := ResolveCorrections(rows)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(kept) {
			t.Fatalf("got %d want %d", len(got), len(kept))
		}
		for _, r := range got {
			if !kept[r.RawID] {
				t.Fatal("incorrect row survives")
			}
			delete(kept, r.RawID)
		}
		if len(kept) != 0 {
			t.Fatal("ordinary row lost")
		}
	})
}

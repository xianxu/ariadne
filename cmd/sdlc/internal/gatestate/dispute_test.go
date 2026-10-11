package gatestate

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// dispute builds a Dispute on id, recorded after round `after`.
func dispute(id string, after int) Dispute {
	return Dispute{ID: id, Rationale: "the finding is wrong", Evidence: "go test ./... -run X: ok", After: after, Timestamp: testTimestamp}
}

// The literal the dispute stream writes is the model's implementer disposition.
func TestDisputedIsTheModelsImplementerDisposition(t *testing.T) {
	m := vocab.Finding()
	if !m.IsDisposition(Disputed) || !m.ImplementerMay(Disputed) || m.JudgeMay(Disputed) {
		t.Fatalf("%q must be a modeled, implementer-only disposition", Disputed)
	}
}

func openIDs(l Ledger) string { return strings.Join(ids(OpenFindings(l)), ",") }

// A dispute settles an Important; on an appeal-only (Critical) it is recorded but the
// finding stays open — the judge alone can settle it.
func TestDisputeSettlesUnlessAppealOnly(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a", "Critical/b")))
	l.Disputes = []Dispute{dispute("PQ-1", 1), dispute("PQ-2", 1)}
	if got := openIDs(l); got != "PQ-2" {
		t.Fatalf("open = %q, want PQ-2 (the Critical is appealed, not settled)", got)
	}
	counts, open := DispositionCounts(l)
	if counts[Disputed] != 1 || open != 1 {
		t.Fatalf("counts = %v open = %d, want disputed 1 / open 1", counts, open)
	}
}

// Ordering: a dispute after round 1 is overridden by round 2's judge answer, and not by
// round 1's own dispositions.
func TestDisputeTimelineOrdering(t *testing.T) {
	cases := []struct {
		name  string
		after int
		r2    []Disposition
		want  string
	}{
		{"judge re-raises after the dispute", 1, dispose("PQ-1", "not-addressed"), "PQ-1"},
		{"judge accepts the dispute", 1, dispose("PQ-1", "withdrawn"), ""},
		{"dispute after round 2 overrides round 2's not-addressed", 2, dispose("PQ-1", "not-addressed"), ""},
		{"unanswered dispute stands", 1, nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := ledgerWith(round(1, nil, findings("Important/a")), round(2, c.r2, nil))
			l.Disputes = []Dispute{dispute("PQ-1", c.after)}
			if got := openIDs(l); got != c.want {
				t.Fatalf("open = %q, want %q", got, c.want)
			}
		})
	}
}

func TestFilterBoundaryCarriesDisputes(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a")))
	l.Rounds[0].Boundary = "M1"
	l.Disputes = []Dispute{dispute("PQ-1", 1)}
	if got := openIDs(FilterBoundary(l, "M1")); got != "" {
		t.Fatalf("a dispute must survive the boundary filter; open = %q", got)
	}
}

// Each refusal row asserts its OWN message, so a shared failure (an empty ledger) cannot
// make every row pass vacuously.
func TestAdmitDispute(t *testing.T) {
	base := func() Ledger {
		return ledgerWith(
			round(1, nil, findings("Important/open one", "Important/fixed one", "Important/reraised")),
			round(2, dispose("PQ-2", "addressed", "PQ-3", "not-addressed"), nil),
		)
	}
	disputed := base()
	disputed.Disputes = []Dispute{dispute("PQ-1", 2)}

	refusals := []struct {
		name string
		l    Ledger
		id   string
		want string
	}{
		{"unknown id", base(), "PQ-9", "no finding PQ-9"},
		{"judge-settled", base(), "PQ-2", "already addressed by the judge"},
		{"duplicate live dispute", disputed, "PQ-1", "already disputed (after round 2)"},
	}
	for _, c := range refusals {
		t.Run(c.name, func(t *testing.T) {
			_, err := AdmitDispute(c.l, c.id)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
	for _, id := range []string{"PQ-1", "PQ-3"} {
		if f, err := AdmitDispute(base(), id); err != nil || f.ID != id {
			t.Errorf("AdmitDispute(%s) = %+v, %v; want admitted", id, f, err)
		}
	}
}

// A judge fence may not carry the implementer's disposition: dropped and named, while
// the valid disposition beside it still applies.
func TestApplyCheckedRejectsImplementerDisposition(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a", "Important/b")))
	got, err := ApplyChecked(l, round(2, dispose("PQ-1", Disputed, "PQ-2", "addressed"), nil))
	if err == nil || !strings.Contains(err.Error(), "implementer-only") {
		t.Fatalf("err = %v, want an implementer-only rejection", err)
	}
	if o := openIDs(got); o != "PQ-1" {
		t.Fatalf("open = %q, want PQ-1 (its judge `disputed` dropped, PQ-2's addressed kept)", o)
	}
}

func TestLiveDisputes(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a", "Important/b")), round(2, dispose("PQ-2", "withdrawn"), nil))
	l.Disputes = []Dispute{dispute("PQ-1", 1), dispute("PQ-2", 1)}
	live := LiveDisputes(l)
	if len(live) != 1 || live[0].Finding.ID != "PQ-1" || live[0].Dispute.Evidence == "" {
		t.Fatalf("live = %+v, want only PQ-1 with its evidence (PQ-2 was answered)", live)
	}
}

// A dispute survives the sidecar round trip.
func TestDisputeRenderParseRoundTrip(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a")))
	l.Disputes = []Dispute{{ID: "PQ-1", Rationale: "\nwrong: see #187", Evidence: "ran `x` # ok", After: 1, Timestamp: testTimestamp, By: "ariadne:2"}}
	back, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Disputes) != 1 || back.Disputes[0].Evidence != "ran `x` # ok" || back.Disputes[0].By != "ariadne:2" {
		t.Fatalf("round trip lost the dispute: %+v", back.Disputes)
	}
	if !strings.Contains(Render(l, "ariadne"), "## Disputes") {
		t.Error("the human projection must show the disputes")
	}
}

// The next judge sees each live dispute with its evidence and is told how to answer it;
// a disputed finding is not in the do-not-re-raise list, and an answered one is gone.
func TestPromptCarriesLiveDisputes(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/byte path unasserted", "Important/b")), round(2, dispose("PQ-2", "withdrawn"), nil))
	l.Disputes = []Dispute{
		{ID: "PQ-1", Rationale: "the full suite asserts it", Evidence: "TestEditorLoopCtrlCExitsZero: exit = 9, want 0", After: 1},
		{ID: "PQ-2", Rationale: "answered", Evidence: "x", After: 1},
	}
	p := RenderPriorFindings(l)
	section, rest, ok := strings.Cut(p, "DISPUTED BY THE IMPLEMENTER")
	if !ok {
		t.Fatalf("prompt has no dispute section:\n%s", p)
	}
	for _, want := range []string{"PQ-1", "the full suite asserts it", "exit = 9, want 0", "`withdrawn`", "`not-addressed`"} {
		if !strings.Contains(rest, want) {
			t.Errorf("dispute section lacks %q", want)
		}
	}
	if strings.Contains(rest, "id: PQ-2") {
		t.Error("an answered dispute must not be put to the judge again")
	}
	if _, disposed, _ := strings.Cut(section, "ALREADY DISPOSED"); strings.Contains(disposed, "PQ-1") {
		t.Error("a live dispute must not be listed as settled — the judge would be told not to touch it")
	}
}

// The operator line counts dispute-settled findings against those raised; an appealed
// Critical still blocks and says why.
func TestDecideReportsDisputes(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/a", "Important/b", "Minor/c")))
	l.Disputes = []Dispute{dispute("PQ-1", 1)}
	l.Rounds = append(l.Rounds, round(2, dispose("PQ-2", "addressed"), nil))
	d := Decide(l, 0)
	if d.Block || len(d.Disputed) != 1 || d.DisputeLine() != "1 of 3 finding(s) settled by implementer DISPUTE, not a fix: PQ-1" {
		t.Fatalf("decision = %+v", d)
	}

	c := ledgerWith(round(1, nil, findings("Critical/a")))
	c.Disputes = []Dispute{dispute("PQ-1", 1)}
	d = Decide(c, 0)
	if !d.Block || len(d.Disputed) != 0 || !strings.Contains(d.Reason, "appeal-only, so it blocks until the judge answers") {
		t.Fatalf("appealed Critical: decision = %+v", d)
	}
}

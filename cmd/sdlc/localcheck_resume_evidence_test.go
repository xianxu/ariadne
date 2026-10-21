package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func resumeCoreFixture(t *testing.T) ResumeProofInput {
	t.Helper()
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	reg, err := flakes.BuildRegistry([]flakes.Card{{ID: "463", Tests: []string{"example.org/p.TestOne"}}}, func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	wall := 1.0
	p := SuiteAttempt{WallSeconds: &wall, Version: 2, ID: "aa", Repo: "x/y", SHA: strings.Repeat("a", 40), Check: "suite", StartedAt: now, FinishedAt: now.Add(time.Second), Exit: 1, LogSHA256: strings.Repeat("b", 64), Failures: []string{"example.org/p.TestOne"}, Outcome: string(flakes.Unknown)}
	return ResumeProofInput{Complete: true, Parent: p, ParentRecordSHA256: strings.Repeat("c", 64), ParentLogSHA256: p.LogSHA256, ParentFacts: SuiteLogFacts{Failed: true, Failures: []string{"example.org/p.TestOne"}}, Policy: flakes.PolicyAccept, Registry: reg}
}
func resumeCoreLink(in ResumeProofInput) *ResumeLink {
	return &ResumeLink{Version: 3, Kind: "resume-link", ID: "bb", Repo: in.Parent.Repo, SHA: in.Parent.SHA, Check: "suite", RerunOf: in.Parent.ID, ParentRecordSHA256: in.ParentRecordSHA256, ParentLogSHA256: in.ParentLogSHA256, ChildID: "cc", Owner: ResumeOwner{PID: 42, PIDStarted: "start", BootID: "boot", Host: "host"}, CreatedAt: in.Parent.FinishedAt.Add(time.Second)}
}
func resumeCorePassed(t *testing.T) ResumeProofInput {
	in := resumeCoreFixture(t)
	in.Link = resumeCoreLink(in)
	in.LinkSHA256 = strings.Repeat("d", 64)
	in.ChildLog = "{\"Action\":\"pass\",\"Package\":\"example.org/p\",\"Test\":\"TestOne\"}\n{\"Action\":\"pass\",\"Package\":\"example.org/p\",\"Elapsed\":1}\n"
	in.ChildLogSHA256 = resumeRecordDigest([]byte(in.ChildLog))
	wall := 1.0
	in.Child = &SuiteAttempt{WallSeconds: &wall, Version: 2, ID: "cc", Repo: in.Parent.Repo, SHA: in.Parent.SHA, Check: suiteRerunCheck, RerunOf: in.Parent.ID, StartedAt: in.Link.CreatedAt, FinishedAt: in.Link.CreatedAt.Add(time.Second), LogSHA256: in.ChildLogSHA256}
	return in
}
func TestResumeEvidenceStrict(t *testing.T) {
	in := resumeCoreFixture(t)
	l := resumeCoreLink(in)
	raw, _ := json.Marshal(l)
	got, receipt, err := decodeResumeEvidence(raw, l.ID+".json")
	if err != nil || got == nil || receipt != nil || got.ChildID != "cc" {
		t.Fatalf("decode: %#v %v", got, err)
	}
	// Future values remain opaque, including numbers outside float64.
	future := strings.TrimSuffix(string(raw), "}") + `,"future":{"huge":1e9999}}`
	if _, _, err := decodeResumeEvidence([]byte(future), "bb.json"); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	for key := range fields {
		t.Run("missing-"+key, func(t *testing.T) {
			var m map[string]json.RawMessage
			_ = json.Unmarshal(raw, &m)
			delete(m, key)
			b, _ := json.Marshal(m)
			if _, _, e := decodeResumeEvidence(b, "bb.json"); e == nil {
				t.Fatal("accepted missing field")
			}
		})
		t.Run("null-"+key, func(t *testing.T) {
			var m map[string]json.RawMessage
			_ = json.Unmarshal(raw, &m)
			m[key] = json.RawMessage("null")
			b, _ := json.Marshal(m)
			if _, _, e := decodeResumeEvidence(b, "bb.json"); e == nil {
				t.Fatal("accepted null")
			}
		})
	}
	for _, bad := range []string{strings.Replace(string(raw), `"version":3`, `"Version":3`, 1), strings.TrimSuffix(string(raw), "}") + `,"id":"bb"}`, strings.Replace(string(raw), `"pid":42`, `"pid":null`, 1), strings.Replace(string(raw), `"child_id":"cc"`, `"child_id":"aa"`, 1), string(raw) + " {}"} {
		if _, _, e := decodeResumeEvidence([]byte(bad), "bb.json"); e == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	if _, _, e := decodeResumeEvidence(raw, "other.json"); e == nil {
		t.Fatal("filename mismatch")
	}
}
func TestResumeProofEligibleForcesRerun(t *testing.T) {
	in := resumeCoreFixture(t)
	p := resolveResumeProof(in)
	if p.State != ResumeEligible || p.Outcome.Kind != flakes.NeedRerun {
		t.Fatalf("%+v", p)
	}
}
func TestResumeProofPassRequiresCompleteBoundEvidence(t *testing.T) {
	in := resumeCorePassed(t)
	if p := resolveResumeProof(in); p.State != ResumePassed || p.Outcome.Kind != flakes.KnownFlakePassed {
		t.Fatalf("%+v", p)
	}
	cases := map[string]func(*ResumeProofInput){"incomplete": func(i *ResumeProofInput) { i.Complete = false }, "record digest": func(i *ResumeProofInput) { i.ParentRecordSHA256 = strings.Repeat("e", 64) }, "log digest": func(i *ResumeProofInput) { i.ParentLogSHA256 = strings.Repeat("e", 64) }, "child digest": func(i *ResumeProofInput) { i.ChildLogSHA256 = strings.Repeat("e", 64) }, "no package pass": func(i *ResumeProofInput) {
		i.ChildLog = strings.Split(i.ChildLog, "\n")[0]
		i.ChildLogSHA256 = resumeRecordDigest([]byte(i.ChildLog))
		i.Child.LogSHA256 = i.ChildLogSHA256
	}, "other parent": func(i *ResumeProofInput) { i.Child.RerunOf = "ff" }, "baseline": func(i *ResumeProofInput) { i.Parent.MainBaseline = true }, "timeout": func(i *ResumeProofInput) { i.Child.TimedOut = true }, "policy off": func(i *ResumeProofInput) { i.Policy = flakes.PolicyOff }, "previous rejected": func(i *ResumeProofInput) { i.Parent.Outcome = string(flakes.RerunInconclusive) }, "facts disagree": func(i *ResumeProofInput) { i.ParentFacts.Failures = nil }, "warnings": func(i *ResumeProofInput) { i.ParentFacts.Warnings = []string{"truncated"} }, "unknown policy": func(i *ResumeProofInput) { i.Policy = "bogus" }, "negative exit": func(i *ResumeProofInput) { i.Parent.Exit = -1 }, "stale registry": func(i *ResumeProofInput) { i.Registry = flakes.Registry{} }, "incidental fail": func(i *ResumeProofInput) {
		i.ChildLog += "{\"Action\":\"fail\",\"Package\":\"example.org/other\",\"Test\":\"TestOther\"}\n"
		i.ChildLogSHA256 = resumeRecordDigest([]byte(i.ChildLog))
		i.Child.LogSHA256 = i.ChildLogSHA256
	}}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			i := resumeCorePassed(t)
			mutate(&i)
			p := resolveResumeProof(i)
			if p.State == ResumePassed || p.State == ResumeEligible || p.Outcome.Kind.Lands() {
				t.Fatalf("unsafe proof %+v", p)
			}
		})
	}
}
func TestResumeProofTerminalReceipt(t *testing.T) {
	in := resumeCoreFixture(t)
	in.Link = resumeCoreLink(in)
	in.LinkSHA256 = strings.Repeat("d", 64)
	if p := resolveResumeProof(in); p.State != ResumeLinked {
		t.Fatalf("%+v", p)
	}
	in.Receipt = &ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: in.Parent.Repo, SHA: in.Parent.SHA, Check: "suite", RerunOf: in.Parent.ID, LinkID: in.Link.ID, LinkSHA256: in.LinkSHA256, ChildID: in.Link.ChildID, Actor: "operator", Note: "verified stopped", Outcome: "interrupted", FinishedAt: in.Link.CreatedAt.Add(time.Second)}
	raw, _ := json.Marshal(in.Receipt)
	l, r, e := decodeResumeEvidence(raw, "dd.json")
	if e != nil || l != nil || r == nil {
		t.Fatalf("receipt decode %v", e)
	}
	if p := resolveResumeProof(in); p.State != ResumeInterrupted {
		t.Fatalf("%+v", p)
	}
	in.Receipt.LinkSHA256 = strings.Repeat("e", 64)
	if p := resolveResumeProof(in); p.State != ResumeRefused {
		t.Fatalf("%+v", p)
	}
	in = resumeCorePassed(t)
	in.Receipt = r
	if p := resolveResumeProof(in); p.State != ResumeRefused {
		t.Fatalf("child/receipt conflict %+v", p)
	}
}
func TestResumeTransitionSequences(t *testing.T) {
	// Exhaustive short event sequences prove that durable link confirmation is
	// required before launch, and every later request cannot launch another child.
	events := []ResumeEvent{ResumeRequested, ResumeLinkPublished, ResumeLaunchConfirmed, ResumeCompletionConfirmed, ResumeChildPublished, ResumeProofValidated, ResumeUncertain, ResumeReconcileAuthorized, ResumeReceiptPublished}
	var visit func(ResumeState, int, int, bool)
	visit = func(state ResumeState, depth, launches int, fenced bool) {
		if depth == 0 {
			return
		}
		for _, event := range events {
			tr := resumeTransition(state, event)
			n := launches
			f := fenced
			if state == ResumePublishing && event == ResumeLinkPublished {
				f = true
			}
			for _, effect := range tr.Effects {
				if effect == ResumeLaunch {
					n++
					if !f || n > 1 {
						t.Fatalf("unsafe launch state=%s event=%s launches=%d fenced=%v", state, event, n, f)
					}
				}
				if effect == ResumeRepost && tr.State != ResumePassed {
					t.Fatal("unproven post")
				}
			}
			visit(tr.State, depth-1, n, f)
		}
	}
	visit(ResumeEligible, 6, 0, false)
	state := ResumeEligible
	for _, e := range events[:6] {
		state = resumeTransition(state, e).State
	}
	if state != ResumePassed {
		t.Fatalf("happy sequence: %s", state)
	}
	for _, s := range []ResumeState{ResumeFailed, ResumeInterrupted, ResumeRefused} {
		for _, e := range events {
			if tr := resumeTransition(s, e); len(tr.Effects) > 0 {
				t.Fatalf("terminal effect %+v", tr)
			}
		}
	}
}
func FuzzResumeEvidence(f *testing.F) {
	seed, _ := json.Marshal(ResumeLink{Version: 3, Kind: "resume-link", ID: "bb", Repo: "x/y", SHA: strings.Repeat("a", 40), Check: "suite", RerunOf: "aa", ParentRecordSHA256: strings.Repeat("b", 64), ParentLogSHA256: strings.Repeat("c", 64), ChildID: "cc", Owner: ResumeOwner{PID: 42, PIDStarted: "start", BootID: "boot", Host: "host"}, CreatedAt: time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)})
	f.Add(seed)
	f.Add([]byte(`{"version":3}`))
	f.Add([]byte(`{"future":1e9999}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		before := string(raw)
		l, r, e := decodeResumeEvidence(raw, "bb.json")
		if string(raw) != before {
			t.Fatal("mutated input")
		}
		if e == nil {
			if (l == nil) == (r == nil) {
				t.Fatal("ambiguous variant")
			}
			if l != nil && (l.ID != "bb" || l.RerunOf == l.ChildID) {
				t.Fatal("invalid identity accepted")
			}
		}
	})
}

func TestResumeProofRequiresCompletedDurations(t *testing.T) {
	for _, which := range []string{"parent", "child"} {
		for _, duration := range []*float64{nil, new(float64), func() *float64 { v := -1.0; return &v }(), func() *float64 { v := math.NaN(); return &v }(), func() *float64 { v := math.Inf(1); return &v }()} {
			t.Run(which, func(t *testing.T) {
				in := resumeCorePassed(t)
				a := &in.Parent
				if which == "child" {
					a = in.Child
				}
				a.WallSeconds = duration
				p := resolveResumeProof(in)
				if duration != nil && *duration == 0 {
					if p.State != ResumePassed {
						t.Fatalf("zero measured duration refused %+v", p)
					}
					return
				}
				if p.State == ResumePassed || p.State == ResumeEligible {
					t.Fatalf("incomplete duration authorized %+v", p)
				}
			})
		}
	}
}
func TestResumeProofChildNeverExplainable(t *testing.T) {
	in := resumeCorePassed(t)
	in.Child.Outcome = string(flakes.RerunInconclusive)
	if p := resolveResumeProof(in); p.State == ResumePassed {
		t.Fatalf("rejected child outcome regained authority %+v", p)
	}
}

func TestResumeTransitionRepostRequiresFreshValidation(t *testing.T) {
	if got := resumeTransition(ResumePassed, ResumeRequested); len(got.Effects) != 0 {
		t.Fatalf("unvalidated repost: %+v", got)
	}
	if got := resumeTransition(ResumePassed, ResumeProofRejected); got.State != ResumeFailed || len(got.Effects) != 0 {
		t.Fatalf("invalidated proof remained passing: %+v", got)
	}
	if got := resumeTransition(ResumePassed, ResumeUncertain); got.State != ResumeRefused || len(got.Effects) != 0 {
		t.Fatalf("uncertain proof remained passing: %+v", got)
	}
}

func TestResumeTransitionCommittedProofCannotBeReconciled(t *testing.T) {
	for _, state := range []ResumeState{ResumeCompleted, ResumePassed} {
		state = resumeTransition(state, ResumeUncertain).State
		tr := resumeTransition(state, ResumeReconcileAuthorized)
		if len(tr.Effects) != 0 {
			t.Fatalf("committed child allowed terminal receipt: %+v", tr)
		}
	}
}
func TestResumeEvidenceRequiredTypesAndReceipt(t *testing.T) {
	in := resumeCoreFixture(t)
	link := resumeCoreLink(in)
	receipt := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: "dd", Repo: link.Repo, SHA: link.SHA, Check: link.Check, RerunOf: link.RerunOf, LinkID: link.ID, LinkSHA256: strings.Repeat("d", 64), ChildID: link.ChildID, Actor: "operator", Note: "verified no descendants remain", Outcome: "interrupted", FinishedAt: link.CreatedAt}
	for _, value := range []any{link, receipt} {
		raw, _ := json.Marshal(value)
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		var id string
		_ = json.Unmarshal(fields["id"], &id)
		for key := range fields {
			for _, replacement := range []json.RawMessage{json.RawMessage("null"), json.RawMessage("[]"), json.RawMessage("true")} {
				var changed map[string]json.RawMessage
				_ = json.Unmarshal(raw, &changed)
				changed[key] = replacement
				b, _ := json.Marshal(changed)
				if _, _, e := decodeResumeEvidence(b, id+".json"); e == nil {
					t.Fatalf("%T accepted %s=%s", value, key, replacement)
				}
			}
			var missing map[string]json.RawMessage
			_ = json.Unmarshal(raw, &missing)
			delete(missing, key)
			b, _ := json.Marshal(missing)
			if _, _, e := decodeResumeEvidence(b, id+".json"); e == nil {
				t.Fatalf("%T missing %s", value, key)
			}
		}
	}
	raw, _ := json.Marshal(link)
	for _, key := range []string{"pid", "pid_started", "boot_id", "host"} {
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		var owner map[string]json.RawMessage
		_ = json.Unmarshal(fields["owner"], &owner)
		delete(owner, key)
		fields["owner"], _ = json.Marshal(owner)
		b, _ := json.Marshal(fields)
		if _, _, e := decodeResumeEvidence(b, "bb.json"); e == nil {
			t.Fatalf("accepted missing owner %s", key)
		}
	}
	// Unknown fields must not conceal duplicate identity at any depth.
	bad := strings.TrimSuffix(string(raw), "}") + `,"future":{"x":1,"X":2}}`
	if _, _, e := decodeResumeEvidence([]byte(bad), "bb.json"); e == nil {
		t.Fatal("accepted nested duplicate")
	}
}

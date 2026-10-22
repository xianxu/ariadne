package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func TestCarWaitSelectionClipsPinnedOwnershipWithoutQueueEvidence(t *testing.T) {
	at := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	id := landingline.WaitIdentity{Repo: "github.com/x/ariadne", Root: "/owner", Slot: "ariadne:4", Head: strings.Repeat("a", 40)}
	wait := landingline.WaitInterval{Seq: 3, Issue: 547, Identity: id, Start: at, End: at.Add(time.Hour), Trains: []int64{2}, RecordDigest: strings.Repeat("b", 64)}
	engagements := []tracker.EngagementWindow{{Start: at.Add(5 * time.Minute), End: at.Add(55 * time.Minute), Owners: []issue.Claimant{{Workspace: id.Slot, Worktree: id.Root}}}}
	sessions := []transcripts.SessionEvidence{{Harness: "codex", ID: "owner", File: "owner", CWD: id.Root}, {Harness: "claude", ID: "foreign", File: "foreign", CWD: "/other"}}
	windows := map[string][]activetime.ActivityWindow{"owner": {{Start: at.Add(10 * time.Minute), End: at.Add(50 * time.Minute)}}, "foreign": {{Start: at, End: wait.End}}}
	slots := map[string]string{"owner": id.Slot, "foreign": "ariadne:2"}
	members := map[string]queueHeadMembership{id.Head: {Known: true, Issues: []int{547}}}
	got, notes := selectCarWaits([]landingline.WaitInterval{wait}, members, engagements, id.Repo, 547, at, wait.End, sessions, windows, slots)
	if len(notes) != 0 || len(got) != 1 || got[0].Start != at.Add(10*time.Minute).Format(time.RFC3339Nano) || got[0].End != at.Add(50*time.Minute).Format(time.RFC3339Nano) || strings.Join(got[0].Sources, ",") != "codex/owner" {
		t.Fatalf("waits=%+v notes=%v", got, notes)
	}
	for _, mutate := range []func(*landingline.WaitInterval){func(w *landingline.WaitInterval) { w.Issue = 548 }, func(w *landingline.WaitInterval) { w.Identity.Repo = "github.com/other/ariadne" }, func(w *landingline.WaitInterval) { w.Identity.Root = "/other" }, func(w *landingline.WaitInterval) { w.Identity.Slot = "ariadne:2" }} {
		bad := wait
		mutate(&bad)
		if got, _ := selectCarWaits([]landingline.WaitInterval{bad}, members, engagements, id.Repo, 547, at, wait.End, sessions, windows, slots); len(got) != 0 {
			t.Fatalf("foreign identity admitted: %+v", got)
		}
	}
	delete(members, id.Head)
	if got, notes := selectCarWaits([]landingline.WaitInterval{wait}, members, engagements, id.Repo, 547, at, wait.End, sessions, windows, slots); len(got) != 0 || len(notes) == 0 {
		t.Fatalf("unknown head trusted: %+v %v", got, notes)
	}
	members[id.Head] = queueHeadMembership{Known: true, Issues: []int{547}}
	if got, notes := selectCarWaits([]landingline.WaitInterval{wait}, members, engagements, id.Repo, 547, at, wait.End, sessions, windows, nil); len(got) != 0 || len(notes) == 0 {
		t.Fatalf("missing source binding was silent: %+v %v", got, notes)
	}
}

func TestActualTrainCoverageKeepsSameSourceWorkAndReplay(t *testing.T) {
	for _, harness := range []string{"claude", "codex"} {
		for _, minutes := range []int{30, 60} {
			t.Run(fmt.Sprintf("%s/%d", harness, minutes), func(t *testing.T) {
				at := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
				until := at.Add(time.Duration(minutes+20) * time.Minute)
				t.Setenv("GIT_AUTHOR_DATE", at.Format(time.RFC3339))
				t.Setenv("GIT_COMMITTER_DATE", at.Format(time.RFC3339))
				root := testfix.Repo(t, testfix.InitialCommit())
				t.Setenv("GIT_AUTHOR_DATE", until.Format(time.RFC3339))
				t.Setenv("GIT_COMMITTER_DATE", until.Format(time.RFC3339))
				testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#547: work around a retained wait")
				head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
				var lines []string
				for minute := 0; minute <= minutes+20; minute += 5 {
					stamp := at.Add(time.Duration(minute) * time.Minute).Format(time.RFC3339)
					if harness == "claude" {
						lines = append(lines, fmt.Sprintf(`{"timestamp":%q,"type":"assistant","message":{"content":[{"type":"tool_use","id":%q,"name":"Edit","input":{"file_path":"work.go"}},{"type":"text","text":"#547 status and substantive work"}]}}`, stamp, fmt.Sprint(minute)))
					} else {
						lines = append(lines, fmt.Sprintf(`{"timestamp":%q,"type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":%q,"input":"work and arbitrary wrapped polls #547"}}`, stamp, fmt.Sprint(minute)))
					}
				}
				data := []byte(strings.Join(lines, "\n") + "\n")
				file := filepath.Join(t.TempDir(), "session.jsonl")
				if err := os.WriteFile(file, data, 0600); err != nil {
					t.Fatal(err)
				}
				session := transcripts.SessionEvidence{Harness: harness, ID: "owner", CWD: root, File: file, Started: at, Length: int64(len(data)), Digest: fmt.Sprintf("%x", sha256.Sum256(data))}
				windows := []activetime.ActivityWindow{{Start: at, End: until, Began: at}}
				pin := ownedMeasurementSource(session, windows)
				pin.Slot = "ariadne:4"
				m := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodEngagementV1, Repository: "github.com/x/ariadne", CodeHead: head, TrackerHead: head, Since: at.Format(time.RFC3339), Until: until.Format(time.RFC3339), GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true, Sources: []gatestate.MeasurementSource{pin}, Engagements: []gatestate.MeasurementEngagement{{Start: at.Format(time.RFC3339), End: until.Format(time.RFC3339), Began: at.Format(time.RFC3339), Roots: []string{root}}}}
				in := &ownedActualInputs{Snapshot: m, Files: []string{file}, Prefixes: map[string]activetime.SourcePrefix{file: {Bytes: session.Length, Digest: session.Digest}}, Windows: map[string][]activetime.ActivityWindow{file: windows}, SourceSlots: map[string]string{file: pin.Slot}, Scope: activetime.Scope{Head: head, Issue: "547"}, Sessions: []transcripts.SessionEvidence{session}, TrackerRefs: []string{head}}
				baseline := computeOwnedActualFor(context.Background(), root, "547", "", in)
				in.CarWaits = []gatestate.MeasurementCarWait{{Ticket: 3, Trains: []int64{2}, Slot: pin.Slot, Head: head, RecordDigest: strings.Repeat("d", 64), Start: at.Add(10 * time.Minute).Format(time.RFC3339), End: at.Add(time.Duration(minutes+10) * time.Minute).Format(time.RFC3339), Sources: []string{harness + "/owner"}}}
				got := computeOwnedActualFor(context.Background(), root, "547", "", in)
				if baseline.Status != actualMeasured || got.Status != baseline.Status || got.Hours != baseline.Hours || got.Inputs == nil || len(got.Inputs.CarWaits) != 1 {
					t.Fatalf("accounting changed: baseline=%+v got=%+v", baseline, got)
				}
				if !strings.Contains(strings.Join(got.Warnings, " "), fmt.Sprintf("unverified coverage: %.2f min", float64(minutes))) {
					t.Fatal(got.Warnings)
				}
				replay, err := ownedReplaySources(*got.Inputs, []transcripts.SessionEvidence{session})
				if err != nil {
					t.Fatal(err)
				}
				replay.Scope.Issue = "547"
				again := computeOwnedActualFor(context.Background(), root, "547", "", replay)
				if again.Status != got.Status || again.Hours != got.Hours || strings.Join(again.Warnings, " ") != strings.Join(got.Warnings, " ") {
					t.Fatalf("replay changed: first=%+v replay=%+v", got, again)
				}
			})
		}
	}
}

type carWaitCountingReader struct {
	r     io.Reader
	bytes int
}

func (r *carWaitCountingReader) Read(p []byte) (int, error) {
	n, e := r.r.Read(p)
	r.bytes += n
	return n, e
}
func TestCarWaitAuditReadsLinearBytes(t *testing.T) {
	for _, n := range []int{8, 16} {
		data := strings.Repeat("{\"kind\":\"enqueue\",\"seq\":1}\n", n)
		r := &carWaitCountingReader{r: strings.NewReader(data)}
		_, notes := scanCarWaitEvidence(context.Background(), r, defaultCarWaitReadLimits)
		if len(notes) != 0 || r.bytes != len(data) {
			t.Fatalf("n=%d bytes=%d expected=%d notes=%v", n, r.bytes, len(data), notes)
		}
	}
}

func TestCarWaitCoverageReportsUnionAndRemainsDescriptive(t *testing.T) {
	base := gatestate.MeasurementCarWait{Ticket: 3, Trains: []int64{2}, Slot: "ariadne:4", Head: strings.Repeat("a", 40), RecordDigest: strings.Repeat("b", 64), Start: "2026-10-22T00:10:00Z", End: "2026-10-22T00:30:00Z", Sources: []string{"codex/root"}}
	other := base
	other.Ticket = 4
	other.Start = "2026-10-22T00:20:00Z"
	other.End = "2026-10-22T00:40:00Z"
	warnings := carWaitCoverageWarnings([]gatestate.MeasurementCarWait{base, other}, nil)
	if !strings.Contains(strings.Join(warnings, " "), "unverified coverage: 30.00 min in a retained train wait") || !strings.Contains(strings.Join(warnings, " "), "no subtraction") {
		t.Fatalf("coverage=%v", warnings)
	}
	other.End = "2026-10-22T01:10:00Z"
	if got := strings.Join(carWaitCoverageWarnings([]gatestate.MeasurementCarWait{base, other}, nil), " "); !strings.Contains(got, "60.00 min") {
		t.Fatal(got)
	}
}

func TestCarWaitAuditBoundsPreserveIncompleteDiagnostic(t *testing.T) {
	file := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(file, []byte("{\"kind\":\"enqueue\",\"seq\":1}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if records, notes := readCarWaitEvidence(context.Background(), file, carWaitReadLimits{Bytes: 8, Records: 10, RecordBytes: 100}); len(records) != 0 || len(notes) == 0 {
		t.Fatalf("cap ignored: %v %v", records, notes)
	}
	if records, notes := readCarWaitEvidence(context.Background(), file, defaultCarWaitReadLimits); len(records) != 0 || len(notes) != 0 {
		t.Fatalf("ordinary audit rejected: %v %v", records, notes)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, notes := readCarWaitEvidence(ctx, file, defaultCarWaitReadLimits); len(notes) == 0 {
		t.Fatal("cancel not observed")
	}
	if _, notes := readCarWaitEvidence(context.Background(), file+"-absent", defaultCarWaitReadLimits); len(notes) != 0 {
		t.Fatal(notes)
	}
}

func TestCarWaitAttachmentBudgetPreservesAccounting(t *testing.T) {
	m := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "repo", CodeHead: strings.Repeat("a", 40), Since: "2026-10-22T00:00:00Z", Until: "2026-10-22T01:00:00Z", GapMinutes: 15, CommitWeight: 1}
	before, _ := json.Marshal(m)
	waits := make([]gatestate.MeasurementCarWait, 17)
	got, notes := attachCarWaitCoverage(m, waits, nil, nil)
	after, _ := json.Marshal(got)
	// A compact optional note may fit; the accounting and numeric facts stay exact.
	noteCount := len(notes) + len(got.CarWaitNotes)
	got.CarWaitNotes = nil
	without, _ := json.Marshal(got)
	if string(before) != string(without) || noteCount == 0 || len(after) == 0 || len(got.CarWaits) != 0 {
		t.Fatalf("budget changed accounting: %+v %v", got, notes)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(before, &fields); err != nil {
		t.Fatal(err)
	}
	fields["future"] = json.RawMessage(`""`)
	empty, _ := json.Marshal(fields)
	fields["future"], _ = json.Marshal(strings.Repeat("x", gatestate.SnapshotLimitBytes-len(empty)-1))
	full, _ := json.Marshal(fields)
	if err := json.Unmarshal(full, &m); err != nil {
		t.Fatal(err)
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	got, notes = attachCarWaitCoverage(m, nil, []string{"retained train wait evidence incomplete"}, nil)
	after, _ = json.Marshal(got)
	if string(after) != string(full) || len(notes) == 0 {
		t.Fatalf("full snapshot not preserved: before=%d after=%d notes=%v", len(full), len(after), notes)
	}
}

func TestCarWaitAuditAttachmentUsesImmutableHeadAndPinnedSources(t *testing.T) {
	at := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	t.Setenv("GIT_AUTHOR_DATE", at.Format(time.RFC3339))
	t.Setenv("GIT_COMMITTER_DATE", at.Format(time.RFC3339))
	root := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#547: car work")
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	common := filepath.Join(root, ".git")
	dir := filepath.Join(common, "sdlc", "landing")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	id := landingline.WaitIdentity{Repo: "github.com/x/ariadne", Root: root, Slot: "ariadne:4", Head: head, OperationID: "car-op", PID: 123, PIDStarted: "car-incarnation", Branch: "000547-car"}
	conductor := id
	conductor.Slot = "ariadne:2"
	conductor.OperationID = "conductor-op"
	conductor.PID = 124
	conductor.PIDStarted = "conductor-incarnation"
	events := []landingline.Event{
		{Kind: landingline.EventWaitStart, Seq: 3, Issue: 547, At: at.Add(10 * time.Minute), Wait: &landingline.WaitFact{Version: 1, Identity: id}},
		{Kind: landingline.EventWaitBind, Seq: 3, Issue: 547, Train: 2, PR: 40, At: at.Add(15 * time.Minute), Wait: &landingline.WaitFact{Version: 1, Identity: id, Conductor: &conductor}},
		{Kind: landingline.EventWaitEnd, Seq: 3, Issue: 547, At: at.Add(50 * time.Minute), Wait: &landingline.WaitFact{Version: 1, Identity: id}},
	}
	var data []byte
	for _, e := range events {
		raw, _ := json.Marshal(e)
		data = append(data, raw...)
		data = append(data, '\n')
	}
	file := filepath.Join(dir, landingline.EventsFile)
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	until := at.Add(time.Hour)
	engagements := []tracker.EngagementWindow{{Start: at, End: until, Owners: []issue.Claimant{{Workspace: id.Slot, Worktree: root}}}}
	fresh := func() *ownedActualInputs {
		return &ownedActualInputs{Snapshot: gatestate.MeasurementInputs{Since: at.Format(time.RFC3339), Until: until.Format(time.RFC3339)}, Sessions: []transcripts.SessionEvidence{{Harness: "codex", ID: "root", File: "source", CWD: root}}, Windows: map[string][]activetime.ActivityWindow{"source": {{Start: at, End: until}}}, SourceSlots: map[string]string{"source": id.Slot}}
	}
	// Current HEAD is a foreign split prefix; only the original car head is authoritative.
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#999: later split prefix")
	out := fresh()
	attachCarWaitInputs(context.Background(), root, common, id.Repo, 547, engagements, out)
	if len(out.CarWaits) != 1 || len(out.CarWaitNotes) != 0 || out.CarWaits[0].Head != head || out.CarWaits[0].RecordDigest == "" {
		t.Fatalf("selection=%+v notes=%v", out.CarWaits, out.CarWaitNotes)
	}
	if err := os.WriteFile(file, append(data, []byte(`{"kind":"wait_end"`)...), 0600); err != nil {
		t.Fatal(err)
	}
	out = fresh()
	attachCarWaitInputs(context.Background(), root, common, id.Repo, 547, engagements, out)
	if len(out.CarWaits) != 0 || len(out.CarWaitNotes) == 0 {
		t.Fatalf("partial scan advertised numeric coverage: %+v", out)
	}
}

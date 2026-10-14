package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// The kind/dispute rule must survive production ledger loading, prompt rendering,
// dispatch admission, and close finalization, not just the pure validator.
func TestCloseUnverifiableDisputeLifecycle(t *testing.T) {
	t.Setenv("WF_PLANS_DIR", "")
	const id = 364
	r, cardPath, detailPath := closeReady(t, id)
	plans := resolvePlansDir("")
	ledgerPath := boundaryGatePath(plans, filepath.Base(detailPath))
	const reviewedRef = "refs/sdlc/reviewed/000364/close"
	const hostEvidence = "host: go test ./pkg/probe -run TestProcessLookup -count=1 passed; process listing returned the expected parent"
	first, _ := stubJudge(t, "VERDICT: REWORK\n```findings\nfindings:\n  - id: new\n    kind: unverifiable\n    severity: Important\n    title: Process lookup cannot be verified here\n    detail: go test ./pkg/probe failed because sandbox process lookup was denied; a passing host run would answer it.\n```\n")
	if _, stderr, err := executeSDLCTestCommand(closeArgs("364")...); err == nil || *first != 1 {
		t.Fatalf("first review must accept one blocking round: attempts=%d err=%v\n%s", *first, err, stderr)
	}
	raised := readDisputeLedger(t, ledgerPath, id)
	if len(raised.Rounds) != 1 || len(raised.Rounds[0].New) != 1 || raised.Rounds[0].New[0].ID != "BR-1" || raised.Rounds[0].New[0].Kind != "unverifiable" || raised.Rounds[0].Reviewed != "" {
		t.Fatalf("accepted first finding missing or advanced reviewed head: %+v", raised)
	}
	if _, stderr, err := executeSDLCTestCommand("finding", "dispute", "--issue", "364", "--id", "BR-1", "--rationale", "the restriction is specific to the reviewer sandbox", "--evidence", hostEvidence); err != nil {
		t.Fatalf("production finding dispute: %v\n%s", err, stderr)
	}
	disputed := readDisputeLedger(t, ledgerPath, id)
	if len(disputed.Disputes) != 1 || disputed.Disputes[0].ID != "BR-1" || disputed.Disputes[0].Evidence != hostEvidence || disputed.Disputes[0].After != 1 {
		t.Fatalf("host evidence not attached to stable finding ID: %+v", disputed)
	}
	head := r.git("rev-parse", "HEAD")
	card := r.card(cardPath)
	detailBefore, err := os.ReadFile(filepath.Join(r.root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	if gitSucceeds(r.root, "rev-parse", "-q", "--verify", reviewedRef) {
		t.Fatal("REWORK/dispute advanced reviewed ref")
	}
	assertPrompt := func(t *testing.T, prompt string) {
		t.Helper()
		for _, want := range []string{"BR-1", "kind: unverifiable", hostEvidence} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("next reviewer did not receive %q:\n%s", want, prompt)
			}
		}
	}
	for _, tc := range []struct{ name, body string }{
		{"unanswered", "findings: []"},
		{"noteless", "dispose:\n  - id: BR-1\n    disposition: withdrawn"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attempts, prompt := stubJudge(t, "VERDICT: SHIP\n```findings\n"+tc.body+"\n```\n")
			_, stderr, err := executeSDLCTestCommand(closeArgs("364")...)
			assertPrompt(t, *prompt)
			if err == nil || *attempts != 2 || !strings.Contains(stderr, "protocol failed after two attempts") {
				t.Fatalf("invalid dispute answer admitted: attempts=%d err=%v\n%s", *attempts, err, stderr)
			}
			if after := readDisputeLedger(t, ledgerPath, id); !reflect.DeepEqual(after, disputed) {
				t.Fatalf("invalid attempts advanced accepted ledger:\nbefore=%+v\nafter=%+v", disputed, after)
			}
			if r.git("rev-parse", "HEAD") != head || gitSucceeds(r.root, "rev-parse", "-q", "--verify", reviewedRef) {
				t.Fatal("invalid attempts advanced HEAD/reviewed ref")
			}
			detailAfter, err := os.ReadFile(filepath.Join(r.root, detailPath))
			if err != nil {
				t.Fatal(err)
			}
			if r.card(cardPath) != card || string(detailAfter) != string(detailBefore) {
				t.Fatal("invalid attempts finalized or changed the issue")
			}
		})
	}
	if t.Failed() {
		return
	}
	const note = "The passing host process-lookup test answers the sandbox restriction; the invariant is verified."
	attempts, prompt := stubJudge(t, "VERDICT: SHIP\n```findings\ndispose:\n  - id: BR-1\n    disposition: withdrawn\n    note: "+note+"\n```\n")
	if _, stderr, err := executeSDLCTestCommand(closeArgs("364")...); err != nil {
		t.Fatalf("evidence-answering disposition should close: %v\n%s", err, stderr)
	}
	assertPrompt(t, *prompt)
	accepted := readDisputeLedger(t, ledgerPath, id)
	if *attempts != 1 || len(accepted.Rounds) != 2 {
		t.Fatalf("accepted attempts=%d ledger=%+v", *attempts, accepted)
	}
	last := accepted.Rounds[1]
	if len(last.Dispositions) != 1 || last.Dispositions[0].ID != "BR-1" || last.Dispositions[0].State != "withdrawn" || last.Dispositions[0].Note != note || last.Reviewed != head {
		t.Fatalf("accepted disposition/reviewed head lost: %+v", last)
	}
	if len(gatestate.LiveDisputes(accepted)) != 0 || len(gatestate.OpenFindings(accepted)) != 0 {
		t.Fatalf("accepted answer left dispute open: %+v", accepted)
	}
	// Tracker closes pin the evidence commit, which descends from the reviewed
	// head; the card and ledger retain the actual reviewed head as authority.
	completion, ok, err := issue.CardCompletion([]byte(r.card(cardPath)))
	if err != nil || !ok || completion.ReviewedHEAD != head || completion.EvidenceCommit == "" {
		t.Fatalf("completion lost reviewed/evidence identity: %+v ok=%v err=%v", completion, ok, err)
	}
	if r.git("rev-parse", reviewedRef) != completion.EvidenceCommit || !strings.Contains(r.card(cardPath), "status: codecomplete") {
		t.Fatalf("valid answer did not pin evidence and close issue: completion=%+v card=%s", completion, r.card(cardPath))
	}
}

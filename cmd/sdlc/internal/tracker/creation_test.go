package tracker

import (
	"fmt"
	"strings"
	"testing"
)

func operationSpec() ReceiptSpec {
	return ReceiptSpec{Token: "operation-1", Repository: "github.com/example/repo", IssueID: "000001", CardPath: "workshop/issue-cards/000001-example.md", SourcePath: "workshop/issues/000001-example.md", DestinationPath: "workshop/issues/000001-example.md", SourceBranch: "refs/heads/work", SourceBase: strings.Repeat("1", 40), SourceHEAD: strings.Repeat("2", 40), SourceBlob: strings.Repeat("3", 40), CardOID: strings.Repeat("4", 40), TrackerBase: strings.Repeat("5", 40), MainBase: strings.Repeat("6", 40), Source: LocalSource}
}

func TestCreationReservesBeforeMaterializingAndNeverReallocatesAfterReserve(t *testing.T) {
	s, err := NewCreation(operationSpec())
	if err != nil {
		t.Fatal(err)
	}
	s, effects, err := StepCreation(s, Event{Kind: EventBegin, Binding: s.Binding()})
	if err != nil || effects[0].Kind != PrepareCandidate {
		t.Fatalf("begin: %v %v", effects, err)
	}
	s, effects, err = StepCreation(s, Event{Kind: EventCandidatePrepared, Binding: s.Binding(), CandidateOID: strings.Repeat("7", 40)})
	if err != nil || effects[0].Kind != PersistReceipt {
		t.Fatalf("prepare: %v %v", effects, err)
	}
	s, effects, err = StepCreation(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
	if err != nil || effects[0].Kind != PublishCard {
		t.Fatalf("save: %v %v", effects, err)
	}
	s, effects, err = StepCreation(s, Event{Kind: EventConfirmed, Binding: s.Binding(), CandidateOID: strings.Repeat("7", 40), ResultCardOID: strings.Repeat("8", 40)})
	if err != nil || effects[0].Kind != PrepareCandidate || effects[0].Stage != "creation.detail" {
		t.Fatalf("reserve: %v %v", effects, err)
	}
	if _, _, err := StepCreation(s, Event{Kind: EventRefRace, Binding: s.Binding()}); err == nil {
		t.Fatal("local materialization must not reallocate")
	}
	if s.Binding().CardOID != strings.Repeat("8", 40) {
		t.Fatal("reserved generation lost")
	}
}

func TestCreationRetryBoundAndAllocationReplacement(t *testing.T) {
	s, _ := NewCreation(operationSpec())
	var effects []Effect
	var err error
	s, _, _ = StepCreation(s, Event{Kind: EventBegin, Binding: s.Binding()})
	for i := 0; i < 3; i++ {
		s, _, err = StepCreation(s, Event{Kind: EventCandidatePrepared, Binding: s.Binding(), CandidateOID: fmt.Sprintf("%040x", 100+i)})
		if err != nil {
			t.Fatal(err)
		}
		s, _, err = StepCreation(s, Event{Kind: EventReceiptSaved, Binding: s.Binding()})
		if err != nil {
			t.Fatal(err)
		}
		next, got, raceErr := StepCreation(s, Event{Kind: EventRefRace, Binding: s.Binding()})
		if i == 2 {
			if raceErr == nil {
				t.Fatal("third rejection granted a fourth publication attempt")
			}
			return
		}
		if raceErr != nil || got[0].Kind != RefreshInputs {
			t.Fatalf("race: %v %v", got, raceErr)
		}
		s = next
		spec := s.Receipt().Spec()
		spec.IssueID = fmt.Sprintf("%06d", i+2)
		spec.CardPath = fmt.Sprintf("workshop/issue-cards/%s-example.md", spec.IssueID)
		spec.SourcePath = fmt.Sprintf("workshop/issues/%s-example.md", spec.IssueID)
		spec.DestinationPath = spec.SourcePath
		s, effects, err = StepCreation(s, Event{Kind: EventRevalidated, Binding: s.Binding(), Replacement: &spec})
		if err != nil || effects[0].Kind != PrepareCandidate {
			t.Fatalf("revalidate: %v %v", effects, err)
		}
	}
}

package acquire

import "testing"

func TestAcquisitionEventSequences(t *testing.T) {
	events := []acquisitionEvent{existingVerified, absentRemote, stageCreated, cloneSucceeded, checksSucceeded, destinationAbsent, publishSucceeded, operationFailed, publicationUncertain, destinationAppeared}
	// Exhaust short event sequences: publication needs the entire ordered proof.
	var walk func(acquisitionState, int)
	walk = func(state acquisitionState, depth int) {
		if depth == 0 {
			return
		}
		for _, event := range events {
			next, err := transition(state, event)
			if err != nil {
				continue
			}
			if next == published && (state != publishing || event != publishSucceeded) {
				t.Fatalf("premature publication: %v %v", state, event)
			}
			if next == stagedVerified && (state != checking || event != checksSucceeded) {
				t.Fatal("unchecked stage")
			}
			walk(next, depth-1)
		}
	}
	walk(uninspected, 7)
	state := uninspected
	for _, event := range []acquisitionEvent{absentRemote, stageCreated, cloneSucceeded, checksSucceeded, destinationAbsent, publishSucceeded} {
		var err error
		state, err = transition(state, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	if state != published {
		t.Fatal(state)
	}
	for _, event := range events {
		if _, err := transition(published, event); err == nil {
			t.Fatalf("published checkout allows mutation event %v", event)
		}
	}
}

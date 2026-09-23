package gitx

import (
	"errors"
	"testing"
)

func TestPublicationStep(t *testing.T) {
	for _, tc := range []struct {
		push         pushObservation
		confirmation confirmationObservation
		want         publicationAction
	}{
		{pushAccepted, publicationUnconfirmed, publicationSucceeded},
		{pushRejected, publicationConfirmed, publicationSucceeded},
		{pushUnknown, publicationConfirmed, publicationSucceeded},
		{pushRejected, publicationAbsentMoved, publicationRetry},
		{pushRejected, publicationAbsentSame, publicationRefuse},
		{pushRejected, publicationUnconfirmed, publicationUncertain},
		{pushUnknown, publicationAbsentMoved, publicationUncertain},
		{pushUnknown, publicationAbsentSame, publicationUncertain},
		{pushUnknown, publicationUnconfirmed, publicationUncertain},
	} {
		if got := publicationStep(tc.push, tc.confirmation); got != tc.want {
			t.Errorf("%v,%v = %v want %v", tc.push, tc.confirmation, got, tc.want)
		}
	}
}

func TestPublicationStep_ReceivePackCASRejection(t *testing.T) {
	old := "1111111111111111111111111111111111111111"
	next := "2222222222222222222222222222222222222222"
	out := []byte("!\tcandidate:refs/heads/main\t[remote rejected] (failed to update ref)\nremote: error: cannot lock ref 'refs/heads/main': is at " + next + " but expected " + old + "\n")
	if got := observedPush(errors.New("push rejected"), out); got != pushRejected {
		t.Fatalf("receive-pack CAS loss classified %v, want rejected", got)
	}
}

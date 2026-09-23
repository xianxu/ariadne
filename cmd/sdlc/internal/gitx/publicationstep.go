package gitx

import "strings"

// The retry decision describes knowledge, not desired remote state. A transport
// failure never becomes permission to replay simply because a later ref differs.
type pushObservation uint8

const (
	pushUnknown pushObservation = iota
	pushAccepted
	pushRejected
)

type confirmationObservation uint8

const (
	publicationUnconfirmed confirmationObservation = iota
	publicationConfirmed
	publicationAbsentSame
	publicationAbsentMoved
)

type publicationAction uint8

const (
	publicationUncertain publicationAction = iota
	publicationSucceeded
	publicationRetry
	publicationRefuse
)

func publicationStep(push pushObservation, confirmation confirmationObservation) publicationAction {
	if push == pushAccepted || confirmation == publicationConfirmed {
		return publicationSucceeded
	}
	if push == pushRejected {
		switch confirmation {
		case publicationAbsentMoved:
			return publicationRetry
		case publicationAbsentSame:
			return publicationRefuse
		}
	}
	return publicationUncertain
}
func observedPush(err error, out []byte) pushObservation {
	// Git reports up-to-date without enforcing a stale lease. No write won.
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "=\t") {
			return pushRejected
		}
	}
	if err == nil {
		return pushAccepted
	}
	if pushRejectedExplicitly(out) {
		return pushRejected
	}
	return pushUnknown
}
func observedConfirmation(base, now string, confirmed bool, err error) confirmationObservation {
	if err != nil {
		return publicationUnconfirmed
	}
	if confirmed {
		return publicationConfirmed
	}
	if now == base {
		return publicationAbsentSame
	}
	return publicationAbsentMoved
}

package main

import "fmt"

// integrationState is advanced only by transitionIntegration. IO reports typed
// events; it cannot separately advance the head, lease base or retry counter.
type integrationState struct {
	head, oldHead, main string
	pr                  landingPR
	attempts            int
	resume              bool
	phase               integrationStep
}

type integrationEventKind uint8

const (
	integrationMainObserved integrationEventKind = iota
	integrationMerged
	integrationMergeFailed
	integrationReplayAccepted
	integrationReplayRejected
	integrationPushRegistered
	integrationPushFailed
	integrationGatesChecked
	integrationGatesRejected
	integrationRecoveryCompleted
)

type integrationEvent struct {
	kind       integrationEventKind
	main, head string
	pr         landingPR
	included   bool
}

func newIntegrationState(head string, pr landingPR, resume bool) integrationState {
	return integrationState{head: head, pr: pr, resume: resume, phase: integrationGates}
}

func transitionIntegration(s integrationState, e integrationEvent) (integrationState, integrationStep, error) {
	invalid := func() (integrationState, integrationStep, error) {
		return s, integrationRefuse, fmt.Errorf("invalid integration event %d in phase %s", e.kind, s.phase)
	}
	switch e.kind {
	case integrationMainObserved, integrationGatesChecked:
		if e.kind == integrationMainObserved && s.phase == integrationPush {
			s.main = e.main
			if !e.included {
				s.phase = integrationRetryReset
			}
			return s, s.phase, nil
		}
		if s.phase != integrationGates && !(e.kind == integrationMainObserved && s.phase == integrationMerge) {
			return invalid()
		}
		s.main = e.main
		if e.kind == integrationGatesChecked && e.included {
			s.phase = integrationLand
			break
		}
		if s.resume {
			s.oldHead = s.pr.HeadOID
			s.phase = integrationReplay
		} else if !e.included {
			if s.attempts >= integrationLimit {
				s.phase = integrationRefuse
			} else {
				s.oldHead = s.head
				s.phase = integrationMerge
			}
		} else {
			s.phase = integrationGates
		}
	case integrationMerged:
		if s.phase != integrationMerge || e.head == "" {
			return invalid()
		}
		s.head = e.head
		s.phase = integrationReplay
	case integrationMergeFailed:
		if s.phase != integrationMerge {
			return invalid()
		}
		s.phase = integrationAbort
	case integrationReplayAccepted, integrationReplayRejected:
		if s.phase != integrationReplay {
			return invalid()
		}
		if e.kind == integrationReplayAccepted {
			s.phase = integrationPush
		} else {
			s.phase = integrationReset
		}
	case integrationPushRegistered:
		if s.phase != integrationPush || e.pr.HeadOID != s.head {
			return invalid()
		}
		s.pr = e.pr
		s.attempts++
		s.resume = false
		s.phase = integrationGates
	case integrationGatesRejected:
		if s.phase != integrationGates {
			return invalid()
		}
		s.phase = integrationRefuse
	case integrationRecoveryCompleted:
		if s.phase != integrationAbort && s.phase != integrationReset && s.phase != integrationRetryReset {
			return invalid()
		}
		s.head = s.oldHead
		if s.phase == integrationRetryReset {
			s.attempts++
			s.resume = false
			s.phase = integrationGates
		} else {
			s.phase = integrationRefuse
		}
	case integrationPushFailed:
		if s.phase != integrationPush {
			return invalid()
		}
		s.phase = integrationRefuse
	default:
		return invalid()
	}
	return s, s.phase, nil
}

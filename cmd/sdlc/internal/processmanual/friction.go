package processmanual

import (
	"regexp"
	"strings"
)

// GateEvent is one classified bypass-ACK or gate-refusal observed in an sdlc
// invocation's output (#172 friction audit).
type GateEvent struct {
	Kind          GateEventKind
	Gate          string // flag name, e.g. "no-atlas"
	Command       string // the sdlc verb whose output this came from
	ViaForce      bool   // bypassed via --force (vs the specific --no-<gate>)
	Observability Observability
}

type GateEventKind int

const (
	GateBypass GateEventKind = iota
	GateRefusal
)

// Observability records how completely a gate's events can be measured.
type Observability int

const (
	ObsFull       Observability = iota // ACK/refusal is emitted and names the flag
	ObsForceOnly                       // change-code: bypass observable only via --force (silent alone)
	ObsFlagOmitted                     // merge/push: refusal never names the flag (best-effort attribution)
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// contamination markers: a line bearing any of these is source code, a cat-n file
// read, or a format-string template — NOT a real runtime ACK/refusal.
var (
	catnPrefixRE     = regexp.MustCompile(`^\s*\d+\t`) // "944\t…" cat-n read
	contamSubstrings = []string{
		"%s", "%d", "%q", "%v", // unexpanded format verbs
		"cwarn(", "cinfo(", "cok(", "Sprintf(", "Fprint", "append(", "stderr,",
	}
)

func isContamination(line string) bool {
	if catnPrefixRE.MatchString(line) {
		return true
	}
	for _, m := range contamSubstrings {
		if strings.Contains(line, m) {
			return true
		}
	}
	return false
}

// classifyOutputLine classifies ONE line of an sdlc invocation's output as a gate
// bypass, a gate refusal, or neither, given the invocation's verb. Anchoring to the
// verb (plus the runtime-reset requirement for ACKs and the grammar-anchored refusal
// patterns) is what separates a real event from the source/log-read contamination
// that saturates this repo's transcripts.
func classifyOutputLine(line, verb string) (GateEvent, bool) {
	if isContamination(line) {
		return GateEvent{}, false
	}
	hasReset := strings.Contains(line, "\x1b[0m ") // runtime cwarn/cinfo/cok marker
	stripped := ansiRE.ReplaceAllString(line, "")

	for i := range GateCatalog {
		g := &GateCatalog[i]
		if !contains(g.Commands, verb) {
			continue
		}
		// Bypass ACK — requires the runtime reset (source restatements lack it).
		if hasReset && g.ackRE != nil && g.ackRE.MatchString(stripped) {
			return GateEvent{
				Kind: GateBypass, Gate: g.Flag, Command: verb,
				ViaForce:      g.Grammar == grammarG2 || strings.Contains(stripped, "--force:"),
				Observability: bypassObs(g),
			}, true
		}
		// Refusal — grammar+digit-anchored, NOT reset-gated (runtime refusals are
		// plain strings). The exact per-gate pattern rejects the warmup twin.
		if g.HasRefusal && g.refusalRE != nil && g.refusalRE.MatchString(stripped) {
			return GateEvent{
				Kind: GateRefusal, Gate: g.Flag, Command: verb,
				Observability: refusalObs(g),
			}, true
		}
	}
	return GateEvent{}, false
}

func bypassObs(g *GateSig) Observability {
	if g.SilentAlone {
		return ObsForceOnly
	}
	return ObsFull
}

func refusalObs(g *GateSig) Observability {
	if !g.RefusalNamesFlag {
		return ObsFlagOmitted
	}
	return ObsFull
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

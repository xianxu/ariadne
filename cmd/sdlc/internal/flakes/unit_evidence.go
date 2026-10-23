package flakes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const maxUnitChannel = 16 << 20

// Match readLocalCheckLog: a mixed outer log has a separate suite budget.
const maxUnitOuterLog = 64 << 20
const maxUnitLine = 1 << 20
const maxUnitEntries = 10000

// UnitEvent is a control frame written by the process observer, never by a
// child's diagnostics. Exit belongs to the observed process, not its summary.
type UnitEvent struct {
	Version    int    `json:"sdlc_unit"`
	Invocation string `json:"invocation"`
	Unit       string `json:"unit"`
	Event      string `json:"event"`
	Case       string `json:"case,omitempty"`
	Result     string `json:"result,omitempty"`
	Exit       *int   `json:"exit,omitempty"`
	Complete   bool   `json:"complete"`
	Errors     int    `json:"errors"`
}

type CaseResult struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}
type UnitResult struct {
	Passed   bool
	Complete bool
	Exit     int
	Errors   int
	Cases    map[string]string
}
type UnitFacts struct {
	Failures     []string
	Unattributed []string
	Units        map[string]UnitResult
	Seen         bool
	Visits       int
}

type unitPhase uint8

const (
	unitRunning unitPhase = iota
	unitFinished
	unitUnusable
)

type unitState struct {
	phase      unitPhase
	invocation string
	end        UnitEvent
	cases      map[string]string
}

// UnitEvidence owns authoritative state; Finish returns independent snapshots.
type UnitEvidence struct {
	units        map[string]*unitState
	unattributed []string
	seen         bool
	visits       int
}

func NewUnitEvidence() *UnitEvidence { return &UnitEvidence{units: map[string]*unitState{}} }

// reservedUnitMarker recognizes only top-level object keys, including a
// truncated reserved frame. Quoted/nested diagnostic output cannot opt in.
func reservedUnitMarker(line string) bool {
	if len(line) == 0 || line[0] != '{' {
		return false
	}
	depth := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case '"':
			start := i
			i++
			for i < len(line) {
				if line[i] == '\\' {
					i += 2
					continue
				}
				if line[i] == '"' {
					break
				}
				i++
			}
			if i >= len(line) {
				return false
			}
			end := i + 1
			prior := start - 1
			for prior >= 0 && (line[prior] == ' ' || line[prior] == '\t' || line[prior] == '\r') {
				prior--
			}
			keyPosition := prior >= 0 && (line[prior] == '{' || line[prior] == ',')
			if depth == 1 && keyPosition {
				var key string
				if json.Unmarshal([]byte(line[start:end]), &key) == nil && strings.EqualFold(key, "sdlc_unit") {
					return true
				}
			}
		}
	}
	return false
}

// strictUnitObject rejects ambiguity while tolerating additive future members.
// Required fields and canonical spellings remain owned by this protocol.
func strictUnitObject(data []byte, fields map[string]any, required ...string) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("invalid UTF-8 JSON")
	}
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return fmt.Errorf("expected object")
	}
	for _, key := range required {
		if _, ok := raw[key]; !ok {
			return fmt.Errorf("missing %s", key)
		}
	}
	for key, value := range raw {
		target, known := fields[key]
		if !known {
			for canonical := range fields {
				if strings.EqualFold(canonical, key) {
					return fmt.Errorf("noncanonical field %q", key)
				}
			}
			continue
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("null %s", key)
		}
		if err := json.Unmarshal(value, target); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func validateUnitEvent(e UnitEvent) error {
	if e.Version != 1 {
		return fmt.Errorf("unsupported unit event version")
	}
	if len(e.Invocation) != 32 || strings.Trim(e.Invocation, "0123456789abcdef") != "" {
		return fmt.Errorf("invalid invocation")
	}
	unit, err := ParseUnit(e.Unit)
	if err != nil {
		return err
	}
	if e.Errors < 0 || e.Errors > maxUnitEntries {
		return fmt.Errorf("invalid error count")
	}
	switch e.Event {
	case "start":
		if e.Case != "" || e.Result != "" || e.Exit != nil || e.Complete || e.Errors != 0 {
			return fmt.Errorf("invalid start fields")
		}
	case "case":
		if unit.Kind() != KindLua || e.Exit != nil || e.Complete || e.Errors != 0 {
			return fmt.Errorf("invalid case fields")
		}
		if _, err := ParseIdentity(e.Unit + "#" + e.Case); err != nil {
			return err
		}
		if e.Result != "pass" && e.Result != "fail" && e.Result != "skip" {
			return fmt.Errorf("invalid case result")
		}
	case "finish":
		if e.Case != "" || e.Result != "" || e.Exit == nil || *e.Exit < 0 || *e.Exit > 255 {
			return fmt.Errorf("invalid finish fields")
		}
	default:
		return fmt.Errorf("invalid unit event kind")
	}
	return nil
}

// DecodeUnitEvent ignores ordinary/indented diagnostics. A recognized but
// malformed control frame returns an error so it can never erase a failure.
func DecodeUnitEvent(line string) (UnitEvent, bool, error) {
	var e UnitEvent
	recognized := reservedUnitMarker(line)
	if !recognized {
		return e, false, nil
	}
	if len(line) > maxUnitLine {
		return e, true, fmt.Errorf("unit event exceeds 1 MiB")
	}
	err := strictUnitObject([]byte(line), map[string]any{"sdlc_unit": &e.Version, "invocation": &e.Invocation, "unit": &e.Unit, "event": &e.Event, "case": &e.Case, "result": &e.Result, "exit": &e.Exit, "complete": &e.Complete, "errors": &e.Errors}, "sdlc_unit", "invocation", "unit", "event", "complete", "errors")
	if err == nil {
		err = validateUnitEvent(e)
	}
	return e, true, err
}

// DecodeUnitSummary reads only the private Lua summary. It does not infer a
// process exit or grant proof; the observer supplies that independently.
func DecodeUnitSummary(data []byte, unit string) ([]CaseResult, int, error) {
	if len(data) > maxUnitChannel {
		return nil, 0, fmt.Errorf("unit summary exceeds 16 MiB")
	}
	id, err := ParseUnit(unit)
	if err != nil || id.Kind() != KindLua {
		return nil, 0, fmt.Errorf("expected Lua unit")
	}
	var version, errors int
	var bound string
	var complete bool
	var raw []json.RawMessage
	err = strictUnitObject(data, map[string]any{"version": &version, "unit": &bound, "complete": &complete, "cases": &raw, "errors": &errors}, "version", "unit", "complete", "cases", "errors")
	if err != nil {
		return nil, 0, err
	}
	if version != 1 || bound != unit || !complete || errors < 0 || errors > maxUnitEntries || len(raw) > maxUnitEntries {
		return nil, 0, fmt.Errorf("invalid unit summary metadata")
	}
	out := make([]CaseResult, 0, len(raw))
	seen := map[string]bool{}
	for _, data := range raw {
		var c CaseResult
		if err := strictUnitObject(data, map[string]any{"name": &c.Name, "result": &c.Result}, "name", "result"); err != nil {
			return nil, 0, err
		}
		if _, err := ParseIdentity(unit + "#" + c.Name); err != nil {
			return nil, 0, err
		}
		if (c.Result != "pass" && c.Result != "fail") || seen[c.Name] {
			return nil, 0, fmt.Errorf("invalid or duplicate case")
		}
		seen[c.Name] = true
		out = append(out, c)
	}
	return out, errors, nil
}

// Apply consumes each event exactly once. Any lifecycle conflict permanently
// revokes that unit's proof, including conflicts arriving after its terminal.
func (m *UnitEvidence) Apply(e UnitEvent) error {
	m.seen = true
	m.visits++
	if m.units == nil {
		m.units = map[string]*unitState{}
	}
	invalid := func(err error) error {
		state := m.units[e.Unit]
		if state == nil && len(m.units) < maxUnitEntries {
			if _, unitErr := ParseUnit(e.Unit); unitErr == nil {
				state = &unitState{phase: unitUnusable, invocation: e.Invocation, cases: map[string]string{}}
				m.units[e.Unit] = state
			}
		}
		if state != nil {
			state.phase = unitUnusable
		}
		m.unattributed = append(m.unattributed, err.Error())
		return err
	}
	if err := validateUnitEvent(e); err != nil {
		return invalid(err)
	}
	state := m.units[e.Unit]
	if state == nil {
		if len(m.units) >= maxUnitEntries {
			return invalid(fmt.Errorf("too many execution units"))
		}
		if e.Event != "start" {
			return invalid(fmt.Errorf("%s before start for %s", e.Event, e.Unit))
		}
		m.units[e.Unit] = &unitState{phase: unitRunning, invocation: e.Invocation, cases: map[string]string{}}
		return nil
	}
	if state.phase == unitUnusable {
		return fmt.Errorf("unusable unit %s", e.Unit)
	}
	if state.invocation != e.Invocation || state.phase == unitFinished {
		return invalid(fmt.Errorf("conflicting lifecycle for %s", e.Unit))
	}
	switch e.Event {
	case "start":
		return invalid(fmt.Errorf("duplicate start for %s", e.Unit))
	case "case":
		if _, ok := state.cases[e.Case]; ok || len(state.cases) >= maxUnitEntries {
			return invalid(fmt.Errorf("duplicate or excessive case for %s", e.Unit))
		}
		state.cases[e.Case] = e.Result
	case "finish":
		state.phase = unitFinished
		state.end = e
		exit := *e.Exit
		state.end.Exit = &exit
	}
	return nil
}

func (m *UnitEvidence) Finish() UnitFacts {
	f := UnitFacts{Seen: m.seen, Visits: m.visits, Units: map[string]UnitResult{}, Unattributed: append([]string(nil), m.unattributed...)}
	for unit, state := range m.units {
		result := UnitResult{Cases: map[string]string{}}
		failed := []string{}
		for name, status := range state.cases {
			result.Cases[name] = status
			if status == "fail" {
				failed = append(failed, unit+"#"+name)
			}
		}
		if state.end.Exit != nil {
			result.Exit = *state.end.Exit
			result.Errors = state.end.Errors
		}
		if state.phase == unitUnusable {
			f.Units[unit] = result
			continue
		}
		if state.phase != unitFinished {
			f.Unattributed = append(f.Unattributed, "incomplete execution for "+unit)
			f.Units[unit] = result
			continue
		}
		isScript := strings.HasPrefix(unit, "script:")
		if isScript && result.Exit > 0 {
			f.Failures = append(f.Failures, unit)
		}
		consistent := state.end.Complete && result.Errors == 0
		if !isScript {
			consistent = consistent && ((result.Exit == 0 && len(failed) == 0) || (result.Exit > 0 && len(failed) > 0))
		}
		if !consistent {
			f.Unattributed = append(f.Unattributed, "inconsistent completion for "+unit)
		} else {
			result.Complete = true
			result.Passed = result.Exit == 0
			if !isScript {
				f.Failures = append(f.Failures, failed...)
			}
		}
		f.Units[unit] = result
	}
	sort.Strings(f.Failures)
	sort.Strings(f.Unattributed)
	return f
}

// ParseUnitEvidence scans the bounded outer channel once. It never searches
// inside a Go Output string or treats diagnostic echoes as control frames.
func ParseUnitEvidence(log string) UnitFacts {
	m := NewUnitEvidence()
	if len(log) > maxUnitOuterLog {
		m.unattributed = append(m.unattributed, "unit evidence exceeds 64 MiB")
		return m.Finish()
	}
	for len(log) > 0 {
		line, rest, found := strings.Cut(log, "\n")
		log = rest
		if len(line) > maxUnitLine {
			m.unattributed = append(m.unattributed, "unit evidence line exceeds 1 MiB")
		} else if e, recognized, err := DecodeUnitEvent(line); recognized {
			if err != nil {
				m.seen = true
				m.visits++
				m.unattributed = append(m.unattributed, "malformed unit event: "+err.Error())
			} else {
				_ = m.Apply(e)
			}
		}
		if !found {
			break
		}
	}
	return m.Finish()
}

// RequiredUnitProof returns exact selected non-Go identities without positive
// complete proof. Any unattributed execution blocks all typed acceptance.
func RequiredUnitProof(facts UnitFacts, tests []string) []string {
	var missing []string
	for _, test := range tests {
		id, err := ParseIdentity(test)
		if err == nil && id.Kind() == KindGo {
			continue
		}
		if err != nil {
			missing = append(missing, test)
			continue
		}
		result, ok := facts.Units[id.Unit()]
		if len(facts.Unattributed) > 0 || !ok || !result.Passed || !result.Complete || (id.Kind() == KindLua && result.Cases[id.Case()] != "pass") {
			missing = append(missing, test)
		}
	}
	return missing
}

package flakes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// UnitRunner is operator-declared argv authority for one kind of execution.
// Unknown additive members survive a declaration read/modify/write.
type UnitRunner struct {
	Command []string `json:"command"`
	Results string   `json:"results,omitempty"`
	kind    IdentityKind
	extra   map[string]json.RawMessage
}

type UnitRunners map[IdentityKind]UnitRunner

func ParseUnitRunners(data []byte) (UnitRunners, error) {
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("flake_runners exceeds 1 MiB")
	}
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || len(raw) > 2 {
		return nil, fmt.Errorf("flake_runners must name script and/or lua runners")
	}
	runners := UnitRunners{}
	for key, value := range raw {
		kind := IdentityKind(key)
		if kind != KindScript && kind != KindLua {
			return nil, fmt.Errorf("unsupported flake runner kind %q", key)
		}
		runner, err := parseUnitRunner(kind, value)
		if err != nil {
			return nil, fmt.Errorf("%s runner: %w", kind, err)
		}
		runners[kind] = runner
	}
	return runners, nil
}

func parseUnitRunner(kind IdentityKind, data []byte) (UnitRunner, error) {
	runner := UnitRunner{kind: kind, extra: map[string]json.RawMessage{}}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return runner, fmt.Errorf("expected runner object")
	}
	if _, ok := raw["command"]; !ok {
		return runner, fmt.Errorf("missing command")
	}
	for key, value := range raw {
		var target any
		switch key {
		case "command":
			target = &runner.Command
		case "results":
			target = &runner.Results
		default:
			if strings.EqualFold(key, "command") || strings.EqualFold(key, "results") {
				return runner, fmt.Errorf("noncanonical runner field %q", key)
			}
			runner.extra[key] = bytes.Clone(value)
			continue
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return runner, fmt.Errorf("null %s", key)
		}
		if err := json.Unmarshal(value, target); err != nil {
			return runner, fmt.Errorf("%s: %w", key, err)
		}
	}
	if err := runner.validate(); err != nil {
		return runner, err
	}
	return runner, nil
}

func (runner UnitRunner) validate() error {
	if len(runner.Command) < 2 || len(runner.Command) > 32 {
		return fmt.Errorf("command must have 2..32 arguments and one literal {path} operand")
	}
	paths := 0
	for i, arg := range runner.Command {
		if arg == "" || len(arg) > 4096 || !validText(arg) {
			return fmt.Errorf("invalid command argument %d", i)
		}
		if arg == "{path}" && i > 0 {
			paths++
		} else if strings.Contains(arg, "{path}") {
			return fmt.Errorf("{path} must be a complete non-executable argument")
		}
	}
	if paths != 1 {
		return fmt.Errorf("command must contain exactly one literal {path} argument")
	}
	if runner.kind == KindLua && runner.Results != "plenary-events-v1" || runner.kind == KindScript && runner.Results != "" {
		return fmt.Errorf("lua requires results plenary-events-v1; script uses observed exit")
	}
	return nil
}

// Project substitutes a validated unit path as one literal argv operand.
func (runner UnitRunner) Project(unit Identity) ([]string, error) {
	if unit.Kind() != runner.kind || unit.Case() != "" {
		return nil, fmt.Errorf("runner kind does not match exact execution unit %q", unit.String())
	}
	if _, err := ParseUnit(unit.String()); err != nil {
		return nil, err
	}
	if err := runner.validate(); err != nil {
		return nil, err
	}
	argv := slices.Clone(runner.Command)
	for i := range argv {
		if argv[i] == "{path}" {
			argv[i] = unit.Path()
		}
	}
	return argv, nil
}

func (runners *UnitRunners) UnmarshalJSON(data []byte) error {
	parsed, err := ParseUnitRunners(data)
	if err == nil {
		*runners = parsed
	}
	return err
}

func (runner UnitRunner) MarshalJSON() ([]byte, error) {
	type known struct {
		Command []string `json:"command"`
		Results string   `json:"results,omitempty"`
	}
	data, err := json.Marshal(known{runner.Command, runner.Results})
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for key, value := range runner.extra {
		raw[key] = value
	}
	return json.Marshal(raw)
}

// ContractDigest binds the full declared runner contract, including opaque
// members that a newer producer may interpret. No argv is learned from a log.
func (runners UnitRunners) ContractDigest() string {
	data, err := json.Marshal(runners)
	if err != nil || len(runners) == 0 {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

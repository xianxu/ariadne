package fleet

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// PolicyDeclarationPath derives the repository-local declaration location from
// the fleet-policy vocabulary authority.
func PolicyDeclarationPath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(vocab.FleetPolicy().DeclarationPath))
}

func LoadPolicyFile(declarationPath string) PolicyCapability {
	raw, err := os.ReadFile(declarationPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return policyFailure(DiagnosticMissingPolicy, declarationPath, nil, "fleet policy declaration is missing")
		}
		return policyFailure(DiagnosticInvalidPolicy, declarationPath, nil, fmt.Sprintf("read fleet policy: %v", err))
	}
	return DecodePolicy(declarationPath, raw)
}

func DecodePolicy(declarationPath string, raw []byte) PolicyCapability {
	version, err := inspectPolicyJSON(raw)
	if err != nil {
		return policyFailure(DiagnosticInvalidPolicy, declarationPath, version, err.Error())
	}
	declaration, err := decodePolicyDeclaration(raw)
	if err != nil {
		return policyFailure(DiagnosticInvalidPolicy, declarationPath, version, fmt.Sprintf("decode fleet policy: %v", err))
	}
	if err := validatePolicy(declaration); err != nil {
		return policyFailure(DiagnosticInvalidPolicy, declarationPath, version, err.Error())
	}

	roots := make([]string, len(declaration.Admission.Key.Roots))
	copy(roots, declaration.Admission.Key.Roots)
	sort.Strings(roots)
	declaration.Admission.Key.Roots = roots
	canonical, err := json.Marshal(declaration)
	if err != nil {
		return policyFailure(DiagnosticInvalidPolicy, declarationPath, &declaration.Version, fmt.Sprintf("canonicalize fleet policy: %v", err))
	}
	sum := sha256.Sum256(canonical)
	value := &PolicyCapabilityValue{
		PolicyVersion: declaration.Version,
		PolicyDigest:  hex.EncodeToString(sum[:]),
		KeyKind:       declaration.Admission.Key.Kind,
		Roots:         roots,
		Capacity:      declaration.Admission.Capacity,
	}
	if declaration.Admission.OnCapacity != nil {
		value.OnCapacity = *declaration.Admission.OnCapacity
	}
	return PolicyCapability{OK: true, Value: value}
}

type fleetPolicyWire struct {
	Version   json.RawMessage `json:"version"`
	Admission json.RawMessage `json:"admission"`
}

type admissionWire struct {
	Key        json.RawMessage `json:"key"`
	Capacity   json.RawMessage `json:"capacity"`
	OnCapacity json.RawMessage `json:"onCapacity"`
}

type keyWire struct {
	Kind  json.RawMessage `json:"kind"`
	Roots json.RawMessage `json:"roots"`
}

type capacityWire struct {
	Kind  json.RawMessage `json:"kind"`
	Limit json.RawMessage `json:"limit"`
}

func decodePolicyDeclaration(raw []byte) (*fleetPolicyDeclaration, error) {
	var top fleetPolicyWire
	if err := strictUnmarshal(raw, &top); err != nil {
		return nil, err
	}
	var declaration fleetPolicyDeclaration
	if err := decodeRequired(top.Version, "version", &declaration.Version); err != nil {
		return nil, err
	}

	var admission admissionWire
	if err := decodeRequired(top.Admission, "admission", &admission); err != nil {
		return nil, err
	}
	var key keyWire
	if err := decodeRequired(admission.Key, "admission.key", &key); err != nil {
		return nil, err
	}
	if err := decodeRequired(key.Kind, "admission.key.kind", &declaration.Admission.Key.Kind); err != nil {
		return nil, err
	}
	if err := decodeRequired(key.Roots, "admission.key.roots", &declaration.Admission.Key.Roots); err != nil {
		return nil, err
	}

	var capacity capacityWire
	if err := decodeRequired(admission.Capacity, "admission.capacity", &capacity); err != nil {
		return nil, err
	}
	if err := decodeRequired(capacity.Kind, "admission.capacity.kind", &declaration.Admission.Capacity.Kind); err != nil {
		return nil, err
	}
	if len(capacity.Limit) != 0 {
		var limit int
		if err := decodeRequired(capacity.Limit, "admission.capacity.limit", &limit); err != nil {
			return nil, err
		}
		declaration.Admission.Capacity.Limit = &limit
	}
	if len(admission.OnCapacity) != 0 {
		var action string
		if err := decodeRequired(admission.OnCapacity, "admission.onCapacity", &action); err != nil {
			return nil, err
		}
		declaration.Admission.OnCapacity = &action
	}
	return &declaration, nil
}

func decodeRequired(raw json.RawMessage, field string, target any) error {
	if len(raw) == 0 {
		return fmt.Errorf("missing required field %s", field)
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("required field %s must not be null", field)
	}
	if err := strictUnmarshal(raw, target); err != nil {
		return fmt.Errorf("field %s: %w", field, err)
	}
	return nil
}

func validatePolicy(declaration *fleetPolicyDeclaration) error {
	model := vocab.FleetPolicy()
	if !model.SupportsVersion(declaration.Version) {
		return fmt.Errorf("unsupported fleet policy version %d", declaration.Version)
	}
	key := declaration.Admission.Key
	if !model.IsKeyKind(key.Kind) {
		return fmt.Errorf("unknown admission key kind %q", key.Kind)
	}
	switch key.Kind {
	case "repo", "worktree":
		if len(key.Roots) != 0 {
			return fmt.Errorf("admission key kind %q requires roots to be empty", key.Kind)
		}
	case "declared-root":
		if err := validateRootRules(key.Roots); err != nil {
			return err
		}
	}

	capacity := declaration.Admission.Capacity
	if !model.IsCapacityKind(capacity.Kind) {
		return fmt.Errorf("unknown capacity kind %q", capacity.Kind)
	}
	switch capacity.Kind {
	case "bounded":
		if capacity.Limit == nil || *capacity.Limit <= 0 {
			return errors.New("bounded capacity requires a positive limit")
		}
		if declaration.Admission.OnCapacity == nil || !model.IsAction(*declaration.Admission.OnCapacity) {
			return errors.New("bounded capacity requires a supported onCapacity action")
		}
	case "unbounded":
		if capacity.Limit != nil {
			return errors.New("unbounded capacity forbids limit")
		}
		if declaration.Admission.OnCapacity != nil {
			return errors.New("unbounded capacity forbids onCapacity")
		}
	}
	return nil
}

func validateRootRules(rules []string) error {
	if len(rules) == 0 {
		return errors.New("declared-root admission requires at least one root rule")
	}
	prefixes := make([]string, 0, len(rules))
	for _, rule := range rules {
		if !declaredRootRulePattern.MatchString(rule) {
			return fmt.Errorf("invalid declared-root rule %q", rule)
		}
		prefix := strings.TrimSuffix(rule, "/*")
		prefixes = append(prefixes, prefix)
	}
	for i := range prefixes {
		for j := i + 1; j < len(prefixes); j++ {
			if prefixes[i] == prefixes[j] || strings.HasPrefix(prefixes[i]+"/", prefixes[j]+"/") || strings.HasPrefix(prefixes[j]+"/", prefixes[i]+"/") {
				return fmt.Errorf("overlapping declared-root rules %q and %q", rules[i], rules[j])
			}
		}
	}
	return nil
}

var declaredRootRulePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]*(?:/[A-Za-z0-9][A-Za-z0-9._ -]*)*/\*$`)

func policyFailure(code, declarationPath string, version *int, message string) PolicyCapability {
	return PolicyCapability{Diagnostic: &PolicyDiagnostic{Code: code, Message: message, Path: declarationPath, PolicyVersion: version}}
}

func requireJSONEOF(dec *json.Decoder) error {
	var extra any
	err := dec.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("decode fleet policy: trailing JSON value")
	}
	return fmt.Errorf("decode fleet policy: %v", err)
}

func inspectPolicyJSON(raw []byte) (*int, error) {
	version := extractTopLevelPolicyVersion(raw)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := scanJSONNoDuplicateKeys(dec); err != nil {
		return version, fmt.Errorf("decode fleet policy: %v", err)
	}
	if err := requireJSONEOF(dec); err != nil {
		return version, err
	}
	return version, nil
}

// extractTopLevelPolicyVersion is deliberately independent of duplicate-key
// validation. A duplicate elsewhere in the object must not make a unique,
// well-formed schema version disappear merely because it appeared later.
func extractTopLevelPolicyVersion(raw []byte) *int {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return nil
	}
	var version *int
	seenVersion := false
	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return version
		}
		key, ok := keyToken.(string)
		if !ok {
			return version
		}
		if key == "version" {
			if seenVersion {
				return nil
			}
			seenVersion = true
		}
		valueToken, err := dec.Token()
		if err != nil {
			return version
		}
		if key == "version" {
			if number, ok := valueToken.(json.Number); ok {
				if value, err := strconv.Atoi(number.String()); err == nil {
					version = &value
				}
			}
		}
		if err := consumeJSONToken(dec, valueToken); err != nil {
			return version
		}
	}
	return version
}

func consumeJSONToken(dec *json.Decoder, token json.Token) error {
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		for dec.More() {
			if _, err := dec.Token(); err != nil {
				return err
			}
			value, err := dec.Token()
			if err != nil {
				return err
			}
			if err := consumeJSONToken(dec, value); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	case '[':
		for dec.More() {
			value, err := dec.Token()
			if err != nil {
				return err
			}
			if err := consumeJSONToken(dec, value); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	default:
		return fmt.Errorf("unexpected delimiter %q", delim)
	}
}

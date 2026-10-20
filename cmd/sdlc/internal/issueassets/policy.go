package issueassets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

type Mode string

const (
	ModeAllowed  Mode = "allowed"
	ModeTextOnly Mode = "text-only"
	PolicyPath        = "workshop/asset-policy.json"
)

type Policy struct {
	Version int  `json:"version"`
	Policy  Mode `json:"policy"`
}
type Decision struct {
	Allowed bool
	Source  string
	Reason  string
}

func ParsePolicy(raw []byte) (Policy, error) {
	fail := func() (Policy, error) {
		return Policy{}, fmt.Errorf("invalid %s: ask the operator to set exactly version 1 and policy allowed or text-only", PolicyPath)
	}
	// Ordinary Unmarshal accepts duplicate keys and other spellings of a field:
	// refuse duplicates at any depth, then require the exact names.
	if jsonstrict.RejectDuplicateKeys(raw) != nil {
		return fail()
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	tok, e := d.Token()
	if e != nil || tok != json.Delim('{') {
		return fail()
	}
	for d.More() {
		tok, e = d.Token()
		if e != nil {
			return fail()
		}
		key, ok := tok.(string)
		if !ok || (key != "version" && key != "policy") {
			return fail()
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return fail()
		}
	}
	if _, e = d.Token(); e != nil {
		return fail()
	}
	if _, e = d.Token(); e != io.EOF {
		return fail()
	}
	var p Policy
	if json.Unmarshal(raw, &p) != nil || p.Version != 1 || (p.Policy != ModeAllowed && p.Policy != ModeTextOnly) {
		return fail()
	}
	return p, nil
}
func ResolvePolicy(p *Policy, visibility string) Decision {
	if p != nil {
		if p.Version == 1 && p.Policy == ModeAllowed {
			return Decision{true, PolicyPath, "explicit operator policy allows assets"}
		}
		return Decision{false, PolicyPath, "repo policy is text-only: ask the operator to change " + PolicyPath}
	}
	if visibility == "private" {
		return Decision{true, "host visibility: private", "private repository permits assets"}
	}
	if visibility != "public" {
		visibility = "unknown"
	}
	return Decision{false, "host visibility: " + visibility, "repo policy is text-only (" + visibility + "): ask the operator to change " + PolicyPath}
}

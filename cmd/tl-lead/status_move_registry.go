package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// syncRegistration is the read-only identity evidence used before deleting
// source files. Lookup does not acknowledge edits or deliveries.
type syncRegistration struct{ ID, File, Owner string }

func (c couchSessionSync) LookupRegistration(file string) (*syncRegistration, error) {
	raw, err := c.run("lookup", file, "--json")
	if err != nil {
		return nil, err
	}
	return decodeMoveRegistration(raw)
}

// RegistrationSnapshot reads one coherent list without acknowledging edits or
// deliveries. Owner is Couch's persisted fallback owner; lookup resolves current frontmatter.
func (c couchSessionSync) RegistrationSnapshot() (map[string]*syncRegistration, error) {
	raw, err := c.run("list", "--json")
	if err != nil {
		return nil, err
	}
	return decodeMoveRegistrationSnapshot(raw)
}

func decodeMoveRegistrationSnapshot(raw []byte) (map[string]*syncRegistration, error) {
	fields, err := moveJSONFields(raw, "Code", "Registrations")
	if err != nil {
		return nil, err
	}
	var code string
	if err := json.Unmarshal(fields["Code"], &code); err != nil || code != "ok" {
		return nil, fmt.Errorf("couch list unavailable: Code is not ok")
	}
	list, ok := fields["Registrations"]
	list = bytes.TrimSpace(list)
	if !ok || len(list) == 0 || list[0] != '[' {
		return nil, fmt.Errorf("couch list invalid Registrations: expected array")
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(list, &entries); err != nil {
		return nil, fmt.Errorf("couch list invalid Registrations: %w", err)
	}
	out := make(map[string]*syncRegistration, len(entries))
	ids := make(map[string]bool, len(entries))
	for i, entry := range entries {
		r, err := decodeMoveRegistrationObject(entry)
		if err != nil {
			return nil, fmt.Errorf("couch list registration %d: %w", i, err)
		}
		if !filepath.IsAbs(r.File) || filepath.Clean(r.File) != r.File || strings.ContainsAny(r.File, "\x00\r\n") {
			return nil, fmt.Errorf("couch list invalid canonical File %q", r.File)
		}
		if _, exists := out[r.File]; exists {
			return nil, fmt.Errorf("couch list duplicate registration path %s", r.File)
		}
		if ids[r.ID] {
			return nil, fmt.Errorf("couch list duplicate registration identity %s", r.ID)
		}
		out[r.File], ids[r.ID] = r, true
	}
	return out, nil
}

func decodeMoveRegistration(raw []byte) (*syncRegistration, error) {
	fields, err := moveJSONFields(raw, "Code", "Registration")
	if err != nil {
		return nil, err
	}
	var code string
	if err := json.Unmarshal(fields["Code"], &code); err != nil || code != "ok" {
		return nil, fmt.Errorf("couch lookup unavailable: Code is not ok")
	}
	raw, ok := fields["Registration"]
	if !ok {
		return nil, fmt.Errorf("couch lookup omitted Registration")
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	return decodeMoveRegistrationObject(raw)
}

func decodeMoveRegistrationObject(raw []byte) (*syncRegistration, error) {
	fields, err := moveJSONFields(raw, "ID", "File", "Owner")
	if err != nil {
		return nil, err
	}
	r := &syncRegistration{}
	for k, dst := range map[string]*string{"ID": &r.ID, "File": &r.File, "Owner": &r.Owner} {
		if err := json.Unmarshal(fields[k], dst); err != nil || *dst == "" {
			return nil, fmt.Errorf("couch registration invalid %s", k)
		}
	}
	return r, nil
}

// Decode only known members. Raw unknown values remain opaque, including
// numbers beyond float64. Duplicate known members cannot confer authority.
func moveJSONFields(raw []byte, known ...string) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, fmt.Errorf("expected JSON object")
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("invalid JSON key")
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, err
		}
		for _, k := range known {
			if strings.EqualFold(key, k) {
				if _, exists := fields[k]; exists {
					return nil, fmt.Errorf("duplicate %s", k)
				}
				fields[k] = value
			}
		}
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON data")
	}
	return fields, nil
}

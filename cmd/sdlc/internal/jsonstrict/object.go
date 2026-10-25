package jsonstrict

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
)

// A persisted record written by one sdlc and read by another is a boundary (ARCH-SECURE), and each rewrite must carry a
// later sdlc's members forward (ARCH-EVOLVE). DecodeObject and EncodeWithExtra are the one codec for such records:
// strict where the schema is known, tolerant and preserving where it isn't (the landing line's sidecars, #616's
// missed-mapping ledger).

// DecodeObject parses one JSON object: no duplicate member, every known member (by exact name, never encoding/json's
// case-insensitive match) of its exact type and non-null, no trailing data. Unknown members are returned raw.
func DecodeObject(data []byte, fields map[string]any) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, errors.New("not one JSON object")
	}
	var extra map[string]json.RawMessage
	seen := map[string]bool{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key := tok.(string) // object keys are strings
		if seen[key] {
			return nil, fmt.Errorf("duplicate member %q", key)
		}
		seen[key] = true
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		field, known := fields[key]
		if !known {
			if extra == nil {
				extra = map[string]json.RawMessage{}
			}
			extra[key] = raw
			continue
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, fmt.Errorf("member %q is null", key)
		}
		if err := json.Unmarshal(raw, field); err != nil {
			return nil, fmt.Errorf("member %q: %w", key, err)
		}
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return nil, errors.New("object is not closed")
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing data after the object")
	}
	return extra, nil
}

// EncodeWithExtra appends extra's members, sorted, to base (an encoded object), skipping any that names a known
// member: extra never overrides.
func EncodeWithExtra(base []byte, extra map[string]json.RawMessage, known func(string) bool) []byte {
	keys := make([]string, 0, len(extra))
	for k := range extra {
		if !known(k) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return base
	}
	slices.Sort(keys)
	var b bytes.Buffer
	b.Write(base[:len(base)-1]) // drop the closing brace
	for _, k := range keys {
		name, _ := json.Marshal(k)
		b.WriteByte(',')
		b.Write(name)
		b.WriteByte(':')
		b.Write(extra[k])
	}
	b.WriteByte('}')
	return b.Bytes()
}

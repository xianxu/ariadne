package queuegate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

// inspectJSON uses Number tokens so unknown additive numeric values are opaque
// even when their magnitude exceeds float64. Duplicate keys at any depth remain
// malformed, and depth is bounded independently of the standard decoder.
func inspectJSON(data []byte) error {
	if len(data) == 0 || len(data) > MaxJSON || !utf8.Valid(data) {
		return errors.New("JSON is empty or exceeds 1 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 128 {
			return errors.New("JSON nesting exceeds 128 levels")
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				s, ok := key.(string)
				if !ok {
					return errors.New("invalid JSON member")
				}
				if seen[s] {
					return fmt.Errorf("duplicate member %q", s)
				}
				seen[s] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim('}') {
				return errors.New("unclosed JSON object")
			}
		case '[':
			for d.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return errors.New("unclosed JSON array")
			}
		default:
			return errors.New("unexpected JSON delimiter")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	return nil
}

// StrictObject is the shared structural boundary for persisted gate metadata
// and read-only legacy owner records. It preserves opaque member values while
// rejecting duplicate keys at every depth, malformed JSON and bounded-input
// violations. Schema-specific required fields and types belong to the caller.
func StrictObject(data []byte) (map[string]json.RawMessage, error) {
	if err := inspectJSON(data); err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return nil, errors.New("expected JSON object")
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func decodeObject(data []byte, fields map[string]any, optional map[string]bool) (map[string]json.RawMessage, error) {
	raw, err := StrictObject(data)
	if err != nil {
		return nil, err
	}
	extra := map[string]json.RawMessage{}
	for key, v := range raw {
		dst, known := fields[key]
		if !known {
			for name := range fields {
				if strings.EqualFold(name, key) {
					return nil, fmt.Errorf("noncanonical member %q", key)
				}
			}
			extra[key] = append(json.RawMessage(nil), v...)
			continue
		}
		if bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("member %q is null", key)
		}
		if err := json.Unmarshal(v, dst); err != nil {
			return nil, fmt.Errorf("member %q: %w", key, err)
		}
	}
	for key := range fields {
		if _, ok := raw[key]; !ok && !optional[key] {
			return nil, fmt.Errorf("missing member %q", key)
		}
	}
	return extra, nil
}

func encodeObject(fields map[string]json.RawMessage, extra map[string]json.RawMessage) ([]byte, error) {
	for key, v := range extra {
		if !utf8.ValidString(key) {
			return nil, errors.New("extra member name has invalid UTF-8")
		}
		for name := range fields {
			if strings.EqualFold(name, key) {
				return nil, fmt.Errorf("extra member conflicts with %q", name)
			}
		}
		if err := inspectJSON(v); err != nil {
			return nil, fmt.Errorf("extra member %q: %w", key, err)
		}
	}
	keys := make([]string, 0, len(fields)+len(extra))
	for k := range fields {
		keys = append(keys, k)
	}
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		key, _ := json.Marshal(k)
		b.Write(key)
		b.WriteByte(':')
		v, ok := fields[k]
		if !ok {
			v = extra[k]
		}
		b.Write(v)
	}
	b.WriteByte('}')
	if b.Len() > MaxJSON {
		return nil, errors.New("encoded JSON exceeds 1 MiB")
	}
	return b.Bytes(), nil
}

func rawValue(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

func DecodeControl(data []byte) (Control, error) {
	var c Control
	var owners []json.RawMessage
	extra, err := decodeObject(data, map[string]any{"schema_version": &c.SchemaVersion, "pause_owners": &owners}, nil)
	if err != nil {
		return c, err
	}
	c.Extra = extra
	c.PauseOwners = make([]PauseOwner, 0, len(owners))
	if len(owners) > MaxOwners {
		return c, errors.New("pause owner capacity exceeded")
	}
	for _, raw := range owners {
		var p PauseOwner
		p.Extra, err = decodeObject(raw, map[string]any{"owner": &p.Owner, "reason": &p.Reason, "cancel_running": &p.CancelRunning}, nil)
		if err != nil {
			return c, err
		}
		c.PauseOwners = append(c.PauseOwners, p)
	}
	return c, validateControl(c)
}

func EncodeControl(c Control) ([]byte, error) {
	if err := validateControl(c); err != nil {
		return nil, err
	}
	var owners bytes.Buffer
	owners.WriteByte('[')
	for i, p := range c.PauseOwners {
		b, err := encodeObject(map[string]json.RawMessage{"owner": rawValue(p.Owner), "reason": rawValue(p.Reason), "cancel_running": rawValue(p.CancelRunning)}, p.Extra)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			owners.WriteByte(',')
		}
		owners.Write(b)
	}
	owners.WriteByte(']')
	return encodeObject(map[string]json.RawMessage{"schema_version": rawValue(c.SchemaVersion), "pause_owners": owners.Bytes()}, c.Extra)
}

func operationFields(o *Operation) map[string]any {
	return map[string]any{"schema_version": &o.SchemaVersion, "id": &o.ID, "workspace": &o.Workspace, "head": &o.Head, "kinds": &o.Kinds, "phase": &o.Phase, "pid": &o.PID, "pid_started": &o.PIDStarted, "managed": &o.Managed, "started_at": &o.StartedAt, "boot_id": &o.BootID, "root": &o.Root, "detail": &o.Detail}
}

func DecodeOperation(data []byte) (Operation, error) {
	var o Operation
	var err error
	o.Extra, err = decodeObject(data, operationFields(&o), map[string]bool{"head": true})
	if err != nil {
		return o, err
	}
	return o, validateOperation(o)
}

func EncodeOperation(o Operation) ([]byte, error) {
	if err := validateOperation(o); err != nil {
		return nil, err
	}
	// Standard encoding formats the known timestamp; opaque additions are
	// appended separately so their original values are not normalized.
	base, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(base, &fields); err != nil {
		return nil, err
	}
	// Optional known fields still reserve their names against future extras.
	if _, ok := fields["head"]; !ok {
		fields["head"] = rawValue(o.Head)
	}
	return encodeObject(fields, o.Extra)
}

// Package qodertranscript projects Qoder records into bounded, content-free
// evidence shared by transcript discovery, usage accounting and agent identity.
// It performs no IO and does not interpret native usage amounts.
package qodertranscript

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"
)

const (
	MaxRecordBytes  = 4 << 20
	MaxStringBytes  = 4096
	MaxConfigEpochs = 4096
)

var (
	ErrInvalidRecord = errors.New("invalid qoder record")
	ErrLimit         = errors.New("qoder record limit exceeded")
	ErrIncomplete    = errors.New("incomplete qoder session metadata")
	ErrConflict      = errors.New("conflicting qoder session evidence")
)

// Record contains only metadata, never message content or message DAG ancestry.
type Record struct {
	Type            string
	SessionID       string
	Timestamp       time.Time
	Directories     []string
	Model           string
	ReasoningEffort *string
	Message         *Message
}

type Message struct {
	ID         string
	Model      string
	StopReason string
	// Usage retains the exact reported JSON values. The accounting consumer owns
	// numeric and billing validation; nil distinguishes absent usage evidence.
	Usage map[string]json.RawMessage
}

// object rejects duplicate keys in an interpreted object. Diagnostics never
// contain payload bytes, JSON decoder errors, or user-controlled field names.
func object(b []byte) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	tok, e := d.Token()
	if e != nil || tok != json.Delim('{') {
		return nil, ErrInvalidRecord
	}
	m := make(map[string]json.RawMessage)
	for d.More() {
		tok, e = d.Token()
		if e != nil {
			return nil, ErrInvalidRecord
		}
		k, ok := tok.(string)
		if !ok {
			return nil, ErrInvalidRecord
		}
		if _, ok = m[k]; ok {
			return nil, ErrInvalidRecord
		}
		var v json.RawMessage
		if d.Decode(&v) != nil {
			return nil, ErrInvalidRecord
		}
		m[k] = v
	}
	if _, e = d.Token(); e != nil {
		return nil, ErrInvalidRecord
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, ErrInvalidRecord
	}
	return m, nil
}
func field(m map[string]json.RawMessage, k string) (string, error) {
	b, ok := m[k]
	if !ok {
		return "", nil
	}
	var s string
	if bytes.Equal(b, []byte("null")) || json.Unmarshal(b, &s) != nil {
		return "", ErrInvalidRecord
	}
	if len(s) > MaxStringBytes {
		return "", ErrLimit
	}
	return s, nil
}

// ParseRecord whitelists supported record fields, rejecting malformed typed
// metadata while leaving usage-number semantics to the accounting consumer.
func ParseRecord(b []byte) (Record, error) {
	var r Record
	if len(b) > MaxRecordBytes {
		return r, ErrLimit
	}
	m, e := object(b)
	if e != nil {
		return r, e
	}
	if r.Type, e = field(m, "type"); e != nil {
		return Record{}, e
	}
	switch r.Type {
	case "workspace-directories", "runtime-config", "user", "assistant":
	default:
		return r, nil
	}
	if r.SessionID, e = field(m, "sessionId"); e != nil {
		return Record{}, e
	}
	if r.SessionID == "" {
		return Record{}, ErrInvalidRecord
	}
	if r.Type == "workspace-directories" {
		if raw, ok := m["directories"]; ok {
			if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &r.Directories) != nil {
				return Record{}, ErrInvalidRecord
			}
			if len(r.Directories) > MaxStringBytes {
				return Record{}, ErrLimit
			}
			for _, v := range r.Directories {
				if len(v) > MaxStringBytes {
					return Record{}, ErrLimit
				}
			}
		}
		return r, nil
	}
	if r.Type == "runtime-config" {
		var ms int64
		raw, ok := m["timestamp"]
		if !ok || bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &ms) != nil || ms <= 0 || ms > 253402300799999 {
			return Record{}, ErrInvalidRecord
		}
		r.Timestamp = time.UnixMilli(ms).UTC()
		if r.Model, e = field(m, "model"); e != nil {
			return Record{}, e
		}
		if raw, ok := m["reasoningEffort"]; ok && !bytes.Equal(raw, []byte("null")) {
			v, e := field(m, "reasoningEffort")
			if e != nil {
				return Record{}, e
			}
			r.ReasoningEffort = &v
		}
		return r, nil
	}
	ts, e := field(m, "timestamp")
	if e != nil {
		return Record{}, e
	}
	r.Timestamp, e = time.Parse(time.RFC3339Nano, ts)
	if e != nil || r.Timestamp.IsZero() || r.Timestamp.Year() < 1 {
		return Record{}, ErrInvalidRecord
	}
	r.Timestamp = r.Timestamp.UTC()
	if r.Type == "assistant" {
		if raw, ok := m["message"]; ok && !bytes.Equal(raw, []byte("null")) {
			msg, e := object(raw)
			if e != nil {
				return Record{}, e
			}
			v := new(Message)
			if v.ID, e = field(msg, "id"); e != nil {
				return Record{}, e
			}
			if v.Model, e = field(msg, "model"); e != nil {
				return Record{}, e
			}
			if raw, ok := msg["stop_reason"]; ok && !bytes.Equal(raw, []byte("null")) {
				if v.StopReason, e = field(msg, "stop_reason"); e != nil {
					return Record{}, e
				}
			}
			if raw, ok := msg["usage"]; ok && !bytes.Equal(raw, []byte("null")) {
				u, e := object(raw)
				if e != nil {
					return Record{}, e
				}
				v.Usage = make(map[string]json.RawMessage)
				for _, k := range []string{"credits", "original_credits", "billable", "input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"} {
					if raw, ok := u[k]; ok {
						if len(raw) > MaxStringBytes {
							return Record{}, ErrLimit
						}
						if !bytes.Equal(raw, []byte("null")) {
							if k == "billable" {
								if !bytes.Equal(raw, []byte("true")) && !bytes.Equal(raw, []byte("false")) {
									return Record{}, ErrInvalidRecord
								}
							} else if len(raw) == 0 || (raw[0] != '-' && (raw[0] < '0' || raw[0] > '9')) {
								return Record{}, ErrInvalidRecord
							}
						}
						v.Usage[k] = append(json.RawMessage(nil), raw...)
					}
				}
			}
			r.Message = v
		}
	}
	return r, nil
}

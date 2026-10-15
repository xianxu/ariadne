package transcripts

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CodexExecutingID reads the FIRST session_meta payload's executing identity.
// Callers must never replace that first result (including an error) with later
// copied metadata. session_id is intentionally ignored: it may name an ancestor.
func CodexExecutingID(payload json.RawMessage) (string, error) {
	var meta struct {
		ID       string `json:"id"`
		ThreadID string `json:"thread_id"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return "", fmt.Errorf("invalid Codex executing identity: %w", err)
	}
	if meta.ID != "" && meta.ThreadID != "" && meta.ID != meta.ThreadID {
		return "", fmt.Errorf("conflicting Codex executing identity aliases")
	}
	id := meta.ID
	if id == "" {
		id = meta.ThreadID
	}
	if id == "" || strings.TrimSpace(id) != id || strings.ContainsAny(id, "\r\n\x00") {
		return "", fmt.Errorf("missing or malformed Codex executing identity")
	}
	return id, nil
}

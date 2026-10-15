package agentid

import (
	"bytes"
	"encoding/json"
)

// claudeFromTail reads the model and effort of the latest assistant record in
// the tail of a Claude Code session transcript. The tail is input this process
// did not write (ARCH-SECURE): a cut first line or a record of an unknown
// shape is skipped, and nothing found means empty parts, never a guess.
func claudeFromTail(tail []byte) (model, effort string) {
	var rec struct {
		Type    string `json:"type"`
		Effort  string `json:"effort"`
		Message struct {
			Model string `json:"model"`
		} `json:"message"`
	}
	lastRecord(tail, func(line []byte) bool {
		rec.Type, rec.Effort, rec.Message.Model = "", "", ""
		if json.Unmarshal(line, &rec) != nil || rec.Type != "assistant" || rec.Message.Model == "" {
			return false
		}
		model, effort = rec.Message.Model, rec.Effort
		return true
	})
	return model, effort
}

// codexFromTail reads the model and effort of the latest turn_context record
// in the tail of a Codex rollout transcript.
func codexFromTail(tail []byte) (model, effort string) {
	var rec struct {
		Type    string `json:"type"`
		Payload struct {
			Model  string `json:"model"`
			Effort string `json:"effort"`
		} `json:"payload"`
	}
	lastRecord(tail, func(line []byte) bool {
		rec.Type, rec.Payload.Model, rec.Payload.Effort = "", "", ""
		if json.Unmarshal(line, &rec) != nil || rec.Type != "turn_context" || rec.Payload.Model == "" {
			return false
		}
		model, effort = rec.Payload.Model, rec.Payload.Effort
		return true
	})
	return model, effort
}

// lastRecord offers the tail's lines to take, newest first, until one is taken.
func lastRecord(tail []byte, take func([]byte) bool) {
	lines := bytes.Split(tail, []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		if line := bytes.TrimSpace(lines[i]); len(line) > 0 && take(line) {
			return
		}
	}
}

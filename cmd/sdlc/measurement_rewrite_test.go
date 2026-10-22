package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

func TestCorrectionJSONPreservesMeasurementExtensionsOnly(t *testing.T) {

	type envelope struct {
		Inputs gatestate.MeasurementInputs `json:"inputs"`
		Flag   bool                        `json:"flag"`
	}
	raw := `{"inputs":{"version":1,"method":"legacy-v1","repository":"github.com/xianxu/ariadne","code_head":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","since":"2026-10-01T01:00:00Z","until":"2026-10-01T02:00:00Z","gap_minutes":15,"commit_weight":1,"include_assistant":true,"future_member":{"number":9007199254740993}},"flag":true}`
	var got envelope
	if err := strictCorrectionJSON([]byte(raw), &got); err != nil {
		t.Fatalf("valid measurement extension refused: %v", err)
	}
	encoded, err := json.Marshal(got)
	if err != nil || !strings.Contains(string(encoded), `"future_member":{"number":9007199254740993}`) {
		t.Fatalf("accepted measurement extension lost: %s %v", encoded, err)
	}
	for _, invalid := range []string{
		strings.Replace(raw, `"flag":true`, `"flag":true,"unknown_outer":1`, 1),
		strings.Replace(raw, `"flag":true`, `"flag":null`, 1),
		strings.Replace(raw, `"gap_minutes":15`, `"gap_minutes":"15"`, 1),
		strings.Replace(raw, `"version":1`, `"version":1,"VERSION":1`, 1),
	} {
		if err := strictCorrectionJSON([]byte(invalid), &got); err == nil {
			t.Fatalf("weakened correction envelope or known measurement validation: %s", invalid)
		}
	}
}

func TestGateLedgerRewritePreservesUnknownMeasurementMembers(t *testing.T) {
	dir := t.TempDir()
	issue := "000544-legacy.md"
	raw := `---
gate: boundary-review
issue: 544
id_prefix: BR
rounds:
  - n: 1
    actual_inputs:
      version: 2
      method: active-time-v3-engagement-legacy-prefix-v1
      repository: github.com/xianxu/ariadne
      code_head: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
      since: "2026-10-01T01:00:00Z"
      until: "2026-10-01T02:00:00Z"
      tracker_head: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
      legacy_cutover: cccccccccccccccccccccccccccccccccccccccc
      engagements:
        - start: "2026-10-01T01:30:00Z"
          end: "2026-10-01T02:00:00Z"
          began: "2026-10-01T01:30:00Z"
          roots: [/repo]
      legacy_prefix:
        version: 1
        method: legacy-v1
        repository: github.com/xianxu/ariadne
        code_head: dddddddddddddddddddddddddddddddddddddddd
        since: "2026-10-01T01:00:00Z"
        until: "2026-10-01T01:30:00Z"
        gap_minutes: 15
        commit_weight: 1
        future_prefix: retained
        sources:
          - harness: codex
            id: prefix
            cwd: /repo
            digest: eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
            started: "2026-10-01T01:00:00Z"
            bytes: 10
            future_source: retained
            windows:
              - start: "2026-10-01T01:00:00Z"
                end: "2026-10-01T01:30:00Z"
                future_window: retained
      gap_minutes: 15
      commit_weight: 1
      include_assistant: true
      future_member:
        token: preserved
---
`
	path := planGateKind.path(dir, issue)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	l, err := readGateLedger(planGateKind, dir, issue, 544)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeGateLedger(planGateKind, dir, issue, l, "ariadne"); err != nil {
		t.Fatal(err)
	}
	again, err := readGateLedger(planGateKind, dir, issue, 544)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(again.Rounds[0].ActualInputs)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	root, ok := fields["future_member"].(map[string]any)
	if !ok || root["token"] != "preserved" {
		t.Fatalf("root extension lost: %s", data)
	}
	prefix := fields["legacy_prefix"].(map[string]any)
	source := prefix["sources"].([]any)[0].(map[string]any)
	window := source["windows"].([]any)[0].(map[string]any)
	for name, value := range map[string]any{"prefix": prefix["future_prefix"], "source": source["future_source"], "window": window["future_window"]} {
		if value != "retained" {
			t.Fatalf("%s extension lost: %s", name, data)
		}
	}
}

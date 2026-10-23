package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

// Fields follow a native --actors / --peek --json --transcripts capture from
// this development slot; paths/IDs below are synthetic conformance fixtures.
func benchCouchFixture(t *testing.T) (benchCouchDeps, map[string]any, map[string]any) {
	t.Helper()
	root := t.TempDir()
	binding := map[string]any{"Slot": "bench:0", "Repository": filepath.Join(root, ".git"), "Scope": "scope", "Tag": "tag", "Session": "terminal", "Nonce": "generation", "Agent": "codex", "PID": 123, "Start": "start-token"}
	actors := map[string]any{"Code": "ok", "Error": "", "Actors": []any{map[string]any{"Binding": binding, "Known": true, "Supported": true}}}
	peek := map[string]any{"ref": "bench:0", "agent": "codex", "working_path": root, "native_session": "native", "session": map[string]any{"name": "terminal", "scope_key": "scope", "tag": "tag", "start_nonce": "generation"}, "transcripts": []string{filepath.Join(t.TempDir(), "native.jsonl")}}
	d := benchCouchDeps{Command: func(_ context.Context, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "--actors" {
			return json.Marshal(actors)
		}
		if len(args) > 0 && args[0] == "--peek" {
			return json.Marshal(peek)
		}
		return nil, errors.New("unexpected runtime effect")
	}}
	return d, actors, peek
}

func TestBenchCouchProbeBindsWholeNativeIncarnation(t *testing.T) {
	d, _, peek := benchCouchFixture(t)
	profile := bench.AgentProfile{Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}
	a, err := probeBenchActor(context.Background(), d, "bench:0", peek["working_path"].(string), profile)
	if err != nil || a.Identity.Session != "terminal" || a.Identity.Generation != "generation" || a.NativeSession != "native" || len(a.Transcripts) != 1 || a.PID != 123 {
		t.Fatalf("native binding %+v %v", a, err)
	}
	// The model is labelled provisioning evidence, never alleged to have been
	// read from Couch: source logs are cross-checked during collection.
	if a.Identity.Model != profile.Model {
		t.Fatal("provisioned profile lost")
	}
	peek["session"].(map[string]any)["start_nonce"] = "replacement"
	if _, err := probeBenchActor(context.Background(), d, "bench:0", a.Identity.Root, profile); err == nil {
		t.Fatal("replacement incarnation admitted")
	}
}

func TestBenchCouchProbeRefusesIncompleteAndCrossRootEvidence(t *testing.T) {
	for _, kind := range []string{"root", "agent", "session", "transcripts", "native", "unknown", "unsupported", "repository"} {
		t.Run(kind, func(t *testing.T) {
			d, actors, peek := benchCouchFixture(t)
			root := peek["working_path"].(string)
			actor := actors["Actors"].([]any)[0].(map[string]any)
			switch kind {
			case "root":
				peek["working_path"] = t.TempDir()
			case "agent":
				peek["agent"] = "claude"
			case "session":
				peek["session"].(map[string]any)["name"] = "other"
			case "transcripts":
				peek["transcripts"] = []string{}
			case "native":
				peek["native_session"] = ""
			case "unknown":
				actor["Known"] = false
			case "unsupported":
				actor["Supported"] = false
			case "repository":
				actor["Binding"].(map[string]any)["Repository"] = t.TempDir()
			}
			if _, err := probeBenchActor(context.Background(), d, "bench:0", root, bench.AgentProfile{Agent: "codex", Model: "m", Effort: "medium"}); err == nil {
				t.Fatal("unproved native runtime admitted")
			}
		})
	}
}

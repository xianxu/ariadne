package bench

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func validTask() Task {
	p := strings.Repeat("a", 40)
	l := strings.Repeat("b", 40)
	return Task{Version: 1, CorpusVersion: "v1", Repo: "parley.nvim", ID: "000403", Type: "bugfix", Repeat: 1, SourceRepo: "/private/tmp/source", ParentOID: p, LandingOID: l, OriginalIssueOID: strings.Repeat("c", 40), Spec: "Fix the error", DoneWhen: "Regression passes", OriginalHours: 0.14, Worker: AgentProfile{Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}, TL: AgentProfile{Agent: "claude", Model: "claude-opus-5-5", Effort: "high"}, Reviewer: AgentProfile{Agent: "claude", Model: "claude-opus-5-5", Effort: "high"}, Acceptance: AcceptanceSpec{Patch: "/private/tmp/observer/acceptance.patch", PatchSHA256: strings.Repeat("d", 64), Paths: []string{"tests/regression.lua"}, Argv: []string{"make", "test"}, ParentOID: p, LandingOID: l}}
}

func TestTaskValidAndPublicSeparation(t *testing.T) {
	task := validTask()
	if err := ValidateTask(task); err != nil {
		t.Fatal(err)
	}
	task.Unknown = map[string]json.RawMessage{"future_secret": json.RawMessage(`{"secret":"observer only"}`)}
	pub := ProjectPublic(task)
	raw, err := json.Marshal(pub)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{task.SourceRepo, task.ParentOID, task.LandingOID, task.OriginalIssueOID, "regression.lua", "observer only", "acceptance", "original_hours"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("public projection leaked %q: %s", secret, raw)
		}
	}
	if pub.Spec != task.Spec || pub.Worker.Model != task.Worker.Model || pub.ID != task.ID {
		t.Fatal("missing public identity/instructions")
	}
	task.Worker.Unknown = map[string]json.RawMessage{"private_future": json.RawMessage(`"secret"`)}
	raw, _ = json.Marshal(ProjectPublic(task))
	if strings.Contains(string(raw), "secret") {
		t.Fatal("nested observer extension leaked")
	}
}

func TestTaskValidationRefusesInvalidAuthority(t *testing.T) {
	cases := map[string]func(*Task){
		"version": func(x *Task) { x.Version = 2 }, "identity": func(x *Task) { x.ID = "../403" }, "corpus": func(x *Task) { x.CorpusVersion = "" }, "repeat": func(x *Task) { x.Repeat = 0 }, "relative source": func(x *Task) { x.SourceRepo = "relative" }, "unclean source": func(x *Task) { x.SourceRepo = "/tmp/../source" }, "short pin": func(x *Task) { x.ParentOID = "abcdef" }, "same landing": func(x *Task) { x.LandingOID = x.ParentOID; x.Acceptance.LandingOID = x.ParentOID }, "spec": func(x *Task) { x.Spec = " " }, "done": func(x *Task) { x.DoneWhen = "" }, "hours": func(x *Task) { x.OriginalHours = -1 }, "model": func(x *Task) { x.Worker.Model = "" }, "review effort": func(x *Task) { x.Reviewer.Effort = "" }, "acceptance pin": func(x *Task) { x.Acceptance.ParentOID = strings.Repeat("f", 40) }, "empty argv": func(x *Task) { x.Acceptance.Argv = nil }, "nul argv": func(x *Task) { x.Acceptance.Argv = []string{"make\x00"} }, "escape": func(x *Task) { x.Acceptance.Paths = []string{"../test.go"} }, "code path": func(x *Task) { x.Acceptance.Paths = []string{"src/main.lua"} }, "patch hash": func(x *Task) { x.Acceptance.PatchSHA256 = "short" }, "no patch": func(x *Task) { x.Acceptance.Patch = "" }, "duplicate paths": func(x *Task) { x.Acceptance.Paths = append(x.Acceptance.Paths, x.Acceptance.Paths[0]) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			task := validTask()
			mutate(&task)
			if ValidateTask(task) == nil {
				t.Fatal("invalid descriptor accepted")
			}
		})
	}
}

func TestTaskStrictJSONAndNestedFutureRoundTrip(t *testing.T) {
	task := validTask()
	task.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"large":1e400,"values":[1,2]}`)}
	task.Worker.Unknown = map[string]json.RawMessage{"future_profile": json.RawMessage(`{"nested":true}`)}
	task.Acceptance.Unknown = map[string]json.RawMessage{"future_acceptance": json.RawMessage(`{"nested":true}`)}
	raw, err := EncodeTask(task)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeTask(raw)
	if err != nil {
		t.Fatal(err)
	}
	again, err := EncodeTask(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(again) {
		t.Fatalf("future extensions changed:\n%s\n%s", raw, again)
	}
	base, _ := EncodeTask(validTask())
	bad := map[string]string{"duplicate": strings.Replace(string(base), `"version":1`, `"version":1,"version":1`, 1), "nested duplicate": strings.Replace(string(base), `"agent":"codex"`, `"agent":"codex","AGENT":"codex"`, 1), "wrong type": strings.Replace(string(base), `"repeat":1`, `"repeat":"1"`, 1), "missing": strings.Replace(string(base), `"version":1,`, "", 1), "null": strings.Replace(string(base), `"spec":"Fix the error"`, `"spec":null`, 1), "alias": strings.Replace(string(base), `"version":1`, `"VERSION":1`, 1), "trailing": string(base) + ` {}`, "null object": "null", "oversize": strings.Repeat(" ", TaskRecordLimit+1)}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeTask([]byte(raw)); err == nil {
				t.Fatal("invalid JSON accepted")
			}
		})
	}
	decoded.Worker.Unknown["model"] = json.RawMessage(`"override"`)
	if _, err := EncodeTask(decoded); err == nil {
		t.Fatal("unknown field overrides known authority")
	}
}

func TestTaskLiteralArgvAndSHA256(t *testing.T) {
	task := validTask()
	task.ParentOID = strings.Repeat("a", 64)
	task.LandingOID = strings.Repeat("b", 64)
	task.OriginalIssueOID = strings.Repeat("c", 64)
	task.Acceptance.ParentOID = task.ParentOID
	task.Acceptance.LandingOID = task.LandingOID
	task.Acceptance.Argv = []string{"make", "target with spaces", "$(literal)"}
	if err := ValidateTask(task); err != nil {
		t.Fatal(err)
	}
	raw, _ := EncodeTask(task)
	got, err := DecodeTask(raw)
	if err != nil || got.Acceptance.Argv[2] != "$(literal)" {
		t.Fatalf("literal argv changed: %+v %v", got.Acceptance.Argv, err)
	}
}

func TestTaskBoundsAndUnknownStrictness(t *testing.T) {
	task := validTask()
	task.Spec = strings.Repeat("x", TaskRecordLimit)
	if _, err := EncodeTask(task); err == nil {
		t.Fatal("aggregate oversized task accepted")
	}
	task = validTask()
	task.Acceptance.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"x":1,"x":2}`)}
	if _, err := EncodeTask(task); err == nil {
		t.Fatal("duplicate nested extension keys accepted")
	}
	task = validTask()
	raw, _ := EncodeTask(task)
	for _, replacement := range []string{`"worker":null`, `"worker":{"agent":"codex","model":"x"}`, `"worker":{"agent":"codex","model":null,"effort":"high"}`} {
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil {
			t.Fatal(err)
		}
		members["worker"] = json.RawMessage(strings.TrimPrefix(replacement, `"worker":`))
		bad, _ := json.Marshal(members)
		before := validTask()
		if err := json.Unmarshal(bad, &before); err == nil {
			t.Fatal("invalid nested profile accepted")
		}
		if before.ID != "000403" || before.Worker.Model != "gpt-6.1-sol" {
			t.Fatal("failed decode partially changed destination")
		}
	}
	bad := append([]byte(nil), raw...)
	bad = bytes.Replace(bad, []byte("Fix the error"), []byte{0xff}, 1)
	if _, err := DecodeTask(bad); err == nil {
		t.Fatal("invalid UTF8 silently repaired")
	}
}

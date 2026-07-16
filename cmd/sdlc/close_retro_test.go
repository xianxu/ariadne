package main

import "testing"

func TestShouldNudgeProjectRetro(t *testing.T) {
	doc := func(status, log string) string {
		return "---\ntype: project\nstatus: " + status + "\n---\n## Log\n" + log + "\n"
	}
	if !shouldNudgeProjectRetro(doc("executing", ""), "2026-07-20", false) {
		t.Error("executing project without retro was not nudged")
	}
	if !shouldNudgeProjectRetro(doc("'executing'", ""), "2026-07-20", false) {
		t.Error("single-quoted executing status was not YAML-decoded for nudge")
	}
	if !shouldNudgeProjectRetro(doc("paused", "### 2026-07-01 — retro"), "2026-07-20", false) {
		t.Error("paused stale project was not nudged")
	}
	if shouldNudgeProjectRetro(doc("ideation", ""), "2026-07-20", false) {
		t.Error("forming project was nudged")
	}
	if shouldNudgeProjectRetro("---\ntype: project\nstatus: executing\n---\n# legacy\n", "2026-07-20", false) {
		t.Error("legacy no-Log project was nudged")
	}
	if shouldNudgeProjectRetro(doc("executing", ""), "2026-07-20", true) {
		t.Error("--no-project did not suppress nudge")
	}
}

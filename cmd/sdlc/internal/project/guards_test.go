package project

import (
	"strings"
	"testing"
)

func guardDoc(t *testing.T, fm, prd, estimate, log string) *Doc {
	t.Helper()
	d, err := ParseDoc("---\ntype: project\nname: demo\nstatus: ideation\n" + fm + "---\n## PRD\n" + prd + "\n## Estimate\n" + estimate + "\n## Breakdown\n- [ ]\n## Log\n" + log + "\n")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestProjectGuards(t *testing.T) {
	guards := Guards()
	for _, name := range []string{"prd-present", "phase-a-estimate", "baseline-set", "reality-check", "issues-cover-prd", "retro-recorded", "fog-factor-recorded"} {
		if guards[name] == nil {
			t.Fatalf("guard %q is not registered", name)
		}
	}
	ctx := GuardCtx{Evidence: map[string]string{}, Today: "2026-07-16"}
	if err := guards["prd-present"](guardDoc(t, "", "\n", "", ""), ctx); err == nil {
		t.Error("empty PRD passed")
	}
	if err := guards["prd-present"](guardDoc(t, "", "\nA real requirement.\n", "", ""), ctx); err != nil {
		t.Errorf("prose PRD failed: %v", err)
	}
	if err := guards["phase-a-estimate"](guardDoc(t, "", "ok", "\n**phase-a:** 3.5h\n", ""), ctx); err != nil {
		t.Errorf("phase-a failed: %v", err)
	}
	if err := guards["phase-a-estimate"](guardDoc(t, "", "ok", "\n**phase-a:** TBD\n", ""), ctx); err == nil {
		t.Error("non-numeric phase-a passed")
	}
	if err := guards["baseline-set"](guardDoc(t, "deadline: 2026-09-01\nplanned_finish: 2026-08-20\n", "ok", "", ""), ctx); err != nil {
		t.Errorf("baseline failed: %v", err)
	}
	if err := guards["baseline-set"](guardDoc(t, "deadline: 2026-09-01\n", "ok", "", ""), ctx); err == nil {
		t.Error("partial baseline passed")
	}
	for _, name := range []string{"reality-check", "issues-cover-prd"} {
		if err := guards[name](guardDoc(t, "", "ok", "", ""), ctx); err == nil {
			t.Errorf("%s passed without evidence", name)
		}
		with := GuardCtx{Evidence: map[string]string{name: "checked"}}
		if err := guards[name](guardDoc(t, "", "ok", "", ""), with); err != nil {
			t.Errorf("%s rejected evidence: %v", name, err)
		}
	}
	if err := guards["retro-recorded"](guardDoc(t, "", "ok", "", "\n### 2026-07-16 — retro\nLearned.\n"), ctx); err != nil {
		t.Errorf("retro failed: %v", err)
	}
	if err := guards["retro-recorded"](guardDoc(t, "", "ok", "", "\n### 2026-07-16\nNot retro.\n"), ctx); err == nil {
		t.Error("non-retro heading passed")
	}
	if err := guards["fog-factor-recorded"](guardDoc(t, "", "ok", "", ""), ctx); err == nil || !strings.Contains(err.Error(), "project close") {
		t.Errorf("fog guard = %v", err)
	}
}

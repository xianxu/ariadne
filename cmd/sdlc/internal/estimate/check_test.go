package estimate

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, section string) Block {
	t.Helper()
	b, err := ParseBlock(section)
	if err != nil {
		t.Fatalf("ParseBlock: %v", err)
	}
	return b
}

func TestCheck_GreenReconciles(t *testing.T) {
	b := mustParse(t, greenSection)
	if fs := Check(b, 3.4); len(fs) != 0 {
		t.Errorf("expected no failures, got %v", fs)
	}
}

func TestCheck_TotalMismatch(t *testing.T) {
	b := mustParse(t, greenSection)
	b.Total = 5.0 // recomputed ~3.41, well outside tol
	fs := Check(b, 5.0)
	if !hasFailureContaining(fs, "recomputed") {
		t.Errorf("expected a total≠recomputed failure, got %v", fs)
	}
}

func TestCheck_EstimateHoursMismatch(t *testing.T) {
	b := mustParse(t, greenSection)
	fs := Check(b, 7.0) // block total 3.4 vs frontmatter 7.0
	if !hasFailureContaining(fs, "estimate_hours") {
		t.Errorf("expected an estimate_hours mismatch failure, got %v", fs)
	}
}

func TestCheck_UnknownPrimitive(t *testing.T) {
	section := fence + "estimate\n" +
		"model: estimate-logic-v2\n" +
		"item: not-a-primitive design=0.0 impl=0.2\n" +
		"total: 0.2\n" + fence + "\n"
	fs := Check(mustParse(t, section), 0.2)
	if !hasFailureContaining(fs, "unknown primitive") {
		t.Errorf("expected unknown-primitive failure, got %v", fs)
	}
}

func TestCheck_UnknownModel(t *testing.T) {
	section := fence + "estimate\n" +
		"model: vibes\n" +
		"item: atlas-docs design=0.0 impl=0.2\n" +
		"total: 0.2\n" + fence + "\n"
	fs := Check(mustParse(t, section), 0.2)
	if !hasFailureContaining(fs, "unknown model") {
		t.Errorf("expected unknown-model failure, got %v", fs)
	}
}

func TestCheck_WithinTolerance(t *testing.T) {
	// total 3.4 vs recomputed 3.41 vs estimate_hours 3.4 — all within tol.
	b := mustParse(t, greenSection)
	if fs := Check(b, 3.4); len(fs) != 0 {
		t.Errorf("rounding within tol should pass, got %v", fs)
	}
}

func hasFailureContaining(fs []Failure, substr string) bool {
	for _, f := range fs {
		if strings.Contains(f.Message, substr) {
			return true
		}
	}
	return false
}

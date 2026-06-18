package issue

import (
	"strings"
	"testing"
)

func TestEstimateSection(t *testing.T) {
	body := "# T\n\n## Spec\n\nstuff\n\n## Estimate\n\nprose\n\n```estimate\ntotal: 3.4\n```\n\n## Plan\n\n- [ ] x\n"
	got, ok := EstimateSection(body)
	if !ok {
		t.Fatal("expected an Estimate section")
	}
	if !strings.Contains(got, "```estimate") || !strings.Contains(got, "total: 3.4") {
		t.Errorf("Estimate body missing the fenced block: %q", got)
	}
	if strings.Contains(got, "## Plan") {
		t.Error("Estimate body bled into the next section")
	}
}

func TestEstimateSection_Absent(t *testing.T) {
	if _, ok := EstimateSection("# T\n\n## Spec\n\nstuff\n"); ok {
		t.Error("expected no Estimate section")
	}
}

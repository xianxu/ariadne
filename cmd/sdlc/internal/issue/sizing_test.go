package issue

import (
	"strings"
	"testing"
)

func TestComputeSizing_FromContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Sizing
	}{
		{
			name: "small flat plan",
			text: joinDoc(
				`---`,
				`id: 000001`,
				`estimate_hours: 0.75`,
				`related: [a.go, b.go]`,
				`---`,
				"# x",
				"## Spec",
				strings.Repeat("word ", 60),
				"## Plan",
				"- [ ] do A",
				"- [ ] do B",
				"- [x] do C",
			),
			want: Sizing{
				EstimateHours: 0.75,
				PlanItems:     3,
				Milestones:    0,
				SpecWords:     60,
				RelatedFiles:  2,
				Bucket:        BucketSmall,
			},
		},
		{
			name: "large multi-milestone",
			text: joinDoc(
				`---`,
				`id: 000018`,
				`estimate_hours: 12`,
				`related: [a, b, c, d, e]`,
				`---`,
				"# x",
				"## Spec",
				strings.Repeat("word ", 200),
				"## Plan",
				"- [ ] M1: design",
				"- [ ] M2: scaffold",
				"- [ ] M3: implement",
				"- [ ] M4: test",
				"- [ ] do unrelated thing",
			),
			want: Sizing{
				EstimateHours: 12,
				PlanItems:     5,
				Milestones:    4,
				SpecWords:     200,
				RelatedFiles:  5,
				Bucket:        BucketLarge,
			},
		},
		{
			name: "medium — estimate boundary 2h, no milestones",
			text: joinDoc(
				`---`,
				`id: 000010`,
				`estimate_hours: 2`,
				`---`,
				"# x",
				"## Spec",
				strings.Repeat("word ", 80),
				"## Plan",
				"- [ ] one",
				"- [ ] two",
				"- [ ] three",
			),
			want: Sizing{
				EstimateHours: 2,
				PlanItems:     3,
				Milestones:    0,
				SpecWords:     80,
				RelatedFiles:  0,
				Bucket:        BucketMedium,
			},
		},
		{
			name: "large by milestone count alone (3 milestones, small estimate)",
			text: joinDoc(
				`---`,
				`id: 000020`,
				`estimate_hours: 1`,
				`---`,
				"# x",
				"## Spec",
				strings.Repeat("word ", 80),
				"## Plan",
				"- [ ] M1: a",
				"- [ ] M2: b",
				"- [ ] M3: c",
			),
			want: Sizing{
				EstimateHours: 1,
				PlanItems:     3,
				Milestones:    3,
				SpecWords:     80,
				RelatedFiles:  0,
				Bucket:        BucketLarge,
			},
		},
		{
			name: "no frontmatter — zero defaults, bucket small",
			text: "# Just a title with no frontmatter.\n",
			want: Sizing{
				EstimateHours: 0,
				PlanItems:     0,
				Milestones:    0,
				SpecWords:     0,
				RelatedFiles:  0,
				Bucket:        BucketSmall,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeSizingFromContent(tt.text)
			if got != tt.want {
				t.Errorf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestSizing_Format_ContainsKeyFields(t *testing.T) {
	s := Sizing{
		EstimateHours: 0.75,
		PlanItems:     5,
		Milestones:    0,
		SpecWords:     130,
		RelatedFiles:  3,
		Bucket:        BucketSmall,
	}
	out := s.Format("000040", "Judge classifier: structured VERDICT line")
	for _, want := range []string{"000040", "Judge classifier", "0.75", "5", "130", "small"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format output missing %q:\n%s", want, out)
		}
	}
}

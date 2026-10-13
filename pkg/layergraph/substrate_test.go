package layergraph

import "testing"

func TestChooseSubstrate(t *testing.T) {
	const owner, primary = "/f/worktree/leaf/br", "/f/leaf"
	local, cand := "/f/worktree/leaf/base", "/f/base"
	cases := []struct {
		name, target, primary string
		disk                  map[string]Probe
		want                  Resolution
		asked                 []string
	}{
		{"local layer wins", "../base", primary, map[string]Probe{local: Layer, cand: Layer},
			Resolution{Path: local, Local: local, Probe: Layer}, []string{local}},
		{"invalid local is not bypassed", "../base", primary, map[string]Probe{local: NotLayer, cand: Layer},
			Resolution{Path: local, Local: local, Probe: NotLayer}, []string{local}},
		{"no anchor stays local", "../base", "", map[string]Probe{cand: Layer},
			Resolution{Path: local, Local: local, Probe: Absent}, []string{local}},
		{"absent local falls back", "../base", primary, map[string]Probe{cand: Layer},
			Resolution{Path: cand, Local: local, Via: primary, Probe: Layer}, []string{local, cand}},
		{"invalid fallback is reported", "../base", primary, map[string]Probe{cand: NotLayer},
			Resolution{Path: cand, Local: local, Via: primary, Probe: NotLayer}, []string{local, cand}},
		{"absent everywhere stays local", "../base", primary, nil,
			Resolution{Path: local, Local: local, Probe: Absent}, []string{local, cand}},
		{"absolute never falls back", "/abs/base", primary, map[string]Probe{"/f/abs/base": Layer},
			Resolution{Path: "/abs/base", Local: "/abs/base", Probe: Absent}, []string{"/abs/base"}},
		{"spaces are kept", "../my base", primary, map[string]Probe{"/f/my base": Layer},
			Resolution{Path: "/f/my base", Local: "/f/worktree/leaf/my base", Via: primary, Probe: Layer},
			[]string{"/f/worktree/leaf/my base", "/f/my base"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var asked []string
			got := ChooseSubstrate(owner, c.target, c.primary, func(p string) Probe {
				asked = append(asked, p)
				return c.disk[p]
			})
			if got != c.want {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
			if len(asked) != len(c.asked) {
				t.Fatalf("probed %v, want %v", asked, c.asked)
			}
			for i := range asked {
				if asked[i] != c.asked[i] {
					t.Errorf("probed %v, want %v", asked, c.asked)
				}
			}
		})
	}
}

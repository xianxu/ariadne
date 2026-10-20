package landingline

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// #475: today's measured costs pack as the operator's example: ariadne alone
// at :00 on both slots, pair and parley sharing :15; then the property: at
// every minute the summed weight of reserved suites fits the slots, whenever
// the costs can fit at all.
func TestPackOffsets(t *testing.T) {
	min := func(f float64) time.Duration { return time.Duration(f * float64(time.Minute)) }
	got, over := PackOffsets([]RepoCost{
		{"parley.nvim", 1, min(2.7)},
		{"pair", 1, min(12.5)},
		{"ariadne", 2, min(9.5)},
	}, time.Hour, 2)
	if len(over) != 0 {
		t.Errorf("today's costs fit, yet over %v", over)
	}
	want := map[string]time.Duration{"ariadne": 0, "pair": 15 * time.Minute, "parley.nvim": 15 * time.Minute}
	for repo, off := range want {
		if got[repo] != off {
			t.Errorf("%s at %v, want %v (all %v)", repo, got[repo], off, got)
		}
	}

	rng := rand.New(rand.NewSource(475))
	for i := 0; i < 500; i++ {
		var costs []RepoCost
		total := 0
		for r := 0; r < 1+rng.Intn(5); r++ {
			c := RepoCost{Repo: string(rune('a' + r)), Weight: 1 + rng.Intn(2), Duration: min(float64(1 + rng.Intn(15)))}
			span := int((c.Duration+PlanStep-1)/PlanStep)*5 + 5
			total += c.Weight * span
			costs = append(costs, c)
		}
		if total > 60 { // more than one slot-hour reserved: not always packable
			continue
		}
		offs, _ := PackOffsets(costs, time.Hour, 2)
		load := make([]int, 60)
		for _, c := range costs {
			span := int((c.Duration+PlanStep-1)/PlanStep)*5 + 5
			for m := 0; m < span; m++ {
				load[(int(offs[c.Repo]/time.Minute)+m)%60] += c.Weight
			}
		}
		for m, l := range load {
			if l > 2 {
				t.Fatalf("costs %+v offsets %v: %d slots at minute %d", costs, offs, l, m)
			}
		}
	}
}

// #475: the plan file decodes strictly, refuses what it can't use, and keeps
// unknown members, top-level and per repository, through a rewrite.
func TestMachinePlanCodec(t *testing.T) {
	raw := `{"version":1,"slots":2,"later":{"x":1},"repos":{"pair":{"offset_minutes":15,"basis":"p90 12.5m x1","note":"kept"}}}`
	p, err := DecodePlan([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if off, ok := p.Offset("pair", time.Hour); !ok || off != 15*time.Minute {
		t.Fatalf("pair offset %v %v", off, ok)
	}
	if _, ok := p.Offset("ariadne", time.Hour); ok {
		t.Fatal("an absent repo was placed")
	}
	if off, _ := p.Offset("pair", 10*time.Minute); off != 5*time.Minute {
		t.Errorf("offset not reduced modulo a shorter interval: %v", off)
	}
	n := p.WithOffsets(2, map[string]PlanEntry{"pair": {OffsetMinutes: 20, Basis: "new"}, "ariadne": {OffsetMinutes: 0, Basis: "p90 9.5m x2"}})
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"later"`, `"note":"kept"`, `"offset_minutes":20`, `"ariadne"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("rewrite lost %s:\n%s", want, data)
		}
	}
	back, err := DecodePlan(data)
	if err != nil || back.Repos["pair"].OffsetMinutes != 20 || back.Repos["pair"].extra["note"] == nil || back.extra["later"] == nil {
		t.Fatalf("round trip: %+v %v", back, err)
	}
	for _, bad := range []string{
		`[]`, `{"version":1,"slots":2} {}`, `{"version":2,"slots":2,"repos":{}}`, `{"version":1,"slots":0,"repos":{}}`,
		`{"version":1,"slots":2,"repos":{"pair":{"basis":"x"}}}`, `{"version":1,"slots":2,"repos":{"pair":{"offset_minutes":1440}}}`,
		`{"version":1,"slots":2,"repos":{"pair":{"offset_minutes":"15"}}}`,
		`{"version":1,"version":1,"slots":2,"repos":{}}`,
		`{"version":1,"slots":2,"repos":{"pair":{"offset_minutes":15},"pair":{"offset_minutes":20}}}`,
		`{"version":1,"slots":2,"repos":{"pair":{"offset_minutes":15,"offset_minutes":20}}}`,
	} {
		if _, err := DecodePlan([]byte(bad)); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

// #475: costs that cannot fit the slots are packed as well as possible and
// named, with their peak, never silently over the envelope.
func TestPackOffsetsNamesAnOverload(t *testing.T) {
	var costs []RepoCost
	for _, r := range []string{"a", "b", "c", "d", "e"} {
		costs = append(costs, RepoCost{Repo: r, Weight: 2, Duration: 20 * time.Minute})
	}
	offs, over := PackOffsets(costs, time.Hour, 2)
	if len(offs) != 5 || len(over) == 0 {
		t.Fatalf("offsets %v over %v: five 25-minute double-slot suites can't fit an hour of 2 slots", offs, over)
	}
	for r, peak := range over {
		if peak <= 2 {
			t.Errorf("%s named over with peak %d", r, peak)
		}
	}
}

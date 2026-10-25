package main

import (
	"bytes"
	"errors"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func mustDay(t *testing.T, s string) day {
	t.Helper()
	d, ok := parseDay(s)
	if !ok {
		t.Fatalf("bad day %q", s)
	}
	return d
}

// Over generated card sets, for every week after the first: bugs opened minus
// bugs closed is the change in open bugs (no reopens), and the weeks are
// consecutive Mondays. The todays straddle US DST changes and an ISO week-year
// boundary, which civil-day arithmetic must not notice.
func TestWeekStatsReconcilesOpenBugs(t *testing.T) {
	for _, todayS := range []string{"2026-03-09", "2026-11-02", "2026-12-31", "2027-01-04", "2026-10-25"} {
		today := mustDay(t, todayS)
		for seed := int64(0); seed < 50; seed++ {
			r := rand.New(rand.NewSource(seed))
			var cards []cardDays
			for i := 0; i < 60; i++ {
				created := today - day(r.Intn(200))
				c := cardDays{kind: []string{"bug", "feature", ""}[r.Intn(3)], created: created, close: -1}
				if r.Intn(2) == 0 {
					c.close = created + day(r.Intn(int(today-created)+1))
				}
				cards = append(cards, c)
			}
			weeks := 1 + r.Intn(16)
			rows := weekStats(cards, today, weeks)
			if len(rows) != weeks || rows[weeks-1].Start != today.monday() {
				t.Fatalf("%s seed %d: rows %d, last week %s, want %s", todayS, seed, len(rows), rows[len(rows)-1].Start, today.monday())
			}
			for i := 1; i < weeks; i++ {
				if rows[i].Start-rows[i-1].Start != 7 || time.Unix(int64(rows[i].Start)*86400, 0).UTC().Weekday() != time.Monday {
					t.Fatalf("%s seed %d: week %d starts %s", todayS, seed, i, rows[i].Start)
				}
				if got, want := rows[i].BugsOpened-rows[i].BugsClosed, rows[i].OpenBugs-rows[i-1].OpenBugs; got != want {
					t.Fatalf("%s seed %d week %s: opened-closed %d, open changed %d", todayS, seed, rows[i].Start, got, want)
				}
			}
		}
	}
}

func TestWeekStatsCases(t *testing.T) {
	today := mustDay(t, "2026-10-25") // a Sunday: its week starts 2026-10-19
	cards := []cardDays{
		{kind: "bug", created: mustDay(t, "2026-10-20"), close: mustDay(t, "2026-10-22")},     // opened and closed this week
		{kind: "bug", created: mustDay(t, "2026-01-05"), close: mustDay(t, "2026-02-01")},     // closed before the window
		{kind: "bug", created: mustDay(t, "2026-09-01"), close: -1},                           // open throughout
		{kind: "feature", created: mustDay(t, "2026-10-19"), close: mustDay(t, "2026-10-25")}, // a feature, both ends
		{kind: "", created: mustDay(t, "2026-10-21"), close: -1},                              // unclassified: in no column
	}
	rows := weekStats(cards, today, 2)
	last := rows[1]
	if last.Start.String() != "2026-10-19" || last.BugsOpened != 1 || last.BugsClosed != 1 || last.OpenBugs != 1 || last.FeaturesOpened != 1 || last.FeaturesClosed != 1 {
		t.Fatalf("this week: %+v", last)
	}
	if rows[0].OpenBugs != 1 || rows[0].BugsOpened != 0 {
		t.Fatalf("last week: %+v", rows[0])
	}
	if o, c := unclassified(cards); o != 1 || c != 0 {
		t.Fatalf("unclassified %d open %d closed", o, c)
	}
}

func TestSlotVelocityAndQuota(t *testing.T) {
	today := mustDay(t, "2026-10-25") // window: the 4 weeks from 2026-09-21 to 2026-10-18
	done := func(ws, closed string) cardDays {
		return cardDays{kind: "bug", created: mustDay(t, "2026-09-01"), close: mustDay(t, closed), done: true, workspace: ws}
	}
	cards := []cardDays{
		done("ariadne:2", "2026-09-22"), done("ariadne:2", "2026-09-23"), done("ariadne:2", "2026-09-24"), // 3 in week 1
		done("ariadne:3", "2026-09-30"),                                  // 1 in week 2
		done("ariadne:2", "2026-10-06"), done("ariadne:2", "2026-10-07"), // 2 in week 3
		done("ariadne:4", "2026-10-20"), // this week: outside the window
		done("ariadne:4", "2026-09-10"), // before the window
		done("", "2026-09-25"),          // no claimant: not a slot's close
		{kind: "bug", created: mustDay(t, "2026-09-01"), close: mustDay(t, "2026-09-25"), workspace: "ariadne:5"}, // wontfix/punt, not done
	}
	median, samples := slotVelocity(cards, today)
	if samples != 3 || median != 2 || bugQuota(median) != 2 {
		t.Fatalf("median %v over %d slot-weeks, quota %d; want 2 over 3, quota 2", median, samples, bugQuota(median))
	}
	if m, n := slotVelocity(cards[3:4], today); m != 1 || n != 1 {
		t.Fatalf("one slot-week: %v %d", m, n)
	}
	if m, n := slotVelocity(nil, today); n != 0 || bugQuota(m) != 1 {
		t.Fatalf("nothing measured: %v %d, quota %d", m, n, bugQuota(m))
	}
	if bugQuota(0.5) != 1 || bugQuota(2.5) != 2 {
		t.Fatal("the quota rounds down and never drops below one")
	}
}

func TestDecodeRepoCards(t *testing.T) {
	raw := []byte(`{"schema_version":1,"tracker":true,"stale":true,"new_field":1,"cards":[{"id":"000622","status":"done","kind":"bug","created":"2026-10-20","updated":"2026-10-25","claimant_workspace":"ariadne:2","later_field":2},{"id":"000700","status":"unreadable","unreadable":"bad card"}]}`)
	got, err := decodeRepoCards(raw)
	if err != nil || !got.Tracker || !got.Stale || len(got.Cards) != 2 || got.Cards[0].Workspace != "ariadne:2" || got.Cards[0].Kind != "bug" || got.Cards[1].Unreadable != "bad card" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := decodeRepoCards([]byte(`{"schema_version":2,"cards":[]}`)); err == nil || !strings.Contains(err.Error(), "schema_version 2") {
		t.Fatalf("a newer schema was read: %v", err)
	}
	if _, err := decodeRepoCards([]byte("not json")); err == nil {
		t.Fatal("accepted garbage")
	}
}

// The report covers each claimed repo; a failed read prints `?` and the rest
// continue; unreadable cards and a tracker-less repo are said, never hidden.
func TestTriageStats(t *testing.T) {
	f := &fakeSDLC{
		fleet: Fleet{Slots: []FleetSlot{{Repo: "ariadne", Slot: 0, HostPath: "/w/ariadne"}, {Repo: "pair", Slot: 0, HostPath: "/w/pair"}, {Repo: "echo", Slot: 0, HostPath: "/w/echo"}}},
		cards: map[string]RepoCards{
			"/w/ariadne": {Tracker: true, Cards: []RepoCard{
				{ID: "000622", Status: "done", Kind: "bug", Created: "2026-10-10", Updated: "2026-10-14", Workspace: "ariadne:2"},
				{ID: "000633", Status: "working", Kind: "feature", Created: "2026-10-25"},
				{ID: "000500", Status: "open", Created: "2026-08-01"},
				{ID: "000501", Status: "unreadable", Unreadable: "bad card"},
			}},
			"/w/echo": {Tracker: false, Stale: true, Cards: []RepoCard{{ID: "000001", Status: "open", Kind: "bug", Created: "2026-10-01"}}},
		},
		cardsErr: map[string]error{"/w/pair": errors.New("tracker down")},
	}
	h := hub{Name: "ops", Claims: []hubClaim{{"ariadne", -1}, {"pair", -1}, {"echo", -1}}}
	var out bytes.Buffer
	triageStats(&out, f, "/w/ops", h, "", mustDay(t, "2026-10-25"), 3)
	for _, want := range []string{
		"ariadne: 3 weeks to 2026-10-25",
		"2026-10-05",
		"unclassified (no kind): 1 open, 0 closed; unreadable: 1",
		"bug quota: 1 per week (median 1 issues closed per slot per week, 1 slot-weeks",
		"pair: ? (tracker down)",
		"echo: 3 weeks", "no tracker: open counts only", "stale: the tracker could not be fetched",
		"bug quota: 1 per week (no slot closed an issue",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("lacks %q:\n%s", want, out.String())
		}
	}
	out.Reset()
	triageStats(&out, f, "/w/ops", h, "echo", mustDay(t, "2026-10-25"), 1)
	if strings.Contains(out.String(), "ariadne") || !strings.Contains(out.String(), "echo:") {
		t.Errorf("--repo echo printed others:\n%s", out.String())
	}
}

// Today is the operator's local date, not UTC's.
func TestDayOfIsTheLocalDate(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Skip(err)
	}
	if got := dayOf(time.Date(2026, 10, 25, 23, 30, 0, 0, la)).String(); got != "2026-10-25" {
		t.Fatalf("dayOf late evening in Los Angeles: %s", got)
	}
}

package landingline

import (
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC)

func slot(n int) *int { return &n }

func grant(issue int, granted time.Time, hours int) Grant {
	return Grant{Issue: issue, Grantor: Grantor{Address: "ops:0", Slot: slot(0), Name: "TL"}, Reason: "fleet-blocking", Granted: granted, Expires: granted.Add(time.Duration(hours) * time.Hour)}
}

func seqs(ts []Ticket) []int64 {
	var out []int64
	for _, t := range ts {
		out = append(out, t.Seq)
	}
	return out
}

func TestNextHead(t *testing.T) {
	a, b, c := Ticket{Seq: 3, Issue: 10}, Ticket{Seq: 1, Issue: 11}, Ticket{Seq: 2, Issue: 12}
	cases := []struct {
		name      string
		tickets   []Ticket
		grants    map[int]Grant
		head      int64
		displaced []int64
		ok        bool
	}{
		{"empty line", nil, nil, 0, nil, false},
		{"FIFO by Seq, not slice order", []Ticket{a, b, c}, nil, 1, nil, true},
		{"fleet-blocking overtakes older normal tickets", []Ticket{a, b, c}, map[int]Grant{10: grant(10, t0, 4)}, 3, []int64{1, 2}, true},
		{"FIFO among fleet-blocking", []Ticket{a, b, c}, map[int]Grant{10: grant(10, t0, 4), 12: grant(12, t0, 4)}, 2, []int64{1}, true},
		{"a grant for an issue not in line changes nothing", []Ticket{a, c}, map[int]Grant{99: grant(99, t0, 4)}, 2, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			head, displaced, ok := NextHead(tc.tickets, tc.grants)
			if ok != tc.ok || (ok && head.Seq != tc.head) {
				t.Fatalf("head %v ok %v, want %d %v", head.Seq, ok, tc.head, tc.ok)
			}
			if got := seqs(displaced); len(got) != len(tc.displaced) || (len(got) > 0 && got[0] != tc.displaced[0]) || (len(got) > 1 && got[1] != tc.displaced[1]) {
				t.Fatalf("displaced %v, want %v", got, tc.displaced)
			}
		})
	}
}

func TestValidGrants(t *testing.T) {
	now := t0.Add(time.Hour)
	worker := grant(5, t0, 4)
	worker.Grantor.Slot = slot(3)
	noGrantor := grant(6, t0, 4)
	noGrantor.Grantor = Grantor{}
	noReason := grant(7, t0, 4)
	noReason.Reason = ""
	revoked := grant(8, t0, 4)
	revoked.Revoked = true
	shell := grant(9, t0, 4)
	shell.Grantor.Slot = nil
	cases := []struct {
		name    string
		entries []Grant
		valid   []int
	}{
		{"a live :0 grant counts", []Grant{grant(1, t0, 4)}, []int{1}},
		{"an operator shell (no slot) counts", []Grant{shell}, []int{9}},
		{"expired", []Grant{grant(2, t0.Add(-5*time.Hour), 4)}, nil},
		{"longer than 24 h", []Grant{grant(3, t0, 25)}, nil},
		{"a numbered-slot grantor confers nothing", []Grant{worker}, nil},
		{"no grantor", []Grant{noGrantor}, nil},
		{"no reason", []Grant{noReason}, nil},
		{"revoked", []Grant{revoked}, nil},
		{"the latest entry per issue wins: revoke after grant", []Grant{grant(8, t0, 4), revoked}, nil},
		{"the latest entry per issue wins: renewal after revoke", []Grant{revoked, grant(8, t0, 4)}, []int{8}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidGrants(tc.entries, now)
			if len(got) != len(tc.valid) {
				t.Fatalf("valid %v, want %v", got, tc.valid)
			}
			for _, issue := range tc.valid {
				if _, ok := got[issue]; !ok {
					t.Fatalf("issue %d not valid: %v", issue, got)
				}
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	cases := []struct {
		name         string
		caller, tree *int
		ok           bool
	}{
		{"operator shell outside a workspace", nil, nil, true},
		{"a :0 slot (the TL's hub)", slot(0), nil, true},
		{"a :0 caller inside a :0 Couch tree", slot(0), slot(0), true},
		{"a numbered caller", slot(2), nil, false},
		{"a :0 checkout entered from a numbered Couch slot", slot(0), slot(2), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Authorize(tc.caller, tc.tree); (err == nil) != tc.ok {
				t.Fatalf("Authorize = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

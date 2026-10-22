package main

import (
	"strings"
	"testing"
)

func leadDrainSection(t *testing.T, args ...string) string {
	t.Helper()
	help := strings.Join(strings.Fields(strings.ReplaceAll(helpOutput(t, args...), "`", "")), " ")
	_, section, ok := strings.Cut(help, "DRAIN THE FLEET ")
	if !ok {
		t.Errorf("tl-lead %v: missing DRAIN THE FLEET procedure", args)
		return ""
	}
	section, _, ok = strings.Cut(section, " TEAMS ")
	if !ok {
		t.Error("missing TEAMS boundary after drain procedure")
		return ""
	}
	return section
}

func TestLeadDrainFleetSafetyAndRecoveryAuthority(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {}} {
		help := leadDrainSection(t, args...)
		for _, want := range []string{
			"steward", "sdlc queue pause --reason TEXT --owner TOKEN",
			"couch --drain [--fast]", "--cancel-running", "only running suites",
			"admitted merge", "safe boundary", "status", "receipts", "peek",
			"durable checkpoints", "managed work outcomes", "restart safety",
			"Queued work blocked by a pause remains outstanding",
			"quiet screens", "drain flags", "do not prove restart safety",
			"After restart", "start recovery", "one at a time", "owned slots",
			"couch --undrain [--message TEXT]", "sdlc queue resume --owner TOKEN",
			"only your hold", "does not restart interrupted work", "tl-worker --help", "couch --skill",
		} {
			if !strings.Contains(help, want) {
				t.Errorf("tl-lead %v: missing drain safeguard %q", args, want)
			}
		}
		for _, order := range [][]string{
			{"sdlc queue pause --reason TEXT --owner TOKEN", "couch --drain [--fast]", "durable checkpoints", "restart safety"},
			{"After restart", "start recovery", "one at a time"},
			{"couch --undrain [--message TEXT]", "sdlc queue resume --owner TOKEN"},
		} {
			previous := -1
			for _, step := range order {
				next := strings.Index(help, step)
				if next < 0 || next <= previous {
					t.Errorf("tl-lead %v: drain procedure must order %q", args, order)
					break
				}
				previous = next
			}
		}
		for _, obsolete := range []string{"couch --drain --ack", "wait for every drained reply", "automatically recover on restart"} {
			if strings.Contains(help, obsolete) {
				t.Errorf("tl-lead %v: obsolete drain/recovery authority %q", args, obsolete)
			}
		}
	}
}

func TestLeadAutomationIsCapabilityQualifiedAndNotRecoveryAuthority(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {}} {
		help := leadDrainSection(t, args...)
		for _, want := range []string{
			"when Couch provides couch --automation", "couch --automation status",
			"first genuine operator", "console", "session file", "side doc", "keep it off",
			"couch --automation on", "own hub", "never change another hub",
			"worker reports", "restart", "not permission", "no automatic recovery parsing or replay",
			"operator", "start recovery",
		} {
			if !strings.Contains(help, want) {
				t.Errorf("tl-lead %v: missing automation safeguard %q", args, want)
			}
		}
		qualifier, status, enable := strings.Index(help, "when Couch provides couch --automation"), strings.Index(help, "couch --automation status"), strings.Index(help, "couch --automation on")
		input := strings.Index(help, "first genuine operator")
		if qualifier < 0 || status <= qualifier || input <= status || enable <= input {
			t.Errorf("tl-lead %v: capability check and status must precede operator input and own-hub automation enable", args)
		}
	}
}

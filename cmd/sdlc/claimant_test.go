package main

import (
	"context"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestParseMachineIdentitySources(t *testing.T) {
	ioreg := `+-o J316sAP  <class IOPlatformExpertDevice>
    {
      "IOPlatformSerialNumber" = "C02XXXX"
      "IOPlatformUUID" = "2f1c0a3e-1111-2222-3333-444455556666"
    }`
	if id, ok := parseIOPlatformUUID(ioreg); !ok || id != "2F1C0A3E-1111-2222-3333-444455556666" {
		t.Fatalf("ioreg: %q %v", id, ok)
	}
	if _, ok := parseIOPlatformUUID(`"IOPlatformSerialNumber" = "C02XXXX"`); ok {
		t.Fatal("ioreg without a UUID accepted")
	}
	if id, ok := parseMachineIDFile("4c4c4544004b4d10804bc4c04f4e4d32\n"); !ok || id != "4c4c4544004b4d10804bc4c04f4e4d32" {
		t.Fatalf("machine-id: %q %v", id, ok)
	}
	for _, bad := range []string{"", "uninitialized\n", "00000000000000000000000000000000", "4C4C"} {
		if _, ok := parseMachineIDFile(bad); ok {
			t.Errorf("machine-id %q accepted", bad)
		}
	}
}

// The real identity on this host: a fingerprint (never the raw ID), the
// canonical worktree and the publication repository.
func TestResolveClaimantIdentityOnThisHost(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("machine identity is supported on darwin and linux")
	}
	if _, err := machineID(); err != nil {
		t.Skipf("host machine ID unreadable here: %v", err)
	}
	cardPath, card, detailPath, detail := seededIssue(t, "000501", "who")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	c, err := resolveClaimantIdentity(env)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := machineID()
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(c.Machine) || c.Machine != issue.MachineFingerprint(raw) {
		t.Fatalf("machine %q", c.Machine)
	}
	if c.Worktree != canonRoot(r.root) || c.Repository != env.target.Repository || c.Operator == "" || c.MachineName == "" {
		t.Fatalf("identity %+v", c)
	}
	if c.Workspace != "" {
		t.Fatalf("a plain test clone got a slot label %q", c.Workspace)
	}
}

// withClaimant makes this test's workspace identity c (in-process only).
func withClaimant(t *testing.T, c issue.Claimant) {
	t.Helper()
	prev := claimantIdentity
	claimantIdentity = func(*trackerEnv) (issue.Claimant, error) { return c, nil }
	t.Cleanup(func() { claimantIdentity = prev })
}

// hostClaimant is the identity a claim from root would record on this host —
// for built-binary fixtures, whose subprocess cannot see withClaimant.
func hostClaimant(t *testing.T, root string) issue.Claimant {
	t.Helper()
	raw, err := machineID()
	if err != nil {
		t.Skipf("claims need the host machine ID, unreadable here: %v", err)
	}
	target, err := gitx.ResolvePublicationTarget(context.Background(), root, "main")
	if err != nil {
		t.Fatal(err)
	}
	name := strings.TrimSpace(testfix.Capture(t, root, "config", "--get", "user.name"))
	return issue.Claimant{Operator: name, Machine: issue.MachineFingerprint(raw), MachineName: machineName(),
		Worktree: canonRoot(root), Repository: target.Repository}
}

// withOwner returns card with c as its recorded claimant.
func withOwner(t *testing.T, card []byte, c issue.Claimant) []byte {
	t.Helper()
	out, err := issue.SetCardClaimant(card, c)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func mustSplitCard(t *testing.T, full string) []byte {
	t.Helper()
	card, _, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	return card
}

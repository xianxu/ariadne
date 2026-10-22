package main

import (
	"bytes"
	"context"
	"errors"
	"net/netip"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/netcheck"
)

type netCommandProbe struct {
	mu    sync.Mutex
	calls []string
	fail  bool
}

func (p *netCommandProbe) Lookup(_ context.Context, host string, path netcheck.Path) ([]netip.Addr, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, host+" "+string(path))
	if p.fail {
		return nil, errors.New("unavailable")
	}
	return []netip.Addr{netip.MustParseAddr("192.0.2.1")}, nil
}
func (*netCommandProbe) HTTPS(context.Context, string, netip.Addr) (int, error) { return 403, nil }

func runNetCommand(t *testing.T, p *netCommandProbe, args ...string) (string, error) {
	t.Helper()
	r := &cobra.Command{Use: "sdlc"}
	r.AddCommand(newNetCmd(nil, p)) // --host must not need a repository or Git runner.
	var b bytes.Buffer
	r.SetOut(&b)
	r.SetErr(&b)
	r.SetArgs(append([]string{"net", "check"}, args...))
	err := r.Execute()
	return b.String(), err
}

func TestNetCheckCommandExplicitHostAndBothPaths(t *testing.T) {
	p := &netCommandProbe{}
	out, err := runNetCommand(t, p, "--host", "github.com")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"github.com", "api.github.com", "system", "direct", "HTTPS"} {
		if !strings.Contains(out, s) {
			t.Errorf("missing %q in %s", s, out)
		}
	}
	if len(p.calls) != 4 {
		t.Fatalf("DNS calls=%v", p.calls)
	}
}

func TestNetCheckCommandNoProbeForInvalidInputOrHelp(t *testing.T) {
	for _, args := range [][]string{{"--host", "https://user:secret@example.com/x"}, {"--host", "example.com;echo x"}, {"--host", "example.com", "--remote", "origin"}, {"--host", "example.com", "extra"}, {"--help"}} {
		p := &netCommandProbe{}
		out, err := runNetCommand(t, p, args...)
		if args[0] != "--help" && err == nil {
			t.Errorf("accepted %v", args)
		}
		if len(p.calls) != 0 {
			t.Errorf("%v probed %v", args, p.calls)
		}
		if strings.Contains(out, "secret") {
			t.Fatal("credential echoed")
		}
	}
}

func TestNetCheckCommandPartialFailureStillPrintsReport(t *testing.T) {
	p := &netCommandProbe{fail: true}
	out, err := runNetCommand(t, p, "--host", "example.com")
	if err == nil || !strings.Contains(out, "system") || !strings.Contains(out, "direct") {
		t.Fatalf("out=%s err=%v", out, err)
	}
}

func TestNetCheckSignalOwnershipIsNarrow(t *testing.T) {
	r := buildRoot()
	for _, tc := range []struct {
		args []string
		want bool
	}{{[]string{"net", "check"}, true}, {[]string{"workspace"}, false}, {[]string{"landing", "status"}, false}} {
		c, _, err := r.Find(tc.args)
		if err != nil {
			t.Fatal(err)
		}
		if got := cliOwnsReviewSignals(c); got != tc.want {
			t.Errorf("%v signal ownership=%v", tc.args, got)
		}
	}
}

type netConfigRunner struct {
	execGitRunner
	values map[string]string
}

func (r netConfigRunner) Git(args ...string) ([]byte, error) {
	if value, ok := r.values[strings.Join(args, " ")]; ok {
		return []byte(value), nil
	}
	return nil, errors.New("unexpected config read")
}

func TestNetCheckEffectiveRemoteHostsAndNoCredentials(t *testing.T) {
	r := netConfigRunner{values: map[string]string{
		"remote get-url --all -- upstream":        "https://user:secret@fetch.example/repo?credential=secret",
		"remote get-url --push --all -- upstream": "ssh://git@push.example/owner/repo",
	}}
	host, other, err := netCheckRemote(r, "upstream")
	if err != nil || host != "fetch.example" || other != "push.example" {
		t.Fatalf("%q %q %v", host, other, err)
	}
	p := &netCommandProbe{}
	c := newNetCmd(r, p)
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs([]string{"check", "--remote", "upstream"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "sdlc net check --host push.example") || strings.Contains(out.String(), "secret") {
		t.Fatalf("unsafe or missing recovery: %s", out.String())
	}
	r.values["remote get-url --all -- upstream"] += "\nhttps://other.example/repo"
	_, _, err = netCheckRemote(r, "upstream")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("ambiguous configuration=%v", err)
	}
}

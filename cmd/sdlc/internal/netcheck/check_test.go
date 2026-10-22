package netcheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"
)

type stateProbe struct {
	mu                  sync.Mutex
	fail                bool
	active, peak, calls int
	allowed             map[netip.Addr]bool
	bad                 bool
	block               bool
}

func (p *stateProbe) Lookup(ctx context.Context, host string, path Path) ([]netip.Addr, error) {
	p.mu.Lock()
	p.active++
	p.calls++
	if p.active > p.peak {
		p.peak = p.active
	}
	fail, block := p.fail, p.block
	p.mu.Unlock()
	defer func() { p.mu.Lock(); p.active--; p.mu.Unlock() }()
	if block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if fail {
		return nil, errors.New("https://user:secret@evil/path")
	}
	ip := netip.MustParseAddr("192.0.2.1")
	if path == Direct {
		ip = netip.MustParseAddr("192.0.2.2")
	}
	p.mu.Lock()
	p.allowed[ip] = true
	p.mu.Unlock()
	return []netip.Addr{ip, netip.MustParseAddr("192.0.2.3")}, nil
}
func (p *stateProbe) HTTPS(ctx context.Context, host string, ip netip.Addr) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if !p.allowed[ip] {
		p.bad = true
	}
	return 403, nil
}
func TestCheckRecoveryAndSafeReport(t *testing.T) {
	p := &stateProbe{fail: true, allowed: map[netip.Addr]bool{}}
	r := Check(context.Background(), "github.com", "api.github.com", p)
	if r.Healthy() {
		t.Fatal("failed DNS healthy")
	}
	var out bytes.Buffer
	if _, err := r.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "secret") || strings.Contains(out.String(), "evil") {
		t.Fatal(out.String())
	}
	p.fail = false
	p.calls = 0
	r = Check(context.Background(), "github.com", "api.github.com", p)
	if !r.Healthy() || p.calls != 6 || p.bad || p.peak > 4 {
		t.Fatalf("report=%+v fake=%+v", r, p)
	}
	out.Reset()
	_, _ = r.WriteTo(&out)
	if !strings.Contains(out.String(), "403") || !strings.Contains(out.String(), "no-proxy") {
		t.Fatal(out.String())
	}
}
func TestCheckCancelJoinsAndInvalidTargetsHaveNoEffects(t *testing.T) {
	p := &stateProbe{block: true, allowed: map[netip.Addr]bool{}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	r := Check(ctx, "example.com", "api.example.com", p)
	if r.Healthy() || p.active != 0 || p.calls != 4 {
		t.Fatalf("not joined: %+v", p)
	}
	for _, host := range []string{"https://user:secret@example.com", "bad\nname", "-x", ""} {
		p.calls = 0
		r = Check(context.Background(), host, "", p)
		if r.Healthy() || p.calls != 0 {
			t.Fatal(host)
		}
	}
}
func TestParseAddresses(t *testing.T) {
	got := ParseAddresses([]byte(";; SERVER: 127.0.0.1#53\nName: evil\nip_address: 192.0.2.1\n192.0.2.1 STREAM\n2001:db8::1\n"))
	if fmt.Sprint(got) != "[192.0.2.1 2001:db8::1]" {
		t.Fatal(got)
	}
}
func FuzzParseAddresses(f *testing.F) {
	f.Add("ip_address: 192.0.2.1\n")
	f.Fuzz(func(t *testing.T, s string) {
		for _, ip := range ParseAddresses([]byte(s)) {
			if !ip.IsValid() || ip.Zone() != "" {
				t.Fatal(ip)
			}
		}
	})
}

type deadlineProbe struct {
	t     *testing.T
	mu    sync.Mutex
	https map[netip.Addr]int
}

func (p *deadlineProbe) Lookup(ctx context.Context, host string, path Path) ([]netip.Addr, error) {
	d, ok := ctx.Deadline()
	if !ok || time.Until(d) > 8*time.Second {
		p.t.Error("missing total budget")
	}
	ip := netip.MustParseAddr("192.0.2.41")
	if path == Direct {
		ip = netip.MustParseAddr("192.0.2.42")
	}
	return []netip.Addr{ip, netip.MustParseAddr("192.0.2.99")}, nil
}
func (p *deadlineProbe) HTTPS(ctx context.Context, host string, ip netip.Addr) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.https[ip]++
	return 200, nil
}
func TestCheckSamplesOneAddressIndependentlyPerPath(t *testing.T) {
	p := &deadlineProbe{t: t, https: map[netip.Addr]int{}}
	r := Check(context.Background(), "example.com", "", p)
	if !r.Healthy() {
		t.Fatal(r)
	}
	for _, s := range []string{"192.0.2.41", "192.0.2.42"} {
		if p.https[netip.MustParseAddr(s)] != 1 {
			t.Fatal(p.https)
		}
	}
	if len(p.https) != 2 {
		t.Fatal(p.https)
	}
}
func TestInvalidAPIHasNoEffects(t *testing.T) {
	p := &stateProbe{}
	r := Check(context.Background(), "example.com", "https://user:secret@evil", p)
	if r.Healthy() || p.calls != 0 {
		t.Fatal(r)
	}
}

type partialProbe struct{ deadlineProbe }

func (p *partialProbe) Lookup(ctx context.Context, host string, path Path) ([]netip.Addr, error) {
	if path == System {
		return nil, errors.New("https://user:secret@evil")
	}
	return p.deadlineProbe.Lookup(ctx, host, path)
}
func TestCheckRetainsIndependentPartialPaths(t *testing.T) {
	p := &partialProbe{deadlineProbe{t: t, https: map[netip.Addr]int{}}}
	r := Check(context.Background(), "example.com", "", p)
	if r.Healthy() || r.Observations[0].DNS != "failed" || r.Observations[1].Status != 200 || len(p.https) != 1 {
		t.Fatalf("%+v", r)
	}
}

type errorHTTPS struct{ deadlineProbe }

func (p *errorHTTPS) HTTPS(context.Context, string, netip.Addr) (int, error) {
	return 0, errors.New("https://user:secret@evil")
}
func TestCheckDoesNotRenderHTTPSSecrets(t *testing.T) {
	p := &errorHTTPS{deadlineProbe{t: t}}
	r := Check(context.Background(), "example.com", "", p)
	var out bytes.Buffer
	_, _ = r.WriteTo(&out)
	if r.Healthy() || strings.Contains(out.String(), "secret") || strings.Contains(out.String(), "evil") {
		t.Fatal(out.String())
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestReportWriterCountAndFailure(t *testing.T) {
	r := Report{Invalid: true}
	var b bytes.Buffer
	n, err := r.WriteTo(&b)
	if err != nil || n != int64(b.Len()) {
		t.Fatalf("%d %d %v", n, b.Len(), err)
	}
	_, err = r.WriteTo(failedWriter{})
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}

type cancellingSuccess struct {
	deadlineProbe
	cancel context.CancelFunc
}

func (p *cancellingSuccess) Lookup(ctx context.Context, h string, path Path) ([]netip.Addr, error) {
	p.cancel()
	return p.deadlineProbe.Lookup(ctx, h, path)
}
func TestCheckCancelledSuccessIsIncomplete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &cancellingSuccess{deadlineProbe: deadlineProbe{t: t, https: map[netip.Addr]int{}}, cancel: cancel}
	r := Check(ctx, "example.com", "", p)
	if r.Healthy() {
		t.Fatal("cancelled success became healthy")
	}
	var b bytes.Buffer
	_, _ = r.WriteTo(&b)
	if !strings.Contains(b.String(), "cancelled") {
		t.Fatal(b.String())
	}
}

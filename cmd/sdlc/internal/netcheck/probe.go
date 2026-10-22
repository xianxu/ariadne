package netcheck

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

var errUnavailable = errors.New("observation unavailable")

const outputLimit = 64 * 1024

// nativeProber keeps only per-instance test seams; production has no mutable
// global resolver or transport state.
type nativeProber struct {
	goos           string
	commandTimeout time.Duration
	roots          *tls.Config
	port           string
}

func NewProber() Prober {
	return &nativeProber{goos: runtime.GOOS, commandTimeout: 2 * time.Second, port: "443"}
}

type cappedOutput struct {
	buf      bytes.Buffer
	exceeded bool
}

func (b *cappedOutput) Write(p []byte) (int, error) {
	n := len(p)
	if b.buf.Len()+n > outputLimit {
		b.exceeded = true
		p = p[:outputLimit-b.buf.Len()]
	}
	_, _ = b.buf.Write(p)
	return n, nil
}
func (p *nativeProber) command(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, p.commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = 100 * time.Millisecond
	var out cappedOutput
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if errors.Is(err, exec.ErrNotFound) {
		return nil, errUnavailable
	}
	if out.exceeded {
		return nil, errors.New("output limit exceeded")
	}
	return out.buf.Bytes(), err
}

// ParseAddresses accepts only address fields in native utility output. DNS
// comments and arbitrary text cannot introduce destinations.
func ParseAddresses(raw []byte) []netip.Addr {
	var result []netip.Addr
	seen := map[netip.Addr]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		token := fields[0]
		if token == "ip_address:" && len(fields) == 2 {
			token = fields[1]
		}
		ip, err := netip.ParseAddr(token)
		if err != nil || ip.Zone() != "" {
			continue
		}
		ip = ip.Unmap()
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}
	return result
}
func (p *nativeProber) Lookup(ctx context.Context, host string, path Path) ([]netip.Addr, error) {
	host, err := gitx.NormalizeNetworkHost(host)
	if err != nil {
		return nil, errUnavailable
	}
	if path == System {
		var out []byte
		switch p.goos {
		case "darwin":
			out, err = p.command(ctx, "dscacheutil", "-q", "host", "-a", "name", host)
		case "linux":
			out, err = p.command(ctx, "getent", "ahosts", host)
		default:
			return nil, errUnavailable
		}
		return ParseAddresses(out), err
	}
	if path != Direct {
		return nil, errUnavailable
	}
	// Sequential A/AAAA keeps two-host checks within four concurrent effects.
	var ips []netip.Addr
	for _, kind := range []string{"A", "AAAA"} {
		out, e := p.command(ctx, "dig", "+short", "+time=1", "+tries=1", host, kind)
		if e != nil {
			return ips, e
		}
		ips = append(ips, ParseAddresses(out)...)
	}
	return ips, nil
}
func (p *nativeProber) HTTPS(ctx context.Context, host string, ip netip.Addr) (int, error) {
	host, err := gitx.NormalizeNetworkHost(host)
	if err != nil || !ip.IsValid() || ip.Zone() != "" {
		return 0, errUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	config := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host}
	if p.roots != nil {
		config = p.roots.Clone()
		config.ServerName = host
		config.InsecureSkipVerify = false
	}
	// Own dialing and TLS synchronously. http.Transport normally detaches its
	// dial context to reuse sockets for later requests, which would outlive this
	// one-shot diagnostic. Only an already-owned connection enters transport.
	dialer := &tls.Dialer{NetDialer: &net.Dialer{}, Config: config}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), p.port))
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	available := make(chan net.Conn, 1)
	available <- conn
	tr := &http.Transport{Proxy: nil, DisableKeepAlives: true, DialTLSContext: func(context.Context, string, string) (net.Conn, error) {
		select {
		case c := <-available:
			return c, nil
		default:
			return nil, errUnavailable
		}
	}}
	defer tr.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://"+net.JoinHostPort(host, p.port)+"/", nil)
	if err != nil {
		return 0, errUnavailable
	}
	client := &http.Client{Transport: tr, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

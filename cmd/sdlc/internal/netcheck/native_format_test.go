package netcheck

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParseAddressesNativeFormats(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		want      []string
	}{
		{"mac-ipv4", "name: example.test\nip_address: 192.0.2.1\n", []string{"192.0.2.1"}},
		{"mac-ipv6-only", "name: example.test\nipv6_address: ::1\n", []string{"::1"}},
		{"mac-mixed", "ip_address: 192.0.2.1\nipv6_address: 2001:db8::1\nipv6_address: 2001:db8::1\n", []string{"192.0.2.1", "2001:db8::1"}},
		{"mac-malformed", "ip_address:\nipv6_address:\nip_address: invalid\nipv6_address: invalid\nip_address: 192.0.2.2 extra\nipv6_address: ::1 extra\nipv6_address: fe80::1%en0\nipv6_address:: ::1\n", nil},
		{"getent", "192.0.2.1 STREAM example.test\n192.0.2.1 DGRAM\n2001:db8::1 RAW\n", []string{"192.0.2.1", "2001:db8::1"}},
		{"dig", "alias.example.test.\n192.0.2.1\n2001:db8::1\n", []string{"192.0.2.1", "2001:db8::1"}},
		{"untrusted-format-fields", "name: 192.0.2.1\naddress: ::1\n;; SERVER: 192.0.2.1#53\n;; ANSWER SECTION:\nexample.test. 60 IN AAAA ::1\nnot-an-ip STREAM example.test\nfe80::1%en0 STREAM\n", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var want []netip.Addr
			for _, s := range tc.want {
				want = append(want, netip.MustParseAddr(s))
			}
			if got := ParseAddresses([]byte(tc.raw)); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
}

type nativeFormatProbe struct {
	native *nativeProber
	mu     sync.Mutex
	https  []netip.Addr
}

func (p *nativeFormatProbe) Lookup(ctx context.Context, host string, path Path) ([]netip.Addr, error) {
	return p.native.Lookup(ctx, host, path)
}
func (p *nativeFormatProbe) HTTPS(_ context.Context, _ string, ip netip.Addr) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.https = append(p.https, ip)
	return 200, nil
}
func TestNativeDarwinFormatsEligibleForHTTPS(t *testing.T) {
	// Captured with dscacheutil -q host -a name localhost on macOS. Replay the
	// actual native fields through Lookup; no public resolver or TLS is used.
	capture, err := os.ReadFile("testdata/dscacheutil-localhost.txt")
	if err != nil {
		t.Fatal(err)
	}
	var ipv6Only string
	for _, block := range strings.Split(string(capture), "\n\n") {
		if strings.Contains(block, "ipv6_address:") {
			ipv6Only = block + "\n\n"
		}
	}
	if ipv6Only == "" {
		t.Fatal("captured fixture lacks IPv6 observation")
	}

	for _, tc := range []struct {
		name, output string
		want         []netip.Addr
	}{
		{"ipv6-only", ipv6Only, []netip.Addr{netip.MustParseAddr("::1")}},
		{"captured-mixed", string(capture), []netip.Addr{netip.MustParseAddr("::1"), netip.MustParseAddr("127.0.0.1")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("PATH", dir)
			script := "#!/bin/sh\nif [ \"$*\" != '-q host -a name localhost' ]; then exit 44; fi\nprintf '%s' '" + tc.output + "'\n"
			if err := os.WriteFile(filepath.Join(dir, "dscacheutil"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			dig := "#!/bin/sh\ncase \"$*\" in\n'+short +time=1 +tries=1 localhost A') exit 0;;\n'+short +time=1 +tries=1 localhost AAAA') echo ::1;;\n*) exit 44;;\nesac\n"
			if err := os.WriteFile(filepath.Join(dir, "dig"), []byte(dig), 0700); err != nil {
				t.Fatal(err)
			}
			p := &nativeFormatProbe{native: &nativeProber{goos: "darwin", commandTimeout: time.Second}}
			ips, err := p.Lookup(context.Background(), "localhost", System)
			if err != nil || !reflect.DeepEqual(ips, tc.want) {
				t.Fatalf("native DNS = %v, %v; want %v", ips, err, tc.want)
			}
			r := Check(context.Background(), "localhost", "", p)
			if !r.Healthy() || len(p.https) != 2 || r.Observations[0].DNS != "ok" || r.Observations[0].HTTPS != "ok" {
				t.Fatalf("report %+v HTTPS addresses %v", r, p.https)
			}
			if !reflect.DeepEqual(r.Observations[0].Addresses, tc.want) {
				t.Fatal(r.Observations[0])
			}
		})
	}
}

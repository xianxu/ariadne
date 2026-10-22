package netcheck

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNativeHTTPSPrivateTLS(t *testing.T) {
	var calls atomic.Int32
	closed := make(chan struct{}, 8)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "HEAD" || r.Header.Get("Authorization") != "" || !strings.HasPrefix(r.Host, "example.com:") {
			t.Errorf("unexpected request: %+v", r)
		}
		w.Header().Set("Location", "https://never-contact.invalid")
		w.WriteHeader(302)
	}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateClosed {
			closed <- struct{}{}
		}
	}
	server.StartTLS()
	defer server.Close()
	host, port, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	p := &nativeProber{port: port, roots: &tls.Config{RootCAs: roots}}
	t.Setenv("HTTPS_PROXY", "http://user:secret@127.0.0.1:1")
	status, err := p.HTTPS(context.Background(), "example.com", netip.MustParseAddr(host))
	if err != nil || status != 302 || calls.Load() != 1 {
		t.Fatalf("%d %v calls=%d", status, err, calls.Load())
	}
	// The successful request must close its connection before return, without
	// depending on server shutdown to hide a leaked transport.
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("HTTPS connection remained open")
	}
	_, err = p.HTTPS(context.Background(), "wrong.invalid", netip.MustParseAddr(host))
	if err == nil {
		t.Fatal("accepted wrong hostname")
	}
	p.roots = nil
	_, err = p.HTTPS(context.Background(), "example.com", netip.MustParseAddr(host))
	if err == nil {
		t.Fatal("accepted untrusted certificate")
	}

}
func TestNativeUtilitiesMissingOutputAndConfiguredQueries(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	p := &nativeProber{goos: "linux", commandTimeout: time.Second}
	if _, err := p.Lookup(context.Background(), "example.com", System); !errors.Is(err, errUnavailable) {
		t.Fatal(err)
	}
	script := "#!/bin/sh\ncase \"$*\" in\n'+short +time=1 +tries=1 example.com A') echo 192.0.2.1;;\n'+short +time=1 +tries=1 example.com AAAA') echo 2001:db8::1;;\n*) exit 44;;\nesac\n"
	if err := os.WriteFile(filepath.Join(dir, "dig"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	ips, err := p.Lookup(context.Background(), "example.com", Direct)
	if err != nil || fmt.Sprint(ips) != "[192.0.2.1 2001:db8::1]" {
		t.Fatalf("%v %v", ips, err)
	}
	out, err := p.command(context.Background(), "dig", "unexpected")
	if err == nil || len(out) != 0 {
		t.Fatal("unexpected argv passed")
	}
	// A native utility exceeding its cap cannot manufacture a successful result.
	flood := "#!/bin/sh\ni=0\nwhile [ $i -lt 7000 ]; do echo 192.0.2.1; i=$((i+1)); done\n"
	if err := os.WriteFile(filepath.Join(dir, "getent"), []byte(flood), 0700); err != nil {
		t.Fatal(err)
	}
	if ips, err := p.Lookup(context.Background(), "example.com", System); err == nil || len(ips) != 0 {
		t.Fatalf("uncapped output: %d %v", len(ips), err)
	}
}

func TestNativeHTTPSCancellationClosesStalledTLS(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	closed := make(chan error, 1)
	ready := make(chan struct{})
	go func() {
		conn, e := listener.Accept()
		if e != nil {
			closed <- e
			return
		}
		defer conn.Close()
		close(ready)
		_, e = io.Copy(io.Discard, conn)
		closed <- e
	}()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	p := &nativeProber{port: port}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := p.HTTPS(ctx, "example.com", netip.MustParseAddr(host)); done <- e }()
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("TLS connection never started")
	}
	cancel()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("cancelled HTTPS succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("TLS handshake not bounded")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("TLS socket survived cancellation")
	}
}

func TestNativeDirectRetainsPartialAddresses(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	script := "#!/bin/sh\ncase \"$*\" in\n*A) echo 192.0.2.1;;\nesac\ncase \"$*\" in\n*AAAA) exit 1;;\nesac\n"
	if err := os.WriteFile(filepath.Join(dir, "dig"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	p := &nativeProber{commandTimeout: time.Second}
	ips, err := p.Lookup(context.Background(), "example.com", Direct)
	if err == nil || len(ips) == 0 || ips[0] != netip.MustParseAddr("192.0.2.1") {
		t.Fatalf("%v %v", ips, err)
	}
}

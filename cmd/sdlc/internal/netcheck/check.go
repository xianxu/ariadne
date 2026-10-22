// Package netcheck makes bounded, read-only observations of resolver paths and
// direct HTTPS reachability. It does not diagnose an upstream outage.
package netcheck

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

type Path string

const (
	System Path = "system"
	Direct Path = "direct"
)

type Prober interface {
	Lookup(context.Context, string, Path) ([]netip.Addr, error)
	HTTPS(context.Context, string, netip.Addr) (int, error)
}

// Observation retains only validated addresses and fixed reason classes.
type Observation struct {
	Host      string
	Path      Path
	Addresses []netip.Addr
	DNS       string
	HTTPS     string
	Status    int
}
type Report struct {
	Observations []Observation
	Invalid      bool
	interrupted  string
}

func (r Report) Healthy() bool {
	if r.Invalid || r.interrupted != "" || len(r.Observations) == 0 {
		return false
	}
	for _, o := range r.Observations {
		if o.DNS != "ok" || (o.HTTPS != "" && o.HTTPS != "ok") {
			return false
		}
	}
	return true
}

// WriteTo implements io.WriterTo without retaining arbitrary probe errors.
func (r Report) WriteTo(w io.Writer) (int64, error) {
	counted := &countWriter{w: w}
	err := r.render(counted)
	return counted.n, err
}

type countWriter struct {
	w io.Writer
	n int64
}

func (w *countWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.n += int64(n)
	return n, err
}
func (r Report) render(w io.Writer) error {
	if r.Invalid {
		_, err := fmt.Fprintln(w, "network check: invalid host; use a hostname or IP address")
		return err
	}
	for _, o := range r.Observations {
		if _, err := fmt.Fprintf(w, "%s %s DNS: %s %v\n", o.Host, o.Path, o.DNS, o.Addresses); err != nil {
			return err
		}
		if o.HTTPS != "" {
			if _, err := fmt.Fprintf(w, "%s %s no-proxy HTTPS (one address sampled): %s", o.Host, o.Path, o.HTTPS); err != nil {
				return err
			}
			if o.Status != 0 {
				if _, err := fmt.Fprintf(w, " (HTTP %d)", o.Status); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
	}
	if r.interrupted != "" {
		if _, err := fmt.Fprintf(w, "network check: %s\n", r.interrupted); err != nil {
			return err
		}
	}
	if !r.Healthy() {
		_, err := fmt.Fprintln(w, "Observations incomplete or failed; cause unknown.")
		return err
	}
	return nil
}
func safeReason(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, errUnavailable):
		return "unavailable"
	default:
		return "failed"
	}
}

// Check joins all effects before returning. Four path jobs bound concurrency;
// each performs DNS then at most one HTTPS attempt from that same path.
func Check(ctx context.Context, host, apiHost string, p Prober) Report {
	host, err := gitx.NormalizeNetworkHost(host)
	if err != nil {
		return Report{Invalid: true}
	}
	if apiHost != "" {
		apiHost, err = gitx.NormalizeNetworkHost(apiHost)
		if err != nil {
			return Report{Invalid: true}
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	hosts := []string{host}
	if apiHost != "" && apiHost != host {
		hosts = append(hosts, apiHost)
	}
	r := Report{Observations: make([]Observation, len(hosts)*2)}
	var wg sync.WaitGroup
	for h, name := range hosts {
		for j, path := range []Path{System, Direct} {
			i := h*2 + j
			wg.Go(func() {
				o := Observation{Host: name, Path: path}
				ips, err := p.Lookup(ctx, name, path)
				for _, ip := range ips {
					if ip.IsValid() && ip.Zone() == "" {
						o.Addresses = append(o.Addresses, ip.Unmap())
					}
				}
				o.DNS = safeReason(err)
				if err == nil && len(o.Addresses) == 0 {
					o.DNS = "unknown"
				}
				if name == host {
					o.HTTPS = "unavailable"
					if ctx.Err() != nil {
						o.HTTPS = safeReason(ctx.Err())
					}
					if o.DNS == "ok" && ctx.Err() == nil {
						status, e := p.HTTPS(ctx, name, o.Addresses[0])
						o.HTTPS = safeReason(e)
						if e == nil {
							if status < 100 || status > 599 {
								o.HTTPS = "unknown"
							} else {
								o.Status = status
							}
						}
					}
				}
				r.Observations[i] = o
			})
		}
	}
	wg.Wait()
	if ctx.Err() != nil {
		r.interrupted = safeReason(ctx.Err())
	}
	return r
}

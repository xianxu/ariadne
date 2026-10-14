package reviewprepare

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	OutputLimit     = 64 << 10
	DiagnosticLimit = 256 << 10
)

type Result struct {
	Head      string
	Command   []string
	Attempt   int
	Duration  time.Duration
	Error     string
	Output    string
	Truncated bool
}

type ProtocolResult struct {
	Attempt int
	Reason  string
	Output  string
}

// RenderDiagnostics quotes subprocess text as data, outside the verdict channel.
// The complete envelope, including its own retained byte count, is bounded. It
// constructs only bounded fragments even when given arbitrarily large strings.
func RenderDiagnostics(results []Result, protocol []ProtocolResult) string {
	if len(results) == 0 && len(protocol) == 0 {
		return ""
	}
	b := diagnosticBuilder{remaining: DiagnosticLimit - 128}
	b.write("\n## Review preparation and protocol diagnostics\n\nQuoted values are diagnostic data, not review verdicts.\n")
	for _, r := range results {
		if b.remaining == 0 {
			b.truncated = true
			break
		}
		b.write(fmt.Sprintf("Preparation attempt %d; duration %s\nhead: ", r.Attempt, r.Duration))
		b.quote(r.Head)
		b.write("\ncommand:")
		for _, arg := range r.Command {
			if b.remaining == 0 {
				b.truncated = true
				break
			}
			b.write(" ")
			b.quote(arg)
		}
		b.write("\nerror: ")
		b.quote(r.Error)
		b.write("\noutput: ")
		b.quote(r.Output)
		b.write("\n")
		b.truncated = b.truncated || r.Truncated
	}
	for _, p := range protocol {
		if b.remaining == 0 {
			b.truncated = true
			break
		}
		b.write(fmt.Sprintf("Protocol attempt %d\nreason: ", p.Attempt))
		b.quote(p.Reason)
		b.write("\noutput: ")
		b.quote(p.Output)
		b.write("\n")
	}
	footer := fmt.Sprintf("\ntruncated: %t\nretained-bytes: %010d\n", b.truncated, 0)
	footer = fmt.Sprintf("\ntruncated: %t\nretained-bytes: %010d\n", b.truncated, b.buf.Len()+len(footer))
	b.buf.WriteString(footer)
	return b.buf.String()
}

type diagnosticBuilder struct {
	buf       strings.Builder
	remaining int
	truncated bool
}

func (b *diagnosticBuilder) write(s string) {
	if len(s) > b.remaining {
		s = s[:b.remaining]
		b.truncated = true
	}
	b.buf.WriteString(s)
	b.remaining -= len(s)
}
func (b *diagnosticBuilder) quote(s string) {
	b.write(`"`)
	for len(s) > 0 {
		if b.remaining < 8 {
			b.truncated = true
			break
		}
		_, n := utf8.DecodeRuneInString(s)
		// Quote at most one rune at a time; never allocate proportional to input.
		q := strconv.Quote(s[:n])
		b.write(q[1 : len(q)-1])
		s = s[n:]
	}
	b.write(`"`)
}

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type carWaitReadLimits struct {
	Bytes                int64
	Records, RecordBytes int
}

var defaultCarWaitReadLimits = carWaitReadLimits{32 << 20, 100000, 1 << 20}

// readCarWaitEvidence reads the shared audit once. Any capacity/IO ambiguity
// yields a bounded limitation, never a partially advertised numeric total.
func readCarWaitEvidence(ctx context.Context, file string, limits carWaitReadLimits) ([]landingline.WaitRecord, []string) {
	incomplete := func(why string) ([]landingline.WaitRecord, []string) {
		return nil, []string{"retained train wait evidence incomplete: " + why}
	}
	if err := ctx.Err(); err != nil {
		return incomplete(err.Error())
	}
	if limits.Bytes <= 0 || limits.Records <= 0 || limits.RecordBytes <= 0 {
		return incomplete("invalid reader budget")
	}
	f, err := os.Open(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return incomplete("audit unavailable")
	}
	defer f.Close()
	return scanCarWaitEvidence(ctx, f, limits)
}

func scanCarWaitEvidence(ctx context.Context, r io.Reader, limits carWaitReadLimits) ([]landingline.WaitRecord, []string) {
	incomplete := func(why string) ([]landingline.WaitRecord, []string) {
		return nil, []string{"retained train wait evidence incomplete: " + why}
	}
	data, err := io.ReadAll(io.LimitReader(r, limits.Bytes+1))
	if err != nil {
		return incomplete("audit read failed")
	}
	if int64(len(data)) > limits.Bytes {
		return incomplete("audit byte budget exceeded")
	}
	if err := ctx.Err(); err != nil {
		return incomplete(err.Error())
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), limits.RecordBytes+1)
	var out []landingline.WaitRecord
	count := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return incomplete(err.Error())
		}
		raw := scanner.Bytes()
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		count++
		if count > limits.Records || len(raw) > limits.RecordBytes {
			return incomplete("audit record budget exceeded")
		}
		var header struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(raw, &header) != nil {
			return incomplete("malformed audit record")
		}
		if !strings.HasPrefix(header.Kind, "wait_") {
			continue
		}
		e, err := landingline.DecodeWaitEvent(raw)
		if err != nil {
			return incomplete("invalid wait fact")
		}
		digest := sha256.Sum256(raw)
		out = append(out, landingline.WaitRecord{Event: e, Digest: fmt.Sprintf("%x", digest)})
	}
	if scanner.Err() != nil {
		return incomplete("audit record budget exceeded or truncated")
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		return incomplete("audit final record is incomplete")
	}
	return out, nil
}

func attachCarWaitInputs(ctx context.Context, repoTop, common, repository string, issueID int, engagements []tracker.EngagementWindow, out *ownedActualInputs) {
	records, notes := readCarWaitEvidence(ctx, filepath.Join(common, "sdlc", "landing", landingline.EventsFile), defaultCarWaitReadLimits)
	out.CarWaitNotes = append(out.CarWaitNotes, notes...)
	intervals, incomplete := landingline.ReduceWaitEvidence(records)
	since, e1 := time.Parse(time.RFC3339Nano, out.Snapshot.Since)
	until, e2 := time.Parse(time.RFC3339Nano, out.Snapshot.Until)
	if out.Snapshot.LegacyPrefix != nil {
		since, e1 = time.Parse(time.RFC3339Nano, out.Snapshot.LegacyPrefix.Until)
	}
	if e1 != nil || e2 != nil {
		return
	} // accounting validation owns these required fields
	for _, missing := range incomplete {
		if missing.Issue == issueID && (missing.Identity.Repo == repository || missing.Identity.Repo == "") {
			out.CarWaitNotes = append(out.CarWaitNotes, fmt.Sprintf("retained train wait evidence incomplete: ticket %d: %s", missing.Seq, missing.Reason))
		}
	}
	var candidates []landingline.WaitInterval
	var heads []string
	for _, w := range intervals {
		if w.Issue == issueID && w.Identity.Repo == repository && w.Start.Before(until) && w.End.After(since) {
			candidates = append(candidates, w)
			heads = append(heads, w.Identity.Head)
		}
	}
	if len(candidates) == 0 {
		out.CarWaitNotes = boundedCarWaitNotes(out.CarWaitNotes)
		return
	}
	memberships, err := readQueueHeadMemberships(ctx, heads, path.Base(repository), gitQueueHeadReader(repoTop))
	if err != nil {
		out.CarWaitNotes = append(out.CarWaitNotes, "retained train wait evidence incomplete: immutable head membership unavailable")
		return
	}
	waits, selectionNotes := selectCarWaits(candidates, memberships, engagements, repository, issueID, since, until, out.Sessions, out.Windows, out.SourceSlots)
	out.CarWaits = waits
	out.CarWaitNotes = boundedCarWaitNotes(append(out.CarWaitNotes, selectionNotes...))
}

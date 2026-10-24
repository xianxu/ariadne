package activetime

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	HeadHistoryByteLimit   = 4 << 20
	HeadHistoryRecordLimit = 20000
)

type headBranchKind uint8

const (
	headUnknown headBranchKind = iota
	headAttached
	headDetached
)

type headBranch struct {
	kind headBranchKind
	name string
}

type headHistoryRecord struct {
	at             time.Time
	old, next, msg string
}

func historyOID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil && strings.ToLower(s) == s
}

func historyBranch(s string) headBranch {
	s = strings.TrimPrefix(s, "refs/heads/")
	if s == "" || s == "HEAD" || historyOID(s) {
		return headBranch{kind: headDetached}
	}
	if strings.ContainsAny(s, " \t\r\n\x00~^:?*[\\") || strings.HasPrefix(s, "-") {
		return headBranch{}
	}
	return headBranch{kind: headAttached, name: s}
}

// historyTransition describes the branch facts a native log action actually
// records. Unknown actions are fences; an issue mention never supplies a branch.
func historyTransition(msg string) (before, after headBranch, preserve bool) {
	for _, prefix := range []string{"checkout: moving from ", "Branch: renamed "} {
		if rest, ok := strings.CutPrefix(msg, prefix); ok {
			parts := strings.Fields(rest)
			if len(parts) == 3 && parts[1] == "to" {
				return historyBranch(parts[0]), historyBranch(parts[2]), false
			}
			return headBranch{}, headBranch{}, false
		}
	}
	for _, prefix := range []string{"rebase (finish): returning to ", "rebase (abort): returning to "} {
		if branch, ok := strings.CutPrefix(msg, prefix); ok {
			return headBranch{}, historyBranch(branch), false
		}
	}
	if strings.HasPrefix(msg, "rebase (start): checkout ") {
		return headBranch{}, headBranch{kind: headDetached}, false
	}
	for _, prefix := range []string{"commit: ", "commit (initial): ", "commit (amend): ", "commit (merge): ", "reset: ", "merge ", "pull ", "sdlc ", "rebase (pick): ", "rebase (continue): ", "rebase (squash): ", "rebase (fixup): ", "rebase (reword): "} {
		if strings.HasPrefix(msg, prefix) {
			return headBranch{}, headBranch{}, true
		}
	}
	return headBranch{}, headBranch{}, false
}

func decodeHeadHistory(raw []byte) ([]headHistoryRecord, error) {
	if len(raw) == 0 || len(raw) > HeadHistoryByteLimit || raw[len(raw)-1] != '\n' {
		return nil, fmt.Errorf("HEAD history missing, truncated or exceeds %d bytes", HeadHistoryByteLimit)
	}
	lines := bytes.Split(raw[:len(raw)-1], []byte{'\n'})
	if len(lines) > HeadHistoryRecordLimit {
		return nil, fmt.Errorf("HEAD history exceeds %d records", HeadHistoryRecordLimit)
	}
	rows := make([]headHistoryRecord, 0, len(lines))
	for _, line := range lines {
		metadata, msg, ok := bytes.Cut(line, []byte{'\t'})
		fields := strings.Fields(string(metadata))
		if len(fields) < 6 || !historyOID(fields[0]) || !historyOID(fields[1]) || len(fields[0]) != len(fields[1]) {
			return nil, fmt.Errorf("malformed HEAD history record")
		}
		if !ok && (len(rows) != 0 || fields[0] != strings.Repeat("0", len(fields[0]))) {
			return nil, fmt.Errorf("HEAD history message is missing outside native creation")
		}
		zone := fields[len(fields)-1]
		stamp, err := strconv.ParseInt(fields[len(fields)-2], 10, 64)
		if err != nil || stamp <= 0 || len(zone) != 5 || zone[0] != '+' && zone[0] != '-' {
			return nil, fmt.Errorf("malformed HEAD history time")
		}
		offset, err := strconv.Atoi(zone[1:])
		if err != nil || offset/100 > 23 || offset%100 > 59 {
			return nil, fmt.Errorf("malformed HEAD history zone")
		}
		row := headHistoryRecord{at: time.Unix(stamp, 0).UTC(), old: fields[0], next: fields[1], msg: string(msg)}
		if len(rows) > 0 {
			previous := rows[len(rows)-1]
			if previous.next != row.old || row.at.Before(previous.at) {
				return nil, fmt.Errorf("HEAD history has a broken object chain or reversed time")
			}
		}
		rows = append(rows, row)
	}
	// Native branch rename publishes a delete/create pair in this worktree's
	// HEAD log. Only an exact adjacent pair proves one rename, never a partial
	// deletion or another operation's zero OID.
	var normalized []headHistoryRecord
	for i := 0; i < len(rows); i++ {
		row := rows[i]
		if row.next == strings.Repeat("0", len(row.next)) {
			if !strings.HasPrefix(row.msg, "Branch: renamed ") || i+1 == len(rows) {
				return nil, fmt.Errorf("HEAD history contains an incomplete deletion")
			}
			next := rows[i+1]
			if next.msg != row.msg || !next.at.Equal(row.at) || next.old != row.next || next.next != row.old {
				return nil, fmt.Errorf("HEAD history contains an incomplete rename pair")
			}
			row.next = next.next
			i++
		}
		normalized = append(normalized, row)
	}
	return normalized, nil
}

// ParseHeadHistory walks a complete captured HEAD-log prefix backwards from a
// simultaneously verified symbolic HEAD. Only explicit transitions and known
// branch-preserving actions propagate that witness. Expired prefixes leave an
// unknown interval; no present-day branch is projected before the first record.
func ParseHeadHistory(raw []byte, currentBranch, currentHead string, since, until time.Time) ([]BranchInterval, error) {
	if since.IsZero() || !until.After(since) {
		return nil, fmt.Errorf("invalid HEAD history measurement interval")
	}
	rows, err := decodeHeadHistory(raw)
	if err != nil {
		return nil, err
	}
	if rows[len(rows)-1].next != currentHead {
		return nil, fmt.Errorf("HEAD history tip disagrees with captured HEAD")
	}
	state := historyBranch(currentBranch)
	end := until
	var reversed []BranchInterval
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		before, after, preserve := historyTransition(row.msg)
		if i == 0 && row.msg == "" && row.old == strings.Repeat("0", len(row.old)) {
			preserve = true // native worktree creation, no branch change
		}
		if !preserve {
			if after.kind != headUnknown && state.kind != headUnknown && after != state {
				return nil, fmt.Errorf("HEAD history branch transition contradicts its following witness")
			}
			state = after
		}
		start := row.at
		if start.Before(since) {
			start = since
		}
		if start.Before(end) {
			reversed = append(reversed, BranchInterval{Start: start, End: end, Branch: state.name, Known: state.kind == headAttached})
			end = start
		}
		if !preserve {
			state = before
		}
		if !row.at.After(since) {
			break
		}
	}
	if since.Before(end) {
		reversed = append(reversed, BranchInterval{Start: since, End: end})
	}
	slices.Reverse(reversed)
	var out []BranchInterval
	for _, part := range reversed {
		last := len(out) - 1
		if last >= 0 && out[last].End.Equal(part.Start) && out[last].Known == part.Known && out[last].Branch == part.Branch {
			out[last].End = part.End
		} else {
			out = append(out, part)
		}
	}
	return out, nil
}

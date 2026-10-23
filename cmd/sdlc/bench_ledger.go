package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

const benchLedgerRowLimit = 10000
const benchLedgerByteLimit int64 = 160 << 20

type benchLedgerScan struct {
	rows []bench.LedgerRow
	byID map[string]int
	size int64
}

// scanBenchLedger reads each framed row exactly once. The decoder seam permits
// counted complexity tests; production always uses the strict bench codec.
func scanBenchLedger(ctx context.Context, input io.Reader, decode func([]byte) (bench.LedgerRow, error)) (benchLedgerScan, error) {
	scan := benchLedgerScan{byID: map[string]int{}}
	if decode == nil {
		decode = bench.DecodeLedgerRow
	}
	reader := bufio.NewReaderSize(input, bench.LedgerRowLimit+2)
	for {
		if err := ctx.Err(); err != nil {
			return scan, err
		}
		line, err := reader.ReadSlice('\n')
		if len(line) == 0 && err == io.EOF {
			return scan, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) || len(line) > bench.LedgerRowLimit+1 {
			return scan, fmt.Errorf("bench ledger row %d exceeds 16 KiB", len(scan.rows)+1)
		}
		if err != nil {
			if err == io.EOF {
				return scan, fmt.Errorf("bench ledger has partial final row; retained without repair")
			}
			return scan, err
		}
		scan.size += int64(len(line))
		if scan.size > benchLedgerByteLimit || len(scan.rows) >= benchLedgerRowLimit {
			return scan, fmt.Errorf("bench ledger exceeds row or byte capacity")
		}
		body := line[:len(line)-1]
		if len(bytes.TrimSpace(body)) == 0 {
			return scan, fmt.Errorf("bench ledger row %d is blank", len(scan.rows)+1)
		}
		row, err := decode(body)
		if err != nil {
			return scan, fmt.Errorf("bench ledger row %d: %w", len(scan.rows)+1, err)
		}
		if _, exists := scan.byID[row.RunID]; exists {
			return scan, fmt.Errorf("bench ledger contains duplicate run ID %q", row.RunID)
		}
		scan.byID[row.RunID] = len(scan.rows)
		scan.rows = append(scan.rows, row)
	}
}

// benchLedgerParent requires a physical, already existing directory. The
// caller owns policy about separation from source/agent roots; this owner never
// creates directories and never silently canonicalizes a supplied alias.
func benchLedgerParent(path string) (os.FileInfo, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == string(filepath.Separator) || strings.ContainsRune(path, 0) {
		return nil, errors.New("bench ledger path must be absolute and clean")
	}
	parent := filepath.Dir(path)
	physical, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return nil, fmt.Errorf("bench ledger parent must already exist: %w", err)
	}
	if physical != parent {
		return nil, errors.New("bench ledger parent must be canonical without symlink ancestry")
	}
	info, err := os.Stat(parent)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("bench ledger parent is not a directory")
	}
	return info, nil
}

func checkBenchLedgerFile(path string, f *os.File, parent os.FileInfo) error {
	now, err := benchLedgerParent(path)
	if err != nil {
		return err
	}
	if !os.SameFile(parent, now) {
		return errors.New("bench ledger parent identity changed")
	}
	named, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("bench ledger path changed: %w", err)
	}
	opened, err := f.Stat()
	if err != nil {
		return err
	}
	if !named.Mode().IsRegular() || !opened.Mode().IsRegular() || !os.SameFile(named, opened) {
		return errors.New("bench ledger path is not the opened regular file")
	}
	return nil
}

func lockBenchLedger(ctx context.Context, f *os.File) error {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := flock.Exclusive(f, true)
		if err == nil {
			return ctx.Err()
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// openBenchLedger owns only an existing regular inode or its own O_EXCL-created
// file. O_NONBLOCK prevents a raced-in FIFO from blocking before fstat refuses it.
// The returned file holds the lock until Close; no sidecar lock is created.
func openBenchLedger(ctx context.Context, path string, create bool) (file *os.File, created bool, retErr error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	parent, err := benchLedgerParent(path)
	if err != nil {
		return nil, false, err
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}
	absent := errors.Is(err, os.ErrNotExist)
	if !absent && !info.Mode().IsRegular() {
		return nil, false, errors.New("bench ledger must be a regular file, not a symlink or special file")
	}
	if absent && !create {
		return nil, false, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	flags := os.O_RDONLY | syscall.O_NOFOLLOW | syscall.O_NONBLOCK
	if create {
		flags = os.O_RDWR | os.O_APPEND | syscall.O_NOFOLLOW | syscall.O_NONBLOCK
	}
	if absent && create {
		file, err = os.OpenFile(path, flags|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			created = true
		} else if errors.Is(err, os.ErrExist) {
			file, err = os.OpenFile(path, flags, 0)
		}
	} else {
		file, err = os.OpenFile(path, flags, 0)
	}
	if err != nil {
		return nil, false, err
	}
	defer func() {
		if retErr != nil {
			_ = file.Close()
		}
	}()
	opened, err := file.Stat()
	if err != nil {
		return file, created, err
	}
	if !opened.Mode().IsRegular() {
		return file, created, errors.New("bench ledger opened a nonregular file")
	}
	if err = lockBenchLedger(ctx, file); err != nil {
		return file, created, err
	}
	if err = checkBenchLedgerFile(path, file, parent); err != nil {
		return file, created, err
	}
	if opened.Size() > benchLedgerByteLimit {
		return file, created, errors.New("bench ledger exceeds 160 MiB capacity")
	}
	return file, created, nil
}

func scanBenchLedgerFile(ctx context.Context, path string, f *os.File) (benchLedgerScan, error) {
	parent, err := benchLedgerParent(path)
	if err != nil {
		return benchLedgerScan{}, err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return benchLedgerScan{}, err
	}
	scan, err := scanBenchLedger(ctx, f, bench.DecodeLedgerRow)
	if err != nil {
		return scan, err
	}
	if err = checkBenchLedgerFile(path, f, parent); err != nil {
		return scan, err
	}
	// A writer that ignores flock cannot silently extend the scanned byte range.
	info, err := f.Stat()
	if err != nil {
		return scan, err
	}
	if info.Size() != scan.size {
		return scan, errors.New("bench ledger changed while scanning")
	}
	return scan, ctx.Err()
}

func readBenchLedger(ctx context.Context, path string) ([]bench.LedgerRow, error) {
	f, _, err := openBenchLedger(ctx, path, false)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, nil
	}
	defer f.Close()
	scan, err := scanBenchLedgerFile(ctx, path, f)
	if err != nil {
		return nil, err
	}
	return scan.rows, nil
}

// preflightBenchLedger refuses corruption or exhausted capacity before a new
// run. Missing files are admissible and remain absent; reads create nothing.
func preflightBenchLedger(ctx context.Context, path string) error {
	f, _, err := openBenchLedger(ctx, path, false)
	if err != nil {
		return err
	}
	if f == nil {
		return nil
	}
	defer f.Close()
	scan, err := scanBenchLedgerFile(ctx, path, f)
	if err != nil {
		return err
	}
	if len(scan.rows) >= benchLedgerRowLimit || scan.size >= benchLedgerByteLimit {
		return errors.New("bench ledger is full; admission refused")
	}
	return ctx.Err()
}

// recordBenchRow appends once under the inode's lock. Any write/sync failure is
// unconfirmed publication: retained bytes are inspected on retry, never erased.
func recordBenchRow(ctx context.Context, path string, row bench.LedgerRow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := bench.EncodeLedgerRow(row)
	if err != nil {
		return err
	}
	f, created, err := openBenchLedger(ctx, path, true)
	if err != nil {
		return err
	}
	defer f.Close()
	scan, err := scanBenchLedgerFile(ctx, path, f)
	if err != nil {
		return err
	}
	if index, exists := scan.byID[row.RunID]; exists {
		previous := scan.rows[index]
		old, err := bench.EncodeLedgerRow(previous)
		if err != nil {
			return err
		}
		if previous.RawDigest != row.RawDigest || !bytes.Equal(raw, old) {
			return fmt.Errorf("bench ledger run %q has conflicting content", row.RunID)
		}
		// Sync on an exact retry also discharges an earlier lost sync reply.
		if err := ctx.Err(); err != nil {
			return err
		}
		if err = f.Sync(); err != nil {
			return err
		}
		return syncBenchLedgerParent(ctx, path)
	}
	if len(scan.rows) >= benchLedgerRowLimit || scan.size+int64(len(raw))+1 > benchLedgerByteLimit {
		return errors.New("bench ledger capacity exhausted")
	}
	parent, err := benchLedgerParent(path)
	if err != nil {
		return err
	}
	if err = checkBenchLedgerFile(path, f, parent); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	line := append(raw, '\n')
	n, err := f.Write(line)
	if err != nil {
		return fmt.Errorf("bench ledger append unconfirmed; retained bytes: %w", err)
	}
	if n != len(line) {
		return fmt.Errorf("bench ledger append unconfirmed: %w", io.ErrShortWrite)
	}
	if err = f.Sync(); err != nil {
		return fmt.Errorf("bench ledger sync unconfirmed: %w", err)
	}
	if err = checkBenchLedgerFile(path, f, parent); err != nil {
		return fmt.Errorf("bench ledger append identity unconfirmed: %w", err)
	}
	if created {
		if err = syncBenchLedgerParent(ctx, path); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func syncBenchLedgerParent(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	expected, err := benchLedgerParent(path)
	if err != nil {
		return err
	}
	parent, err := os.OpenFile(filepath.Dir(path), os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	defer parent.Close()
	opened, err := parent.Stat()
	if err != nil {
		return err
	}
	if !opened.IsDir() || !os.SameFile(expected, opened) {
		return errors.New("bench ledger parent changed before sync")
	}
	if err = parent.Sync(); err != nil {
		return fmt.Errorf("bench ledger parent sync unconfirmed: %w", err)
	}
	current, err := benchLedgerParent(path)
	if err != nil {
		return err
	}
	if !os.SameFile(expected, current) {
		return errors.New("bench ledger parent changed during sync")
	}
	return ctx.Err()
}

package activetime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// SourcePrefix binds replay to the bytes available at measurement time.
type SourcePrefix struct {
	Bytes  int64
	Digest string
}

func readSessionPrefix(path string, pins []SourcePrefix) ([]byte, error) {
	return readSessionPrefixContext(nil, path, pins)
}

func readSessionPrefixContext(ctx context.Context, path string, pins []SourcePrefix) ([]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	if len(pins) == 0 && ctx == nil {
		return os.ReadFile(path)
	}
	if len(pins) > 0 && (len(pins) != 1 || pins[0].Bytes <= 0) {
		return nil, fmt.Errorf("invalid source prefix for %s", path)
	}
	var pin SourcePrefix
	if len(pins) > 0 {
		pin = pins[0]
		want, err := hex.DecodeString(pin.Digest)
		if err != nil || len(want) != sha256.Size {
			return nil, fmt.Errorf("invalid source digest for %s", path)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var source io.Reader = &contextReader{ctx: ctx, reader: f}
	if len(pins) > 0 {
		source = io.LimitReader(source, pin.Bytes)
	}
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, err
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	if len(pins) == 0 {
		return data, nil
	}
	if int64(len(data)) != pin.Bytes {
		return nil, fmt.Errorf("source %s shortened since measurement", path)
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != pin.Digest {
		return nil, fmt.Errorf("source %s changed since measurement", path)
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	return data, nil
}

// contextReader checks cancellation between bounded regular-file reads.
type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := contextError(r.ctx); err != nil {
		return 0, err
	}
	if len(p) > 32*1024 {
		p = p[:32*1024]
	}
	return r.reader.Read(p)
}
func contextError(ctx context.Context) error {
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

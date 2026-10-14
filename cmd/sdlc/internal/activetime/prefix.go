package activetime

import (
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
	if len(pins) == 0 {
		return os.ReadFile(path)
	}
	if len(pins) != 1 || pins[0].Bytes <= 0 {
		return nil, fmt.Errorf("invalid source prefix for %s", path)
	}
	pin := pins[0]
	want, err := hex.DecodeString(pin.Digest)
	if err != nil || len(want) != sha256.Size {
		return nil, fmt.Errorf("invalid source digest for %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, pin.Bytes))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) != pin.Bytes {
		return nil, fmt.Errorf("source %s shortened since measurement", path)
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != pin.Digest {
		return nil, fmt.Errorf("source %s changed since measurement", path)
	}
	return data, nil
}

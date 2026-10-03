package layergraph

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// DeclarationLimit is the byte bound for one construct/deps document.
const DeclarationLimit int64 = 1 << 20

// ReadDeclaration reads a dependency document safely: it must be an ordinary
// file (never a symlink, never a FIFO that would block), at most maxBytes,
// checked before and after opening so a replacement in between is refused.
// The one reader weave and sdlc share (#289).
func ReadDeclaration(path string, maxBytes int64) ([]byte, error) {
	st, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("dependency declaration %s must be an ordinary file", path)
	}
	if st.Size() > maxBytes {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, maxBytes)
	}
	// Avoid following a replacement symlink or blocking on a replacement FIFO.
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open dependency declaration %s: %w", path, err)
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	if st, err = f.Stat(); err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("dependency declaration %s must be an ordinary file", path)
	}
	if st.Size() > maxBytes {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, maxBytes)
	}
	content, err := io.ReadAll(io.LimitReader(f, maxBytes))
	if err != nil {
		return nil, err
	}
	var extra [1]byte
	n, err := f.Read(extra[:])
	if n > 0 {
		return nil, fmt.Errorf("dependency declaration %s exceeds byte limit %d", path, maxBytes)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return content, nil
}

package staging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

var ErrSetupInUse = errors.New("environment setup is active")

// AcquireSetup excludes other cooperative setup calls in this environment.
// The lock file is permanent: removing it would allow two independently locked
// inodes. Close the returned file, never explicitly unlock it, so inherited
// descriptors keep excluding setup after their parent exits.
func AcquireSetup(environment string) (*os.File, error) {
	path := filepath.Join(environment, ".weave-setup.lock")
	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, fmt.Errorf("open setup lease for %s: %w", environment, err)
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("inspect setup lease for %s: %w", environment, err)
		}
		return nil, fmt.Errorf("invalid setup lease %s: expected regular file", path)
	}
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("%w in %s; retry after the active setup finishes", ErrSetupInUse, environment)
		}
		return nil, fmt.Errorf("lock setup lease for %s: %w", environment, err)
	}
	return f, nil
}

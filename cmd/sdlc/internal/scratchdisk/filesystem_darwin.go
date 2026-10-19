package scratchdisk

import "syscall"

// Darwin exposes the filesystem name directly; no capability writes are needed.
func fileSystemName(fs *syscall.Statfs_t) string {
	var name []byte
	for _, c := range fs.Fstypename {
		if c == 0 {
			break
		}
		name = append(name, byte(c))
	}
	return string(name)
}

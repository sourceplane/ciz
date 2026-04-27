//go:build darwin

package runner

import (
	"golang.org/x/sys/unix"
)

// ioCloneFile uses APFS clonefile(2) for true copy-on-write on macOS. Falls
// back to error so the caller does a byte copy on non-APFS or cross-fs cases.
func ioCloneFile(src, dst string) error {
	return unix.Clonefile(src, dst, 0)
}

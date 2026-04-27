//go:build linux

package runner

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// ioCloneFile uses the FICLONE ioctl (btrfs, xfs reflinks, bcachefs, …) for
// copy-on-write on Linux. Returns an error on filesystems that don't support
// it so the caller falls back to byte copy.
func ioCloneFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if err := unix.IoctlFileClone(int(out.Fd()), int(in.Fd())); err != nil {
		out.Close()
		_ = os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(dst)
		return err
	}
	return nil
}

var _ = errors.New

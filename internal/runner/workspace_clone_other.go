//go:build !darwin && !linux

package runner

import "errors"

// ioCloneFile is unsupported on this platform; caller will byte-copy.
func ioCloneFile(src, dst string) error {
	return errors.New("clonefile unsupported on this platform")
}

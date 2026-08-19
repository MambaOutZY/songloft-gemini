//go:build !windows

package fileutil

import (
	"errors"
	"syscall"
)

func isCrossDeviceError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.EXDEV
	}
	return false
}

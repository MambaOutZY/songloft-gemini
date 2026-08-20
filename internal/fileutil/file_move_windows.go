//go:build windows

package fileutil

import (
	"errors"
	"syscall"
)

const errNotSameDevice = syscall.Errno(17) // Windows ERROR_NOT_SAME_DEVICE

func isCrossDeviceError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.EXDEV || errno == errNotSameDevice
	}
	return false
}

//go:build linux

package main

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// vrfControl returns a Control function that binds sockets to a VRF device
// via SO_BINDTODEVICE. Returns nil when vrfName is empty.
func vrfControl(vrfName string) func(network, address string, c syscall.RawConn) error {
	if vrfName == "" {
		return nil
	}
	return func(network, address string, c syscall.RawConn) error {
		var sysErr error
		err := c.Control(func(fd uintptr) {
			sysErr = unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, vrfName)
		})
		if err != nil {
			return err
		}
		return sysErr
	}
}

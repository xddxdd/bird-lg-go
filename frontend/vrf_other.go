//go:build !linux

package main

import "syscall"

// vrfControl returns a Control function that binds sockets to a VRF device.
// On non-Linux platforms, VRF is not supported; panics if vrfName is non-empty.
func vrfControl(vrfName string) func(network, address string, c syscall.RawConn) error {
	if vrfName == "" {
		return nil
	}
	panic("VRF binding is only supported on Linux")
}

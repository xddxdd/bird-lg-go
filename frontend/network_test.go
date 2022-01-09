package main

import (
	"net"
	"testing"
	"time"
)

const (
	NETWORK_UNKNOWN = 0
	NETWORK_DOWN = 1
	NETWORK_UP = 2
)

var networkState int = NETWORK_UNKNOWN
func checkNetwork(t *testing.T) {
	if networkState == NETWORK_UNKNOWN {
		conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 1*time.Second)
		if err != nil {
			networkState = NETWORK_DOWN
		} else {
			networkState = NETWORK_UP
			conn.Close()
		}
	}

	if networkState == NETWORK_DOWN {
		t.Skipf("Test skipped for network error")
	}
}

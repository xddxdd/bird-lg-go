package main

import (
	"io"
	"net"
	"time"
)

// Send a whois request
func whois(s string) string {
	if setting.whoisServer == "" {
		return ""
	}

	conn, err := net.DialTimeout("tcp", setting.whoisServer+":43", 5*time.Second)
	if err != nil {
		return err.Error()
	}
	defer conn.Close()

	conn.Write([]byte(s + "\r\n"))

	buf := make([]byte, 65536)
	n, err := io.ReadFull(conn, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err.Error()
	}
	return string(buf[:n])
}

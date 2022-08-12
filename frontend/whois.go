package main

import (
	"io"
	"net"
	"os/exec"
	"strings"
	"time"
)

// Send a whois request
func whois(s string) string {
	if setting.whoisServer == "" {
		return ""
	}

	if strings.HasPrefix(setting.whoisServer, "/") {
		cmd := exec.Command(setting.whoisServer, s)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return err.Error()
		}
		if len(output) > 65535 {
			output = output[:65535]
		}
		return string(output)
	} else {
		buf := make([]byte, 65536)
		conn, err := net.DialTimeout("tcp", setting.whoisServer+":43", 5*time.Second)
		if err != nil {
			return err.Error()
		}
		defer conn.Close()

		conn.Write([]byte(s + "\r\n"))

		n, err := io.ReadFull(conn, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return err.Error()
		}
		return string(buf[:n])
	}

}

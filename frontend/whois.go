package main

import (
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
)

// Send a whois request
func whois(s string) string {
	if setting.whoisServer == "" {
		return ""
	}

	if strings.HasPrefix(setting.whoisServer, "/") {
		args, err := shlex.Split(setting.whoisServer)
		if err != nil {
			return err.Error()
		}
		args = append(args, s)

		cmd := exec.Command(args[0], args[1:]...)
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

		whoisServer := setting.whoisServer
		if !strings.Contains(whoisServer, ":") {
			whoisServer = whoisServer + ":43"
		}

		conn, err := net.DialTimeout("tcp", whoisServer, 5*time.Second)
		if err != nil {
			return err.Error()
		}
		defer conn.Close()

		conn.Write([]byte(s + "\r\n"))

		n, err := io.ReadFull(conn, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err.Error()
		}
		return string(buf[:n])
	}

}

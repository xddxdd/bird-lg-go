package main

import (
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
)

// addDefaultWhoisPort adds the default whois port (43) if not specified.
// Handles IPv4, IPv6 (bare and bracketed), and domain names.
func addDefaultWhoisPort(server string) string {
	if _, _, err := net.SplitHostPort(server); err != nil {
		// No port specified, add default whois port
		// Strip brackets from IPv6 addresses like [::1] before JoinHostPort adds them back
		host := server
		if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
			host = host[1 : len(host)-1]
		}
		return net.JoinHostPort(host, "43")
	}
	return server
}

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
		if len(output) > 65535 {
			output = output[:65535]
		}
		if err != nil {
			return err.Error() + "\n" + string(output)
		} else {
			return string(output)
		}
	} else {
		buf := make([]byte, 65536)

		whoisServer := addDefaultWhoisPort(setting.whoisServer)

		conn, err := (&net.Dialer{Timeout: 5 * time.Second, Control: vrfControl(setting.vrf)}).Dial("tcp", whoisServer)
		if err != nil {
			return err.Error()
		}
		defer conn.Close()

		conn.Write([]byte(s + "\r\n"))

		n, err := io.ReadFull(conn, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err.Error() + "\n" + string(buf[:n])
		}
		return string(buf[:n])
	}

}

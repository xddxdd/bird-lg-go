package main

import (
	"context"
	"net"
	"os"
	"strings"
)

type settingType struct {
	servers           []string
	serversDisplay    []string
	domain            string
	proxyPort         int
	whoisServer       string
	listen            []string
	dnsInterface      string
	netSpecificMode   string
	titleBrand        string
	navBarBrand       string
	navBarBrandURL    string
	navBarAllServer   string
	navBarAllURL      string
	bgpmapInfo        string
	telegramBotName   string
	protocolFilter    []string
	nameFilter        string
	timeOut           int
	connectionTimeOut int
	trustProxyHeaders bool
	vrf               string
}

var setting settingType

func main() {
	parseSettings()
	ImportTemplates()

	for _, listenAddr := range setting.listen {
		go func(listenAddr string) {
			var l net.Listener
			var err error

			if strings.HasPrefix(listenAddr, "/") {
				// Delete existing socket file, ignore errors (will fail later anyway)
				os.Remove(listenAddr)
				l, err = net.Listen("unix", listenAddr)
			} else {
				if !strings.Contains(listenAddr, ":") {
					listenAddr = ":" + listenAddr
				}
				lc := net.ListenConfig{Control: vrfControl(setting.vrf)}
				l, err = lc.Listen(context.Background(), "tcp", listenAddr)
			}

			if err != nil {
				panic(err)
			}

			webServerStart(l)
		}(listenAddr)
	}

	select {}
}

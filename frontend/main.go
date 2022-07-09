package main

import (
	"net"
	"os"
	"strings"
)

type settingType struct {
	servers         []string
	serversDisplay  []string
	domain          string
	proxyPort       int
	whoisServer     string
	listen          string
	dnsInterface    string
	netSpecificMode string
	titleBrand      string
	navBarBrand     string
	navBarBrandURL  string
	navBarAllServer string
	navBarAllURL    string
	bgpmapInfo      string
	telegramBotName string
	protocolFilter  []string
	nameFilter      string
	timeOut         int
}

var setting settingType

func main() {
	parseSettings()
	ImportTemplates()

	var l net.Listener
	var err error

	if strings.HasPrefix(setting.listen, "/") {
		// Delete existing socket file, ignore errors (will fail later anyway)
		os.Remove(setting.listen)
		l, err = net.Listen("unix", setting.listen)
	} else {
		listenAddr := setting.listen
		if !strings.Contains(listenAddr, ":") {
			listenAddr = ":" + listenAddr
		}
		l, err = net.Listen("tcp", listenAddr)
	}

	if err != nil {
		panic(err)
	}

	webServerStart(l)
}

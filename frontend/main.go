package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type settingType struct {
	servers     []string
	domain      string
	proxyPort   int
	whoisServer string
	listen      string
}

var setting settingType

func main() {
	var settingDefault = settingType{
		[]string{""}, "", 8000, "whois.verisign-grs.com", ":5000",
	}

	if serversEnv := os.Getenv("BIRDLG_SERVERS"); serversEnv != "" {
		settingDefault.servers = strings.Split(serversEnv, ",")
	}
	if domainEnv := os.Getenv("BIRDLG_DOMAIN"); domainEnv != "" {
		settingDefault.domain = domainEnv
	}
	if proxyPortEnv := os.Getenv("BIRDLG_PROXY_PORT"); proxyPortEnv != "" {
		var err error
		if settingDefault.proxyPort, err = strconv.Atoi(proxyPortEnv); err != nil {
			panic(err)
		}
	}
	if whoisEnv := os.Getenv("BIRDLG_WHOIS"); whoisEnv != "" {
		settingDefault.whoisServer = whoisEnv
	}
	if listenEnv := os.Getenv("BIRDLG_LISTEN"); listenEnv != "" {
		settingDefault.listen = listenEnv
	}

	serversPtr := flag.String("servers", strings.Join(settingDefault.servers, ","), "server name prefixes, separated by comma")
	domainPtr := flag.String("domain", settingDefault.domain, "server name domain suffixes")
	proxyPortPtr := flag.Int("proxy-port", settingDefault.proxyPort, "port bird-lgproxy is running on")
	whoisPtr := flag.String("whois", settingDefault.whoisServer, "whois server for queries")
	listenPtr := flag.String("listen", settingDefault.listen, "address bird-lg is listening on")
	flag.Parse()

	if *serversPtr == "" {
		panic("no server set")
	} else if *domainPtr == "" {
		panic("no base domain set")
	}

	setting = settingType{
		strings.Split(*serversPtr, ","),
		*domainPtr,
		*proxyPortPtr,
		*whoisPtr,
		*listenPtr,
	}

	webServerStart()
}

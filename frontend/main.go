package main

import (
	"flag"
	"os"
	"strconv"
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
	telegramBotName string
	protocolFilter  []string
}

var setting settingType

func main() {
	var settingDefault = settingType{
		servers:         []string{""},
		proxyPort:       8000,
		whoisServer:     "whois.verisign-grs.com",
		listen:          ":5000",
		dnsInterface:    "asn.cymru.com",
		titleBrand:      "Bird-lg Go",
		navBarBrand:     "Bird-lg Go",
		telegramBotName: "",
		protocolFilter:  []string{},
	}

	if env := os.Getenv("BIRDLG_SERVERS"); env != "" {
		settingDefault.servers = strings.Split(env, ",")
	}
	if env := os.Getenv("BIRDLG_DOMAIN"); env != "" {
		settingDefault.domain = env
	}
	if env := os.Getenv("BIRDLG_PROXY_PORT"); env != "" {
		var err error
		if settingDefault.proxyPort, err = strconv.Atoi(env); err != nil {
			panic(err)
		}
	}
	if env := os.Getenv("BIRDLG_WHOIS"); env != "" {
		settingDefault.whoisServer = env
	}
	if env := os.Getenv("BIRDLG_LISTEN"); env != "" {
		settingDefault.listen = env
	}
	if env := os.Getenv("BIRDLG_DNS_INTERFACE"); env != "" {
		settingDefault.dnsInterface = env
	}
	if env := os.Getenv("BIRDLG_NET_SPECIFIC_MODE"); env != "" {
		settingDefault.netSpecificMode = env
	}
	if env := os.Getenv("BIRDLG_TITLE_BRAND"); env != "" {
		settingDefault.titleBrand = env
		settingDefault.navBarBrand = env
	}
	if env := os.Getenv("BIRDLG_NAVBAR_BRAND"); env != "" {
		settingDefault.navBarBrand = env
	}
	if env := os.Getenv("BIRDLG_TELEGRAM_BOT_NAME"); env != "" {
		settingDefault.telegramBotName = env
	}
	if env := os.Getenv("BIRDLG_PROTOCOL_FILTER"); env != "" {
		settingDefault.protocolFilter = strings.Split(env, ",")
	}

	serversPtr := flag.String("servers", strings.Join(settingDefault.servers, ","), "server name prefixes, separated by comma")
	domainPtr := flag.String("domain", settingDefault.domain, "server name domain suffixes")
	proxyPortPtr := flag.Int("proxy-port", settingDefault.proxyPort, "port bird-lgproxy is running on")
	whoisPtr := flag.String("whois", settingDefault.whoisServer, "whois server for queries")
	listenPtr := flag.String("listen", settingDefault.listen, "address bird-lg is listening on")
	dnsInterfacePtr := flag.String("dns-interface", settingDefault.dnsInterface, "dns zone to query ASN information")
	netSpecificModePtr := flag.String("net-specific-mode", settingDefault.netSpecificMode, "network specific operation mode, [(none)|dn42]")
	titleBrandPtr := flag.String("title-brand", settingDefault.titleBrand, "prefix of page titles in browser tabs")
	navBarBrandPtr := flag.String("navbar-brand", settingDefault.navBarBrand, "brand to show in the navigation bar")
	telegramBotNamePtr := flag.String("telegram-bot-name", settingDefault.telegramBotName, "telegram bot name (used to filter @bot commands)")
	protocolFilterPtr := flag.String("protocol-filter", strings.Join(settingDefault.protocolFilter, ","),
		"protocol types to show in summary tables (comma separated list); defaults to all if not set")
	flag.Parse()

	if *serversPtr == "" {
		panic("no server set")
	}

	servers := strings.Split(*serversPtr, ",")
	serversDisplay := strings.Split(*serversPtr, ",")

	protocolFilter := []string{}
	// strings.Split returns [""] for empty inputs; we want the list to remain empty in these cases
	if len(*protocolFilterPtr) > 0 {
		protocolFilter = strings.Split(*protocolFilterPtr, ",")
	}

	// Split server names of the form "DisplayName<Hostname>"
	for i, server := range servers {
		pos := strings.Index(server, "<")
		if pos != -1 {
			serversDisplay[i] = server[0:pos]
			servers[i] = server[pos+1 : len(server)-1]
		}
	}

	setting = settingType{
		servers,
		serversDisplay,
		*domainPtr,
		*proxyPortPtr,
		*whoisPtr,
		*listenPtr,
		*dnsInterfacePtr,
		strings.ToLower(*netSpecificModePtr),
		*titleBrandPtr,
		*navBarBrandPtr,
		*telegramBotNamePtr,
		protocolFilter,
	}

	ImportTemplates()
	webServerStart()
}

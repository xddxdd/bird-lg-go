package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type viperSettingType struct {
	Servers         string `mapstructure:"servers"`
	Domain          string `mapstructure:"domain"`
	ProxyPort       int    `mapstructure:"proxy_port"`
	WhoisServer     string `mapstructure:"whois"`
	Listen          string `mapstructure:"listen"`
	DNSInterface    string `mapstructure:"dns_interface"`
	NetSpecificMode string `mapstructure:"net_specific_mode"`
	TitleBrand      string `mapstructure:"title_brand"`
	NavBarBrand     string `mapstructure:"navbar_brand"`
	NavBarBrandURL  string `mapstructure:"navbar_brand_url"`
	NavBarAllServer string `mapstructure:"navbar_all_servers"`
	NavBarAllURL    string `mapstructure:"navbar_all_url"`
	BgpmapInfo      string `mapstructure:"bgpmap_info"`
	TelegramBotName string `mapstructure:"telegram_bot_name"`
	ProtocolFilter  string `mapstructure:"protocol_filter"`
	NameFilter      string `mapstructure:"name_filter"`
	TimeOut         int    `mapstructure:"timeout"`
}

// Parse settings with viper, and convert to legacy setting format
func parseSettings() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/bird-lg")
	viper.SetConfigName("bird-lg")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("birdlg")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	pflag.String("servers", "", "server name prefixes, separated by comma")
	viper.BindPFlag("servers", pflag.Lookup("servers"))

	pflag.String("domain", "", "server name domain suffixes")
	viper.BindPFlag("domain", pflag.Lookup("domain"))

	pflag.Int("proxy-port", 8000, "port bird-lgproxy is running on")
	viper.BindPFlag("proxy_port", pflag.Lookup("proxy-port"))

	pflag.String("whois", "whois.verisign-grs.com", "whois server for queries")
	viper.BindPFlag("whois", pflag.Lookup("whois"))

	pflag.String("listen", "5000", "address or unix socket bird-lg is listening on")
	viper.BindPFlag("listen", pflag.Lookup("listen"))

	pflag.String("dns-interface", "asn.cymru.com", "dns zone to query ASN information")
	viper.BindPFlag("dns_interface", pflag.Lookup("dns-interface"))

	pflag.String("net-specific-mode", "", "network specific operation mode, [(none)|dn42]")
	viper.BindPFlag("net_specific-mode", pflag.Lookup("net-specific-mode"))

	pflag.String("title-brand", "Bird-lg Go", "prefix of page titles in browser tabs")
	viper.BindPFlag("title_brand", pflag.Lookup("title-brand"))

	pflag.String("navbar-brand", "", "brand to show in the navigation bar")
	viper.BindPFlag("navbar_brand", pflag.Lookup("navbar-brand"))

	pflag.String("navbar-brand-url", "/", "the url of the brand to show in the navigation bar")
	viper.BindPFlag("navbar_brand_url", pflag.Lookup("navbar-brand-url"))

	pflag.String("navbar-all-servers", "All Servers", "the text of \"All servers\" button in the navigation bar")
	viper.BindPFlag("navbar_all_servers", pflag.Lookup("navbar-all-servers"))

	pflag.String("navbar-all-url", "all", "the URL of \"All servers\" button")
	viper.BindPFlag("navbar_all_url", pflag.Lookup("navbar-all-url"))

	pflag.String("bgpmap-info", "asn,as-name,ASName,descr", "the infos displayed in bgpmap, separated by comma, start with \":\" means allow multiline")
	viper.BindPFlag("bgpmap_info", pflag.Lookup("bgpmap-info"))

	pflag.String("telegram-bot-name", "", "telegram bot name (used to filter @bot commands)")
	viper.BindPFlag("telegram_bot_name", pflag.Lookup("telegram-bot-name"))

	pflag.String("protocol-filter", "",
		"protocol types to show in summary tables (comma separated list); defaults to all if not set")
	viper.BindPFlag("protocol_filter", pflag.Lookup("protocol-filter"))

	pflag.String("name-filter", "", "protocol name regex to hide in summary tables (RE2 syntax); defaults to none if not set")
	viper.BindPFlag("name_filter", pflag.Lookup("name-filter"))

	pflag.Int("time-out", 120, "time before request timed out, in seconds; defaults to 120 if not set")
	viper.BindPFlag("timeout", pflag.Lookup("time-out"))

	pflag.Parse()

	if err := viper.ReadInConfig(); err != nil {
		println("Warning on reading config: " + err.Error())
	}

	viperSettings := viperSettingType{}
	if err := viper.Unmarshal(&viperSettings); err != nil {
		panic(err)
	}

	setting.servers = strings.Split(viperSettings.Servers, ",")
	setting.serversDisplay = strings.Split(viperSettings.Servers, ",")
	// Split server names of the form "DisplayName<Hostname>"
	for i, server := range setting.servers {
		pos := strings.Index(server, "<")
		if pos != -1 {
			setting.serversDisplay[i] = server[0:pos]
			setting.servers[i] = server[pos+1 : len(server)-1]
		}
	}

	setting.domain = viperSettings.Domain
	setting.proxyPort = viperSettings.ProxyPort
	setting.whoisServer = viperSettings.WhoisServer
	setting.listen = viperSettings.Listen
	setting.dnsInterface = viperSettings.DNSInterface
	setting.netSpecificMode = viperSettings.NetSpecificMode
	setting.titleBrand = viperSettings.TitleBrand

	setting.navBarBrand = viperSettings.NavBarBrand
	if setting.navBarBrand == "" {
		setting.navBarBrand = setting.titleBrand
	}

	setting.navBarBrandURL = viperSettings.NavBarBrandURL
	setting.navBarAllServer = viperSettings.NavBarAllServer
	setting.navBarAllURL = viperSettings.NavBarAllURL
	setting.bgpmapInfo = viperSettings.BgpmapInfo
	setting.telegramBotName = viperSettings.TelegramBotName

	if viperSettings.ProtocolFilter != "" {
		setting.protocolFilter = strings.Split(viperSettings.ProtocolFilter, ",")
	} else {
		setting.protocolFilter = []string{}
	}

	setting.nameFilter = viperSettings.NameFilter
	setting.timeOut = viperSettings.TimeOut

	fmt.Printf("%#v\n", setting)
}

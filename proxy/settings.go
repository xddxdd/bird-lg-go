package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type viperSettingType struct {
	BirdSocket      string   `mapstructure:"bird_socket"`
	Listen          []string `mapstructure:"listen"`
	AllowedNets     string   `mapstructure:"allowed_ips"`
	TracerouteBin   string   `mapstructure:"traceroute_bin"`
	TracerouteFlags string   `mapstructure:"traceroute_flags"`
	TracerouteRaw   bool     `mapstructure:"traceroute_raw"`
}

// Parse settings with viper, and convert to legacy setting format
func parseSettings() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/bird-lg")
	viper.SetConfigName("bird-lgproxy")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("birdlg")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// Legacy environment variables without prefixes
	viper.BindEnv("bird_socket", "BIRD_SOCKET")
	viper.BindEnv("listen", "BIRDLG_LISTEN", "BIRDLG_PROXY_PORT")
	viper.BindEnv("allowed_ips", "ALLOWED_IPS")

	pflag.String("bird", "/var/run/bird/bird.ctl", "socket file for bird, set either in parameter or environment variable BIRD_SOCKET")
	viper.BindPFlag("bird_socket", pflag.Lookup("bird"))

	pflag.StringSlice("listen", []string{"8000"}, "listen address, set either in parameter or environment variable BIRDLG_PROXY_PORT")
	viper.BindPFlag("listen", pflag.Lookup("listen"))

	pflag.String("allowed", "", "IPs or networks allowed to access this proxy, separated by commas. Don't set to allow all IPs.")
	viper.BindPFlag("allowed_ips", pflag.Lookup("allowed"))

	pflag.String("traceroute_bin", "", "traceroute binary file, set either in parameter or environment variable BIRDLG_TRACEROUTE_BIN")
	viper.BindPFlag("traceroute_bin", pflag.Lookup("traceroute_bin"))

	pflag.String("traceroute_flags", "", "traceroute flags, supports multiple flags separated with space.")
	viper.BindPFlag("traceroute_flags", pflag.Lookup("traceroute_flags"))

	pflag.Bool("traceroute_raw", false, "whether to display traceroute outputs raw; set via parameter or environment variable BIRDLG_TRACEROUTE_RAW")
	viper.BindPFlag("traceroute_raw", pflag.Lookup("traceroute_raw"))

	pflag.Parse()

	if err := viper.ReadInConfig(); err != nil {
		println("Warning on reading config: " + err.Error())
	}

	viperSettings := viperSettingType{}
	if err := viper.Unmarshal(&viperSettings); err != nil {
		panic(err)
	}

	setting.birdSocket = viperSettings.BirdSocket
	setting.listen = viperSettings.Listen

	if viperSettings.AllowedNets != "" {
		for _, arg := range strings.Split(viperSettings.AllowedNets, ",") {

			// if argument is an IP address, convert to CIDR by adding a suitable mask
			if !strings.Contains(arg, "/") {
				if strings.Contains(arg, ":") {
					// IPv6 address with /128 mask
					arg += "/128"
				} else {
					// IPv4 address with /32 mask
					arg += "/32"
				}
			}

			// parse the network
			_, netip, err := net.ParseCIDR(arg)
			if err != nil {
				fmt.Printf("Failed to parse CIDR %s: %s\n", arg, err.Error())
				continue
			}
			setting.allowedNets = append(setting.allowedNets, netip)

		}
	} else {
		setting.allowedNets = []*net.IPNet{}
	}

	var err error
	setting.tr_bin = viperSettings.TracerouteBin
	setting.tr_flags, err = shlex.Split(viperSettings.TracerouteFlags)
	if err != nil {
		panic(err)
	}

	setting.tr_raw = viperSettings.TracerouteRaw

	fmt.Printf("%#v\n", setting)
}

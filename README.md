# Bird-lg-go

An alternative implementation for [bird-lg](https://github.com/sileht/bird-lg) written in Go. Both frontend and backend (proxy) are implemented, and can work with either the original Python implementation or the Go implementation.

> The code on master branch no longer support BIRDv1. Branch "bird1" is the last version that supports BIRDv1.

## Table of Contents

   * [Bird-lg-go](#bird-lg-go)
      * [Table of Contents](#table-of-contents)
      * [Frontend](#frontend)
      * [Proxy](#proxy)
      * [Advanced Features](#advanced-features)
         * [API](#api)
         * [Telegram Bot Webhook](#telegram-bot-webhook)
            * [Example of setting the webhook](#example-of-setting-the-webhook)
            * [Supported commands](#supported-commands)
      * [Credits](#credits)
      * [License](#license)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Build Instructions

Run `make` to build binaries for both the frontend and the proxy. You need to have Go installed on your machine.

Optionally run `make install` to install them to `/usr/local/bin`.

Or, you can manually do the building steps:

```bash
# Build frontend binary
cd frontend
go get -u github.com/kevinburke/go-bindata/...
go generate
go build -ldflags "-w -s" -o frontend
cd ..

# Build proxy binary
cd proxy
go build -ldflags "-w -s" -o proxy
cd ..
```

- If you get `undefined: MustAssetString`, you need to uninstall an older version of go-bindata from your machine: see [#11](https://github.com/xddxdd/bird-lg-go/issues/11)

## Frontend

The frontend directory contains the code for the web frontend, where users see BGP states, do traceroutes and whois, etc. It's a replacement for "lg.py" in original bird-lg project.

Features implemented:

- Show peering status (`show protocol` command)
- Query route (`show route for ...`, `show route where net ~ [ ... ]`)
- Whois and traceroute
- Work with both Python proxy (lgproxy.py) and Go proxy (proxy dir of this project)
- Visualize AS paths as picture (bgpmap feature)

Usage: all configuration is done via commandline parameters or environment variables, no config file.

| Parameter | Environment Variable | Description |
| --------- | -------------------- | ----------- |
| --servers | BIRDLG_SERVERS | server name prefixes, separated by comma |
| --domain | BIRDLG_DOMAIN | server name domain suffixes |
| --listen | BIRDLG_LISTEN | address bird-lg is listening on (default ":5000") |
| --proxy-port | BIRDLG_PROXY_PORT | port bird-lgproxy is running on (default 8000) |
| --whois | BIRDLG_WHOIS | whois server for queries (default "whois.verisign-grs.com") |
| --dns-interface | BIRDLG_DNS_INTERFACE | dns zone to query ASN information (default "asn.cymru.com") |
| --title-brand | BIRDLG_TITLE_BRAND | prefix of page titles in browser tabs (default "Bird-lg Go") |
| --navbar-brand | BIRDLG_NAVBAR_BRAND | brand to show in the navigation bar (default "Bird-lg Go") |

Example: the following command starts the frontend with 2 BIRD nodes, with domain name "gigsgigscloud.dn42.lantian.pub" and "hostdare.dn42.lantian.pub", and proxies are running on port 8000 on both nodes.

    ./frontend --servers=gigsgigscloud,hostdare --domain=dn42.lantian.pub --proxy-port=8000

Example: the following docker-compose.yml entry does the same as above, but by starting a Docker container:

    services:
      bird-lg:
        image: xddxdd/bird-lg-go
        container_name: bird-lg
        restart: always
        environment:
          - BIRDLG_SERVERS=gigsgigscloud,hostdare
          - BIRDLG_DOMAIN=dn42.lantian.pub
        ports:
          - "5000:5000"

Demo: https://lg.lantian.pub

## Proxy

The proxy directory contains the code for the "proxy" for bird commands and traceroutes. It's a replacement for "lgproxy.py" in original bird-lg project.

Features implemented:

- Sending queries to BIRD
- Sending "restrict" command to BIRD to prevent unauthorized changes
- Executing traceroute command on Linux, FreeBSD and OpenBSD
- Source IP restriction

Usage: all configuration is done via commandline parameters or environment variables, no config file.

| Parameter | Environment Variable | Description |
| --------- | -------------------- | ----------- |
| --allowed | ALLOWED_IPS | IPs allowed to access this proxy, separated by commas. Don't set to allow all IPs. (default "") |
| --bird | BIRD_SOCKET | socket file for bird, set either in parameter or environment variable BIRD_SOCKET (default "/var/run/bird/bird.ctl") |
| --listen | BIRDLG_LISTEN | listen address, set either in parameter or environment variable BIRDLG_LISTEN (default ":8000") |

Example: start proxy with default configuration, should work "out of the box" on Debian 9 with BIRDv1:

    ./proxy

Example: start proxy with custom bird socket location:

    ./proxy --bird /run/bird.ctl

Example: the following docker-compose.yml entry does the same as above, but by starting a Docker container:

    bird-lgproxy:
      image: xddxdd/bird-lgproxy-go
      container_name: bird-lgproxy
      restart: always
      volumes:
        - "/run/bird.ctl:/var/run/bird/bird.ctl"
      ports:
        - "192.168.0.1:8000:8000"

You can use source IP restriction to increase security. You should also bind the proxy to a specific interface and use an external firewall/iptables for added security.

## Advanced Features

### Display names

The server parameter is composed of server name prefixes, separated by comma. It also supports an extended syntax: It allows to define display names for the user interface that are different from the actual server names.

For instance, the two servers from the basic example can be displayed as "Gigs" and "Hostdare" using the following syntax (as known from email addresses):

    ./frontend --servers="Gigs<gigsgigscloud>,Hostdare<hostdare>" --domain=dn42.lantian.pub

### IP addresses

You may also specify IP addresses as server names when no domain is specified. IPv6 link local addresses are supported, too.

For example:

    ./frontend --servers="Prod<prod.mydomain.local>,Test1<fd88:dead:beef::1>,Test2<fe80::c%wg0>" --domain=

These three servers are displayed as "Prod", "Test1" and "Test2" in the user interface. 

### API

The frontend provides an API for running BIRD/traceroute/whois queries.

See [API docs](API.md) for detailed information.

### Telegram Bot Webhook

The frontend can act as a Telegram Bot webhook endpoint, to add BGP route/traceroute/whois lookup functionality to your tech group.

There is no configuration necessary on the frontend, just start it up normally.

Set your Telegram Bot webhook URL to `https://your.frontend.com/telegram/alpha+beta+gamma`, where `alpha+beta+gamma` is the list of servers to be queried on Telegram commands, separated by `+`.

You may omit `alpha+beta+gamma` to use all your servers, but it is not recommended when you have lots of servers, or the message would be too long and hard to read.

#### Example of setting the webhook

```bash
curl "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook?url=https://your.frontend.com:5000/telegram/alpha+beta+gamma"
```

#### Supported commands

- `path`: Show bird's ASN path to target IP
- `route`: Show bird's preferred route to target IP
- `trace`: Traceroute to target IP/domain
- `whois`: Whois query

## Credits

- Everyone who contributed to this project (see Contributors section on the right)
- Mehdi Abaakouk for creating [the original bird-lg project](https://github.com/sileht/bird-lg)
- [Bootstrap](https://getbootstrap.com/) as web UI framework

## License

GPL 3.0

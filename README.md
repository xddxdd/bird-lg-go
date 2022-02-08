# Bird-lg-go

An alternative implementation for [bird-lg](https://github.com/sileht/bird-lg) written in Go. Both frontend and backend (proxy) are implemented, and can work with either the original Python implementation or the Go implementation.

> The code on master branch no longer support BIRDv1. Branch "bird1" is the last version that supports BIRDv1.

## Table of Contents

- [Bird-lg-go](#bird-lg-go)
  - [Table of Contents](#table-of-contents)
  - [Build Instructions](#build-instructions)
    - [Build Docker Images](#build-docker-images)
  - [Frontend](#frontend)
  - [Proxy](#proxy)
  - [Advanced Features](#advanced-features)
    - [Display names](#display-names)
    - [IP addresses](#ip-addresses)
    - [API](#api)
    - [Telegram Bot Webhook](#telegram-bot-webhook)
  - [Credits](#credits)
  - [License](#license)

## Build Instructions

You need to have **Go 1.16 or newer** installed on your machine.

Run `make` to build binaries for both the frontend and the proxy.

Optionally run `make install` to install them to `/usr/local/bin` (`bird-lg-go` and `bird-lgproxy-go`).

### Build Docker Images

Use the Dockerfiles in `frontend` and `proxy` directory.

Ready-to-use images are available at:

- Frontend: <https://hub.docker.com/r/xddxdd/bird-lg-go>
- Proxy: <https://hub.docker.com/r/xddxdd/bird-lgproxy-go>

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
| --listen | BIRDLG_LISTEN | address bird-lg is listening on (default "5000") |
| --proxy-port | BIRDLG_PROXY_PORT | port bird-lgproxy is running on (default 8000) |
| --whois | BIRDLG_WHOIS | whois server for queries (default "whois.verisign-grs.com") |
| --dns-interface | BIRDLG_DNS_INTERFACE | dns zone to query ASN information (default "asn.cymru.com") |
| --bgpmap-info | BIRDLG_BGPMAP_INFO | the infos displayed in bgpmap, separated by comma, start with `:` means allow multiline (default "asn,as-name,ASName,descr") |
| --title-brand | BIRDLG_TITLE_BRAND | prefix of page titles in browser tabs (default "Bird-lg Go") |
| --navbar-brand | BIRDLG_NAVBAR_BRAND | brand to show in the navigation bar (default "Bird-lg Go") |
| --navbar-brand-url | BIRDLG_NAVBAR_BRAND_URL | the url of the brand to show in the navigation bar (default "/") |
| --navbar-all-servers | BIRDLG_NAVBAR_ALL_SERVERS | the text of "All servers" button in the navigation bar (default "ALL Servers") |
| --navbar-all-url | BIRDLG_NAVBAR_ALL_URL | the URL of "All servers" button (default "all") |
| --net-specific-mode | BIRDLG_NET_SPECIFIC_MODE | apply network-specific changes for some networks, use "dn42" for BIRD in dn42 network |
| --protocol-filter | BIRDLG_PROTOCOL_FILTER | protocol types to show in summary tables (comma separated list); defaults to all if not set |
| --name-filter | BIRDLG_NAME_FILTER | protocol names to hide in summary tables (RE2 syntax); defaults to none if not set |
| --time-out | BIRDLG_TIMEOUT | time before request timed out, in seconds; defaults to 120 if not set |

Example: the following command starts the frontend with 2 BIRD nodes, with domain name "gigsgigscloud.dn42.lantian.pub" and "hostdare.dn42.lantian.pub", and proxies are running on port 8000 on both nodes.

```bash
./frontend --servers=gigsgigscloud,hostdare --domain=dn42.lantian.pub --proxy-port=8000
```

Example: the following docker-compose.yml entry does the same as above, but by starting a Docker container:

```yaml
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
```

Demo: <https://lg.lantian.pub>

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
| --listen | BIRDLG_PROXY_PORT | listen address, set either in parameter or environment variable  BIRDLG_PROXY_PORT(default "8000") |
| --traceroute_bin | BIRDLG_TRACEROUTE_BIN | traceroute binary file, set either in parameter or environment variable  BIRDLG_TRACEROUTE_BIN(default "traceroute") |
| --traceroute_raw | BIRDLG_TRACEROUTE_RAW | whether to display traceroute outputs raw (default false) |

Example: start proxy with default configuration, should work "out of the box" on Debian 9 with BIRDv1:

```bash
./proxy
```

Example: start proxy with custom bird socket location:

```bash
./proxy --bird /run/bird.ctl
```

Example: the following docker-compose.yml entry does the same as above, but by starting a Docker container:

```yaml
services:
  bird-lgproxy:
    image: xddxdd/bird-lgproxy-go
    container_name: bird-lgproxy
    restart: always
    volumes:
      - "/run/bird.ctl:/var/run/bird/bird.ctl"
    ports:
      - "192.168.0.1:8000:8000"
```

You can use source IP restriction to increase security. You should also bind the proxy to a specific interface and use an external firewall/iptables for added security.

## Advanced Features

### Display names

The server parameter is composed of server name prefixes, separated by comma. It also supports an extended syntax: It allows to define display names for the user interface that are different from the actual server names.

For instance, the two servers from the basic example can be displayed as "Gigs" and "Hostdare" using the following syntax (as known from email addresses):

```bash
./frontend --servers="Gigs<gigsgigscloud>,Hostdare<hostdare>" --domain=dn42.lantian.pub
```

### IP addresses

You may also specify IP addresses as server names when no domain is specified. IPv6 link local addresses are supported, too.

For example:

```bash
./frontend --servers="Prod<prod.mydomain.local>,Test1<fd88:dead:beef::1>,Test2<fe80::c%wg0>" --domain=
```

These three servers are displayed as "Prod", "Test1" and "Test2" in the user interface.

### API

The frontend provides an API for running BIRD/traceroute/whois queries.

See [API docs](docs/API.md) for detailed information.

### Telegram Bot Webhook

The frontend can act as a Telegram Bot webhook endpoint, to add BGP route/traceroute/whois lookup functionality to your tech group.

See [Telegram docs](docs/Telegram.md) for detailed information.

## Credits

- Everyone who contributed to this project (see Contributors section on the right)
- Mehdi Abaakouk for creating [the original bird-lg project](https://github.com/sileht/bird-lg)
- [Bootstrap](https://getbootstrap.com/) as web UI framework

## License

GPL 3.0

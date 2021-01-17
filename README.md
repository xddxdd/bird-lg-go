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
            * [Request fields](#request-fields)
            * [Response fields (when type is summary)](#response-fields-when-type-is-summary)
               * [Fields for apiSummaryResultPair](#fields-for-apisummaryresultpair)
               * [Fields for SummaryRowData](#fields-for-summaryrowdata)
               * [Example response](#example-response)
            * [Response fields (when type is bird, traceroute or whois)](#response-fields-when-type-is-bird-traceroute-or-whois)
               * [Fields for apiGenericResultPair](#fields-for-apigenericresultpair)
               * [Example response](#example-response-1)
         * [Telegram Bot Webhook](#telegram-bot-webhook)
            * [Example of setting the webhook](#example-of-setting-the-webhook)
            * [Supported commands](#supported-commands)
      * [Credits](#credits)
      * [License](#license)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

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

### API

The frontend provides an API for running BIRD/traceroute/whois queries.

API Endpoint: `https://your.frontend.com:5000/api/` (the last slash must not be omitted!)

Requests are sent as POSTS with JSON bodies.

#### Request fields

| Name | Type | Value |
| ---- | ---- | -------- |
| `servers` | `[]string` | List of servers to be queried |
| `type` | `string` | Can be `summary`, `bird`, `traceroute` or `whois` |
| `args` | `string` | Arguments to be passed, see below |

Argument examples for each type:

- `summary`: `args` is ignored. Recommended to set to empty string.
- `bird`: `args` is the command to be passed to bird, e.g. `show route for 8.8.8.8`
- `traceroute`: `args` is the traceroute target, e.g. `8.8.8.8` or `google.com`
- `whois`: `args` is the whois target, e.g. `8.8.8.8` or `google.com`

Example request:

```json
{
    "servers": [
        "alpha"
    ],
    "type": "bird",
    "args": "show route for 8.8.8.8"
}
```

#### Response fields (when `type` is `summary`)

| Name | Type | Value |
| ---- | ---- | -------- |
| `error` | `string` | Error message when something is wrong. Empty when everything good |
| `result` | array of `apiSummaryResultPair` | See below |

##### Fields for `apiSummaryResultPair`

| Name | Type | Value |
| ---- | ---- | -------- |
| `server` | `string` | Name of the server |
| `data` | array of `SummaryRowData` | Summaries of the server, see below |

##### Fields for `SummaryRowData`

All fields below is 1:1 correspondent to the output of `birdc show protocols`.

| Name | Type |
| ---- | ---- |
| `name` | `string` |
| `proto` | `string` |
| `table` | `string` |
| `state` | `string` |
| `since` | `string` |
| `info` | `string` |

##### Example response

Request:
```json
{
    "servers": [
        "alpha"
    ],
    "type": "summary",
    "args": ""
}
```

Response:

```json
{
    "error": "",
    "result": [
        {
            "server": "alpha",
            "data": [
                {
                    "name": "bgp1",
                    "proto": "BGP",
                    "table": "---",
                    "state": "start",
                    "since": "2021-01-15 22:40:01",
                    "info": "Active        Socket: Operation timed out"
                },
                {
                    "name": "bgp2",
                    "proto": "BGP",
                    "table": "---",
                    "state": "start",
                    "since": "2021-01-03 08:15:48",
                    "info": "Established"
                }
            ]
        }
    ]
}
```

#### Response fields (when `type` is `bird`, `traceroute` or `whois`)

| Name | Type | Value |
| ---- | ---- | -------- |
| `error` | `string` | Error message, empty when everything is good |
| `result` | array of `apiGenericResultPair` | See below |

##### Fields for `apiGenericResultPair`

| Name | Type | Value |
| ---- | ---- | -------- |
| `server` | `string` | Name of the server; is empty when type is `whois` |
| `data` | `string` | Result from the server |

##### Example response

Request:

```json
{
    "servers": [
        "alpha"
    ],
    "type": "bird",
    "args": "show status"
}
```

Response:

```json
{
    "error": "",
    "result": [
        {
            "server": "alpha",
            "data": "BIRD v2.0.7-137-g61dae32b\nRouter ID is 1.2.3.4\nCurrent server time is 2021-01-17 04:21:14.792\nLast reboot on 2021-01-03 08:15:48.494\nLast reconfiguration on 2021-01-17 00:49:10.573\nDaemon is up and running\n"
        }
    ]
}
```

### Telegram Bot Webhook

The frontend can act as a Telegram Bot webhook endpoint, to add BGP route/traceroute/whois lookup functionality to your tech group.

There is no configuration necessary on the frontend, just start it up normally.

Set your Telegram Bot webhook URL to `https://your.frontend.com:5000/telegram/alpha+beta+gamma`, where `alpha+beta+gamma` is the list of servers to be queried on Telegram commands, separated by `+`.

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

# Bird-lg-go API documentation

The frontend provides an API for running BIRD/traceroute/whois queries.

API Endpoint: `https://your.frontend.com/api/` (the last slash must not be omitted!)

Requests are sent as POSTS with JSON bodies.

## Table of Contents

   * [Bird-lg-go API documentation](#bird-lg-go-api-documentation)
      * [Table of Contents](#table-of-contents)
      * [Request fields](#request-fields)
         * [Example request of type bird](#example-request-of-type-bird)
         * [Example request of type server_list](#example-request-of-type-server_list)
      * [Response fields (when type is summary)](#response-fields-when-type-is-summary)
         * [Fields for apiSummaryResultPair](#fields-for-apisummaryresultpair)
         * [Fields for SummaryRowData](#fields-for-summaryrowdata)
         * [Example response](#example-response)
      * [Response fields (when type is bird, traceroute, whois or server_list)](#response-fields-when-type-is-bird-traceroute-whois-or-server_list)
         * [Fields for apiGenericResultPair](#fields-for-apigenericresultpair)
         * [Example response of type bird](#example-response-of-type-bird)
         * [Example response of type server_list](#example-response-of-type-server_list)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Request fields

| Name | Type | Value |
| ---- | ---- | -------- |
| `servers` | array of `string` | List of servers to be queried |
| `type` | `string` | Can be `summary`, `bird`, `traceroute`, `whois` or `server_list` |
| `args` | `string` | Arguments to be passed, see below |

Argument examples for each type:

- `summary`: `args` is ignored. Recommended to set to empty string.
- `bird`: `args` is the command to be passed to bird, e.g. `show route for 8.8.8.8`
- `traceroute`: `args` is the traceroute target, e.g. `8.8.8.8` or `google.com`
- `whois`: `args` is the whois target, e.g. `8.8.8.8` or `google.com`
- `server_list`: `args` is ignored. In addition, `servers` is also ignored.

### Example request of type `bird`

```json
{
    "servers": [
        "alpha"
    ],
    "type": "bird",
    "args": "show route for 8.8.8.8"
}
```

### Example request of type `server_list`

```json
{
    "servers": [],
    "type": "server_list",
    "args": ""
}
```

## Response fields (when `type` is `summary`)

| Name | Type | Value |
| ---- | ---- | -------- |
| `error` | `string` | Error message when something is wrong. Empty when everything is good |
| `result` | array of `apiSummaryResultPair` | See below |

### Fields for `apiSummaryResultPair`

| Name | Type | Value |
| ---- | ---- | -------- |
| `server` | `string` | Name of the server |
| `data` | array of `SummaryRowData` | Summaries of the server, see below |

### Fields for `SummaryRowData`

All fields below is 1:1 correspondent to the output of `birdc show protocols`.

| Name | Type |
| ---- | ---- |
| `name` | `string` |
| `proto` | `string` |
| `table` | `string` |
| `state` | `string` |
| `since` | `string` |
| `info` | `string` |

### Example response

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

## Response fields (when `type` is `bird`, `traceroute`, `whois` or `server_list`)

| Name | Type | Value |
| ---- | ---- | -------- |
| `error` | `string` | Error message, empty when everything is good |
| `result` | array of `apiGenericResultPair` | See below |

### Fields for `apiGenericResultPair`

| Name | Type | Value |
| ---- | ---- | -------- |
| `server` | `string` | Name of the server; is empty when type is `whois` |
| `data` | `string` | Result from the server; is empty when type is `server_list` |

### Example response of type `bird`

Request:

```json
{
    "servers": [],
    "type": "server_list",
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
            "data": "BIRD v2.0.7-137-g61dae32b\nRouter ID is 1.2.3.4\nCurrent server time is 2021-01-17 04:21:14.792\nLast reboot on 2021-01-03 08:15:48.494\nLast reconfiguration on 2021-01-17 00:49:10.573\nDaemon is up and running\n"
        }
    ]
}
```

### Example response of type `server_list`

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
            "server": "gigsgigscloud",
            "data": ""
        }
    ]
}
```

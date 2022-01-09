# Telegram Bot Webhook

The frontend can act as a Telegram Bot webhook endpoint, to add BGP route/traceroute/whois lookup functionality to your tech group.

There is no configuration necessary on the frontend, just start it up normally.

Set your Telegram Bot webhook URL to `https://your.frontend.com/telegram/alpha+beta+gamma`, where `alpha+beta+gamma` is the list of servers to be queried on Telegram commands, separated by `+`.

You may omit `alpha+beta+gamma` to use all your servers, but it is not recommended when you have lots of servers, or the message would be too long and hard to read.

## Example of setting the webhook

```bash
curl "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook?url=https://your.frontend.com:5000/telegram/alpha+beta+gamma"
```

## Supported commands

- `path`: Show bird's ASN path to target IP
- `route`: Show bird's preferred route to target IP
- `trace`: Traceroute to target IP/domain
- `whois`: Whois query

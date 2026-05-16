SMTP2HTTP (email-to-web)
========================
smtp2http is a simple SMTP server that forwards incoming email to a webhook as an HTTP POST (JSON). Each recipient domain can use its own webhook URL.

Dev
===
- `go mod vendor`
- `go build`

Routes
======
Configure one webhook per domain, either with a JSON file or repeated `-route` flags (CLI routes override the file for the same domain).

**Config file** (`routes.json`):

```json
{
  "example.com": "http://localhost:8080/webhooks/example",
  "other.org": "http://localhost:8080/webhooks/other"
}
```

See `routes.example.json`.

**CLI routes:**

```
smtp2http -route=example.com=http://localhost:8080/hooks/example -route=other.org=http://localhost:8080/hooks/other
```

**Mixed:**

```
smtp2http -config=routes.json -route=example.com=http://override.example/hook
```

Incoming mail is accepted only when the RCPT TO address domain matches a configured route. The message is POSTed to that domain's webhook.

Optional timeouts make local testing with `telnet` easier:

```
smtp2http -config=routes.json -timeout.read=50 -timeout.write=50
```

Telnet example (`telnet localhost 25`):

```
HELO zeus
# smtp answer

MAIL FROM:<email@from.com>
# smtp answer

RCPT TO:<youremail@example.com>
# smtp answer

DATA
your mail content
.

```

Usage
=====
`smtp2http -listen=:25 -config=routes.json`
`smtp2http -help`

Testing locally
===============
1. Copy `routes.example.json` to `routes.json` and adjust webhook URLs.
2. Start a mock webhook (returns HTTP 200):

```
powershell -File scripts/test-webhook.ps1
```

3. In another terminal, run smtp2http on a non-privileged port:

```
smtp2http -listen=:2525 -config=routes.json -timeout.read=60 -timeout.write=60
```

4. Send test mail:

```
python scripts/send-test-mail.py 127.0.0.1 2525 user@example.com
python scripts/send-test-mail.py 127.0.0.1 2525 admin@other.org
```

The mock webhook prints each JSON payload. Mail to an unconfigured domain is rejected.

Contribution
============
Original repo from @alash3al
Thanks to @aranajuan

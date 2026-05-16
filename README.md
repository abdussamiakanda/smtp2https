# smtp2https

**smtp2https** is an SMTP receiver that accepts inbound email and forwards each message to an HTTPS webhook as a JSON `POST`. Routes are defined per recipient domain, so a single instance can serve multiple domains, each with its own endpoint and optional API key.

## Features

- Per-domain webhook routing via JSON config or CLI flags
- Optional `X-Api-Key` header for authenticated webhooks
- Parses headers, body (plain/HTML), and attachments (Base64-encoded in JSON)
- SPF result included in the payload
- Rejects mail when the recipient domain is not configured or the webhook does not return HTTP `200`

## Requirements

- Go 1.13 or newer
- Network access to your webhook endpoints
- For production on port **25**: root privileges or `CAP_NET_BIND_SERVICE` (ports below 1024)

## Installation

```bash
git clone <repository-url>
cd smtp2https
go mod download
go build -o smtp2https .
```

## Configuration

Copy the example routes file and edit it for your environment:

```bash
cp routes.example.json routes.json
```

### `routes.json`

Each key is a **recipient domain** (the part after `@` in the RCPT TO address). The value is either a webhook URL string or an object with `webhook` and an optional `api_key`.

```json
{
  "mail.example.com": "https://api.example.com/email/incoming",
  "other.example.com": {
    "webhook": "https://automation.example.com/webhook/incoming",
    "api_key": "your-secret-key"
  }
}
```

When `api_key` is set, outbound requests include:

```http
X-Api-Key: <api_key>
Content-Type: application/json
```

CLI `-route` entries override the config file for the same domain but do not support API keys; use `routes.json` when authentication is required.

### Command-line flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | *(none)* | Path to `routes.json` |
| `-route` | *(repeatable)* | `domain=webhookURL` (overrides file for that domain) |
| `-listen` | `:smtp` | SMTP listen address (use `:25` in production) |
| `-name` | `smtp2https` | SMTP banner / server name |
| `-msglimit` | `2097152` | Maximum message size (bytes) |
| `-timeout.read` | `5` | Read timeout (seconds) |
| `-timeout.write` | `5` | Write timeout (seconds) |

Run `./smtp2https -help` for the full list.

## Running

### Foreground

```bash
./smtp2https -listen=:25 -config=routes.json
```

### systemd

Create `/etc/systemd/system/smtp2https.service`:

```ini
[Unit]
Description=smtp2https — SMTP to HTTPS forwarder
After=network.target

[Service]
ExecStart=/opt/smtp2https/smtp2https -listen=:25 -config=/opt/smtp2https/routes.json
WorkingDirectory=/opt/smtp2https
Restart=always
User=root

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now smtp2https
sudo journalctl -u smtp2https -f
```

Ensure your domain **MX records** point to this host and that port **25** is open in the firewall.

## Webhook behavior

- Method: `POST`
- Body: JSON representation of the parsed email
- Success: webhook must respond with **HTTP 200** or the SMTP transaction is rejected
- Failure: connection errors and non-`200` responses are logged; the sender receives a generic error

## Local testing

Use port **2525** for SMTP so you do not need administrator/root access (port 25 is restricted on Linux, macOS, and Windows).

1. Copy and edit routes: `cp routes.example.json routes.json` — point each domain at `http://127.0.0.1:8080/...` (or your mock path).
2. **Terminal 1** — start a mock webhook (logs POST bodies, returns HTTP `200`).
3. **Terminal 2** — run smtp2https.
4. **Terminal 3** — send a test message.

The recipient domain in RCPT TO (`user@example.com` → `example.com`) must exist as a key in `routes.json`.

### Linux / macOS

**Terminal 1 — mock webhook:**

```bash
python3 scripts/test-webhook.py 8080
```

**Terminal 2 — smtp2https:**

```bash
./smtp2https -listen=:2525 -config=routes.json -timeout.read=60 -timeout.write=60
```

**Terminal 3 — send mail:**

```bash
python3 scripts/send-test-mail.py 127.0.0.1 2525 user@example.com
python3 scripts/send-test-mail.py 127.0.0.1 2525 admin@other.org
```

Optional webhook check with curl:

```bash
curl -i -X POST http://127.0.0.1:8080/webhooks/example \
  -H "Content-Type: application/json" \
  -d '{"test":true}'
```

### Windows

**Terminal 1 — mock webhook** (PowerShell):

```powershell
powershell -File scripts\test-webhook.ps1
```

Listens on `http://127.0.0.1:8080/` by default. Same behavior as `test-webhook.py`: prints each POST body and returns `200`.

**Terminal 2 — smtp2https:**

```powershell
.\smtp2https.exe -listen=:2525 -config=routes.json -timeout.read=60 -timeout.write=60
```

**Terminal 3 — send mail** (Python):

```powershell
python scripts\send-test-mail.py 127.0.0.1 2525 user@example.com
python scripts\send-test-mail.py 127.0.0.1 2525 admin@other.org
```

On Windows you can use `python` or `python3`, depending on your install.

## Acknowledgments

**smtp2https** is derived from [smtp2http](https://github.com/alash3al/smtp2http) by [@alash3al](https://github.com/alash3al).
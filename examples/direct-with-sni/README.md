# Example: direct-with-sni

This example shows a direct SSH WebSocket tunneling setup with TLS Server Name Indication (SNI).

## Config file

The sample file is `config.json`.

```json
{
  "host": "ssh-server.com",
  "port": 443,
  "auth": {
    "user": "ssh-username",
    "pass": "ssh-password"
  },
  "local": {
    "type": "http",
    "port": 8080
  },
  "payload": "GET / HTTP/1.1\r\nHost: bug-host.com\r\nUpgrade: websocket\r\n\r\n",
  "timeout": 30,
  "tls": {
    "sni": "bug-host.com"
  }
}
```

## What this config does

- `host`: connects to the remote server at `ssh-server.com`.
- `port`: connects over port `443`, which is typically used for HTTPS/TLS.
- `auth.user`: SSH username used by the tunnel.
- `auth.pass`: SSH password used by the tunnel.
- `local.type`: local listener type; `http` means `tunn` exposes an HTTP proxy locally.
- `local.port`: local proxy listens on port `8080`.
- `payload`: raw HTTP request sent to the remote host to initiate a WebSocket connection.
  - `Host: bug-host.com` is the host name used in the WebSocket upgrade request.
  - `Upgrade: websocket` requests a WebSocket tunnel.
- `timeout`: connection timeout in seconds.
- `tls.sni`: sets the TLS Server Name Indication header to `bug-host.com`.
  - This is useful when the remote TLS server expects a specific hostname during the TLS handshake.

## Usage

Run `tunn` with this config:

```bash
tunn -config examples/direct-with-sni/config.json
```

Then point your local HTTP proxy client to `127.0.0.1:8080`.

# Example: direct

This example shows a direct SSH WebSocket tunneling setup using `tunn`.

## Config file

The sample file is `config.json`.

```json
{
  "host": "vps-server.com",
  "port": 80,
  "auth": {
    "user": "ssh-user",
    "pass": "ssh-passwrod"
  },
  "local": {
    "type": "http",
    "port": 8080
  },
  "payload": "GET / HTTP/1.1\r\nHost: bug-host.com\r\nUpgrade: websocket\r\n\r\n",
  "timeout": 30
}
```

## What this config does

- `host`: connects to the remote server at `vps-server.com`.
- `port`: connects on port `80`, using plain HTTP for the WebSocket handshake.
- `auth.user`: SSH username used by the tunnel.
- `auth.pass`: SSH password used by the tunnel.
- `local.type`: local listener type; `http` means `tunn` exposes an HTTP proxy locally.
- `local.port`: local proxy listens on port `8080`.
- `payload`: raw HTTP request sent to the remote host to initiate a WebSocket connection.
  - `Host: bug-host.com` is the host name used in the WebSocket upgrade request.
  - `Upgrade: websocket` requests a WebSocket tunnel.
- `timeout`: connection timeout in seconds.

## Usage

Run `tunn` with this config:

```bash
tunn -config examples/direct/config.json
```

Then point your local HTTP proxy client to `127.0.0.1:8080`.

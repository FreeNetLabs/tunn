# Example: proxy

This example shows a SSH WebSocket tunneling setup over CDN proxy using `tunn`.

## Config file

The sample file is `config.json`.

```json
{
  "host": "bug-host.com",
  "port": 80,
  "auth": {
    "user": "ssh-username",
    "pass": "ssh-possword"
  },
  "local": {
    "type": "http",
    "port": 2080
  },
  "payload": "GET / HTTP/1.1\r\nHost: vps-server.com\r\nUpgrade: websocket\r\n\r\n",
  "timeout": 30
}
```

## What this config does

- `host`: connects to the remote server at `bug-host.com`.
- `port`: connects over port `80`, using plain HTTP for the WebSocket handshake.
- `auth.user`: SSH username used by the tunnel.
- `auth.pass`: SSH password used by the tunnel.
- `local.type`: local listener type; `http` means `tunn` exposes an HTTP proxy locally.
- `local.port`: local proxy listens on port `2080`.
- `payload`: raw HTTP request sent to the remote host to initiate a WebSocket connection.
  - `Host: vps-server.com` is the host name used in the WebSocket upgrade request.
  - `Upgrade: websocket` requests a WebSocket tunnel.
- `timeout`: connection timeout in seconds.

## Usage

Run `tunn` with this config:

```bash
tunn -config examples/proxy/config.json
```

Then point your local HTTP proxy client to `127.0.0.1:2080`.

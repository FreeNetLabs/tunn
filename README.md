# Tunn

A minimal, cross-platform SSH WebSocket tunneling tool for free internet setups.

## Installation

```bash
go install github.com/FreeNetLabs/tunn@latest
```

## Usage

1. Create a `config.json` file:

```json
{
  "host": "example.com",
  "port": 22,
  "auth": {
    "user": "root",
    "pass": "secret"
  },
  "local": {
    "type": "http",
    "port": 8080
  },
  "payload": "GET / HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\n\r\n",
  "timeout": 30,
  "tls": {
    "sni": "example.com"
  }
}
```

2. Run the tunnel:

```bash
tunn -c config.json
```

3. Connect your applications through the local proxy at `127.0.0.1:8080`.

## Examples

Check the `examples/` folder for configs setups and explanations:

- `examples/direct/` — direct HTTP WebSocket tunnel
- `examples/direct-with-sni/` — direct tunnel using TLS SNI
- `examples/proxy/` — proxy-style HTTP WebSocket tunnel
- `examples/proxy-with-sni/` — proxy tunnel using TLS SNI

## License

MIT License

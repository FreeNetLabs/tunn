# Tunn

A minimal, cross-platform SSH WebSocket tunneling tool for free internet.

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
    "port": 1080
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
tunn -config config.json
```

3. Connect your applications through the local proxy at `127.0.0.1:1080`.

## License

MIT License


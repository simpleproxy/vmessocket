# vmessocket
- forked from [v2ray-core](https://github.com/v2fly/v2ray-core)
- generated from the source code [v4.44.0](https://github.com/v2fly/v2ray-core/archive/refs/tags/v4.44.0.zip)
- a stable implementation of vmess and websocket which is mostly used by network providers

## Compilation

### Windows

```bash
go build -o xray.exe -trimpath -ldflags "-s -w -buildid=" ./main
```

### Linux / macOS

```bash
go build -o xray -trimpath -ldflags "-s -w -buildid=" ./main
```

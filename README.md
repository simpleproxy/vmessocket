# vmessocket
- forked from [v2fly](https://github.com/v2fly)
- generated from the source code [v4.44.0](https://github.com/v2fly)
- a stable implementation of vmess and websocket which is mostly used by network providers

## Compilation

### Windows

```bash
go build -o vmessocket.exe -trimpath -ldflags "-s -w -buildid=" ./main
```

### Linux / macOS

```bash
go build -o vmessocket -trimpath -ldflags "-s -w -buildid=" ./main
```
![Go](https://img.shields.io:/github/go-mod/go-version/vmessocket/vmessocket)

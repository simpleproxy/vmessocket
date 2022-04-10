package all

import (
	_ "github.com/vmessocket/vmessocket/app/commander"
	_ "github.com/vmessocket/vmessocket/app/dispatcher"
	_ "github.com/vmessocket/vmessocket/app/log"
	_ "github.com/vmessocket/vmessocket/app/log/command"
	_ "github.com/vmessocket/vmessocket/app/proxyman/command"
	_ "github.com/vmessocket/vmessocket/app/proxyman/inbound"
	_ "github.com/vmessocket/vmessocket/app/proxyman/outbound"
	_ "github.com/vmessocket/vmessocket/main/confloader/external"
	_ "github.com/vmessocket/vmessocket/main/jsonem"
	_ "github.com/vmessocket/vmessocket/proxy/dns"
	_ "github.com/vmessocket/vmessocket/proxy/freedom"
	_ "github.com/vmessocket/vmessocket/proxy/http"
	_ "github.com/vmessocket/vmessocket/proxy/vmess/inbound"
	_ "github.com/vmessocket/vmessocket/proxy/vmess/outbound"
	_ "github.com/vmessocket/vmessocket/transport/internet/tcp"
	_ "github.com/vmessocket/vmessocket/transport/internet/udp"
	_ "github.com/vmessocket/vmessocket/transport/internet/websocket"
)

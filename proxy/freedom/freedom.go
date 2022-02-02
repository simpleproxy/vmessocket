package freedom

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"context"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/dice"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/retry"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/common/signal"
	"github.com/vmessocket/vmessocket/common/task"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/dns"
	"github.com/vmessocket/vmessocket/features/policy"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		h := new(Handler)
		if err := core.RequireFeatures(ctx, func(pm policy.Manager, d dns.Client) error {
			return h.Init(config.(*Config), pm, d)
		}); err != nil {
			return nil, err
		}
		return h, nil
	}))
}

type Handler struct {
	policyManager policy.Manager
	dns           dns.Client
	config        *Config
}

func (h *Handler) Init(config *Config, pm policy.Manager, d dns.Client) error {
	h.config = config
	h.policyManager = pm
	h.dns = d

	return nil
}

func (h *Handler) policy() policy.Session {
	p := h.policyManager.ForLevel(h.config.UserLevel)
	if h.config.Timeout > 0 && h.config.UserLevel == 0 {
		p.Timeouts.ConnectionIdle = time.Duration(h.config.Timeout) * time.Second
	}
	return p
}

func (h *Handler) resolveIP(ctx context.Context, domain string, localAddr net.Address) net.Address {
	lookupFunc := h.dns.LookupIP
	if h.config.DomainStrategy == Config_USE_IP4 || (localAddr != nil && localAddr.Family().IsIPv4()) {
		if lookupIPv4, ok := h.dns.(dns.IPv4Lookup); ok {
			lookupFunc = lookupIPv4.LookupIPv4
		}
	} else if h.config.DomainStrategy == Config_USE_IP6 || (localAddr != nil && localAddr.Family().IsIPv6()) {
		if lookupIPv6, ok := h.dns.(dns.IPv6Lookup); ok {
			lookupFunc = lookupIPv6.LookupIPv6
		}
	}

	ips, err := lookupFunc(domain)
	if err != nil {
		newError("failed to get IP address for domain ", domain).Base(err).WriteToLog(session.ExportIDToError(ctx))
	}
	if len(ips) == 0 {
		return nil
	}
	return net.IPAddress(ips[dice.Roll(len(ips))])
}

func isValidAddress(addr *net.IPOrDomain) bool {
	if addr == nil {
		return false
	}

	a := addr.AsAddress()
	return a != net.AnyIP
}

func (h *Handler) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified.")
	}
	destination := outbound.Target
	if h.config.DestinationOverride != nil {
		server := h.config.DestinationOverride.Server
		if isValidAddress(server.Address) {
			destination.Address = server.Address.AsAddress()
		}
		if server.Port != 0 {
			destination.Port = net.Port(server.Port)
		}
	}
	newError("opening connection to ", destination).WriteToLog(session.ExportIDToError(ctx))

	input := link.Reader
	output := link.Writer

	var conn internet.Connection
	err := retry.ExponentialBackoff(5, 100).On(func() error {
		dialDest := destination
		if h.config.useIP() && dialDest.Address.Family().IsDomain() {
			ip := h.resolveIP(ctx, dialDest.Address.Domain(), dialer.Address())
			if ip != nil {
				dialDest = net.Destination{
					Network: dialDest.Network,
					Address: ip,
					Port:    dialDest.Port,
				}
				newError("dialing to ", dialDest).WriteToLog(session.ExportIDToError(ctx))
			}
		}

		rawConn, err := dialer.Dial(ctx, dialDest)
		if err != nil {
			return err
		}
		conn = rawConn
		return nil
	})
	if err != nil {
		return newError("failed to open connection to ", destination).Base(err)
	}
	defer conn.Close()

	plcy := h.policy()
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, plcy.Timeouts.ConnectionIdle)

	requestDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.DownlinkOnly)

		var writer buf.Writer
		if destination.Network == net.Network_TCP {
			writer = buf.NewWriter(conn)
		} else {
			writer = &buf.SequentialWriter{Writer: conn}
		}

		if err := buf.Copy(input, writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to process request").Base(err)
		}

		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(plcy.Timeouts.UplinkOnly)

		var reader buf.Reader
		if destination.Network == net.Network_TCP {
			reader = buf.NewReader(conn)
		} else {
			reader = buf.NewPacketReader(conn)
		}
		if err := buf.Copy(reader, output, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to process response").Base(err)
		}

		return nil
	}

	if err := task.Run(ctx, requestDone, task.OnSuccess(responseDone, task.Close(output))); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

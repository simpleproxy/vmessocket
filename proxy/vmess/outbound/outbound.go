package outbound

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"hash/crc64"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/platform"
	"github.com/vmessocket/vmessocket/common/protocol"
	"github.com/vmessocket/vmessocket/common/retry"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/common/signal"
	"github.com/vmessocket/vmessocket/common/task"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/policy"
	"github.com/vmessocket/vmessocket/proxy/vmess"
	"github.com/vmessocket/vmessocket/proxy/vmess/encoding"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/internet"
)

var (
	aeadDisabled  = false
	enablePadding = false
)

type Handler struct {
	serverList    *protocol.ServerList
	serverPicker  protocol.ServerPicker
	policyManager policy.Manager
}

func New(ctx context.Context, config *Config) (*Handler, error) {
	serverList := protocol.NewServerList()
	for _, rec := range config.Receiver {
		s, err := protocol.NewServerSpecFromPB(rec)
		if err != nil {
			return nil, newError("failed to parse server spec").Base(err)
		}
		serverList.AddServer(s)
	}
	v := core.MustFromContext(ctx)
	handler := &Handler{
		serverList:    serverList,
		serverPicker:  protocol.NewRoundRobinServerPicker(serverList),
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}
	return handler, nil
}

func shouldEnablePadding(s protocol.SecurityType) bool {
	return enablePadding || s == protocol.SecurityType_AES128_GCM || s == protocol.SecurityType_CHACHA20_POLY1305 || s == protocol.SecurityType_AUTO
}

func (h *Handler) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	var rec *protocol.ServerSpec
	var conn internet.Connection
	err := retry.ExponentialBackoff(5, 200).On(func() error {
		rec = h.serverPicker.PickServer()
		rawConn, err := dialer.Dial(ctx, rec.Destination())
		if err != nil {
			return err
		}
		conn = rawConn
		return nil
	})
	if err != nil {
		return newError("failed to find an available destination").Base(err).AtWarning()
	}
	defer conn.Close()
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified").AtError()
	}
	target := outbound.Target
	newError("tunneling request to ", target, " via ", rec.Destination()).WriteToLog(session.ExportIDToError(ctx))
	command := protocol.RequestCommandTCP
	if target.Network == net.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	user := rec.PickUser()
	request := &protocol.RequestHeader{
		Version: encoding.Version,
		User:    user,
		Command: command,
		Address: target.Address,
		Port:    target.Port,
		Option:  protocol.RequestOptionChunkStream,
	}
	account := request.User.Account.(*vmess.MemoryAccount)
	request.Security = account.Security
	if request.Security == protocol.SecurityType_AES128_GCM || request.Security == protocol.SecurityType_NONE || request.Security == protocol.SecurityType_CHACHA20_POLY1305 {
		request.Option.Set(protocol.RequestOptionChunkMasking)
	}
	if shouldEnablePadding(request.Security) && request.Option.Has(protocol.RequestOptionChunkMasking) {
		request.Option.Set(protocol.RequestOptionGlobalPadding)
	}
	if request.Security == protocol.SecurityType_ZERO {
		request.Security = protocol.SecurityType_NONE
		request.Option.Clear(protocol.RequestOptionChunkStream)
		request.Option.Clear(protocol.RequestOptionChunkMasking)
	}
	if account.AuthenticatedLengthExperiment {
		request.Option.Set(protocol.RequestOptionAuthenticatedLength)
	}
	input := link.Reader
	output := link.Writer
	isAEAD := false
	if !aeadDisabled && len(account.AlterIDs) == 0 {
		isAEAD = true
	}
	hashkdf := hmac.New(sha256.New, []byte("VMessBF"))
	hashkdf.Write(account.ID.Bytes())
	behaviorSeed := crc64.Checksum(hashkdf.Sum(nil), crc64.MakeTable(crc64.ISO))
	session := encoding.NewClientSession(ctx, isAEAD, protocol.DefaultIDHash, int64(behaviorSeed))
	sessionPolicy := h.policyManager.ForLevel(request.User.Level)
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)
		writer := buf.NewBufferedWriter(buf.NewWriter(conn))
		if err := session.EncodeRequestHeader(request, writer); err != nil {
			return newError("failed to encode request").Base(err).AtWarning()
		}
		bodyWriter := session.EncodeRequestBody(request, writer)
		if err := buf.CopyOnceTimeout(input, bodyWriter, time.Millisecond*100); err != nil && err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
			return newError("failed to write first payload").Base(err)
		}
		if err := writer.SetBuffered(false); err != nil {
			return err
		}
		if err := buf.Copy(input, bodyWriter, buf.UpdateActivity(timer)); err != nil {
			return err
		}
		if request.Option.Has(protocol.RequestOptionChunkStream) && !account.NoTerminationSignal {
			if err := bodyWriter.WriteMultiBuffer(buf.MultiBuffer{}); err != nil {
				return err
			}
		}
		return nil
	}
	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)
		reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
		header, err := session.DecodeResponseHeader(reader)
		if err != nil {
			return newError("failed to read header").Base(err)
		}
		h.handleCommand(rec.Destination(), header.Command)
		bodyReader := session.DecodeResponseBody(request, reader)
		return buf.Copy(bodyReader, output, buf.UpdateActivity(timer))
	}
	responseDonePost := task.OnSuccess(responseDone, task.Close(output))
	if err := task.Run(ctx, requestDone, responseDonePost); err != nil {
		return newError("connection ends").Base(err)
	}
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
	const defaultFlagValue = "NOT_DEFINED_AT_ALL"
	paddingValue := platform.NewEnvFlag("vmessocket.vmess.padding").GetValue(func() string { return defaultFlagValue })
	if paddingValue != defaultFlagValue {
		enablePadding = true
	}
	isAeadDisabled := platform.NewEnvFlag("vmessocket.vmess.aead.disabled").GetValue(func() string { return defaultFlagValue })
	if isAeadDisabled == "true" {
		aeadDisabled = true
	}
}

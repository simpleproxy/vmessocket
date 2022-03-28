package dispatcher

import (
	"context"
	"sync"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/log"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/outbound"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/pipe"
)

type cachedReader struct {
	sync.Mutex
	reader *pipe.Reader
	cache  buf.MultiBuffer
}

type DefaultDispatcher struct {
	ohm    outbound.Manager
}

func (r *cachedReader) Cache(b *buf.Buffer) {
	mb, _ := r.reader.ReadMultiBufferTimeout(time.Millisecond * 100)
	r.Lock()
	if !mb.IsEmpty() {
		r.cache, _ = buf.MergeMulti(r.cache, mb)
	}
	b.Clear()
	rawBytes := b.Extend(buf.Size)
	n := r.cache.Copy(rawBytes)
	b.Resize(0, int32(n))
	r.Unlock()
}

func (*DefaultDispatcher) Close() error {
	return nil
}

func (d *DefaultDispatcher) Dispatch(ctx context.Context, destination net.Destination) (*transport.Link, error) {
	if !destination.IsValid() {
		panic("Dispatcher: Invalid destination.")
	}
	ob := &session.Outbound{
		Target: destination,
	}
	ctx = session.ContextWithOutbound(ctx, ob)
	inbound, outbound := d.getLink(ctx)
	content := session.ContentFromContext(ctx)
	if content == nil {
		content = new(session.Content)
		ctx = session.ContextWithContent(ctx, content)
	}
	go func() {
		cReader := &cachedReader{
			reader: outbound.Reader.(*pipe.Reader),
		}
		outbound.Reader = cReader
		d.routedDispatch(ctx, outbound, destination)
	}()
	return inbound, nil
}

func (d *DefaultDispatcher) getLink(ctx context.Context) (*transport.Link, *transport.Link) {
	opt := pipe.OptionsFromContext(ctx)
	uplinkReader, uplinkWriter := pipe.New(opt...)
	downlinkReader, downlinkWriter := pipe.New(opt...)
	inboundLink := &transport.Link{
		Reader: downlinkReader,
		Writer: uplinkWriter,
	}
	outboundLink := &transport.Link{
		Reader: uplinkReader,
		Writer: downlinkWriter,
	}
	return inboundLink, outboundLink
}

func (d *DefaultDispatcher) Init(config *Config, om outbound.Manager) error {
	d.ohm = om
	return nil
}

func (r *cachedReader) Interrupt() {
	r.Lock()
	if r.cache != nil {
		r.cache = buf.ReleaseMulti(r.cache)
	}
	r.Unlock()
	r.reader.Interrupt()
}

func (r *cachedReader) readInternal() buf.MultiBuffer {
	r.Lock()
	defer r.Unlock()
	if r.cache != nil && !r.cache.IsEmpty() {
		mb := r.cache
		r.cache = nil
		return mb
	}
	return nil
}

func (r *cachedReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb := r.readInternal()
	if mb != nil {
		return mb, nil
	}
	return r.reader.ReadMultiBuffer()
}

func (r *cachedReader) ReadMultiBufferTimeout(timeout time.Duration) (buf.MultiBuffer, error) {
	mb := r.readInternal()
	if mb != nil {
		return mb, nil
	}
	return r.reader.ReadMultiBufferTimeout(timeout)
}

func (d *DefaultDispatcher) routedDispatch(ctx context.Context, link *transport.Link, destination net.Destination) {
	var handler outbound.Handler
	if forcedOutboundTag := session.GetForcedOutboundTagFromContext(ctx); forcedOutboundTag != "" {
		ctx = session.SetForcedOutboundTagToContext(ctx, "")
		if h := d.ohm.GetHandler(forcedOutboundTag); h != nil {
			newError("taking platform initialized detour [", forcedOutboundTag, "] for [", destination, "]").WriteToLog(session.ExportIDToError(ctx))
			handler = h
		} else {
			newError("non existing tag for platform initialized detour: ", forcedOutboundTag).AtError().WriteToLog(session.ExportIDToError(ctx))
			common.Close(link.Writer)
			common.Interrupt(link.Reader)
			return
		}
	}
	if handler == nil {
		handler = d.ohm.GetDefaultHandler()
	}
	if handler == nil {
		newError("default outbound handler not exist").WriteToLog(session.ExportIDToError(ctx))
		common.Close(link.Writer)
		common.Interrupt(link.Reader)
		return
	}
	if accessMessage := log.AccessMessageFromContext(ctx); accessMessage != nil {
		if tag := handler.Tag(); tag != "" {
			accessMessage.Detour = tag
		}
		log.Record(accessMessage)
	}
	handler.Dispatch(ctx, link)
}

func (*DefaultDispatcher) Start() error {
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		d := new(DefaultDispatcher)
		if err := core.RequireFeatures(ctx, func(om outbound.Manager) error {
			return d.Init(config.(*Config), om)
		}); err != nil {
			return nil, err
		}
		return d, nil
	}))
}

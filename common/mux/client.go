package mux

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/protocol"
	"github.com/vmessocket/vmessocket/common/session"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/common/task"
	"github.com/vmessocket/vmessocket/proxy"
	"github.com/vmessocket/vmessocket/transport"
	"github.com/vmessocket/vmessocket/transport/internet"
)

var (
	muxCoolAddress = net.DomainAddress("v1.mux.cool")
	muxCoolPort    = net.Port(9527)
)

type ClientWorker struct {
	sessionManager *SessionManager
	link           transport.Link
	done           *done.Instance
}

type ClientWorkerFactory interface {
	Create() (*ClientWorker, error)
}

type DialingWorkerFactory struct {
	Proxy  proxy.Outbound
	Dialer internet.Dialer
	ctx    context.Context
}

type IncrementalWorkerPicker struct {
	Factory     ClientWorkerFactory
	access      sync.Mutex
	workers     []*ClientWorker
	cleanupTask *task.Periodic
}

type WorkerPicker interface {
	PickAvailable() (*ClientWorker, error)
}

func fetchInput(ctx context.Context, s *Session, output buf.Writer) {
	dest := session.OutboundFromContext(ctx).Target
	transferType := protocol.TransferTypeStream
	if dest.Network == net.Network_UDP {
		transferType = protocol.TransferTypePacket
	}
	s.transferType = transferType
	writer := NewWriter(s.ID, dest, output, transferType)
	defer s.Close()
	defer writer.Close()
	newError("dispatching request to ", dest).WriteToLog(session.ExportIDToError(ctx))
	if err := writeFirstPayload(s.input, writer); err != nil {
		newError("failed to write first payload").Base(err).WriteToLog(session.ExportIDToError(ctx))
		writer.hasError = true
		common.Interrupt(s.input)
		return
	}
	if err := buf.Copy(s.input, writer); err != nil {
		newError("failed to fetch all input").Base(err).WriteToLog(session.ExportIDToError(ctx))
		writer.hasError = true
		common.Interrupt(s.input)
		return
	}
}

func writeFirstPayload(reader buf.Reader, writer *Writer) error {
	err := buf.CopyOnceTimeout(reader, writer, time.Millisecond*100)
	if err == buf.ErrNotTimeoutReader || err == buf.ErrReadTimeout {
		return writer.WriteMultiBuffer(buf.MultiBuffer{})
	}
	if err != nil {
		return err
	}
	return nil
}

func (m *ClientWorker) ActiveConnections() uint32 {
	return uint32(m.sessionManager.Size())
}

func (p *IncrementalWorkerPicker) cleanup() {
	var activeWorkers []*ClientWorker
	for _, w := range p.workers {
		if !w.Closed() {
			activeWorkers = append(activeWorkers, w)
		}
	}
	p.workers = activeWorkers
}

func (p *IncrementalWorkerPicker) cleanupFunc() error {
	p.access.Lock()
	defer p.access.Unlock()
	if len(p.workers) == 0 {
		return newError("no worker")
	}
	p.cleanup()
	return nil
}

func (m *ClientWorker) Closed() bool {
	return m.done.Done()
}

func (m *ClientWorker) Dispatch(ctx context.Context, link *transport.Link) bool {
	if m.Closed() {
		return false
	}
	sm := m.sessionManager
	s := sm.Allocate()
	if s == nil {
		return false
	}
	s.input = link.Reader
	s.output = link.Writer
	go fetchInput(ctx, s, m.link.Writer)
	return true
}

func (m *ClientWorker) fetchOutput() {
	defer func() {
		common.Must(m.done.Close())
	}()
	reader := &buf.BufferedReader{Reader: m.link.Reader}
	var meta FrameMetadata
	for {
		err := meta.Unmarshal(reader)
		if err != nil {
			if errors.Cause(err) != io.EOF {
				newError("failed to read metadata").Base(err).WriteToLog()
			}
			break
		}
		switch meta.SessionStatus {
		case SessionStatusKeepAlive:
			err = m.handleStatueKeepAlive(&meta, reader)
		case SessionStatusEnd:
			err = m.handleStatusEnd(&meta, reader)
		case SessionStatusNew:
			err = m.handleStatusNew(&meta, reader)
		case SessionStatusKeep:
			err = m.handleStatusKeep(&meta, reader)
		default:
			status := meta.SessionStatus
			newError("unknown status: ", status).AtError().WriteToLog()
			return
		}
		if err != nil {
			newError("failed to process data").Base(err).WriteToLog()
			return
		}
	}
}

func (m *ClientWorker) handleStatueKeepAlive(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if meta.Option.Has(OptionData) {
		return buf.Copy(NewStreamReader(reader), buf.Discard)
	}
	return nil
}

func (m *ClientWorker) handleStatusEnd(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if s, found := m.sessionManager.Get(meta.SessionID); found {
		if meta.Option.Has(OptionError) {
			common.Interrupt(s.input)
			common.Interrupt(s.output)
		}
		s.Close()
	}
	if meta.Option.Has(OptionData) {
		return buf.Copy(NewStreamReader(reader), buf.Discard)
	}
	return nil
}

func (m *ClientWorker) handleStatusKeep(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if !meta.Option.Has(OptionData) {
		return nil
	}
	s, found := m.sessionManager.Get(meta.SessionID)
	if !found {
		closingWriter := NewResponseWriter(meta.SessionID, m.link.Writer, protocol.TransferTypeStream)
		closingWriter.Close()

		return buf.Copy(NewStreamReader(reader), buf.Discard)
	}
	rr := s.NewReader(reader)
	err := buf.Copy(rr, s.output)
	if err != nil && buf.IsWriteError(err) {
		newError("failed to write to downstream. closing session ", s.ID).Base(err).WriteToLog()
		closingWriter := NewResponseWriter(meta.SessionID, m.link.Writer, protocol.TransferTypeStream)
		closingWriter.Close()
		drainErr := buf.Copy(rr, buf.Discard)
		common.Interrupt(s.input)
		s.Close()
		return drainErr
	}
	return err
}

func (m *ClientWorker) handleStatusNew(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if meta.Option.Has(OptionData) {
		return buf.Copy(NewStreamReader(reader), buf.Discard)
	}
	return nil
}

func (m *ClientWorker) monitor() {
	timer := time.NewTicker(time.Second * 16)
	defer timer.Stop()
	for {
		select {
		case <-m.done.Wait():
			m.sessionManager.Close()
			common.Close(m.link.Writer)
			common.Interrupt(m.link.Reader)
			return
		case <-timer.C:
			size := m.sessionManager.Size()
			if size == 0 && m.sessionManager.CloseIfNoSession() {
				common.Must(m.done.Close())
			}
		}
	}
}

func (m *ClientWorker) TotalConnections() uint32 {
	return uint32(m.sessionManager.Count())
}

package mux

import (
	"io"
	"time"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/transport"
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

type WorkerPicker interface {
	PickAvailable() (*ClientWorker, error)
}

func (m *ClientWorker) ActiveConnections() uint32 {
	return uint32(m.sessionManager.Size())
}

func (m *ClientWorker) Closed() bool {
	return m.done.Done()
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
		default:
			status := meta.SessionStatus
			newError("unknown status: ", status).AtError().WriteToLog()
			return
		}
	}
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

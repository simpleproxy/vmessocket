package outbound

import (
	"context"
	"strings"
	"sync"

	"github.com/vmessocket/vmessocket/app/proxyman"
	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/errors"
	"github.com/vmessocket/vmessocket/core"
	"github.com/vmessocket/vmessocket/features/outbound"
)

type Manager struct {
	access           sync.RWMutex
	defaultHandler   outbound.Handler
	taggedHandler    map[string]outbound.Handler
	untaggedHandlers []outbound.Handler
	running          bool
}

func New(ctx context.Context, config *proxyman.OutboundConfig) (*Manager, error) {
	m := &Manager{
		taggedHandler: make(map[string]outbound.Handler),
	}
	return m, nil
}

func (m *Manager) AddHandler(ctx context.Context, handler outbound.Handler) error {
	m.access.Lock()
	defer m.access.Unlock()
	if m.defaultHandler == nil {
		m.defaultHandler = handler
	}
	m.untaggedHandlers = append(m.untaggedHandlers, handler)
	if m.running {
		return handler.Start()
	}
	return nil
}

func (m *Manager) Close() error {
	m.access.Lock()
	defer m.access.Unlock()
	m.running = false
	var errs []error
	for _, h := range m.taggedHandler {
		errs = append(errs, h.Close())
	}
	for _, h := range m.untaggedHandlers {
		errs = append(errs, h.Close())
	}
	return errors.Combine(errs...)
}

func (m *Manager) GetDefaultHandler() outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()
	if m.defaultHandler == nil {
		return nil
	}
	return m.defaultHandler
}

func (m *Manager) GetHandler(tag string) outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()
	if handler, found := m.taggedHandler[tag]; found {
		return handler
	}
	return nil
}

func (m *Manager) RemoveHandler(ctx context.Context, tag string) error {
	if tag == "" {
		return common.ErrNoClue
	}
	m.access.Lock()
	defer m.access.Unlock()
	delete(m.taggedHandler, tag)
	return nil
}

func (m *Manager) Select(selectors []string) []string {
	m.access.RLock()
	defer m.access.RUnlock()
	tags := make([]string, 0, len(selectors))
	for tag := range m.taggedHandler {
		match := false
		for _, selector := range selectors {
			if strings.HasPrefix(tag, selector) {
				match = true
				break
			}
		}
		if match {
			tags = append(tags, tag)
		}
	}
	return tags
}

func (m *Manager) Start() error {
	m.access.Lock()
	defer m.access.Unlock()
	m.running = true
	for _, h := range m.taggedHandler {
		if err := h.Start(); err != nil {
			return err
		}
	}
	for _, h := range m.untaggedHandlers {
		if err := h.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Type() interface{} {
	return outbound.ManagerType()
}

func init() {
	common.Must(common.RegisterConfig((*proxyman.OutboundConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*proxyman.OutboundConfig))
	}))
	common.Must(common.RegisterConfig((*core.OutboundHandlerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewHandler(ctx, config.(*core.OutboundHandlerConfig))
	}))
}

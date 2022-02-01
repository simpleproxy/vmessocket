package log

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

import (
	"context"
	"sync"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/log"
)

type Instance struct {
	sync.RWMutex
	config       *Config
	accessLogger log.Handler
	errorLogger  log.Handler
	active       bool
}

func New(ctx context.Context, config *Config) (*Instance, error) {
	g := &Instance{
		config: config,
		active: false,
	}
	log.RegisterHandler(g)

	if err := g.startInternal(); err != nil {
		return nil, err
	}

	newError("Logger started").AtDebug().WriteToLog()
	return g, nil
}

func (g *Instance) initAccessLogger() error {
	handler, err := createHandler(g.config.AccessLogType, HandlerCreatorOptions{
		Path: g.config.AccessLogPath,
	})
	if err != nil {
		return err
	}
	g.accessLogger = handler
	return nil
}

func (g *Instance) initErrorLogger() error {
	handler, err := createHandler(g.config.ErrorLogType, HandlerCreatorOptions{
		Path: g.config.ErrorLogPath,
	})
	if err != nil {
		return err
	}
	g.errorLogger = handler
	return nil
}

func (*Instance) Type() interface{} {
	return (*Instance)(nil)
}

func (g *Instance) startInternal() error {
	g.Lock()
	defer g.Unlock()

	if g.active {
		return nil
	}

	g.active = true

	if err := g.initAccessLogger(); err != nil {
		return newError("failed to initialize access logger").Base(err).AtWarning()
	}
	if err := g.initErrorLogger(); err != nil {
		return newError("failed to initialize error logger").Base(err).AtWarning()
	}

	return nil
}

func (g *Instance) Start() error {
	return g.startInternal()
}

func (g *Instance) Handle(msg log.Message) {
	g.RLock()
	defer g.RUnlock()

	if !g.active {
		return
	}

	switch msg := msg.(type) {
	case *log.AccessMessage:
		if g.accessLogger != nil {
			g.accessLogger.Handle(msg)
		}
	case *log.GeneralMessage:
		if g.errorLogger != nil && msg.Severity <= g.config.ErrorLogLevel {
			g.errorLogger.Handle(msg)
		}
	default:
	}
}

func (g *Instance) Close() error {
	newError("Logger closing").AtDebug().WriteToLog()

	g.Lock()
	defer g.Unlock()

	if !g.active {
		return nil
	}

	g.active = false

	common.Close(g.accessLogger)
	g.accessLogger = nil

	common.Close(g.errorLogger)
	g.errorLogger = nil

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}

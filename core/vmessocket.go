package core

import (
	"context"
	"reflect"
	"sync"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/serial"
	"github.com/vmessocket/vmessocket/features"
	"github.com/vmessocket/vmessocket/features/inbound"
	"github.com/vmessocket/vmessocket/features/outbound"
)

type Instance struct {
	access             sync.Mutex
	features           []features.Feature
	running            bool
	ctx                context.Context
}

type Server interface {
	common.Runnable
}

func ServerType() interface{} {
	return (*Instance)(nil)
}

func AddInboundHandler(server *Instance, config *InboundHandlerConfig) error {
	inboundManager := server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	rawHandler, err := CreateObject(server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(inbound.Handler)
	if !ok {
		return newError("not an InboundHandler")
	}
	if err := inboundManager.AddHandler(server.ctx, handler); err != nil {
		return err
	}
	return nil
}

func addInboundHandlers(server *Instance, configs []*InboundHandlerConfig) error {
	for _, inboundConfig := range configs {
		if err := AddInboundHandler(server, inboundConfig); err != nil {
			return err
		}
	}
	return nil
}

func AddOutboundHandler(server *Instance, config *OutboundHandlerConfig) error {
	outboundManager := server.GetFeature(outbound.ManagerType()).(outbound.Manager)
	rawHandler, err := CreateObject(server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return newError("not an OutboundHandler")
	}
	if err := outboundManager.AddHandler(server.ctx, handler); err != nil {
		return err
	}
	return nil
}

func addOutboundHandlers(server *Instance, configs []*OutboundHandlerConfig) error {
	for _, outboundConfig := range configs {
		if err := AddOutboundHandler(server, outboundConfig); err != nil {
			return err
		}
	}
	return nil
}

func getFeature(allFeatures []features.Feature, t reflect.Type) features.Feature {
	for _, f := range allFeatures {
		if reflect.TypeOf(f.Type()) == t {
			return f
		}
	}
	return nil
}

func initInstanceWithConfig(config *Config, server *Instance) (bool, error) {
	if config.Transport != nil {
		features.PrintDeprecatedFeatureWarning("global transport settings")
	}
	if err := config.Transport.Apply(); err != nil {
		return true, err
	}
	for _, appSettings := range config.App {
		settings, err := appSettings.GetInstance()
		if err != nil {
			return true, err
		}
		obj, err := CreateObject(server, settings)
		if err != nil {
			return true, err
		}
		if feature, ok := obj.(features.Feature); ok {
			if err := server.AddFeature(feature); err != nil {
				return true, err
			}
		}
	}
	if err := addInboundHandlers(server, config.Inbound); err != nil {
		return true, err
	}
	if err := addOutboundHandlers(server, config.Outbound); err != nil {
		return true, err
	}
	return false, nil
}

func New(config *Config) (*Instance, error) {
	server := &Instance{ctx: context.Background()}
	done, err := initInstanceWithConfig(config, server)
	if done {
		return nil, err
	}
	return server, nil
}

func NewWithContext(ctx context.Context, config *Config) (*Instance, error) {
	server := &Instance{ctx: ctx}
	done, err := initInstanceWithConfig(config, server)
	if done {
		return nil, err
	}
	return server, nil
}

func RequireFeatures(ctx context.Context, callback interface{}) error {
	v := MustFromContext(ctx)
	return v.RequireFeatures(callback)
}

func (s *Instance) AddFeature(feature features.Feature) error {
	s.features = append(s.features, feature)
	if s.running {
		if err := feature.Start(); err != nil {
			newError("failed to start feature").Base(err).WriteToLog()
		}
		return nil
	}
	return nil
}

func (s *Instance) Close() error {
	s.access.Lock()
	defer s.access.Unlock()
	s.running = false
	var errors []interface{}
	for _, f := range s.features {
		if err := f.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return newError("failed to close all features").Base(newError(serial.Concat(errors...)))
	}
	return nil
}

func (s *Instance) GetFeature(featureType interface{}) features.Feature {
	return getFeature(s.features, reflect.TypeOf(featureType))
}

func (s *Instance) RequireFeatures(callback interface{}) error {
	callbackType := reflect.TypeOf(callback)
	if callbackType.Kind() != reflect.Func {
		panic("not a function")
	}
	var featureTypes []reflect.Type
	for i := 0; i < callbackType.NumIn(); i++ {
		featureTypes = append(featureTypes, reflect.PtrTo(callbackType.In(i)))
	}
	return nil
}

func (s *Instance) Start() error {
	s.access.Lock()
	defer s.access.Unlock()
	s.running = true
	for _, f := range s.features {
		if err := f.Start(); err != nil {
			return err
		}
	}
	newError("vmessocket ", Version(), " started").AtWarning().WriteToLog()
	return nil
}

func (s *Instance) Type() interface{} {
	return ServerType()
}

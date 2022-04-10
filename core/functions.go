package core

import (
	"bytes"
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/net"
	"github.com/vmessocket/vmessocket/features/routing"
)

func CreateObject(v *Instance, config interface{}) (interface{}, error) {
	var ctx context.Context
	if v != nil {
		ctx = toContext(v.ctx, v)
	}
	return common.CreateObject(ctx, config)
}

func Dial(ctx context.Context, v *Instance, dest net.Destination) (net.Conn, error) {
	ctx = toContext(ctx, v)
	dispatcher := v.GetFeature(routing.DispatcherType())
	if dispatcher == nil {
		return nil, newError("routing.Dispatcher is not registered in vmessocket")
	}
	r, err := dispatcher.(routing.Dispatcher).Dispatch(ctx, dest)
	if err != nil {
		return nil, err
	}
	var readerOpt net.ConnectionOption
	if dest.Network == net.Network_TCP {
		readerOpt = net.ConnectionOutputMulti(r.Reader)
	} else {
		readerOpt = net.ConnectionOutputMultiUDP(r.Reader)
	}
	return net.NewConnection(net.ConnectionInputMulti(r.Writer), readerOpt), nil
}

func StartInstance(configFormat string, configBytes []byte) (*Instance, error) {
	config, err := LoadConfig(configFormat, "", bytes.NewReader(configBytes))
	if err != nil {
		return nil, err
	}
	instance, err := New(config)
	if err != nil {
		return nil, err
	}
	if err := instance.Start(); err != nil {
		return nil, err
	}
	return instance, nil
}

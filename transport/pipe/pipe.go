package pipe

import (
	"context"

	"github.com/vmessocket/vmessocket/common/signal"
	"github.com/vmessocket/vmessocket/common/signal/done"
	"github.com/vmessocket/vmessocket/features/policy"
)

type Option func(*pipeOption)

func DiscardOverflow() Option {
	return func(opt *pipeOption) {
		opt.discardOverflow = true
	}
}

func New(opts ...Option) (*Reader, *Writer) {
	p := &pipe{
		readSignal:  signal.NewNotifier(),
		writeSignal: signal.NewNotifier(),
		done:        done.New(),
		option: pipeOption{
			limit: -1,
		},
	}
	for _, opt := range opts {
		opt(&(p.option))
	}
	return &Reader{
			pipe: p,
		}, &Writer{
			pipe: p,
		}
}

func OptionsFromContext(ctx context.Context) []Option {
	var opt []Option
	bp := policy.BufferPolicyFromContext(ctx)
	if bp.PerConnection >= 0 {
		opt = append(opt, WithSizeLimit(bp.PerConnection))
	} else {
		opt = append(opt, WithoutSizeLimit())
	}
	return opt
}

func WithoutSizeLimit() Option {
	return func(opt *pipeOption) {
		opt.limit = -1
	}
}

func WithSizeLimit(limit int32) Option {
	return func(opt *pipeOption) {
		opt.limit = limit
	}
}

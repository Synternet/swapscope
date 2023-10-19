package service

import (
	"fmt"

	svcnats "github.com/SyntropyNet/pubsub-go/pubsub"
	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
)

type Option func(*Options) error

type Options struct {
	prefix     string
	natsSub    *svcnats.NatsService
	natsPub    *svcnats.NatsService
	analytics  analytics.Analytics
	bufferSize int
}

func (o *Options) SetDefaults() {
	o.natsSub = nil
	o.natsPub = nil
	o.analytics = nil
	o.bufferSize = 5000
}

func (o *Options) ParseOptions(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

func WithPrefix(prefix string) Option {
	return func(o *Options) error {
		o.prefix = prefix
		return nil
	}
}

func WithBufferSize(size int) Option {
	return func(o *Options) error {
		o.bufferSize = size
		return nil
	}
}

func WithNATS(snSub *svcnats.NatsService, snPub *svcnats.NatsService) Option {
	return func(o *Options) error {
		if snSub == nil {
			return fmt.Errorf("nats subscriber service must not be nil")
		}
		o.natsSub = snSub

		if snPub == nil {
			return fmt.Errorf("nats publisher service must not be nil")
		}
		o.natsPub = snPub

		return nil
	}
}

func WithAnalytics(a analytics.Analytics) Option {
	return func(o *Options) error {
		if a == nil {
			return fmt.Errorf("analytics must not be nil")
		}
		o.analytics = a
		return nil
	}
}

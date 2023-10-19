package ethereum

import (
	"time"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
)

type (
	Option func(*Options) error

	PriceFetcher interface {
		Price(tokenAddress string) (repository.TokenPrice, error)
	}

	TokenFetcher interface {
		Token(tokenAddress string) (repository.Token, error)
	}

	Options struct {
		eventLogCacheExpirationTime time.Duration
		eventLogCachePurgeTime      time.Duration
		priceFetcher                PriceFetcher
		tokenFetcher                TokenFetcher
	}
)

func (o *Options) ParseOptions(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return err
		}
	}
	return nil
}

func WithEventLogCache(expiration time.Duration, purge time.Duration) Option {
	return func(o *Options) error {
		o.eventLogCacheExpirationTime = expiration
		o.eventLogCachePurgeTime = purge
		return nil
	}
}

func WithTokenPriceFetcher(fetcher PriceFetcher) Option {
	return func(o *Options) error {
		o.priceFetcher = fetcher
		return nil
	}
}

func WithTokenFetcher(fetcher TokenFetcher) Option {
	return func(o *Options) error {
		o.tokenFetcher = fetcher
		return nil
	}
}

package ethereum

import (
	"context"
	"errors"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/patrickmn/go-cache"
)

var (
	mintSig     string
	transferSig string
	burnSig     string
)

const (
	subSubject            = "syntropy.ethereum.log-event"
	uniswapPositionsOwner = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"
	mintEventHeader       = "Mint(address,address,int24,int24,uint128,uint256,uint256)"
	transferEventHeader   = "Transfer(address,address,uint256)"
	burnEventHeader       = "Burn(address,int24,int24,uint128,uint256,uint256)"
)

type Analytics struct {
	Options
	db  repository.Repository
	ctx context.Context

	eventLogCache *cache.Cache
}

func init() {
	mintSig = convertToEventSignature(mintEventHeader)
	transferSig = convertToEventSignature(transferEventHeader)
	burnSig = convertToEventSignature(burnEventHeader)
}

func New(ctx context.Context, db repository.Repository, opts ...Option) (*Analytics, error) {
	ret := &Analytics{
		ctx: ctx,
		db:  db,
	}
	if err := ret.Options.ParseOptions(opts...); err != nil {
		return nil, err
	}

	if ret.priceFetcher == nil {
		return nil, errors.New("price fetcher must be set")
	}
	if ret.tokenFetcher == nil {
		return nil, errors.New("token fetcher must be set")
	}

	ret.eventLogCache = cache.New(ret.Options.eventLogCacheExpirationTime, ret.Options.eventLogCachePurgeTime)

	return ret, nil
}

func (a *Analytics) Handlers() map[string]analytics.Handler {
	return map[string]analytics.Handler{
		subSubject: a.ProcessMessage,
	}
}

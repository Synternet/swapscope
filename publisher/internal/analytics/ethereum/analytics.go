package ethereum

import (
	"context"
	"errors"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/patrickmn/go-cache"
)

var eventsToTrack = []string{
	"Mint(address,address,int24,int24,uint128,uint256,uint256)",
	"Transfer(address,address,uint256)",
}

var eventsToAvoid = []string{
	"Swap(address,address,int256,int256,uint160,uint128,int24)",
	"Swap(address,uint256,uint256,uint256,uint256,address)",
	"Swap(address,address,uint256,uint256)",
	"Collect(address,address,int24,int24,uint128,uint128)",
	"Collect(uint256,address,uint256,uint256)",
	"AssetWithdrawn(address,address,uint256)",
	"Burn(address,uint256,uint256,address)",
	"AdaptorCalled(address,bytes)",
	"Mint(address,address,int24,int24,uint128,uint256,uint256)",
	"TransformedERC20(address,address,address,uint256,uint256)",
}

const (
	subSubject            = "syntropy.ethereum.log-event"
	uniswapPositionsOwner = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"
)

type Analytics struct {
	Options
	db  repository.Repository
	ctx context.Context

	eventLogCache     *cache.Cache
	signatureMint     string
	signatureTransfer string
	signaturesToAvoid map[string]struct{}
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

	ret.signatureMint = convertToEventSignature(eventsToTrack[0])
	ret.signatureTransfer = convertToEventSignature(eventsToTrack[1])
	ret.signaturesToAvoid = convertToListOfEventSignatures(eventsToAvoid)

	return ret, nil
}

func (a *Analytics) Handlers() map[string]analytics.Handler {
	return map[string]analytics.Handler{
		subSubject: a.ProcessMessage,
	}
}

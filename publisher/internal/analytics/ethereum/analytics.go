package ethereum

import (
	"context"
	"errors"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/patrickmn/go-cache"
)

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
	mintSig               = "0x7a53080b" //Mint(address,address,int24,int24,uint128,uint256,uint256)
	transferSig           = "0xddf252ad" //Transfer(address,address,uint256)
	burnSig               = "0x0c396cd9" //Burn(address,int24,int24,uint128,uint256,uint256)
)

type Analytics struct {
	Options
	db  repository.Repository
	ctx context.Context

	eventLogCache     *cache.Cache
	signatureMint     string
	signatureTransfer string
	signatureBurn     string
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

	ret.signatureMint = mintSig
	ret.signatureTransfer = transferSig
	ret.signatureBurn = burnSig
	ret.signaturesToAvoid = convertToListOfEventSignatures(eventsToAvoid)

	return ret, nil
}

func (a *Analytics) Handlers() map[string]analytics.Handler {
	return map[string]analytics.Handler{
		subSubject: a.ProcessMessage,
	}
}

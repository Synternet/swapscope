package ethereum

import (
	"context"
	_ "embed"
	"errors"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/patrickmn/go-cache"
)

var (
	mintSig     string
	transferSig string
	burnSig     string
	collectSig  string

	//go:embed Uniswap_Liquidity_Pool_contract.json
	uniswapLiqPoolsABIJson string
	uniswapLiqPoolsABI     abi.ABI

	//go:embed Uniswap_Liquidity_Position_contract.json
	uniswapLiqPositionsABIJson string
	uniswapLiqPositionsABI     abi.ABI
)

const (
	subSubject            = "syntropy.ethereum.log-event"
	uniswapPositionsOwner = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"
	mintEventHeader       = "Mint(address,address,int24,int24,uint128,uint256,uint256)" // Matches function's name and parameters types list
	mintEvent             = "Mint"                                                      // Matches function's name
	transferEventHeader   = "Transfer(address,address,uint256)"
	transferEvent         = "Transfer"
	burnEventHeader       = "Burn(address,int24,int24,uint128,uint256,uint256)"
	burnEvent             = "Burn"
	collectEventHeader    = "Collect(address,address,int24,int24,uint128,uint128)"
	collectEvent          = "Collect"
	addressWETH           = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2" // https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2
	addressUSDT           = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7
	addressUSDC           = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
)

type Analytics struct {
	Options
	db  repository.Repository
	ctx context.Context

	eventLogCache *EventLogCache
}

type CacheRecord map[string]interface{}
type EventLogCache struct {
	*cache.Cache
}

func init() {
	mintSig = convertToEventSignature(mintEventHeader)
	transferSig = convertToEventSignature(transferEventHeader)
	burnSig = convertToEventSignature(burnEventHeader)
	collectSig = convertToEventSignature(collectEventHeader)
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

	ret.eventLogCache = &EventLogCache{cache.New(ret.Options.eventLogCacheExpirationTime, ret.Options.eventLogCachePurgeTime)}

	uniswapLiqPoolsABI = parseJsonToAbi(uniswapLiqPoolsABIJson)
	uniswapLiqPositionsABI = parseJsonToAbi(uniswapLiqPositionsABIJson)

	return ret, nil
}

func (a *Analytics) Handlers() map[string]analytics.Handler {
	return map[string]analytics.Handler{
		subSubject: a.ProcessMessage,
	}
}

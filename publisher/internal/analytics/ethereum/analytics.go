package ethereum

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/Synternet/swapscope/publisher/pkg/analytics"
	"github.com/Synternet/swapscope/publisher/pkg/repository"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/patrickmn/go-cache"
)

var (
	//go:embed Uniswap_Liquidity_Pool_contract.json
	uniswapLiqPoolsABIJson string
	uniswapLiqPoolsABI     abi.ABI

	//go:embed ERC20_token_contract_abi.json
	ethereumErc20TokenABIJson string
	ethereumErc20TokenABI     abi.ABI

	stableCoins []string
	nativeCoins []string
)

const (
	subSubject            = "synternet.ethereum.log-event"
	uniswapPositionsOwner = "0xC36442b4a4522E871399CD717aBDD847Ab11FE88"
	mintEvent             = "Mint" // Has to match Event's name in respective ABI
	transferEvent         = "Transfer"
	burnEvent             = "Burn"
	collectEvent          = "Collect"
	swapEvent             = "Swap"
	addressWETH           = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2" // https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2
	addressUSDT           = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7
	addressUSDC           = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
)

type Analytics struct {
	Options
	db  repository.Repository
	ctx context.Context

	eventLogCache *EventLogCache

	eventSignature map[string]string
}

type (
	CacheRecord   map[string]interface{}
	EventLogCache struct {
		*cache.Cache
	}
)

func init() {
	stableCoins = []string{strings.ToLower(addressUSDT), strings.ToLower(addressUSDC)}
	nativeCoins = []string{strings.ToLower(addressWETH)}
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
	ethereumErc20TokenABI = parseJsonToAbi(ethereumErc20TokenABIJson)
	ret.eventSignature = make(map[string]string)
	ret.eventSignature[mintEvent] = convertToEventSignature(uniswapLiqPoolsABI.Events[mintEvent].Sig)
	ret.eventSignature[burnEvent] = convertToEventSignature(uniswapLiqPoolsABI.Events[burnEvent].Sig)
	ret.eventSignature[swapEvent] = convertToEventSignature(uniswapLiqPoolsABI.Events[swapEvent].Sig)
	ret.eventSignature[collectEvent] = convertToEventSignature(uniswapLiqPoolsABI.Events[collectEvent].Sig)
	ret.eventSignature[transferEvent] = convertToEventSignature(ethereumErc20TokenABI.Events[transferEvent].Sig)

	return ret, nil
}

func (a *Analytics) Handlers() map[string]analytics.Handler {
	return map[string]analytics.Handler{
		subSubject: a.ProcessMessage,
	}
}

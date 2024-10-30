package fetcher

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Synternet/swapscope/publisher/pkg/repository"
	"github.com/patrickmn/go-cache"
)

type TokenInfoResponse struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	DetailPlatforms map[string]struct {
		DecimalPlaces int `json:"decimal_place"`
	} `json:"detail_platforms"`
	MarketData struct {
		CurrentPrice map[string]float64 `json:"current_price"`
	} `json:"market_data"`
}

// type TokenPriceResponse struct {
// 	// We can't access price straight away because the first key is token address (changes with different tokens)
// 	Data map[string]map[string]float64 `json:""`
// }

type TokenPriceResponse map[string]map[string]float64

type CoingeckoFetcher struct {
	ctx          context.Context
	db           repository.Repository
	baseApiUrl   string
	expiresIn    time.Duration // For price fetching
	priceFetcher RateLimitedFetcher[TokenPriceResponse]
	tokenFetcher RateLimitedFetcher[TokenInfoResponse]
	cache        *cache.Cache // For price fetching
}

const (
	tokenInfoEndpoint  = "/coins/ethereum/contract/"
	tokenPriceEndpoint = "/simple/token_price/ethereum"
	priceBase          = "usd"
	pricePrecision     = 10
)

func NewCoingeckoFetcher(ctx context.Context, db repository.Repository, apiUrl string, expires, purges, timeout time.Duration, rateLimit int) (*CoingeckoFetcher, error) {
	ret := &CoingeckoFetcher{
		baseApiUrl: apiUrl,
		ctx:        ctx,
		expiresIn:  expires,
		db:         db,
	}
	ret.cache = cache.New(expires, purges)
	ret.priceFetcher = RateLimitedFetcher[TokenPriceResponse]{
		Client:    &http.Client{},
		Timeout:   timeout,
		RateLimit: rateLimit,
	}
	ret.tokenFetcher = RateLimitedFetcher[TokenInfoResponse]{
		Client:    &http.Client{},
		Timeout:   timeout,
		RateLimit: rateLimit,
	}
	return ret, nil
}

func (p *CoingeckoFetcher) Price(tokenAddress string) (repository.TokenPrice, error) {
	if price, found := p.cache.Get(tokenAddress); found {
		log.Println("Price found in cache", price, "of token", tokenAddress)
		return price.(repository.TokenPrice), nil
	}

	price, err := p.fetchPrice(tokenAddress)
	if err != nil {
		return repository.TokenPrice{}, err
	}
	p.addTokenPriceToCache(tokenAddress, price)
	return price, nil
}

func (p *CoingeckoFetcher) fetchPrice(tokenAddress string) (repository.TokenPrice, error) {
	queryParams := url.Values{}
	queryParams.Add("contract_addresses", tokenAddress)
	queryParams.Add("vs_currencies", priceBase)
	queryParams.Add("precision", strconv.Itoa(pricePrecision))
	apiURL, _ := url.Parse(p.baseApiUrl)
	apiURL = apiURL.JoinPath(tokenPriceEndpoint)
	apiURL.RawQuery = queryParams.Encode()

	// var result TokenPriceResponse
	res, err := p.priceFetcher.Fetch(p.ctx, apiURL.String())
	if err != nil {
		return repository.TokenPrice{}, fmt.Errorf("failed fetching price for %s: %w", tokenAddress, err)
	}

	tokenPrices, found := res[tokenAddress]

	if !found {
		return repository.TokenPrice{}, fmt.Errorf("response does not contain token address %s", tokenAddress)
	}
	if value, found := tokenPrices[priceBase]; found {
		return repository.TokenPrice{Value: value, Base: priceBase}, nil
	}

	return repository.TokenPrice{}, fmt.Errorf("token %s price for base %s not found", tokenAddress, priceBase)
}

// Token tries to fetch token from Database if it is not there - tries to fetch from CoinGecko API
// If token is present in CoinGecko API - it is put to DB, price is also updated to cache (to not request CoinGecko 2 times)
func (p *CoingeckoFetcher) Token(tokenAddress string) (repository.Token, error) {
	token, found := p.db.GetToken(tokenAddress)
	if found {
		return token, nil
	}
	log.Println("Token", tokenAddress, "not in DB... Try to fetch it from CoinGecko API.")

	token, err := p.fetchToken(tokenAddress)
	if err != nil {
		return repository.Token{}, err
	}
	err = p.db.AddToken(token)
	return token, err
}

// fetchToken tries to fetch token from Database if it is not there - tries to fetch from CoinGecko API
// If token is present in CoinGecko API - it is put to DB, price is also updated to cache (to not request CoinGecko 2 times)
func (p *CoingeckoFetcher) fetchToken(tokenAddress string) (repository.Token, error) {
	apiURL, _ := url.Parse(p.baseApiUrl)
	apiURL = apiURL.JoinPath(tokenInfoEndpoint)
	apiURL = apiURL.JoinPath(tokenAddress)

	response, err := p.tokenFetcher.Fetch(p.ctx, apiURL.String())
	if err != nil {
		return repository.Token{}, fmt.Errorf("failed fetching token %s: %w", tokenAddress, err)
	}

	token := repository.Token{
		Address:     tokenAddress,
		Symbol:      strings.ToUpper(response.Symbol),
		Name:        response.Name,
		Decimals:    response.DetailPlatforms["ethereum"].DecimalPlaces,
		TotalSupply: 0,
	}

	if token.Symbol == "" || token.Decimals == 0 {
		return repository.Token{}, fmt.Errorf("token (%s,%s,%d) not found in API", tokenAddress, token.Symbol, token.Decimals)
	}

	if value, found := response.MarketData.CurrentPrice[priceBase]; found {
		p.addTokenPriceToCache(tokenAddress, repository.TokenPrice{Value: value, Base: priceBase})
	} else {
		log.Println("Fetched token", tokenAddress, " but could not fetch price.")
	}

	return token, nil
}

func (p *CoingeckoFetcher) addTokenPriceToCache(tokenAddress string, tokenPrice repository.TokenPrice) {
	p.cache.Set(tokenAddress, tokenPrice, p.expiresIn)
	log.Println("Added price", tokenPrice.Value, tokenPrice.Base, "of token", tokenAddress, "with expiration time", p.expiresIn)
}

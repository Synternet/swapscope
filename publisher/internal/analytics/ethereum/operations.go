package ethereum

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/SyntropyNet/swapscope/publisher/pkg/types"
	"github.com/patrickmn/go-cache"
)

type AnalyticsInterface interface {
	convertTransferAmount(string, int) float64
	includeTokenPrices(Position) Position
	calculatePosition(EventLog, Position) Position
	handleLiquidityTransfer(EventLog, Position) Position
	checkAndUpdateMissingToken(EventLog, Position) Position
	isAvoidEvent(EventLog) bool
	isTransferEvent(EventLog) bool
}

type DatabaseInterface interface {
	SaveRemoval(repository.Removal) error
	SaveAddition(repository.Addition) error
	GetPoolPairAddresses(string) (string, string, bool)
	GetToken(string) (repository.Token, bool)
}

type CacheInterface interface {
	Get(string) (interface{}, bool)
}

type Operation interface {
	// Common methods shared by Addition and Removal
	Extract(EventLog) (Operation, error)
	String() string
	CanPublish() bool
	Publish(time.Time) error
	Save(time.Time) error
}

func NewAdditionOperation(db DatabaseInterface, cache *cache.Cache, a AnalyticsInterface, sender analytics.Sender) Addition {
	return Addition{
		DatabaseInterface:  db,
		CacheInterface:     cache,
		AnalyticsInterface: a,
		Send:               sender,
	}
}

func NewRemovalOperation(db DatabaseInterface, cache *cache.Cache, a AnalyticsInterface, sender analytics.Sender) Removal {
	return Removal{
		DatabaseInterface:  db,
		AnalyticsInterface: a,
		CacheInterface:     cache,
		Send:               sender,
	}
}

func (rem Removal) Save(ts time.Time) error {
	removal := repository.Removal{
		TimestampReceived: ts,
		LPoolAddress:      rem.Address,
		Token0Symbol:      rem.Token0.Symbol,
		Token1Symbol:      rem.Token1.Symbol,
		Token0Amount:      rem.Token0.Amount,
		Token1Amount:      rem.Token1.Amount,
		LowerRatio:        rem.LowerRatio,
		UpperRatio:        rem.UpperRatio,
		Token0PriceUsd:    rem.Token0.Price,
		Token1PriceUsd:    rem.Token1.Price,
		TxHash:            rem.TxHash,
	}
	return rem.DatabaseInterface.SaveRemoval(removal)
}

func (add Addition) Save(ts time.Time) error {
	addition := repository.Addition{
		TimestampReceived: ts,
		LPoolAddress:      add.Address,
		Token0Symbol:      add.Token0.Symbol,
		Token1Symbol:      add.Token1.Symbol,
		Token0Amount:      add.Token0.Amount,
		Token1Amount:      add.Token1.Amount,
		LowerRatio:        add.LowerRatio,
		UpperRatio:        add.UpperRatio,
		Token0PriceUsd:    add.Token0.Price,
		Token1PriceUsd:    add.Token1.Price,
		TxHash:            add.TxHash,
	}
	return add.DatabaseInterface.SaveAddition(addition)
}

func (p Position) CanPublish() bool {
	if strings.EqualFold(p.Token0.Symbol, "") || strings.EqualFold(p.Token1.Symbol, "") {
		log.Printf("SKIP - token symbol unknown. Tx: %s\n\n", p.TxHash)
		return false
	}
	if p.LowerRatio == 0 && p.UpperRatio == 0 {
		log.Printf("SKIP - actual ratio not calculated. Tx: %s\n\n", p.TxHash)
		return false
	}
	if p.Token0.Amount == 0 && p.Token1.Amount == 0 {
		log.Printf("SKIP - no tokens moved. Tx: %s\n\n", p.TxHash)
		return false
	}

	return true
}

func (rem Removal) Extract(eLog EventLog) (Operation, error) {
	liqPool := eLog.Address
	addr0, addr1, found := rem.DatabaseInterface.GetPoolPairAddresses(liqPool)
	if !found {
		return Removal{}, fmt.Errorf("SKIP - liq. pool is unknown (removal). pool address: %s", liqPool)
	}
	token0, found0 := rem.DatabaseInterface.GetToken(addr0)
	token1, found1 := rem.DatabaseInterface.GetToken(addr1)
	if !found0 || !found1 {
		return Removal{}, fmt.Errorf("SKIP - at least one token is unknown in liquidity removal. pool address: %s", liqPool)
	}

	_, token0HexAmount, token1HexAmount, err := splitLogDatatoHexStrings(eLog.Data)
	if err != nil {
		return Removal{}, err
	}

	remPosition := Position{
		Token0: TokenTransaction{
			Token:  token0,
			Amount: rem.AnalyticsInterface.convertTransferAmount(token0HexAmount, token0.Decimals),
		},
		Token1: TokenTransaction{
			Token:  token1,
			Amount: rem.AnalyticsInterface.convertTransferAmount(token1HexAmount, token1.Decimals),
		},
	}
	remPosition = rem.AnalyticsInterface.includeTokenPrices(remPosition)      // 6) Getting token prices
	remPosition = rem.AnalyticsInterface.calculatePosition(eLog, remPosition) // 7) Save Liquidity Entry and Liquidity Pool
	rem.Position = remPosition
	return rem, nil
}

func (add Addition) Extract(eLog EventLog) (Operation, error) {
	if !isUniswapPositionsNFT(eLog.Data) {
		return Addition{}, fmt.Errorf("not Uniswap Positions NFT.\n")
	}
	// 2) Mint event is found
	var addPosition Position
	txEventsFromCache, _ := add.CacheInterface.Get(eLog.TransactionHash)
	for _, evLog := range txEventsFromCache.([]EventLog) { // Go through all events of this transaction
		if reflect.DeepEqual(evLog, eLog) { // This allows to correctly process many Mint events in one transaction
			break
		}
		if add.AnalyticsInterface.isAvoidEvent(evLog) {
			addPosition = Position{} // Reset gathered records
			continue
		}
		if add.AnalyticsInterface.isTransferEvent(evLog) { // 3) Searching for relevant "Transfer" event(s)
			addPosition = add.AnalyticsInterface.handleLiquidityTransfer(evLog, addPosition) // 4) Decoding / expanding "Transfer" events
		}
	}
	if isEitherTokenUnknown(addPosition) {
		addPosition = add.AnalyticsInterface.checkAndUpdateMissingToken(eLog, addPosition) // 5) Adding missing token if only 1 token transfer was made
	}
	addPosition = add.AnalyticsInterface.includeTokenPrices(addPosition)      // 6) Getting token prices
	addPosition = add.AnalyticsInterface.calculatePosition(eLog, addPosition) // 7) Save Liquidity Entry and Liquidity Pool
	add.Position = addPosition
	return add, nil
}

func (rem Removal) String() string {
	format := "Removing %f of %s and %f of %s from %s between %f and %f\n"
	return fmt.Sprintf(format,
		rem.Token0.Amount,
		rem.Token0.Symbol,
		rem.Token1.Amount,
		rem.Token1.Symbol,
		rem.Address,
		rem.LowerRatio,
		rem.UpperRatio)
}

func (add Addition) String() string {
	format := "Adding %f of %s($%f) and %f of %s($%f) = $%f. To %s between %f and %f while current is %f\n"
	return fmt.Sprintf(format,
		add.Token0.Amount,
		add.Token0.Symbol,
		add.Token0.Price,
		add.Token1.Amount,
		add.Token1.Symbol,
		add.Token1.Price,
		add.TotalValue,
		add.Address,
		add.LowerRatio,
		add.UpperRatio,
		add.CurrentRatio)
}

func (rem Removal) Publish(timestamp time.Time) error {
	removalMessage := types.RemovalMessage{
		Timestamp:         timestamp,
		Address:           rem.Address,
		LowerTokenRatio:   rem.LowerRatio,
		CurrentTokenRatio: rem.CurrentRatio,
		UpperTokenRatio:   rem.UpperRatio,
		ValueRemovedUSD:   rem.TotalValue,
		Pair: [2]types.TokenMessage{
			{Symbol: rem.Token0.Symbol, Amount: rem.Token0.Amount, Price: rem.Token0.Price},
			{Symbol: rem.Token1.Symbol, Amount: rem.Token1.Amount, Price: rem.Token1.Price},
		},
		TxHash: rem.TxHash,
	}

	removalJson, err := json.Marshal(&removalMessage)
	if err != nil {
		return fmt.Errorf("error marshalling Liquidity Removal object into a json message: %s", err)
	}

	streamName := strings.ToLower(fmt.Sprintf("remove.%s.%s", rem.Token0.Symbol, rem.Token1.Symbol))
	return rem.Send(removalJson, streamName)
}

func (add Addition) Publish(timestamp time.Time) error {
	additionMessage := types.AdditionMessage{
		Timestamp:         timestamp,
		Address:           add.Address,
		LowerTokenRatio:   add.LowerRatio,
		CurrentTokenRatio: add.CurrentRatio,
		UpperTokenRatio:   add.UpperRatio,
		ValueAddedUSD:     add.TotalValue,
		Pair: [2]types.TokenMessage{
			{Symbol: add.Token0.Symbol, Amount: add.Token0.Amount, Price: add.Token0.Price},
			{Symbol: add.Token1.Symbol, Amount: add.Token1.Amount, Price: add.Token1.Price},
		},
		TxHash: add.TxHash,
	}

	additionJson, err := json.Marshal(&additionMessage)
	if err != nil {
		return fmt.Errorf("error marshalling Liquidity Addition object into a json message: %s", err)
	}

	streamName := strings.ToLower(fmt.Sprintf("add.%s.%s", add.Token0.Symbol, add.Token1.Symbol))
	return add.Send(additionJson, streamName)
}

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
)

type Operation interface {
	// Common methods shared by Addition and Removal
	ExtractFromEventLogs(EventLog, *Analytics) (Operation, error)
	PrintPretty() string
	CanPublish() bool
	PublishToNATS(time.Time, analytics.Sender) error
	SaveToDB(time.Time, *Analytics) error
}

type EventLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type Removal struct {
	Position
	//TODO: Fees earned and collected
	//TokenEarned0 TokenTransaction
	//TokenEarned1 TokenTransaction
}

type Addition struct {
	Position
}

type Position struct {
	Address      string
	Token0       TokenTransaction
	Token1       TokenTransaction
	LowerRatio   float64
	CurrentRatio float64
	UpperRatio   float64
	TotalValue   float64
	LowerTick    int
	UpperTick    int
	TxHash       string
}

type TokenTransaction struct {
	repository.Token
	Amount float64
	Price  float64
}

func (rem Removal) SaveToDB(ts time.Time, a *Analytics) error {
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
	return a.db.SaveRemoval(removal)
}

func (add Addition) SaveToDB(ts time.Time, a *Analytics) error {
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
	return a.db.SaveAddition(addition)
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

func (r Removal) ExtractFromEventLogs(eLog EventLog, a *Analytics) (Operation, error) {
	liqPool := eLog.Address
	addr0, addr1, found := a.db.GetPoolPairAddresses(liqPool)
	if !found {
		return Removal{}, fmt.Errorf("SKIP - liq. pool is unknown (removal). pool address: %s", liqPool)
	}
	token0, found0 := a.db.GetToken(addr0)
	token1, found1 := a.db.GetToken(addr1)
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
			Amount: a.convertTransferAmount(token0HexAmount, token0.Decimals),
		},
		Token1: TokenTransaction{
			Token:  token1,
			Amount: a.convertTransferAmount(token1HexAmount, token1.Decimals),
		},
	}
	remPosition = a.includeTokenPrices(remPosition)      // 6) Getting token prices
	remPosition = a.calculatePosition(eLog, remPosition) // 7) Save Liquidity Entry and Liquidity Pool

	return Removal{
		Position: remPosition,
	}, nil
}

func (add Addition) ExtractFromEventLogs(eLog EventLog, a *Analytics) (Operation, error) {
	if !isUniswapPositionsNFT(eLog.Data) {
		return Addition{}, fmt.Errorf("not Uniswap Positions NFT.\n")
	}
	// 2) Mint event is found
	var addPosition Position
	txEventsFromCache, _ := a.eventLogCache.Get(eLog.TransactionHash)
	for _, evLog := range txEventsFromCache.([]EventLog) { // Go through all events of this transaction
		if reflect.DeepEqual(evLog, eLog) { // This allows to correctly process many Mint events in one transaction
			break
		}
		if a.isAvoidEvent(evLog) {
			addPosition = Position{} // Reset gathered records
			continue
		}
		if a.isTransferEvent(evLog) { // 3) Searching for relevant "Transfer" event(s)
			addPosition = a.handleLiquidityTransfer(evLog, addPosition) // 4) Decoding / expanding "Transfer" events
		}
	}
	if isEitherTokenUnknown(addPosition) {
		addPosition = a.checkAndUpdateMissingToken(eLog, addPosition) // 5) Adding missing token if only 1 token transfer was made
	}
	addPosition = a.includeTokenPrices(addPosition)      // 6) Getting token prices
	addPosition = a.calculatePosition(eLog, addPosition) // 7) Save Liquidity Entry and Liquidity Pool

	return Addition{
		Position: addPosition,
	}, nil
}

func (rem Removal) PrintPretty() string {
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

func (add Addition) PrintPretty() string {
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

func (rem Removal) PublishToNATS(timestamp time.Time, send analytics.Sender) error {
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
	return send(removalJson, streamName)
}

func (add Addition) PublishToNATS(timestamp time.Time, send analytics.Sender) error {
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
	return send(additionJson, streamName)
}

package ethereum

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/SyntropyNet/swapscope/publisher/pkg/types"
	"github.com/patrickmn/go-cache"
)

type Fetchers struct {
	priceFetcher PriceFetcher
	tokenFetcher TokenFetcher
}

type OperationBase struct {
	db       Database
	cache    Cache
	fetchers Fetchers
}

type Database interface {
	SaveRemoval(repository.Removal) error
	SaveAddition(repository.Addition) error
	GetPoolPairAddresses(string) (string, string, bool)
	GetToken(string) (repository.Token, bool)
	SavePool(repository.Pool) error
}

type Cache interface {
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

type Removal struct {
	Position
	OperationBase
	Send analytics.Sender
	//TODO: Fees earned and collected
	//TokenEarned0 TokenTransaction
	//TokenEarned1 TokenTransaction
}

type Addition struct {
	Position
	OperationBase
	Send analytics.Sender
}

func NewAdditionOperation(db Database, cache *cache.Cache, a *Analytics, sender analytics.Sender) Addition {
	return Addition{
		OperationBase: OperationBase{
			db:    db,
			cache: cache,
			fetchers: Fetchers{
				priceFetcher: a.priceFetcher,
				tokenFetcher: a.tokenFetcher,
			},
		},
		Send: sender,
	}
}

func NewRemovalOperation(db Database, cache *cache.Cache, a *Analytics, sender analytics.Sender) Removal {
	return Removal{
		OperationBase: OperationBase{
			db:    db,
			cache: cache,
			fetchers: Fetchers{
				priceFetcher: a.priceFetcher,
				tokenFetcher: a.tokenFetcher,
			},
		},
		Send: sender,
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
	return rem.db.SaveRemoval(removal)
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
	return add.db.SaveAddition(addition)
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

func (rem Removal) Extract(burn EventLog) (Operation, error) {
	liqPool := burn.Address
	addr0, addr1, found := rem.db.GetPoolPairAddresses(liqPool)
	if !found {
		return Removal{}, fmt.Errorf("SKIP - liq. pool is unknown (removal). pool address: %s", liqPool)
	}
	token0, found0 := rem.db.GetToken(addr0)
	token1, found1 := rem.db.GetToken(addr1)
	if !found0 || !found1 {
		return Removal{}, fmt.Errorf("SKIP - at least one token is unknown in liquidity removal. pool address: %s", liqPool)
	}

	_, token0HexAmount, token1HexAmount, err := splitBurnDatatoHexStrings(burn.Data)
	if err != nil {
		return Removal{}, err
	}

	remPosition := Position{
		Address: burn.Address,
		TxHash:  burn.TransactionHash,
		Token0: TokenTransaction{
			Token:  token0,
			Amount: convertTransferAmount(token0HexAmount, token0.Decimals),
		},
		Token1: TokenTransaction{
			Token:  token1,
			Amount: convertTransferAmount(token1HexAmount, token1.Decimals),
		},
	}
	remPosition = rem.includeTokenPrices(remPosition)  // 6) Getting token prices
	remPosition = calculatePosition(burn, remPosition) // 7) Save Liquidity Entry and Liquidity Pool
	rem.Position = remPosition
	return rem, nil
}

func (add Addition) Extract(mint EventLog) (Operation, error) {
	if !isUniswapPositionsNFT(mint.Data) {
		return Addition{}, fmt.Errorf("not Uniswap Positions NFT.\n")
	}

	txEventsFromCache, _ := add.cache.Get(mint.TransactionHash)
	addPosition := Position{
		Address: mint.Address,
		TxHash:  mint.TransactionHash,
	}

	for _, evLog := range txEventsFromCache.([]EventLog) { // Go through all events of this transaction
		if isTransferEvent(evLog) && !isUniswapPositionsNFT(evLog.Address) {
			addPosition = add.handleLiquidityTransfer(mint, evLog, addPosition)
		}
	}

	if isEitherTokenUnknown(addPosition) {
		addPosition = add.checkAndUpdateMissingToken(mint, addPosition) // 5) Adding missing token if only 1 token transfer was made
	}
	add.savePool(addPosition)

	addPosition = add.includeTokenPrices(addPosition)  // 6) Getting token prices
	addPosition = calculatePosition(mint, addPosition) // 7) Save Liquidity Entry and Liquidity Pool
	add.Position = addPosition

	return add, nil
}

func (add Addition) savePool(addPos Position) {
	if isEitherTokenAmountIsZero(addPos) || isEitherTokenUnknown(addPos) {
		return
	}
	// In this case both tokens were transferred to LP and their order is correct
	var newLiqPoll repository.Pool
	newLiqPoll.Address = addPos.Address
	newLiqPoll.Token0Address = addPos.Token0.Address
	newLiqPoll.Token1Address = addPos.Token1.Address
	add.db.SavePool(newLiqPoll)
	return
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

// checkAndUpdateMissingToken expands Liq. Add. record if only 1 token was transferred
// Second token is found and appended
// The order of tokens is fixed based on historical results (when 2 tokens were transferred for this LP)
func (op OperationBase) checkAndUpdateMissingToken(evLog EventLog, addPos Position) Position {
	liqPoolAddress := strings.ToLower(evLog.Address)

	tok0Address, tok1Address, foundPool := op.db.GetPoolPairAddresses(liqPoolAddress)
	if !foundPool {
		log.Println("Could not get token information of pool", liqPoolAddress)
		return addPos
	}

	if addPos.Token0.Token == (repository.Token{}) {
		t, err := op.getToken(tok0Address)
		if err != nil {
			log.Println("Failed fetching token information: ", err.Error())
			return addPos
		}
		addPos.Token0.Token = t
	}

	if addPos.Token1.Token == (repository.Token{}) {
		t, err := op.getToken(tok1Address)
		if err != nil {
			log.Println("Failed fetching token information: ", err.Error())
			return addPos
		}
		addPos.Token1.Token = t
	}

	log.Printf("Added second missing token from known pool %s", liqPoolAddress)
	return addPos
}

// handleLiquidityTransfer decodes Transfer event.
// Getting token that was transferred and calculating amount transferred.
// Keeping track of tokens involved in current Liq. Add. event.
func (add Addition) handleLiquidityTransfer(mint EventLog, transfer EventLog, addPos Position) Position {
	_, token0HexAmount, token1HexAmount, err := splitMintDatatoHexFields(mint.Data)
	if err != nil {
		log.Println("Could not split mint event into Amount fields: ", err.Error())
		return addPos
	}

	t, err := add.getToken(transfer.Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
		return addPos
	}

	if transfer.Data == token0HexAmount {
		addPos.Token0.Token = t
		addPos.Token0.Amount = convertTransferAmount(token0HexAmount, t.Decimals)
	}

	if transfer.Data == token1HexAmount {
		addPos.Token1.Token = t
		addPos.Token1.Amount = convertTransferAmount(token1HexAmount, t.Decimals)
	}
	return addPos
}

func (op OperationBase) includeTokenPrices(pos Position) Position {
	// Place here to implement price cache?
	if !strings.EqualFold(pos.Token0.Address, "") {
		price, err := op.getPrice(pos.Token0.Address)
		if err != nil {
			log.Println("failed to feetch Token0 price: ", err.Error())
		}
		pos.Token0.Price = price.Value
	}

	if !strings.EqualFold(pos.Token1.Address, "") {
		price, err := op.getPrice(pos.Token1.Address)
		if err != nil {
			log.Println("failed to feetch Token1 price: ", err.Error())
		}
		pos.Token1.Price = price.Value
	}
	return pos
}

func (op OperationBase) getToken(address string) (repository.Token, error) {
	return op.fetchers.tokenFetcher.Token(address)
}

func (op OperationBase) getPrice(address string) (repository.TokenPrice, error) {
	return op.fetchers.priceFetcher.Price(address)
}

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
	Extract(EventLog) error
	String() string
	CanPublish() bool
	Publish(time.Time) error
	Save(time.Time) error
	InitializeOperation(Database, *cache.Cache, *Analytics, analytics.Sender)
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

func (rem *Removal) InitializeOperation(db Database, cache *cache.Cache, a *Analytics, sender analytics.Sender) {
	rem.OperationBase = OperationBase{
		db:    db,
		cache: cache,
		fetchers: Fetchers{
			priceFetcher: a.priceFetcher,
			tokenFetcher: a.tokenFetcher,
		},
	}
	rem.Send = sender
}

func (add *Addition) InitializeOperation(db Database, cache *cache.Cache, a *Analytics, sender analytics.Sender) {
	add.OperationBase = OperationBase{
		db:    db,
		cache: cache,
		fetchers: Fetchers{
			priceFetcher: a.priceFetcher,
			tokenFetcher: a.tokenFetcher,
		},
	}
	add.Send = sender
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

func (rem *Removal) Extract(burn EventLog) error {
	liqPool := burn.Address
	addr0, addr1, found := rem.db.GetPoolPairAddresses(liqPool)
	if !found {
		return fmt.Errorf("SKIP - liq. pool is unknown (removal). pool address: %s", liqPool)
	}
	token0, found0 := rem.db.GetToken(addr0)
	token1, found1 := rem.db.GetToken(addr1)
	if !found0 || !found1 {
		return fmt.Errorf("SKIP - at least one token is unknown in liquidity removal. pool address: %s", liqPool)
	}

	_, token0HexAmount, token1HexAmount, err := splitBurnDatatoHexStrings(burn.Data)
	if err != nil {
		return err
	}

	//TODO: Create a 'position initialization' function
	remPosition := &Position{
		Address:   burn.Address,
		TxHash:    burn.TransactionHash,
		LowerTick: int(convertHexToBigInt(burn.Topics[2]).Int64()),
		UpperTick: int(convertHexToBigInt(burn.Topics[3]).Int64()),
		Token0: TokenTransaction{
			Token:  token0,
			Amount: convertTransferAmount(token0HexAmount, token0.Decimals),
		},
		Token1: TokenTransaction{
			Token:  token1,
			Amount: convertTransferAmount(token1HexAmount, token1.Decimals),
		},
	}
	rem.Position = *remPosition

	rem.OperationBase.includeTokenPrices(&rem.Position) // 6) Getting token prices
	rem.Position.calculatePosition()                    // 7) Save Liquidity Entry and Liquidity Pool

	return nil
}

func (add *Addition) Extract(mint EventLog) error {
	if !isUniswapPositionsNFT(mint.Data) {
		return fmt.Errorf("not Uniswap Positions NFT.\n")
	}

	txEventsFromCache, _ := add.cache.Get(mint.TransactionHash)

	//TODO: Create a 'position initialization' function
	addPosition := &Position{
		Address:   mint.Address,
		TxHash:    mint.TransactionHash,
		LowerTick: int(convertHexToBigInt(mint.Topics[2]).Int64()),
		UpperTick: int(convertHexToBigInt(mint.Topics[3]).Int64()),
	}
	add.Position = *addPosition

	for _, evLog := range txEventsFromCache.([]EventLog) { // Go through all events of this transaction
		if isTransferEvent(evLog) && !isUniswapPositionsNFT(evLog.Address) {
			add.handleLiquidityTransfer(mint, evLog)
		}
	}

	if isEitherTokenUnknown(*addPosition) {
		add.Position.checkAndUpdateMissingToken(mint, add.OperationBase) // 5) Adding missing token if only 1 token transfer was made
	}
	err := add.savePool(*addPosition)
	if err != nil {
		log.Println("error while adding new pool to database:", err.Error())
	}

	add.OperationBase.includeTokenPrices(&add.Position) // 6) Getting token prices
	add.Position.calculatePosition()                    // 7) Save Liquidity Entry and Liquidity Pool

	return nil
}

func (add Addition) savePool(addPos Position) error {
	if isEitherTokenAmountZero(addPos) || isEitherTokenUnknown(addPos) {
		return nil
	}
	// In this case both tokens were transferred to LP and their order is correct
	var newLiqPoll repository.Pool
	newLiqPoll.Address = addPos.Address
	newLiqPoll.Token0Address = addPos.Token0.Address
	newLiqPoll.Token1Address = addPos.Token1.Address
	return add.db.SavePool(newLiqPoll)
}

func (rem Removal) String() string {
	format := "Removing %f of %s and %f of %s from %s between %f and %f"
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
	format := "Adding %f of %s($%f) and %f of %s($%f) = $%f. To %s between %f and %f while current is %f"
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

// handleLiquidityTransfer decodes Transfer event.
// Getting token that was transferred and calculating amount transferred.
// Keeping track of tokens involved in current Liq. Add. event.
func (add *Addition) handleLiquidityTransfer(mint EventLog, transfer EventLog) {
	_, token0HexAmount, token1HexAmount, err := splitMintDatatoHexFields(mint.Data)
	if err != nil {
		log.Println("Could not split mint event into Amount fields: ", err.Error())
	}

	t, err := add.lookupToken(transfer.Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}

	if transfer.Data == token0HexAmount {
		add.Token0.Token = t
		add.Token0.Amount = convertTransferAmount(token0HexAmount, t.Decimals)
	}

	if transfer.Data == token1HexAmount {
		add.Token1.Token = t
		add.Token1.Amount = convertTransferAmount(token1HexAmount, t.Decimals)
	}
}

// ----------------------------------------------------------------------
// --------------- OperationBase methods

func (ob OperationBase) includeTokenPrices(pos *Position) {
	// Place here to implement price cache?
	if !strings.EqualFold(pos.Token0.Address, "") {
		price, err := ob.lookupPrice(pos.Token0.Address)
		if err != nil {
			log.Println("failed to feetch Token0 price: ", err.Error())
		}
		pos.Token0.Price = price.Value
	}

	if !strings.EqualFold(pos.Token1.Address, "") {
		price, err := ob.lookupPrice(pos.Token1.Address)
		if err != nil {
			log.Println("failed to feetch Token1 price: ", err.Error())
		}
		pos.Token1.Price = price.Value
	}
}

func (op OperationBase) lookupToken(address string) (repository.Token, error) {
	return op.fetchers.tokenFetcher.Token(address)
}

func (op OperationBase) lookupPrice(address string) (repository.TokenPrice, error) {
	return op.fetchers.priceFetcher.Price(address)
}

// ----------------------------------------------------------------------
// --------------- Position methods

func (pos *Position) calculateInterval() {

	lowerRatio := convertTickToRatio(pos.LowerTick, pos.Token0.Decimals, pos.Token1.Decimals)
	upperRatio := convertTickToRatio(pos.UpperTick, pos.Token0.Decimals, pos.Token1.Decimals)

	if isStableOrNativeInvolved(*pos) && isOrderCorrect(*pos) {
		lowerRatio = 1 / lowerRatio
		upperRatio = 1 / upperRatio
	}

	if lowerRatio > upperRatio {
		lowerRatio, upperRatio = upperRatio, lowerRatio
	}
	pos.LowerRatio, pos.UpperRatio = lowerRatio, upperRatio
}

func (pos *Position) calculatePosition() {
	if isEitherTokenUnknown(*pos) {
		return
	}

	pos.calculateInterval() // Decoding / expanding "Mint" event
	pos.adjustOrder()

	if pos.Token0.Price > 0 && pos.Token1.Price > 0 {
		pos.CurrentRatio = pos.Token1.Price / pos.Token0.Price
	}
	pos.TotalValue = pos.Token1.Price*pos.Token1.Amount + pos.Token0.Price*pos.Token0.Amount
}

func (pos *Position) adjustOrder() {
	if isStableOrNativeInvolved(*pos) && !isOrderCorrect(*pos) {
		pos.Token1, pos.Token0 = pos.Token0, pos.Token1
	}
}

// checkAndUpdateMissingToken expands Liq. Add. record if only 1 token was transferred
// Second token is found and appended
// The order of tokens is fixed based on historical results (when 2 tokens were transferred for this LP)
func (pos *Position) checkAndUpdateMissingToken(evLog EventLog, op OperationBase) {
	liqPoolAddress := strings.ToLower(evLog.Address)

	tok0Address, tok1Address, foundPool := op.db.GetPoolPairAddresses(liqPoolAddress)
	if !foundPool {
		log.Println("Could not get token information of pool", liqPoolAddress)
	}

	if pos.Token0.Token == (repository.Token{}) {
		t, err := op.lookupToken(tok0Address)
		if err != nil {
			log.Println("Failed fetching token information: ", err.Error())
		}
		pos.Token0.Token = t
	}

	if pos.Token1.Token == (repository.Token{}) {
		t, err := op.lookupToken(tok1Address)
		if err != nil {
			log.Println("Failed fetching token information: ", err.Error())
		}
		pos.Token1.Token = t
	}

	log.Printf("Added second missing token from known pool %s", liqPoolAddress)
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

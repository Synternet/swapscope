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
	"golang.org/x/exp/slices"
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
	GetByTxHashAndLogType(string, string) ([]EventLog, error)
}

type Operation interface {
	// Common methods shared by Addition, Removal and Swap
	Process(WrappedEventLog) error
	String() string
	CanPublish() bool
	Publish(analytics.Sender, string, time.Time) error
	Save(time.Time) error
}

type Removal struct {
	Position
	OperationBase
	Send analytics.Sender
	//TODO: Fees earned and collected
	Token0Earned TokenTransaction
	Token1Earned TokenTransaction
}

type Addition struct {
	Position
	OperationBase
	Send analytics.Sender
}

type Swap struct {
	Position
	From TokenTransaction
	To   TokenTransaction
	OperationBase
	Send analytics.Sender
}

func (sw Swap) Save(ts time.Time) error {
	return nil
}
func (sw *Swap) Process(swap WrappedEventLog) error {
	swapLog := swap.Log
	addr0, addr1, found := sw.db.GetPoolPairAddresses(swapLog.Address)
	if !found {
		unknownAddressOccurrences[strings.ToLower(swapLog.Address)]++
		unknownAddressTotalOccurences++
		return fmt.Errorf("(swap) pool unknown? %s (%v/%v)", swapLog.Address, unknownAddressOccurrences[strings.ToLower(swapLog.Address)], unknownAddressTotalOccurences)
	}
	token0, found0 := sw.db.GetToken(addr0)
	token1, found1 := sw.db.GetToken(addr1)
	if !found0 || !found1 {
		return fmt.Errorf("SKIP (swap) - at least one token is unknown. Pool address: %s", swapLog.Address)
	}

	hexAmount0, hexAmount1, err := convertLogDataToHexAmounts(swapLog.Data, swap.Instructions.Name)
	if err != nil {
		return err
	}

	swapPosition := &Position{
		Address: swapLog.Address,
		TxHash:  swapLog.TransactionHash,
		Token0: TokenTransaction{
			Token:  token0,
			Amount: convertTransferAmount(hexAmount0, token0.Decimals),
		},
		Token1: TokenTransaction{
			Token:  token1,
			Amount: convertTransferAmount(hexAmount1, token1.Decimals),
		},
	}
	sw.Position = *swapPosition
	sw.adjustOrder()

	if sw.Token0.Amount < 0 && sw.Token1.Amount > 0 {
		sw.From, sw.To = sw.Token1, sw.Token0
	} else if sw.Token1.Amount < 0 && sw.Token0.Amount > 0 {
		sw.From, sw.To = sw.Token0, sw.Token1
	} else {
		return fmt.Errorf("both token amounts are below 0 in swap. TX: %s", sw.TxHash)
	}
	sw.To.Amount = sw.To.Amount * (-1)

	return nil
}

func (sw Swap) String() string {
	format := "Swapping %f of %s to %f of %s."
	return fmt.Sprintf(format,
		sw.From.Amount,
		sw.From.Symbol,
		sw.To.Amount,
		sw.To.Symbol)
}

func (sw Swap) CanPublish() bool {
	if strings.EqualFold(sw.Token0.Address, "") || strings.EqualFold(sw.Token1.Address, "") {
		return false
	}
	return true
}

func (sw Swap) Publish(send analytics.Sender, publishTo string, timestamp time.Time) error {
	swapMessage := types.SwapMessage{
		Timestamp: timestamp,
		TxHash:    sw.TxHash,
		Address:   sw.Address,
		From:      types.TokenMessage{Symbol: sw.From.Symbol, Amount: sw.From.Amount},
		To:        types.TokenMessage{Symbol: sw.To.Symbol, Amount: sw.To.Amount},
	}

	removalJson, err := json.Marshal(&swapMessage)
	if err != nil {
		return fmt.Errorf("error marshalling Liquidity Removal object into a json message: %s", err)
	}

	streamName := strings.ToLower(fmt.Sprintf("%s.%s.%s", publishTo, sw.Token0.Symbol, sw.Token1.Symbol))
	return send(removalJson, streamName)
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

func (rem Removal) getCorespondingRemoval(primaryLog WrappedEventLog) (EventLog, error) {
	removalLogs, err := rem.cache.GetByTxHashAndLogType(primaryLog.Log.TransactionHash, burnEvent)
	if err != nil {
		return EventLog{}, err
	}
	var requiredLog EventLog
	for _, renovalLog := range removalLogs {
		if renovalLog.Address == primaryLog.Log.Address { // Match liquidity pool
			requiredLog = renovalLog
			break
		}
	}
	if requiredLog.Address == "" {
		return EventLog{}, fmt.Errorf("fetched burn event has no liquidity pool.")
	}
	return requiredLog, nil
}

func (rem *Removal) Process(collect WrappedEventLog) error {
	liqPool := collect.Log.Address

	burnLog, err := rem.getCorespondingRemoval(collect)
	if err != nil {
		return err
	}

	addr0, addr1, found := rem.db.GetPoolPairAddresses(liqPool)
	if !found {
		return fmt.Errorf("SKIP - liq. pool is unknown (removal). pool address: %s", liqPool)
	}
	token0, found0 := rem.db.GetToken(addr0)
	token1, found1 := rem.db.GetToken(addr1)
	if !found0 || !found1 {
		return fmt.Errorf("SKIP - at least one token is unknown in liquidity removal. pool address: %s", liqPool)
	}

	token0HexAmount, token1HexAmount, err := convertLogDataToHexAmounts(burnLog.Data, burnEvent)
	if err != nil {
		return err
	}

	//TODO: Create a 'position initialization' function
	remPosition := &Position{
		Address:   burnLog.Address,
		TxHash:    burnLog.TransactionHash,
		LowerTick: int(convertHexToBigInt(burnLog.Topics[2]).Int64()),
		UpperTick: int(convertHexToBigInt(burnLog.Topics[3]).Int64()),
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
	rem.Token0.Price = rem.fetchTokenPrice(rem.Token0.Address)
	rem.Token1.Price = rem.fetchTokenPrice(rem.Token1.Address)

	rem.Position.calculate()

	err = rem.calculateFeesEarned(collect.Log, addr0, addr1)
	if err != nil {
		return err
	}
	return nil
}

func (add *Addition) Process(mint WrappedEventLog) error {
	mintLog := mint.Log
	if !isUniswapPositionsNFT(mintLog.Data) {
		return fmt.Errorf("not Uniswap Positions NFT.\n")
	}
	//TODO: Create a 'position initialization' function
	addPosition := &Position{
		Address:   mintLog.Address,
		TxHash:    mintLog.TransactionHash,
		LowerTick: int(convertHexToBigInt(mintLog.Topics[2]).Int64()),
		UpperTick: int(convertHexToBigInt(mintLog.Topics[3]).Int64()),
	}
	add.Position = *addPosition

	transferLogs, err := add.cache.GetByTxHashAndLogType(mintLog.TransactionHash, transferEvent)
	if err != nil {
		return err
	}
	for _, transferLog := range transferLogs { // Go through all transfers of this transaction
		add.handleLiquidityTransfer(mintLog, transferLog)
	}

	if !add.Position.areTokensSet() {
		add.Position.checkAndUpdateMissingToken(mintLog, add.OperationBase) // 5) Adding missing token if only 1 token transfer was made
	}

	err = add.savePool(add.Position)
	if err != nil {
		log.Println("error while adding new pool to database:", err.Error())
	}

	add.Token0.Price = add.fetchTokenPrice(add.Token0.Address)
	add.Token1.Price = add.fetchTokenPrice(add.Token1.Address)
	add.Position.calculate() // 7) Save Liquidity Entry and Liquidity Pool

	return nil
}

func (add Addition) savePool(addPos Position) error {
	if addPos.isEitherTokenAmountZero() || !addPos.areTokensSet() || (addPos.Token0.Address == addPos.Token1.Address) {
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
	format := "Removing %f of %s and %f of %s from %s. Earned %f of %s and %f of %s ($%f)"
	return fmt.Sprintf(format,
		rem.Token0.Amount,
		rem.Token0.Symbol,
		rem.Token1.Amount,
		rem.Token1.Symbol,
		rem.Address,
		rem.Token0Earned.Amount,
		rem.Token0.Symbol,
		rem.Token1Earned.Amount,
		rem.Token1.Symbol,
		rem.Token0Earned.Amount*rem.Token0.Price+rem.Token1Earned.Amount*rem.Token1.Price,
	)
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

func (rem Removal) Publish(send analytics.Sender, publishTo string, timestamp time.Time) error {
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
		Earned: [2]types.TokensEarnedMessage{
			{Symbol: rem.Token0.Symbol, Amount: rem.Token0Earned.Amount, TotalValueUSD: rem.Token0.Price * rem.Token0Earned.Amount},
			{Symbol: rem.Token1.Symbol, Amount: rem.Token1Earned.Amount, TotalValueUSD: rem.Token1.Price * rem.Token1Earned.Amount},
		},
		ValueEarnedUSD: rem.Token0.Price*rem.Token0Earned.Amount + rem.Token1.Price*rem.Token1Earned.Amount,
		TxHash:         rem.TxHash,
	}

	removalJson, err := json.Marshal(&removalMessage)
	if err != nil {
		return fmt.Errorf("error marshalling Liquidity Removal object into a json message: %s", err)
	}

	streamName := strings.ToLower(fmt.Sprintf("%s.%s.%s", publishTo, rem.Token0.Symbol, rem.Token1.Symbol))
	return send(removalJson, streamName)
}

func (add Addition) Publish(send analytics.Sender, publishTo string, timestamp time.Time) error {
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

	streamName := strings.ToLower(fmt.Sprintf("%s.%s.%s", publishTo, add.Token0.Symbol, add.Token1.Symbol))
	return send(additionJson, streamName)
}

// handleLiquidityTransfer decodes Transfer event.
// Getting token that was transferred and calculating amount transferred.
// Keeping track of tokens involved in current Liq. Add. event.
func (add *Addition) handleLiquidityTransfer(mint EventLog, transfer EventLog) {
	token0HexAmount, token1HexAmount, err := convertLogDataToHexAmounts(mint.Data, mintEvent)
	if err != nil {
		log.Println("Could not split mint event into Amount fields: ", err.Error())
		return
	}

	t, err := add.lookupToken(transfer.Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
		return
	}

	if "0x"+convertHexToBigInt(transfer.Data).Text(16) == token0HexAmount && !strings.EqualFold(transfer.Address, add.Token1.Token.Address) {
		add.Token0.Token = t
		add.Token0.Amount = convertTransferAmount(token0HexAmount, t.Decimals)
	}

	if "0x"+convertHexToBigInt(transfer.Data).Text(16) == token1HexAmount && !strings.EqualFold(transfer.Address, add.Token0.Token.Address) {
		add.Token1.Token = t
		add.Token1.Amount = convertTransferAmount(token1HexAmount, t.Decimals)
	}
}

func (rem *Removal) calculateFeesEarned(collectLog EventLog, poolOrderToken0 string, poolOrderToken1 string) error {
	token0HexAmount, token1HexAmount, err := convertLogDataToHexAmounts(collectLog.Data, collectEvent) // Token order original as in liquidity pool
	if err != nil {
		return err
	}

	rem.Token0Earned, rem.Token1Earned = rem.Token0, rem.Token1

	if strings.EqualFold(rem.Token0.Address, poolOrderToken1) && strings.EqualFold(rem.Token1.Address, poolOrderToken0) {
		rem.Token0Earned, rem.Token1Earned = rem.Token1Earned, rem.Token0Earned // Removed tokens were switched during processing - switch earned tokens as well
		token0HexAmount, token1HexAmount = token1HexAmount, token0HexAmount
	}

	rem.Token0Earned.Amount = convertTransferAmount(token0HexAmount, rem.Token0.Decimals) - rem.Token0.Amount
	rem.Token1Earned.Amount = convertTransferAmount(token1HexAmount, rem.Token1.Decimals) - rem.Token1.Amount

	return nil
}

// ----------------------------------------------------------------------
// --------------- OperationBase methods

func (ob OperationBase) fetchTokenPrice(tokAddress string) float64 {
	// Place here to implement price cache?
	if strings.EqualFold(tokAddress, "") {
		return 0.0
	}
	price, err := ob.lookupPrice(tokAddress)
	if err != nil {
		log.Println("failed to fetch Token0 price: ", err.Error())
	}
	return price.Value
}

func (op OperationBase) lookupToken(address string) (repository.Token, error) {
	return op.fetchers.tokenFetcher.Token(address)
}

func (op OperationBase) lookupPrice(address string) (repository.TokenPrice, error) {
	return op.fetchers.priceFetcher.Price(address)
}

// ----------------------------------------------------------------------
// --------------- Position methods

func (pos *Position) calculateRatios() {

	lowerRatio := convertTickToRatio(pos.LowerTick, pos.Token0.Decimals, pos.Token1.Decimals)
	upperRatio := convertTickToRatio(pos.UpperTick, pos.Token0.Decimals, pos.Token1.Decimals)

	if (pos.isAnyTokenOneOf(stableCoins) && !pos.isToken1OneOf(stableCoins)) || // Stable (USDC, USDT) to always be quote token (second)
		(!pos.isAnyTokenOneOf(stableCoins) && pos.isAnyTokenOneOf(nativeCoins) && !pos.isToken1OneOf(nativeCoins)) { // If there is no stable token involved - WETH will always be quoto token
		lowerRatio = 1 / lowerRatio
		upperRatio = 1 / upperRatio
	}
	if lowerRatio > upperRatio {
		lowerRatio, upperRatio = upperRatio, lowerRatio
	}
	pos.LowerRatio, pos.UpperRatio = lowerRatio, upperRatio
}

func (pos *Position) calculate() {
	if !pos.areTokensSet() {
		return
	}

	pos.calculateRatios()
	pos.adjustOrder()

	if pos.Token0.Price > 0 && pos.Token1.Price > 0 {
		pos.CurrentRatio = pos.Token0.Price / pos.Token1.Price
	}
	pos.TotalValue = pos.Token1.Price*pos.Token1.Amount + pos.Token0.Price*pos.Token0.Amount
}

func (pos *Position) adjustOrder() {
	if (pos.isAnyTokenOneOf(nativeCoins) && !pos.isAnyTokenOneOf(stableCoins) && !pos.isToken1OneOf(nativeCoins)) ||
		(pos.isAnyTokenOneOf(stableCoins) && !pos.isToken1OneOf(stableCoins)) {
		pos.Token1, pos.Token0 = pos.Token0, pos.Token1
	}
}

// checkAndUpdateMissingToken expands Liq. Add. record if only 1 token was transferred
// Second token is found and appended
func (pos *Position) checkAndUpdateMissingToken(evLog EventLog, op OperationBase) {
	liqPoolAddress := strings.ToLower(evLog.Address)

	tok0Address, tok1Address, foundPool := op.db.GetPoolPairAddresses(liqPoolAddress)
	if !foundPool {
		log.Println("(at least 1 token is completely unknown) Could not get token information of pool", liqPoolAddress)
		return
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
	if p.Token0.Price == 0.0 || p.Token1.Price == 0.0 {
		log.Printf("SKIP - missing price (could not calculate current ratio). Tx: %s\n\n", p.TxHash)
		return false
	}
	if !p.isAnyTokenOneOf(nativeCoins) && !p.isAnyTokenOneOf(stableCoins) {
		log.Printf("SKIP - no stable or native currency involved. Tx: %s\n\n", p.TxHash)
		return false
	}
	return true
}

func (p Position) areTokensSet() bool {
	return (!strings.EqualFold(p.Token0.Address, "") && !strings.EqualFold(p.Token1.Address, ""))
}

func (p Position) isEitherTokenAmountZero() bool {
	return (p.Token0.Amount == 0 || p.Token1.Amount == 0)
}

func (p Position) isToken1OneOf(tokens []string) bool {
	return slices.Contains(tokens, strings.ToLower(p.Token1.Address))
}

func (p Position) isToken0OneOf(tokens []string) bool {
	return slices.Contains(tokens, strings.ToLower(p.Token0.Address))
}

func (p Position) isAnyTokenOneOf(tokens []string) bool {
	return p.isToken0OneOf(tokens) || p.isToken1OneOf(tokens)
}

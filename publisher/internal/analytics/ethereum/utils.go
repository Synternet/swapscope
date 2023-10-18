package ethereum

import (
	"log"
	"strings"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/patrickmn/go-cache"
)

// checkAndUpdateMissingToken expands Liq. Add. record if only 1 token was transferred
// Second token is found and appended
// The order of tokens is fixed based on historical results (when 2 tokens were transferred for this LP)
func (a *Analytics) checkAndUpdateMissingToken(evLog EventLog, liqAdd LiquidityAddition) LiquidityAddition {
	liqPoolAddress := strings.ToLower(evLog.Address)
	if !strings.EqualFold(liqAdd.Token1.Address, "") {
		return liqAdd
	}

	tok0Address, tok1Address, foundPool := a.db.GetTokenPairAddresses(liqPoolAddress)
	if !foundPool {
		log.Println("Could not get token information of pool", liqPoolAddress)
		return liqAdd
	}

	tokenInOrder0, err := a.tokenFetcher.Token(tok0Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}
	tokenInOrder1, err := a.tokenFetcher.Token(tok1Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}
	liqAdd = updateOrderOfTokens(liqAdd, tokenInOrder0, tokenInOrder1)

	log.Println("Added second missing token from known pool.")
	return liqAdd
}

func (a *Analytics) finalizeLiquidityEntry(evLog EventLog, liqAdd LiquidityAddition) (LiquidityAddition, bool) {
	var lower, upper float64
	liqAdd, lower, upper = a.handleLiquidityPosition(evLog, liqAdd) // Decoding / expanding "Mint" event
	liqAdd.LowerRatio = lower
	liqAdd.UpperRatio = upper
	liqAdd.LPoolAddress = evLog.Address
	liqAdd.TxHash = evLog.TransactionHash
	if liqAdd.Token0.Price > 0 && liqAdd.Token1.Price > 0 {
		liqAdd.CurrentRatio = liqAdd.Token1.Price / liqAdd.Token0.Price
		liqAdd.ValueAdded = liqAdd.Token1.Price*liqAdd.Token1.Amount + liqAdd.Token0.Price*liqAdd.Token0.Amount
	}
	isFilterSatisfied := a.canLiqAddRecordSatisfyFilters(liqAdd)
	if isFilterSatisfied {
		a.addNewLiquidityPool(liqAdd)
		liqAdd.setLiquidityRecordFields()
		a.db.AddLiquidityPoolAddition(liqAdd.LiquidityEntry) // Saving the Liquidity addition record
		log.Printf("Added: %s\n\n", liqAdd.prettyPrintLiquidityRecord())
	}
	return liqAdd, isFilterSatisfied
}

// updateOrderOfTokens updates order of tokens in liquidity add record based on order in actual LP
func updateOrderOfTokens(liqAdd LiquidityAddition, token0OrderInPool repository.Token, token1OrderInPool repository.Token) LiquidityAddition {
	if strings.EqualFold(liqAdd.Token0.Address, token0OrderInPool.Address) { // The received token is first as it should be
		liqAdd.Token1.Amount = 0.0
		liqAdd.Token1.Token = token1OrderInPool
		log.Println("First token received was correct.")
		return liqAdd
	}
	liqAdd.Token1 = liqAdd.Token0
	liqAdd.Token0.Token = token0OrderInPool
	liqAdd.Token0.Amount = 0.0
	return liqAdd
}

// addLogToTxCache adds event log to cache.
// Cache is grouped by transaction hashes.
// When Mint event is met, all logs until then are recovered from cache.
// Cache clears every 1 minute (~5 ETH blocks).
func (a *Analytics) addLogToTxCache(eLog EventLog) error {
	txHash := eLog.TransactionHash
	tempCacheValue, tempFound := a.eventLogCache.Get(txHash)
	var newCacheValue []EventLog
	if !tempFound {
		newCacheValue = []EventLog{eLog} // Initialize
	} else {
		newCacheValue = append(tempCacheValue.([]EventLog), eLog) // Append
	}
	a.eventLogCache.Set(txHash, newCacheValue, cache.DefaultExpiration)
	return nil
}

// updateLiqAddRecordWithTransfer decodes Transfer event.
// Getting token that was transferred and calculating amount transferred.
// Keeping track of tokens involved in current Liq. Add. event.
func (a *Analytics) handleLiquidityTransfer(evLog EventLog, liqAdd LiquidityAddition) LiquidityAddition {
	tokenAddress := evLog.Address
	if isUniswapPositionsNFT(tokenAddress) {
		log.Println("Uniswap positions NFT transfer.")
		return liqAdd
	}

	t, err := a.tokenFetcher.Token(tokenAddress)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}
	amountScaled := a.convertTransferAmount(evLog.Data, t.Decimals)
	log.Printf("Transfer %f of %s(%s)", amountScaled, tokenAddress, t.Symbol)

	// Do not include transactions which transferred 0 (usually there is another just after this one)
	if amountScaled == 0 || strings.EqualFold(liqAdd.Token0.Address, tokenAddress) {
		return liqAdd
	}

	if strings.EqualFold(liqAdd.Token0.Address, "") {
		liqAdd.Token0.Token = t
		liqAdd.Token0.Amount = amountScaled
	} else if strings.EqualFold(liqAdd.Token1.Address, "") {
		liqAdd.Token1.Token = t
		liqAdd.Token1.Amount = amountScaled
	}
	return liqAdd
}

// handleLiquidityPosition decodes Mint event.
// Getting pool address, ticks. Ticks are converted to understandable token ratios.
// Some adjustments are done if required (switching places).
func (a *Analytics) handleLiquidityPosition(evLog EventLog, liqAdd LiquidityAddition) (LiquidityAddition, float64, float64) {
	if isEitherTokenUnknown(liqAdd) {
		return liqAdd, 0, 0
	}
	liqAdd.LowerTick = int(convertHexToBigInt(evLog.Topics[2]).Int64())
	liqAdd.UpperTick = int(convertHexToBigInt(evLog.Topics[3]).Int64())
	lowerCut, upperCut, toReverse := convertTicksToRatios(liqAdd)

	if toReverse {
		liqAdd.Token0, liqAdd.Token1 = liqAdd.Token1, liqAdd.Token0
	}

	if lowerCut > upperCut {
		lowerCut, upperCut = upperCut, lowerCut
	}

	return liqAdd, lowerCut, upperCut
}

func (a *Analytics) addNewLiquidityPool(liqAdd LiquidityAddition) {
	if isEitherTokenAmountIsZero(liqAdd) {
		return
	}
	// In this case both tokens were transferred to LP and their order is correct
	var newLiqPoll repository.LiquidityPool
	newLiqPoll.Address = liqAdd.LPoolAddress
	newLiqPoll.Token0Address = liqAdd.Token0.Address
	newLiqPoll.Token1Address = liqAdd.Token1.Address
	a.db.AddLiquidityPool(newLiqPoll)
	return
}

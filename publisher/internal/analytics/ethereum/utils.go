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
func (a *Analytics) checkAndUpdateMissingToken(evLog EventLog, addPos Position) Position {
	liqPoolAddress := strings.ToLower(evLog.Address)

	tok0Address, tok1Address, foundPool := a.db.GetPoolPairAddresses(liqPoolAddress)
	if !foundPool {
		log.Println("Could not get token information of pool", liqPoolAddress)
		return addPos
	}

	tokenInOrder0, err := a.tokenFetcher.Token(tok0Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}
	tokenInOrder1, err := a.tokenFetcher.Token(tok1Address)
	if err != nil {
		log.Println("Failed fetching token information: ", err.Error())
	}
	addPos = updateOrderOfTokens(addPos, tokenInOrder0, tokenInOrder1)

	log.Printf("Added second missing token from known pool %s", liqPoolAddress)
	return addPos
}

func (a *Analytics) calculatePosition(evLog EventLog, addPos Position) Position {
	addPos.LowerTick = int(convertHexToBigInt(evLog.Topics[2]).Int64())
	addPos.UpperTick = int(convertHexToBigInt(evLog.Topics[3]).Int64())
	addPos.Address = evLog.Address
	addPos.TxHash = evLog.TransactionHash
	a.savePool(addPos)

	addPos = a.decodeLowerUpperTicks(addPos) // Decoding / expanding "Mint" event

	if addPos.Token0.Price > 0 && addPos.Token1.Price > 0 {
		addPos.CurrentRatio = addPos.Token1.Price / addPos.Token0.Price
	}

	addPos.TotalValue = addPos.Token1.Price*addPos.Token1.Amount + addPos.Token0.Price*addPos.Token0.Amount

	return addPos
}

// updateOrderOfTokens updates order of tokens in liquidity add record based on order in actual LP
func updateOrderOfTokens(addPos Position, correctOrderToken0 repository.Token, correctOrderToken1 repository.Token) Position {
	if strings.EqualFold(addPos.Token0.Address, correctOrderToken0.Address) { // The received token is first as it should be
		addPos.Token1.Amount = 0.0
		addPos.Token1.Token = correctOrderToken1
		log.Println("First token received was correct.")
		return addPos
	}
	addPos.Token1 = addPos.Token0
	addPos.Token0.Token = correctOrderToken0
	addPos.Token0.Amount = 0.0
	return addPos
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
func (a *Analytics) handleLiquidityTransfer(evLog EventLog, liqAdd Position) Position {
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
	//log.Printf("Transfer %f of %s(%s)", amountScaled, tokenAddress, t.Symbol)

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

func (a *Analytics) decodeLowerUpperTicks(position Position) Position {
	if isEitherTokenUnknown(position) {
		return position
	}
	toReverse := false
	position.LowerRatio, position.UpperRatio, toReverse = convertTicksToRatios(position)

	if toReverse {
		position.Token0, position.Token1 = position.Token1, position.Token0
	}

	if position.LowerRatio > position.UpperRatio {
		position.LowerRatio, position.UpperRatio = position.UpperRatio, position.LowerRatio
	}
	return position
}

func (a *Analytics) savePool(addPos Position) {
	if isEitherTokenAmountIsZero(addPos) || isEitherTokenUnknown(addPos) {
		return
	}
	// In this case both tokens were transferred to LP and their order is correct
	var newLiqPoll repository.Pool
	newLiqPoll.Address = addPos.Address
	newLiqPoll.Token0Address = addPos.Token0.Address
	newLiqPoll.Token1Address = addPos.Token1.Address
	a.db.SavePool(newLiqPoll)
	return
}

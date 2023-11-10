package ethereum

import (
	"github.com/patrickmn/go-cache"
)

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

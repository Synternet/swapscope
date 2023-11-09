package ethereum

import (
	"github.com/patrickmn/go-cache"
)

func calculatePosition(pos *Position) {
	if isEitherTokenUnknown(*pos) {
		return
	}

	pos.LowerRatio, pos.UpperRatio = calculateInterval(*pos) // Decoding / expanding "Mint" event
	pos.Token0, pos.Token1 = adjustOrder(*pos)

	if pos.Token0.Price > 0 && pos.Token1.Price > 0 {
		pos.CurrentRatio = pos.Token1.Price / pos.Token0.Price
	}
	pos.TotalValue = pos.Token1.Price*pos.Token1.Amount + pos.Token0.Price*pos.Token0.Amount
}

func adjustOrder(pos Position) (TokenTransaction, TokenTransaction) {
	if isStableOrNativeInvolved(pos) && !isOrderCorrect(pos) {
		return pos.Token1, pos.Token0
	}
	return pos.Token0, pos.Token1
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

func calculateInterval(position Position) (float64, float64) {

	lowerRatio := convertTickToRatio(position.LowerTick, position.Token0.Decimals, position.Token1.Decimals)
	upperRatio := convertTickToRatio(position.UpperTick, position.Token0.Decimals, position.Token1.Decimals)

	if isStableOrNativeInvolved(position) && isOrderCorrect(position) {
		lowerRatio = 1 / lowerRatio
		upperRatio = 1 / upperRatio
	}

	if lowerRatio > upperRatio {
		lowerRatio, upperRatio = upperRatio, lowerRatio
	}

	return lowerRatio, upperRatio
}

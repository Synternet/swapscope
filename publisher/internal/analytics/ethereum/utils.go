package ethereum

import (
	"github.com/patrickmn/go-cache"
)

func calculatePosition(evLog EventLog, addPos Position) Position {
	addPos.LowerTick = int(convertHexToBigInt(evLog.Topics[2]).Int64())
	addPos.UpperTick = int(convertHexToBigInt(evLog.Topics[3]).Int64())
	addPos.Address = evLog.Address
	addPos.TxHash = evLog.TransactionHash
	//a.savePool(addPos)

	addPos = decodeLowerUpperTicks(addPos) // Decoding / expanding "Mint" event

	if addPos.Token0.Price > 0 && addPos.Token1.Price > 0 {
		addPos.CurrentRatio = addPos.Token1.Price / addPos.Token0.Price
	}

	addPos.TotalValue = addPos.Token1.Price*addPos.Token1.Amount + addPos.Token0.Price*addPos.Token0.Amount

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

func decodeLowerUpperTicks(position Position) Position {
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

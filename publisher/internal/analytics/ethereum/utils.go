package ethereum

import (
	"github.com/patrickmn/go-cache"
)

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

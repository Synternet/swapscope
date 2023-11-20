package ethereum

import (
	"fmt"

	"github.com/patrickmn/go-cache"
)

// addLogToTxCache adds event log to cache.
// Cache is grouped by transaction hashes.
// When Mint event is met, all logs until then are recovered from cache.
// Cache clears every 1 minute (~5 ETH blocks).
func (a *Analytics) addLogToTxCache(wel WrappedEventLog) error {
	txHash := wel.Log.TransactionHash
	logType := wel.Instructions.Name
	tempCacheValue, tempFound := a.eventLogCache.Get(txHash)
	var newCacheValue CacheRecord

	if !tempFound {
		newCacheValue = CacheRecord{logType: []EventLog{wel.Log}} // Initialize - there is absolutely nothing for this tx hash
		a.eventLogCache.Set(txHash, newCacheValue, cache.DefaultExpiration)
		return nil
	}

	tempCacheRecord := tempCacheValue.(CacheRecord)

	if events, ok := tempCacheRecord[logType].([]EventLog); ok {
		moreEvents := append(events, wel.Log)
		tempCacheRecord[logType] = moreEvents // Append - there is cache for this tx, and this log type
		newCacheValue = tempCacheRecord
	} else {
		tempCacheRecord[logType] = []EventLog{wel.Log} // Initialize - there is cache for this tx, but not for this log type
		newCacheValue = tempCacheRecord
	}

	a.eventLogCache.Set(txHash, newCacheValue, cache.DefaultExpiration)
	return nil
}

func (c EventLogCache) GetByTxHashAndLogType(txHash string, logType string) ([]EventLog, error) {

	txEventsFromCache, found := c.Get(txHash)
	if !found {
		return []EventLog{}, fmt.Errorf("could not find records of tx %s in logs cache", txHash)
	}

	txCacheRecords := txEventsFromCache.(CacheRecord)
	var resFilteredByHashAndType []EventLog
	var ok bool
	if resFilteredByHashAndType, ok = txCacheRecords[logType].([]EventLog); !ok {
		return []EventLog{}, fmt.Errorf("could not find records of type %s in logs cache (tx %s in cache exists)", logType, txHash)
	}

	return resFilteredByHashAndType, nil
}

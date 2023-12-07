package ethereum

import (
	"fmt"
	"strings"

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

func (a *Analytics) newWrappedEventLog(eLog EventLog) WrappedEventLog {
	var wel WrappedEventLog
	wel.Log = eLog

	initOpBase := OperationBase{
		db:    a.db,
		cache: a.eventLogCache,
		fetchers: Fetchers{
			priceFetcher: a.priceFetcher,
			tokenFetcher: a.tokenFetcher,
		},
	}

	switch {
	case a.isTransfer(eLog):
		wel.Instructions = EventInstruction{
			Name:      transferEvent,
			Header:    ethereumErc20TokenABI.Events[transferEvent].Sig,
			Signature: a.eventSignature[transferEvent],
			Operation: nil,
			PublishTo: "",
		}
	case a.isMint(eLog):
		wel.Instructions = EventInstruction{
			Name:      mintEvent,
			Header:    uniswapLiqPoolsABI.Events[mintEvent].Sig,
			Signature: a.eventSignature[mintEvent],
			Operation: &Addition{OperationBase: initOpBase},
			PublishTo: "add",
		}
	case a.isBurn(eLog):
		wel.Instructions = EventInstruction{
			Name:      burnEvent,
			Header:    uniswapLiqPoolsABI.Events[burnEvent].Sig,
			Signature: a.eventSignature[burnEvent],
		}
	case a.isCollect(eLog):
		wel.Instructions = EventInstruction{
			Name:      collectEvent,
			Header:    uniswapLiqPoolsABI.Events[collectEvent].Sig,
			Signature: a.eventSignature[collectEvent],
			Operation: &Removal{OperationBase: initOpBase},
			PublishTo: "remove",
		}
	case a.isSwap(eLog):
		wel.Instructions = EventInstruction{
			Name:      swapEvent,
			Header:    uniswapLiqPoolsABI.Events[swapEvent].Sig,
			Signature: a.eventSignature[swapEvent],
			Operation: &Swap{OperationBase: initOpBase},
			PublishTo: "swap",
		}
	default:
		wel.Instructions = EventInstruction{
			Name: "OTHER",
		}
	}

	return wel
}

func (a *Analytics) isTransfer(el EventLog) bool {
	return strings.HasPrefix(el.Topics[0], a.eventSignature[transferEvent])
}

func (a *Analytics) isMint(el EventLog) bool {
	return strings.HasPrefix(el.Topics[0], a.eventSignature[mintEvent])
}

func (a *Analytics) isBurn(el EventLog) bool {
	return strings.HasPrefix(el.Topics[0], a.eventSignature[burnEvent])
}

func (a *Analytics) isCollect(el EventLog) bool {
	return strings.HasPrefix(el.Topics[0], a.eventSignature[collectEvent])
}

func (a *Analytics) isSwap(el EventLog) bool {
	return strings.HasPrefix(el.Topics[0], a.eventSignature[swapEvent])
}

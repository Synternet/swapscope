package ethereum

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
	"github.com/SyntropyNet/swapscope/publisher/pkg/types"
)

const (
	pubStreamName = "liquiditytest0"
)

func (a *Analytics) ProcessMessage(msg analytics.Message, send analytics.Sender) error {
	eLog := parseEventLogMessage(msg.Data)

	if !hasTopics(eLog) {
		return nil
	}
	// 1) All events are put into cache
	a.addLogToTxCache(eLog)

	if !a.isMintEvent(eLog) || (a.isMintEvent(eLog) && !isUniswapPositionsNFT(eLog.Data)) {
		return nil
	}

	// 2) Mint event is found
	var liquidityEntry LiquidityAddition
	txEventsFromCache, _ := a.eventLogCache.Get(eLog.TransactionHash)
	for _, evLog := range txEventsFromCache.([]EventLog) { // Go through all events of this transaction
		if reflect.DeepEqual(evLog, eLog) { // This allows to correctly process many Mint events in one transaction
			break
		}
		if a.isAvoidEvent(evLog) {
			liquidityEntry = LiquidityAddition{} // Reset gathered records
			continue
		}
		if a.isTransferEvent(evLog) { // 3) Searching for relevant "Transfer" event(s)
			liquidityEntry = a.handleLiquidityTransfer(evLog, liquidityEntry) // 4) Decoding / expanding "Transfer" events
		}
	}
	liquidityEntry = a.checkAndUpdateMissingToken(eLog, liquidityEntry)         // 5) Adding missing token if only 1 token transfer was made
	liquidityEntry = a.includeTokenPrices(liquidityEntry)                       // 6) Getting token prices
	liquidityEntry, doPublish := a.finalizeLiquidityEntry(eLog, liquidityEntry) // 7) Save Liquidity Entry and Liquidity Pool

	if !doPublish {
		return nil
	}
	message, err := makeMessage(liquidityEntry)
	if err != nil {
		log.Println(err)
		return err
	}
	streamName := strings.ToLower(fmt.Sprintf("%s.%s", liquidityEntry.Token0Symbol, liquidityEntry.Token1Symbol))
	err = send(message, streamName)
	return err
}

func makeMessage(liqAdd LiquidityAddition) ([]byte, error) {

	additionMessage := types.AdditionMessage{
		Timestamp:         time.Now().UTC(), // Add a timestamp field
		Address:           liqAdd.LPoolAddress,
		LowerTokenRatio:   liqAdd.LowerRatio,
		CurrentTokenRatio: liqAdd.CurrentRatio,
		UpperTokenRatio:   liqAdd.UpperRatio,
		ValueAddedUSD:     liqAdd.ValueAdded,
		Pair: [2]types.TokenMessage{
			{Symbol: liqAdd.Token0Symbol, Amount: liqAdd.Token0Amount, Price: liqAdd.Token0PriceUsd},
			{Symbol: liqAdd.Token1Symbol, Amount: liqAdd.Token1Amount, Price: liqAdd.Token1PriceUsd},
		},
		TxHash: liqAdd.TxHash,
	}

	additionJson, err := json.Marshal(&additionMessage)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling Liquidity Addition object into a json message: %s", err)
	}

	return additionJson, nil
}

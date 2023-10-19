package ethereum

import (
	"log"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
)

func (a *Analytics) ProcessMessage(msg analytics.Message, send analytics.Sender) error {
	eLog := parseEventLogMessage(msg.Data)

	if !hasTopics(eLog) {
		return nil
	}

	a.addLogToTxCache(eLog) //All events are put into cache

	var operation Operation
	switch {
	case a.isBurnEvent(eLog):
		operation = Removal{}
	case a.isMintEvent(eLog):
		operation = Addition{}
	default:
		return nil
	}

	var err error
	operation, err = operation.ExtractFromEventLogs(eLog, a)
	if err != nil {
		log.Printf("%s\n", err)
		return nil
		//return err //TODO: if error is returned - whole service is stopped
	}

	if !operation.CanPublish() {
		return nil
	}

	log.Println(eLog.TransactionHash)
	log.Println(operation.PrintPretty())

	//return operation.SaveToDB(msg.Timestamp, a) // Do not save liquidity additions and removals to DB
	return operation.PublishToNATS(msg.Timestamp, send)
}

package ethereum

import (
	"log"

	"github.com/SyntropyNet/swapscope/publisher/pkg/analytics"
)

func (a *Analytics) ProcessMessage(msg analytics.Message, send analytics.Sender) error {
	eLog, err := parseEventLogMessage(msg.Data)

	if err != nil {
		log.Println("Failed to parse event log from message: ", err.Error())
		return nil
	}

	a.addLogToTxCache(eLog.Data) // All events are put into cache

	if eLog.Instructions.Operation == nil { // There is no way to turn this log into an operation - processing is done
		return nil
	}

	operation := eLog.Instructions.Operation                      // Set to correct type
	operation.InitializeOperation(a.db, a.eventLogCache, a, send) // Initialize with operation base (db, cache, fetchers) and send function

	err = operation.Extract(eLog.Data)
	if err != nil {
		log.Println("Failed to extract event from logs: ", err.Error())
		return nil
		//return err //TODO: currently if error is returned - whole service (goroutine) is stopped - should not be like this?
	}

	if !operation.CanPublish() {
		return nil
	}

	log.Println("Tx hash:", eLog.Data.TransactionHash)
	log.Println("Operation processed:", operation.String())

	//return operation.Save(msg.Timestamp) // Option to save additions and removals to DB
	return operation.Publish(msg.Timestamp)
}

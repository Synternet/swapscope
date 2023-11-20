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

	wrappedLog := a.newWrappedEventLog(eLog)
	a.addLogToTxCache(wrappedLog) // All events are put into cache

	if wrappedLog.Instructions.Operation == nil { // There is no way to turn this log into an operation - processing is done
		return nil
	}

	operation := wrappedLog.Instructions.Operation // Set to correct type

	err = operation.Process(wrappedLog)
	if err != nil {
		log.Println("Failed to extract event from logs: ", err.Error())
		return nil
		//return err //TODO: currently if error is returned - whole service (goroutine) is stopped - should not be like this?
	}

	if !operation.CanPublish() {
		return nil
	}

	log.Println("Tx hash:", wrappedLog.Log.TransactionHash)
	log.Println("Operation processed:", operation.String())

	//return operation.Save(msg.Timestamp) // Option to save additions and removals to DB
	return operation.Publish(send, wrappedLog.Instructions.PublishTo, msg.Timestamp)
}

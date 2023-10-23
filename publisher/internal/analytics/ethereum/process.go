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
		operation = NewRemovalOperation(a.db, a.eventLogCache, a, send)
	case a.isMintEvent(eLog):
		operation = NewAdditionOperation(a.db, a.eventLogCache, a, send)
	default:
		return nil
	}

	var err error
	operation, err = operation.Extract(eLog)
	if err != nil {
		log.Println("Failed to extract event from logs: ", err.Error())
		return nil
		//return err //TODO: if error is returned - whole service is stopped
	}

	if !operation.CanPublish() {
		return nil
	}

	log.Println("Tx hash:", eLog.TransactionHash)
	log.Println("Operation processed:", operation.String())

	//return operation.Save(msg.Timestamp) // Option to save additions and removals to DB
	return operation.Publish(msg.Timestamp)
}

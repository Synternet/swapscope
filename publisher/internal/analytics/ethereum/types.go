package ethereum

import (
	"github.com/Synternet/swapscope/publisher/pkg/repository"
)

type EventLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type Position struct {
	Address      string
	Token0       TokenTransaction
	Token1       TokenTransaction
	LowerRatio   float64
	CurrentRatio float64
	UpperRatio   float64
	TotalValue   float64
	LowerTick    int
	UpperTick    int
	TxHash       string
}

type TokenTransaction struct {
	repository.Token
	Amount float64
	Price  float64
}

type EventInstruction struct {
	Name      string
	Header    string
	Signature string
	PublishTo string
	Operation
}

type WrappedEventLog struct {
	Log          EventLog
	Instructions EventInstruction
}

func (el *EventLog) hasTopics() bool {
	for _, str := range el.Topics {
		if str != "" {
			return true
		}
	}
	return false
}

package ethereum

import (
	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
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

type Removal struct {
	Position
	AnalyticsInterface
	DatabaseInterface
	CacheInterface
	//TODO: Fees earned and collected
	//TokenEarned0 TokenTransaction
	//TokenEarned1 TokenTransaction
}

type Addition struct {
	Position
	AnalyticsInterface
	DatabaseInterface
	CacheInterface
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

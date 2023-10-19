package ethereum

import (
	"fmt"

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

type LiquidityAddition struct {
	repository.LiquidityEntry
	Token0       TokenAddition
	Token1       TokenAddition
	LowerTick    int
	UpperTick    int
	CurrentRatio float64
	ValueAdded   float64
}

type TokenAddition struct {
	repository.Token
	Amount float64
	Price  float64
}

func (intRec *LiquidityAddition) setLiquidityRecordFields() {
	intRec.Token0Amount = intRec.Token0.Amount
	intRec.Token1Amount = intRec.Token1.Amount
	intRec.Token0PriceUsd = intRec.Token0.Price
	intRec.Token1PriceUsd = intRec.Token1.Price
	intRec.Token0Symbol = intRec.Token0.Symbol
	intRec.Token1Symbol = intRec.Token1.Symbol
}

func (intRec *LiquidityAddition) prettyPrintLiquidityRecord() string {
	format := "%f of %s(%f) and %f of %s(%f) to %s between %f and %f (tx: %s)"
	return fmt.Sprintf(format,
		intRec.Token0Amount,
		intRec.Token0Symbol,
		intRec.Token0PriceUsd,
		intRec.Token1Amount,
		intRec.Token1Symbol,
		intRec.Token1PriceUsd,
		intRec.LPoolAddress,
		intRec.LowerRatio,
		intRec.UpperRatio,
		intRec.TxHash)
}

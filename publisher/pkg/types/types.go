package types

import (
	"time"
)

type AdditionMessage struct {
	Timestamp         time.Time       `json:"timestamp"`
	Address           string          `json:"address"`
	LowerTokenRatio   float64         `json:"lowerTokenRatio"`
	CurrentTokenRatio float64         `json:"currentTokenRatio"`
	UpperTokenRatio   float64         `json:"upperTokenRatio"`
	ValueAddedUSD     float64         `json:"totalValueUSD"`
	Pair              [2]TokenMessage `json:"pair"`
	TxHash            string          `json:"txHash"`
}

type RemovalMessage struct {
	Timestamp         time.Time       `json:"timestamp"`
	Address           string          `json:"address"`
	LowerTokenRatio   float64         `json:"lowerTokenRatio"`
	CurrentTokenRatio float64         `json:"currentTokenRatio"`
	UpperTokenRatio   float64         `json:"upperTokenRatio"`
	ValueRemovedUSD   float64         `json:"totalValueUSD"`
	Pair              [2]TokenMessage `json:"pair"`
	TxHash            string          `json:"txHash"`
}

type TokenMessage struct {
	Symbol string  `json:"symbol"`
	Amount float64 `json:"amount"`
	Price  float64 `json:"priceUSD"`
}

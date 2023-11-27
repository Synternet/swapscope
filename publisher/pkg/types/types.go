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
	Timestamp         time.Time              `json:"timestamp"`
	Address           string                 `json:"address"`
	LowerTokenRatio   float64                `json:"lowerTokenRatio"`
	CurrentTokenRatio float64                `json:"currentTokenRatio"`
	UpperTokenRatio   float64                `json:"upperTokenRatio"`
	ValueRemovedUSD   float64                `json:"totalValueUSD"`
	Pair              [2]TokenMessage        `json:"pair"`
	Earned            [2]TokensEarnedMessage `json:"earned"`
	ValueEarnedUSD    float64                `json:"totalEarnedUSD"`
	TxHash            string                 `json:"txHash"`
}

type SwapMessage struct {
	Timestamp time.Time    `json:"timestamp"`
	Address   string       `json:"address"`
	TxHash    string       `json:"txHash"`
	From      TokenMessage `json:"from"`
	To        TokenMessage `json:"to"`
}

type TokenMessage struct {
	Symbol string  `json:"symbol"`
	Amount float64 `json:"amount"`
	Price  float64 `json:"priceUSD,omitempty"`
}

type TokensEarnedMessage struct {
	Symbol        string  `json:"symbol"`
	Amount        float64 `json:"amount"`
	TotalValueUSD float64 `json:"totalValueUSD"`
}

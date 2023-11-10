package repository

import "time"

type TokenPrice struct {
	Value float64
	Base  string
}

type Token struct {
	Address     string
	Symbol      string
	Name        string
	Decimals    int
	TotalSupply float64
}

type Pool struct {
	Address       string
	Token0Address string
	Token1Address string
}

type Addition struct {
	TimestampReceived time.Time
	LPoolAddress      string
	Token0Symbol      string
	Token1Symbol      string
	Token0Amount      float64
	Token1Amount      float64
	LowerRatio        float64
	UpperRatio        float64
	Token0PriceUsd    float64
	Token1PriceUsd    float64
	TxHash            string
}

type Removal struct {
	TimestampReceived time.Time
	LPoolAddress      string
	Token0Symbol      string
	Token1Symbol      string
	Token0Amount      float64
	Token1Amount      float64
	LowerRatio        float64
	UpperRatio        float64
	Token0PriceUsd    float64
	Token1PriceUsd    float64
	TxHash            string
}

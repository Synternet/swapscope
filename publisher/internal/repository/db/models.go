package db

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
}

type Token struct {
	Timestamp_added time.Time `gorm:"autoCreateTime:true"`
	Address         string
	Symbol          string
	Name            string
	Decimals        int
}

type Addition struct {
	TimestampAdded    time.Time `gorm:"autoCreateTime:true"`
	TimestampReceived time.Time
	LPoolAddress      string
	Token0Symbol      string
	Token1Symbol      string
	Token0Amount      float64
	Token1Amount      float64
	LowerActualRatio  float64
	UpperActualRatio  float64
	Token0PriceUsd    float64
	Token1PriceUsd    float64
	TxHash            string
}

type Removal struct {
	TimestampAdded    time.Time `gorm:"autoCreateTime:true"`
	TimestampReceived time.Time
	LPoolAddress      string
	Token0Symbol      string
	Token1Symbol      string
	Token0Amount      float64
	Token1Amount      float64
	LowerActualRatio  float64
	UpperActualRatio  float64
	Token0PriceUsd    float64
	Token1PriceUsd    float64
	TxHash            string
}

type Swap struct {
	TimestampAdded    time.Time `gorm:"autoCreateTime:true"`
	TimestampReceived time.Time
	LPoolAddress      string
	TokenAddressFrom  string
	TokenFromAmount   float64
	TokenAddressTo    string
	TokenToAmount     float64
	TxHash            string
}

type Pool struct {
	Timestamp_added time.Time `gorm:"autoCreateTime:true"`
	Address         string    `gorm:"primaryKey"`
	Token0Address   string
	Token1Address   string
}

package db

import (
	"fmt"
	"log"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository.Repository = (*Repository)(nil)

type Repository struct {
	dbCon *gorm.DB
}

func New(host string, port string, user string, password string, dbname string) (*Repository, error) {
	ret := &Repository{}

	dbCon, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	ret.dbCon = dbCon

	// Create tables for data structures (if table already exists it will not be overwritten)
	dbCon.Table("eth_tokens_local").AutoMigrate(&Token{})
	dbCon.Table("eth_liq_pools_local").AutoMigrate(&Pool{})
	dbCon.Table("eth_liq_adds_local").AutoMigrate(&Addition{})
	dbCon.Table("eth_liq_removals_local").AutoMigrate(&Removal{})
	dbCon.Table("eth_swaps_local").AutoMigrate(&Swap{})
	return ret, nil
}

func (r *Repository) GetToken(address string) (repository.Token, bool) {
	var token Token
	result := r.dbCon.Table("eth_tokens_local").Limit(1).Find(&token, "address = ?", address)
	isTokenFound := result.RowsAffected != 0
	if result.Error != nil {
		log.Println("Error fetching Token from DB:", result.Error)
	}
	return repository.Token{
		Address:  token.Address,
		Symbol:   token.Symbol,
		Name:     token.Name,
		Decimals: token.Decimals,
	}, isTokenFound
}

func (r *Repository) GetPoolPairAddresses(liqPoolAddress string) (string, string, bool) {
	var liqPool Pool
	result := r.dbCon.Table("eth_liq_pools_local").Limit(1).Find(&liqPool, "address = ?", liqPoolAddress)
	isPoolFound := result.RowsAffected != 0
	if result.Error != nil {
		log.Println("Error fetching Liq. Pool from DB:", result.Error)
	}
	return liqPool.Token0Address, liqPool.Token1Address, isPoolFound
}

func (r *Repository) AddToken(token repository.Token) error {
	newToken := Token{
		Address:  token.Address,
		Symbol:   token.Symbol,
		Name:     token.Name,
		Decimals: token.Decimals,
	}
	result := r.dbCon.Table("eth_tokens_local").Create(&newToken)
	return result.Error
}

func (r *Repository) SavePool(pool repository.Pool) error {
	newPool := Pool{
		Address:       pool.Address,
		Token0Address: pool.Token0Address,
		Token1Address: pool.Token1Address,
	}
	result := r.dbCon.Clauses(clause.OnConflict{DoNothing: true}).Table("eth_liq_pools_local").Create(&newPool)
	return result.Error
}

func (r *Repository) SaveAddition(lpAdd repository.Addition) error {
	add := Addition{
		TimestampReceived: lpAdd.TimestampReceived,
		LPoolAddress:      lpAdd.LPoolAddress,
		Token0Symbol:      lpAdd.Token0Symbol,
		Token1Symbol:      lpAdd.Token1Symbol,
		Token0Amount:      lpAdd.Token0Amount,
		Token1Amount:      lpAdd.Token1Amount,
		LowerActualRatio:  lpAdd.LowerRatio,
		UpperActualRatio:  lpAdd.UpperRatio,
		Token0PriceUsd:    lpAdd.Token0PriceUsd,
		Token1PriceUsd:    lpAdd.Token1PriceUsd,
		TxHash:            lpAdd.TxHash,
	}
	result := r.dbCon.Table("eth_liq_adds_local").Create(&add)
	return result.Error
}

func (r *Repository) SaveRemoval(lpRem repository.Removal) error {
	remove := Removal{
		TimestampReceived: lpRem.TimestampReceived,
		LPoolAddress:      lpRem.LPoolAddress,
		Token0Symbol:      lpRem.Token0Symbol,
		Token1Symbol:      lpRem.Token1Symbol,
		Token0Amount:      lpRem.Token0Amount,
		Token1Amount:      lpRem.Token1Amount,
		LowerActualRatio:  lpRem.LowerRatio,
		UpperActualRatio:  lpRem.UpperRatio,
		Token0PriceUsd:    lpRem.Token0PriceUsd,
		Token1PriceUsd:    lpRem.Token1PriceUsd,
		TxHash:            lpRem.TxHash,
	}
	result := r.dbCon.Table("eth_liq_removals_local").Create(&remove)
	return result.Error
}

func (r *Repository) SaveSwap(sw repository.Swap) error {
	remove := Swap{
		TimestampReceived: sw.TimestampReceived,
		LPoolAddress:      sw.LPoolAddress,
		TokenFrom:         sw.TokenFrom,
		TokenFromAmount:   sw.TokenFromAmount,
		TokenTo:           sw.TokenTo,
		TokenToAmount:     sw.TokenToAmount,
		TxHash:            sw.TxHash,
	}
	result := r.dbCon.Table("eth_swaps_local").Create(&remove)
	return result.Error
}

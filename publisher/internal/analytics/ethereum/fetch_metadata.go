package ethereum

import (
	"log"
	"strings"
)

// includeTokenPrices enriches Liquidity Addition with token prices
func (a *Analytics) includeTokenPrices(pos Position) Position {
	// Place here to implement price cache?
	price, err := a.priceFetcher.Price(pos.Token0.Address)
	if err != nil {
		log.Println("failed to feetch Token0 price: ", err.Error())
	}
	pos.Token0.Price = price.Value
	if !strings.EqualFold(pos.Token1.Address, "") {
		price, err := a.priceFetcher.Price(pos.Token1.Address)
		if err != nil {
			log.Println("failed to feetch Token1 price: ", err.Error())
		}
		pos.Token1.Price = price.Value
	}
	return pos
}

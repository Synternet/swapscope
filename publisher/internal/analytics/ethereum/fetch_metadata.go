package ethereum

import (
	"log"
	"strings"
)

// includeTokenPrices enriches Liquidity Entry with token prices
func (a *Analytics) includeTokenPrices(liqAdd LiquidityAddition) LiquidityAddition {
	// Place here to implement price cache?
	price, err := a.priceFetcher.Price(liqAdd.Token0.Address)
	if err != nil {
		log.Println("failed to feetch Token0 price: ", err.Error())
	}
	liqAdd.Token0.Price = price.Value
	if !strings.EqualFold(liqAdd.Token1.Address, "") {
		price, err := a.priceFetcher.Price(liqAdd.Token1.Address)
		if err != nil {
			log.Println("failed to feetch Token1 price: ", err.Error())
		}
		liqAdd.Token1.Price = price.Value
	}
	return liqAdd
}

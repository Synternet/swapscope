package ethereum

import (
	"log"
	"strings"
)

// isUniswapPositionsNFT checks if owner of event is uniswap positions NFT
func isUniswapPositionsNFT(address string) bool {
	if strings.Contains(address, strings.ToLower(uniswapPositionsOwner)[2:]) {
		return true
	}
	return false
}

func (a *Analytics) isAvoidEvent(evLog EventLog) bool {
	topic0 := evLog.Topics[0]
	_, found := a.signaturesToAvoid[topic0[:10]]
	if found {
		log.Println("FOUND EVENT TO AVOID! (all transfers until now are invalid):", topic0[:10])
		return true
	}
	return false
}

func (a *Analytics) isTransferEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], a.signatureTransfer)
}

func (a *Analytics) isMintEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], a.signatureMint)
}

// isLiqAddRecordFiltered checks custom filters if finalized liq. add. record should be published
// or logged elsewhere if some kind of filter is matched
func (a *Analytics) canLiqAddRecordSatisfyFilters(liqAdd LiquidityAddition) bool {
	retRes := true
	if strings.EqualFold(liqAdd.Token0.Symbol, "") || strings.EqualFold(liqAdd.Token1.Symbol, "") {
		retRes = false
		log.Println("Liq. add. record filter - token symbol is unknown. Tx hash:", liqAdd.TxHash)
	}
	if liqAdd.LowerRatio == 0 && liqAdd.UpperRatio == 0 {
		retRes = false
		log.Printf("Liq. add. record filter - could not calculate actual ratio. Tx hash: %s\n\n", liqAdd.TxHash)
	}

	return retRes
}

func hasTopics(evLog EventLog) bool {
	if len(evLog.Topics) == 0 {
		log.Println("Log message of TX", evLog.TransactionHash, "has no topics.")
		return false
	}
	return true
}

func isOrderCorrect(liqAdd LiquidityAddition) bool {
	token0 := liqAdd.Token0.Symbol
	token1 := liqAdd.Token1.Symbol
	return (token1 == "WETH" || token0 == "USDT" || token0 == "USDC")
}

func isStableOrNativeInvolved(liqAdd LiquidityAddition) bool {
	token0 := liqAdd.Token0.Symbol
	token1 := liqAdd.Token1.Symbol
	return (token1 == "WETH" || token0 == "USDT" || token0 == "USDC" || token0 == "WETH" || token1 == "USDT" || token1 == "USDC")
}

func isEitherTokenUnknown(liqAdd LiquidityAddition) bool {
	return (strings.EqualFold(liqAdd.Token0.Address, "") || strings.EqualFold(liqAdd.Token1.Address, ""))
}

func isEitherTokenAmountIsZero(liqAdd LiquidityAddition) bool {
	return (liqAdd.Token0.Amount == 0 || liqAdd.Token1.Amount == 0)
}

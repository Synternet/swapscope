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

func (a *Analytics) isBurnEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], a.signatureBurn)
}

func hasTopics(evLog EventLog) bool {
	if len(evLog.Topics) == 0 {
		log.Println("Log message of TX", evLog.TransactionHash, "has no topics.")
		return false
	}
	return true
}

func isOrderCorrect(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (token1.Symbol == "WETH" || token0.Symbol == "USDT" || token0.Symbol == "USDC")
}

func isStableOrNativeInvolved(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (token1.Symbol == "WETH" || token0.Symbol == "USDT" || token0.Symbol == "USDC" || token0.Symbol == "WETH" || token1.Symbol == "USDT" || token1.Symbol == "USDC")
}

func isEitherTokenUnknown(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (strings.EqualFold(token0.Address, "") || strings.EqualFold(token1.Address, ""))
}

func isEitherTokenAmountIsZero(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (token0.Amount == 0 || token1.Amount == 0)
}

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

func isTransferEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], transferSig)
}

func (a *Analytics) isMintEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], mintSig)
}

func (a *Analytics) isBurnEvent(evLog EventLog) bool {
	return strings.HasPrefix(evLog.Topics[0], burnSig)
}

func hasTopics(evLog EventLog) bool {
	for _, str := range evLog.Topics {
		if str != "" {
			return true
		}
	}
	log.Println("Log message of TX", evLog.TransactionHash, "has no topics.")
	return false
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

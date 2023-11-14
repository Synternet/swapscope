package ethereum

import (
	"strings"
)

// isUniswapPositionsNFT checks if owner of event is uniswap positions NFT
func isUniswapPositionsNFT(address string) bool {
	if strings.Contains(address, strings.ToLower(uniswapPositionsOwner)[2:]) {
		return true
	}
	return false
}

func isOrderCorrect(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return strings.EqualFold(token1.Address, addressWETH) || strings.EqualFold(token0.Address, addressUSDC) || strings.EqualFold(token0.Address, addressUSDT)
}

func isStableOrNativeInvolved(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	for _, address := range []string{addressWETH, addressUSDC, addressUSDT} {
		if strings.EqualFold(token1.Address, address) || strings.EqualFold(token0.Address, address) {
			return true
		}
	}
	return false
}

func isEitherTokenUnknown(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (strings.EqualFold(token0.Address, "") || strings.EqualFold(token1.Address, ""))
}

func isEitherTokenAmountZero(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	return (token0.Amount == 0 || token1.Amount == 0)
}

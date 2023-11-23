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

func isStableInvolved(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	for _, address := range stableCoins {
		if strings.EqualFold(token1.Address, address) || strings.EqualFold(token0.Address, address) {
			return true
		}
	}
	return false
}

func isNativeInvolved(position Position) bool {
	token0 := position.Token0
	token1 := position.Token1
	for _, address := range nativeCoins {
		if strings.EqualFold(token1.Address, address) || strings.EqualFold(token0.Address, address) {
			return true
		}
	}
	return false
}

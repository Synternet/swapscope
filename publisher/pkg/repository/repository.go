package repository

type Repository interface {
	// GetToken returns the latest token record for the given token address
	GetToken(address string) (Token, bool)
	// GetTokenPairAddresses returns the addresses of the tokens that are used in the liquidity pool
	GetTokenPairAddresses(lpAddress string) (string, string, bool)

	AddToken(newToken Token) error
	AddLiquidityPool(newPool LiquidityPool) error
	AddLiquidityPoolAddition(newLiqAdd LiquidityEntry) error
}

package repository

type Repository interface {
	// GetToken returns the latest token record for the given token address
	GetToken(address string) (Token, bool)
	// GetPoolPairAddresses returns the addresses of the tokens that are used in the liquidity pool
	GetPoolPairAddresses(lpAddress string) (string, string, bool)

	AddToken(newToken Token) error
	SavePool(pool Pool) error
	SaveAddition(add Addition) error
	SaveRemoval(rem Removal) error
}

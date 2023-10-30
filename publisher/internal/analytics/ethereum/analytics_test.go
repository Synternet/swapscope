package ethereum

import (
	"math"
	"math/big"
	"testing"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
)

var knownTokens = map[string]TokenTransaction{
	"WETH":  {repository.Token{"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "WETH", "Wrapped Ether", 18, 0.0}, 0, 0},
	"WBTC":  {repository.Token{"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599", "WBTC", "Wrapped BTC", 8, 0.0}, 0, 0},   // WBTC / WETH
	"MATIC": {repository.Token{"0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0", "MATIC", "Matic Token", 18, 0.0}, 0, 0}, // MATIC / WETH
	"USDC":  {repository.Token{"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "USDC", "USDC", 6, 0.0}, 0, 0},          // USDC / WETH
	"USDT":  {repository.Token{"0xdac17f958d2ee523a2206206994597c13d831ec7", "USDT", "USDT", 6, 0.0}, 0, 0},          // WETH / USDT
}

const tolerance = 1e-10

func Test_tickConversion(t *testing.T) {
	tests := []struct {
		input          Position
		trueLowerRatio float64
		trueUpperRatio float64
		trueToReverse  bool
	}{
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 259720, UpperTick: 259750}, 0.052452336044, 0.052609921433, false},          // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/591227
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 258310, UpperTick: 259930}, 0.051516686880, 0.060575933233, false},          //
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -79980, UpperTick: -79500}, 2834.4481085213397, 2973.812642817363, false},  // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587211
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -90240, UpperTick: -78420}, 2544.292581085861, 8296.166586725507, false},   // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587556
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 186220, UpperTick: 201460}, 1782.956728761764, 8184.129686609503, false},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593545
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201450, UpperTick: 203190}, 1499.724941901797, 1784.7404880350452, false},   // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593547
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201090, UpperTick: 201650}, 1749.4020078533288, 1850.158331324582, false},   // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593603
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -201510, UpperTick: -201300}, 1774.064638465059, 1811.7120276590485, true},  // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593575
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -204660, UpperTick: -197760}, 1294.7130255963305, 2581.2004232087493, true}, // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593557
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -201480, UpperTick: -201470}, 1779.3945567691965, 1781.1747522670805, true}, // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593392
	}

	for _, test := range tests {
		resLowerRatio, resUpperRatio, resToReverse := convertTicksToRatios(test.input)
		if math.Abs(resLowerRatio-test.trueLowerRatio)*1.0 > tolerance ||
			math.Abs(resUpperRatio-test.trueUpperRatio)*1.0 > tolerance ||
			resToReverse != test.trueToReverse {
			t.Errorf("convertTicksToRatios(%v) = (%v,%v,%v); expected (%v,%v,%v)", test.input,
				resLowerRatio, resUpperRatio, resToReverse,
				test.trueLowerRatio, test.trueUpperRatio, test.trueToReverse)
		}
	}
}

func Test_eventSignatureConversion(t *testing.T) {
	tests := []struct {
		inputHeader  string
		trueEventSig string
	}{
		{"Mint(address,address,int24,int24,uint128,uint256,uint256)", "0x7a53080b"},
		{"Transfer(address,address,uint256)", "0xddf252ad"},
		{"Burn(address,int24,int24,uint128,uint256,uint256)", "0x0c396cd9"},
	}
	for _, test := range tests {
		resEventSignature := convertToEventSignature(test.inputHeader)
		if resEventSignature != test.trueEventSig {
			t.Errorf("convertToEventSignature(%v) = (%v); expected (%v)", test.inputHeader, resEventSignature, test.trueEventSig)
		}
	}
}

func Test_hexToBigIntConversion(t *testing.T) {
	tests := []struct {
		inputHex         string
		trueBigIntString string
	}{
		{"0x0000000000000000000000000000000000000003ff0aefc357bb2bcd5150a760", "316616386554458346478543873888"},
		{"0x0000000000000000000000000000000000000000000000111942bdafcf1baace", "315414875015265102542"},
		{"0x000000000000000000000000000000000000000000000000015e6fc4aea528f1", "98639132383062257"},
		{"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffceabe", "-202050"},
		{"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0c40", "-62400"},
		{"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2764c", "-887220"},
		{"0x00000000000000000000000000000000000000000000000000000000000d89b4", "887220"},
	}
	for _, test := range tests {
		resBigInt := convertHexToBigInt(test.inputHex)
		trueBigInt, ok := new(big.Int).SetString(test.trueBigIntString, 10)
		result := resBigInt.Cmp(trueBigInt)
		t.Logf("convertHexToBigInt(%v) = (%v); expected (%v)", test.inputHex, resBigInt, trueBigInt)
		if result != 0 || ok == false {
			t.Errorf("convertHexToBigInt(%v) = (%v); expected (%v)", test.inputHex, resBigInt, trueBigInt)
		}
	}
}

package ethereum

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/Synternet/swapscope/publisher/pkg/repository"
)

var knownTokens = map[string]TokenTransaction{
	"WETH":  {Token: repository.Token{Address: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", Symbol: "WETH", Decimals: 18}},
	"WBTC":  {Token: repository.Token{Address: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", Symbol: "WBTC", Decimals: 8}},   // WBTC / WETH
	"MATIC": {Token: repository.Token{Address: "0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0", Symbol: "MATIC", Decimals: 18}}, // MATIC / WETH
	"USDC":  {Token: repository.Token{Address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", Symbol: "USDC", Decimals: 6}},   // USDC / WETH
	"USDT":  {Token: repository.Token{Address: "0xdAC17F958D2ee523a2206206994597C13D831ec7", Symbol: "USDT", Decimals: 6}},   // WETH / USDT
	"PEPE":  {Token: repository.Token{Address: "0x6982508145454Ce325dDbE47a25d4ec3d2311933", Symbol: "PEPE", Decimals: 18}},  // WETH / PEPE
	"FINE":  {Token: repository.Token{Address: "0x4e6415a5727ea08aae4580057187923aec331227", Symbol: "FINE", Decimals: 18}},  // WETH / FINE
}

const tolerance = 1e-10

func Test_tickConversion(t *testing.T) {
	tests := []struct {
		name           string
		input          *Position
		trueLowerRatio float64
		trueUpperRatio float64
	}{
		{"Custom/native: Ticks > 0, res ratio < 1, low range", &Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 259720, UpperTick: 259750}, 19.007821581, 19.064927807},   // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/591227
		{"Custom/native: Ticks > 0, res ratio < 1, big range", &Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 258310, UpperTick: 259930}, 16.5082062564, 19.4111861717}, //
		{"Custom/native: Ticks < 0, low range", &Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -79980, UpperTick: -79500}, 0.0003362686625, 0.00035280236635},          // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587211
		{"Custom/native: Ticks < 0, big range", &Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -90240, UpperTick: -78420}, 0.00012053759884, 0.000393036558},           // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587556
		{"Stable/native: Ticks > 0, big range", &Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 186220, UpperTick: 201460}, 1782.956728761764, 8184.129686609503},        // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593545
		{"Stable/native: Ticks > 0, big range", &Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201450, UpperTick: 203190}, 1499.724941901797, 1784.7404880350452},       // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593547
		{"Stable/native: Ticks > 0, low range", &Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201090, UpperTick: 201650}, 1749.4020078533288, 1850.158331324582},       // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593603
		{"Native/stable: Ticks < 0, big range", &Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -204660, UpperTick: -197760}, 1294.7130255963305, 2581.2004232087493},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593557
		{"Native/stable: Ticks < 0, low range", &Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -201480, UpperTick: -201470}, 1779.3945567691965, 1781.1747522670805},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593392
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.input.calculate()
			resLowerRatio := test.input.LowerRatio
			resUpperRatio := test.input.UpperRatio
			if math.Abs(resLowerRatio-test.trueLowerRatio)*1.0 > tolerance ||
				math.Abs(resUpperRatio-test.trueUpperRatio)*1.0 > tolerance {
				t.Errorf("convertTicksToRatios(%v) = (%v,%v); expected (%v,%v)",
					test.input, resLowerRatio, resUpperRatio, test.trueLowerRatio, test.trueUpperRatio)
			}
		})
	}
}

func Test_eventSignatureConversion(t *testing.T) {
	tests := []struct {
		name         string
		inputHeader  string
		trueEventSig string
	}{
		{"convert Mint header to sig", "Mint(address,address,int24,int24,uint128,uint256,uint256)", "0x7a53080b"},
		{"convert Transfer header to sig", "Transfer(address,address,uint256)", "0xddf252ad"},
		{"convert Burn header to sig", "Burn(address,int24,int24,uint128,uint256,uint256)", "0x0c396cd9"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resEventSignature := convertToEventSignature(test.inputHeader)
			if resEventSignature != test.trueEventSig {
				t.Errorf("convertToEventSignature(%v) = (%v); expected (%v)", test.inputHeader, resEventSignature, test.trueEventSig)
			}
		})
	}
}

func Test_hexToBigIntConversion(t *testing.T) {
	tests := []struct {
		name             string
		inputHex         string
		trueBigIntString string
	}{
		{"30 digits > 0 number", "0x0000000000000000000000000000000000000003ff0aefc357bb2bcd5150a760", "316616386554458346478543873888"},
		{"17 digits > 0 number", "0x000000000000000000000000000000000000000000000000015e6fc4aea528f1", "98639132383062257"},
		{"5 digits < 0 number (0xfff...)", "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0c40", "-62400"},
		{"6 digits < 0 number (0xfff...)", "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2764c", "-887220"},
		{"same 6 digits > 0 number", "0x00000000000000000000000000000000000000000000000000000000000d89b4", "887220"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resBigInt := convertHexToBigInt(test.inputHex)
			trueBigInt, ok := new(big.Int).SetString(test.trueBigIntString, 10)
			result := resBigInt.Cmp(trueBigInt)
			if result != 0 || ok == false {
				t.Errorf("convertHexToBigInt(%v) = (%v); expected (%v)", test.inputHex, resBigInt, trueBigInt)
			}
		})
	}
}

func Test_hexToTokenAmountConversion(t *testing.T) {
	setTestName := func(dec int) string { return fmt.Sprintf("%d decimals", dec) }

	tests := []struct {
		name            string
		inputHex        string
		token           repository.Token
		trueTokenAmount float64
	}{
		{setTestName(knownTokens["WETH"].Decimals), "0x000000000000000000000000000000000000000000000000738231e99b3c1256", knownTokens["WETH"].Token, 8.323269940735644},
		{setTestName(knownTokens["WETH"].Decimals), "0x0000000000000000000000000000000000000000000000000f0a72da3457cf14", knownTokens["WETH"].Token, 1.0838049418426325},
		{setTestName(knownTokens["USDC"].Decimals), "0x0000000000000000000000000000000000000000000000000000000074f62ca3", knownTokens["USDC"].Token, 1962.290339},
		{setTestName(knownTokens["USDC"].Decimals), "0x00000000000000000000000000000000000000000000000000000f31422dbd0a", knownTokens["USDC"].Token, 16704238.107914},
		{setTestName(knownTokens["WBTC"].Decimals), "0x0000000000000000000000000000000000000000000000000000000000d3492a", knownTokens["WBTC"].Token, 0.13846826},
		{setTestName(knownTokens["WBTC"].Decimals), "0x000000000000000000000000000000000000000000000000000000046fba09be", knownTokens["WBTC"].Token, 190.5433235},
		{setTestName(knownTokens["PEPE"].Decimals), "0x0000000000000000000000000000000000000000003e8f827edf295722270278", knownTokens["PEPE"].Token, 75631106.441958173},
		{setTestName(knownTokens["FINE"].Decimals), "0x0000000000000000000000000000000000000017dd19ad760cd7ddd3ea2fc824", knownTokens["FINE"].Token, 1890674967291.1301660},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resTokenAmount := convertTransferAmount(test.inputHex, test.token.Decimals)
			if math.Abs(resTokenAmount-test.trueTokenAmount)*1.0 > tolerance {
				t.Errorf("convertTransferAmount(%v) = (%v); expected (%v)", test.inputHex, resTokenAmount, test.trueTokenAmount)
			}
		})
	}
}

func Test_isUniswapPositionsNFT(t *testing.T) {
	tests := []struct {
		name         string
		inputAddress string
		trueRes      bool
	}{
		{"Uniswap address", "0xc36442b4a4522e871399cd717abdd847ab11fe88", true},
		{"Empty string", "", false},
		{"Empty hex", "0x", false},
		{"Not full hex", "C36442b4a4522E871399CD717aBDD847Ab11FE88", false},
		{"Different address", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isUniswapPositionsNFT(test.inputAddress)
			if res != test.trueRes {
				t.Errorf("isUniswapPositionsNFT(%v) = (%v); expected (%v)", test.inputAddress, res, test.trueRes)
			}
		})
	}
}

func Test_hasTopics(t *testing.T) {
	tests := []struct {
		name          string
		inputEventLog EventLog
		trueRes       bool
	}{
		{"1 topic", EventLog{Topics: []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}}, true},
		{"3 topics", EventLog{Topics: []string{"0x0c396c", "0x0c396cd989a39f4", "0x0c396cd989a30c396c9f4"}}, true},
		{"1 empty topic", EventLog{Topics: []string{""}}, false},
		{"No topics", EventLog{Topics: []string{}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.inputEventLog.hasTopics()
			if res != test.trueRes {
				t.Errorf("hasTopics(%v) = (%v); expected (%v)", test.inputEventLog, res, test.trueRes)
			}
		})
	}
}

func Test_isToken1Native(t *testing.T) {
	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"native/custom", Position{Token0: knownTokens["WETH"], Token1: knownTokens["WBTC"]}, false},
		{"custom/native", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, true},
		{"custom/native", Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"]}, true},
		{"stable/native", Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{"native/stable", Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDC"]}, false},
		{"native/stable", Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"]}, false},
		{"stable/stable", Position{Token0: knownTokens["USDC"], Token1: knownTokens["USDT"]}, false},
		{"stable/stable", Position{Token0: knownTokens["USDT"], Token1: knownTokens["USDC"]}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.isToken1OneOf(nativeCoins)
			if res != test.trueRes {
				t.Errorf("isToken1Native(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

func Test_isToken1Stable(t *testing.T) {
	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"native/custom", Position{Token0: knownTokens["WETH"], Token1: knownTokens["WBTC"]}, false},
		{"custom/native", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, false},
		{"custom/native", Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"]}, false},
		{"stable/native", Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, false},
		{"native/stable", Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDC"]}, true},
		{"native/stable", Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"]}, true},
		{"stable/stable", Position{Token0: knownTokens["USDC"], Token1: knownTokens["USDT"]}, true},
		{"stable/stable", Position{Token0: knownTokens["USDT"], Token1: knownTokens["USDC"]}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.isToken1OneOf(stableCoins)
			if res != test.trueRes {
				t.Errorf("isToken1Stable(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

func Test_isNativeInvolved(t *testing.T) {
	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"custom/native", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, true},
		{"stable/native", Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{"custom/custom", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["PEPE"]}, false},
		{"stable/custom", Position{Token0: knownTokens["USDC"], Token1: knownTokens["PEPE"]}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.isAnyTokenOneOf(nativeCoins)
			if res != test.trueRes {
				t.Errorf("isNativeInvolved(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

func Test_isStableInvolved(t *testing.T) {
	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"custom/native", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, false},
		{"stable/native", Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{"custom/custom", Position{Token0: knownTokens["WBTC"], Token1: knownTokens["PEPE"]}, false},
		{"stable/custom", Position{Token0: knownTokens["USDC"], Token1: knownTokens["PEPE"]}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.isAnyTokenOneOf(stableCoins)
			if res != test.trueRes {
				t.Errorf("isStableInvolved(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

func Test_areTokensSet(t *testing.T) {
	unknownToken := TokenTransaction{Token: repository.Token{}}
	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"unknown/native", Position{Token0: unknownToken, Token1: knownTokens["WETH"]}, false},
		{"custom/unknown", Position{Token0: knownTokens["MATIC"], Token1: unknownToken}, false},
		{"stable/unknown", Position{Token0: knownTokens["USDC"], Token1: unknownToken}, false},
		{"stable/native", Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{"native/unknown", Position{Token0: knownTokens["WETH"], Token1: unknownToken}, false},
		{"custom/custom", Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WBTC"]}, true},
		{"unknown/custom", Position{Token0: unknownToken, Token1: knownTokens["PEPE"]}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.areTokensSet()
			if res != test.trueRes {
				t.Errorf("isEitherTokenUnknown(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

func Test_isEitherTokenAmountIsZero(t *testing.T) {
	setAmount := func(tt TokenTransaction, amt float64) TokenTransaction {
		tt.Amount = amt
		return tt
	}

	tests := []struct {
		name    string
		input   Position
		trueRes bool
	}{
		{"0 USDC", Position{Token0: setAmount(knownTokens["USDC"], 0), Token1: setAmount(knownTokens["WETH"], 0.0001)}, true},
		{"Amount not set", Position{Token0: setAmount(knownTokens["MATIC"], 5.35), Token1: knownTokens["WBTC"]}, true},
		{" 0 < amount < 1", Position{Token0: setAmount(knownTokens["WBTC"], 0.00001), Token1: setAmount(knownTokens["PEPE"], 99999999)}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.input.isEitherTokenAmountZero()
			if res != test.trueRes {
				t.Errorf("isEitherTokenAmountIsZero(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
			}
		})
	}
}

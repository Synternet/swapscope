package ethereum

import (
	"math"
	"math/big"
	"testing"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
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
		input          Position
		trueLowerRatio float64
		trueUpperRatio float64
	}{
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 259720, UpperTick: 259750}, 0.052452336044, 0.052609921433},           // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/591227
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"], LowerTick: 258310, UpperTick: 259930}, 0.051516686880, 0.060575933233},           //
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -79980, UpperTick: -79500}, 2834.4481085213397, 2973.812642817363},   // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587211
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"], LowerTick: -90240, UpperTick: -78420}, 2544.292581085861, 8296.166586725507},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/587556
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 186220, UpperTick: 201460}, 1782.956728761764, 8184.129686609503},     // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593545
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201450, UpperTick: 203190}, 1499.724941901797, 1784.7404880350452},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593547
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"], LowerTick: 201090, UpperTick: 201650}, 1749.4020078533288, 1850.158331324582},    // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593603
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -201510, UpperTick: -201300}, 1774.064638465059, 1811.7120276590485},  // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593575
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -204660, UpperTick: -197760}, 1294.7130255963305, 2581.2004232087493}, // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593557
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"], LowerTick: -201480, UpperTick: -201470}, 1779.3945567691965, 1781.1747522670805}, // https://etherscan.io/nft/0xc36442b4a4522e871399cd717abdd847ab11fe88/593392
	}

	for _, test := range tests {
		resLowerRatio, resUpperRatio := calculateInterval(test.input)
		if math.Abs(resLowerRatio-test.trueLowerRatio)*1.0 > tolerance ||
			math.Abs(resUpperRatio-test.trueUpperRatio)*1.0 > tolerance {
			t.Errorf("convertTicksToRatios(%v) = (%v,%v); expected (%v,%v)",
				test.input, resLowerRatio, resUpperRatio, test.trueLowerRatio, test.trueUpperRatio)
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
		if result != 0 || ok == false {
			t.Errorf("convertHexToBigInt(%v) = (%v); expected (%v)", test.inputHex, resBigInt, trueBigInt)
		}
	}
}

func Test_hexToTokenAmountConversion(t *testing.T) {
	tests := []struct {
		inputHex        string
		token           repository.Token
		trueTokenAmount float64
	}{
		{"0x000000000000000000000000000000000000000000000000738231e99b3c1256", knownTokens["WETH"].Token, 8.323269940735644},
		{"0x0000000000000000000000000000000000000000000000000f0a72da3457cf14", knownTokens["WETH"].Token, 1.0838049418426325},
		{"0x0000000000000000000000000000000000000000000000000000000074f62ca3", knownTokens["USDC"].Token, 1962.290339},
		{"0x00000000000000000000000000000000000000000000000000000f31422dbd0a", knownTokens["USDC"].Token, 16704238.107914},
		{"0x0000000000000000000000000000000000000000000000000000000000d3492a", knownTokens["WBTC"].Token, 0.13846826},
		{"0x000000000000000000000000000000000000000000000000000000046fba09be", knownTokens["WBTC"].Token, 190.5433235},
		{"0x00000000000000000000000000000000000000002c1ac359d7b172c9fb4dc763", knownTokens["PEPE"].Token, 13649695022.21587552},
		{"0x0000000000000000000000000000000000000000003e8f827edf295722270278", knownTokens["PEPE"].Token, 75631106.441958173},
		{"0x0000000000000000000000000000000000000017dd19ad760cd7ddd3ea2fc824", knownTokens["FINE"].Token, 1890674967291.1301660},
	}
	for _, test := range tests {
		resTokenAmount := convertTransferAmount(test.inputHex, test.token.Decimals)
		if math.Abs(resTokenAmount-test.trueTokenAmount)*1.0 > tolerance {
			t.Errorf("convertTransferAmount(%v) = (%v); expected (%v)", test.inputHex, resTokenAmount, test.trueTokenAmount)
		}
	}
}

func Test_isUniswapPositionsNFT(t *testing.T) {
	tests := []struct {
		inputAddress string
		trueRes      bool
	}{
		{"0xc36442b4a4522e871399cd717abdd847ab11fe88", true},
		{"", false},
		{"0x", false},
		{"C36442b4a4522E871399CD717aBDD847Ab11FE81", false},
		{"2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", false},
	}
	for _, test := range tests {
		res := isUniswapPositionsNFT(test.inputAddress)
		if res != test.trueRes {
			t.Errorf("isUniswapPositionsNFT(%v) = (%v); expected (%v)", test.inputAddress, res, test.trueRes)
		}
	}
}

func Test_isTransferEvent(t *testing.T) {
	tests := []struct {
		inputEventLog EventLog
		trueRes       bool
	}{
		{EventLog{Topics: []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}}, true},
		{EventLog{Topics: []string{"0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c"}}, false},
		{EventLog{Topics: []string{"0xddf252ad1be2c89"}}, true},
		{EventLog{Topics: []string{"0xddf252"}}, false},
		{EventLog{Topics: []string{""}}, false},
	}
	for _, test := range tests {
		res := isTransferEvent(test.inputEventLog)
		if res != test.trueRes {
			t.Errorf("isTransferEvent(%v) = (%v); expected (%v)", test.inputEventLog, res, test.trueRes)
		}
	}
}

func Test_isMintEvent(t *testing.T) {
	tests := []struct {
		inputEventLog EventLog
		trueRes       bool
	}{
		{EventLog{Topics: []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}}, false},
		{EventLog{Topics: []string{"0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde"}}, true},
		{EventLog{Topics: []string{"0x7a53080ba414158"}}, true},
		{EventLog{Topics: []string{"0x7a53"}}, false},
		{EventLog{Topics: []string{""}}, false},
	}
	a := Analytics{}
	for _, test := range tests {
		res := a.isMintEvent(test.inputEventLog)
		if res != test.trueRes {
			t.Errorf("isMintEvent(%v) = (%v); expected (%v)", test.inputEventLog, res, test.trueRes)
		}
	}
}

func Test_isBurnEvent(t *testing.T) {
	tests := []struct {
		inputEventLog EventLog
		trueRes       bool
	}{
		{EventLog{Topics: []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}}, false},
		{EventLog{Topics: []string{"0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c"}}, true},
		{EventLog{Topics: []string{"0x0c396cd989a39f4"}}, true},
		{EventLog{Topics: []string{"0x0c396c"}}, false},
		{EventLog{Topics: []string{""}}, false},
	}
	a := Analytics{}
	for _, test := range tests {
		res := a.isBurnEvent(test.inputEventLog)
		if res != test.trueRes {
			t.Errorf("isBurnEvent(%v) = (%v); expected (%v)", test.inputEventLog, res, test.trueRes)
		}
	}
}

func Test_hasTopics(t *testing.T) {
	tests := []struct {
		inputEventLog EventLog
		trueRes       bool
	}{
		{EventLog{Topics: []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}}, true},
		{EventLog{Topics: []string{"0x0c396c", "0x0c396cd989a39f4", "0x0c396cd989a30c396c9f4"}}, true},
		{EventLog{Topics: []string{"0"}}, true},
		{EventLog{Topics: []string{""}}, false},
		{EventLog{Topics: []string{}}, false},
	}
	for _, test := range tests {
		res := hasTopics(test.inputEventLog)
		if res != test.trueRes {
			t.Errorf("hasTopics(%v) = (%v); expected (%v)", test.inputEventLog, res, test.trueRes)
		}
	}
}

func Test_isOrderCorrect(t *testing.T) {
	tests := []struct {
		input   Position
		trueRes bool
	}{
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["WBTC"]}, false},
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDC"]}, false},
		{Position{Token0: knownTokens["WETH"], Token1: knownTokens["USDT"]}, false},
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["USDT"]}, true},
		{Position{Token0: knownTokens["USDT"], Token1: knownTokens["USDC"]}, true},
	}
	for _, test := range tests {
		res := isOrderCorrect(test.input)
		if res != test.trueRes {
			t.Errorf("isOrderCorrect(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
		}
	}
}

func Test_isStableOrNativeInvolved(t *testing.T) {
	tests := []struct {
		input   Position
		trueRes bool
	}{
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["USDT"], Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["RLB"]}, false},
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["PEPE"]}, false},
	}
	for _, test := range tests {
		res := isStableOrNativeInvolved(test.input)
		if res != test.trueRes {
			t.Errorf("isStableOrNativeInvolved(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
		}
	}
}

func Test_isEitherTokenUnknown(t *testing.T) {
	unknownToken := TokenTransaction{Token: repository.Token{}}
	tests := []struct {
		input   Position
		trueRes bool
	}{
		{Position{Token0: unknownToken, Token1: knownTokens["WETH"]}, true},
		{Position{Token0: knownTokens["MATIC"], Token1: unknownToken}, true},
		{Position{Token0: knownTokens["USDC"], Token1: unknownToken}, true},
		{Position{Token0: knownTokens["USDC"], Token1: knownTokens["WETH"]}, false},
		{Position{Token0: knownTokens["WETH"], Token1: unknownToken}, true},
		{Position{Token0: knownTokens["MATIC"], Token1: knownTokens["WBTC"]}, false},
		{Position{Token0: knownTokens["WBTC"], Token1: knownTokens["PEPE"]}, false},
	}
	for _, test := range tests {
		res := isEitherTokenUnknown(test.input)
		if res != test.trueRes {
			t.Errorf("isEitherTokenUnknown(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
		}
	}
}

func Test_isEitherTokenAmountIsZero(t *testing.T) {
	setAmount := func(tt TokenTransaction, amt float64) TokenTransaction {
		tt.Amount = amt
		return tt
	}

	tests := []struct {
		input   Position
		trueRes bool
	}{
		{Position{Token0: setAmount(knownTokens["USDC"], 0), Token1: setAmount(knownTokens["WETH"], 0.0001)}, true},
		{Position{Token0: setAmount(knownTokens["MATIC"], 5.35), Token1: knownTokens["WBTC"]}, true},
		{Position{Token0: setAmount(knownTokens["WBTC"], 0.00001), Token1: setAmount(knownTokens["PEPE"], 99999999)}, false},
	}
	for _, test := range tests {
		res := isEitherTokenAmountIsZero(test.input)
		if res != test.trueRes {
			t.Errorf("isEitherTokenAmountIsZero(%v) = (%v); expected (%v)", test.input, res, test.trueRes)
		}
	}
}

package ethereum

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

// convertHexToBigInt converts hex string into big int.
func convertHexToBigInt(hexStr string) *big.Int {
	resultInt := new(big.Int)
	hexClean := hexStr[2:]
	if strings.HasPrefix(hexClean, "fff") {
		bit_length := len(hexClean) * 4 // Calculate the number of bits in the hexadecimal value
		resultInt.SetString(hexClean, 16)
		resultInt = big.NewInt(0).Sub(resultInt, new(big.Int).Lsh(big.NewInt(1), uint(bit_length)))
	} else {
		resultInt.SetString(hexClean, 16)
	}
	return resultInt
}

// convertToEventSignature converts event header into an event signature.
func convertToEventSignature(header string) string {
	input := []byte(header)
	hash := make([]byte, 32) // Keccak-256 produces a 32-byte hash
	keccakHash := sha3.NewLegacyKeccak256()
	// Write the input data to the hash object
	_, err := keccakHash.Write(input)
	if err != nil {
		log.Println("--- ERROR: PrintData() keccakHash:", err)
	}
	// Calculate the hash and store it in the 'hash' slice
	keccakHash.Sum(hash[:0])
	signature := fmt.Sprintf("0x%x", hash)[:10]
	log.Println(header[:strings.IndexByte(header, '(')], "signature: ", signature)

	return signature
}

// convertToListOfEventSignatures converts a list of event headers into a list of event signatures.
// Used only to generate a list of events to "avoid" - to consider "Transfer" events accordingly.
// All "Transfer" events before any of "avoid" events are no more considered for "Mint" event.
func convertToListOfEventSignatures(eventHeaders []string) map[string]struct{} {
	hexValues := make(map[string]struct{})
	for _, eventHeaderString := range eventHeaders {
		hexValue := convertToEventSignature(eventHeaderString)
		hexValues[hexValue] = struct{}{}
	}
	return hexValues
}

// convertTicksToRatios converts ticks received from "Mint" event to understandable token ratios.
// More info: http://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf
func convertTicksToRatios(liqAdd LiquidityAddition) (float64, float64, bool) {
	reverseOrder := false
	tokDec0, tokDec1 := liqAdd.Token0.Decimals, liqAdd.Token1.Decimals
	lowerTickRatio, upperTickRatio := 0.0, 0.0

	lowerTickRatio = math.Pow(1.0001, float64(liqAdd.LowerTick)) / math.Pow(10, float64(tokDec1-tokDec0))
	upperTickRatio = math.Pow(1.0001, float64(liqAdd.UpperTick)) / math.Pow(10, float64(tokDec1-tokDec0))

	if isStableOrNativeInvolved(liqAdd) && isOrderCorrect(liqAdd) {
		lowerTickRatio = 1 / lowerTickRatio
		upperTickRatio = 1 / upperTickRatio
	} else if isStableOrNativeInvolved(liqAdd) {
		reverseOrder = true
	}
	return lowerTickRatio, upperTickRatio, reverseOrder
}

func parseEventLogMessage(data []byte) EventLog {
	var eLog EventLog
	err := json.Unmarshal(data, &eLog)
	if err != nil {
		log.Println("--- ERROR: Encountered an error when unmarshaling EventLog:", err)
	}
	return eLog
}

// convertTransferAmount converts Transfer's hex amount into scaled actual amount of tokens
func (a *Analytics) convertTransferAmount(amountHex string, decimals int) float64 {
	amount := convertHexToBigInt(amountHex)
	amountFloat := new(big.Float).SetInt(amount)
	scaleDecFactor := new(big.Float).SetFloat64(math.Pow10(decimals))
	amountScaled, _ := new(big.Float).Quo(amountFloat, scaleDecFactor).Float64() // amount / 10^decimals
	return amountScaled
}

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

// convertToEventSignatures converts a slice ofstrings (event headers) into a map of event signatures.
// Used only to generate a list of events to "avoid" - to consider "Transfer" events accordingly.
// All "Transfer" events before any of "avoid" events are no more considered for "Mint" event.
func convertToEventSignatures(eventHeaders []string) map[string]struct{} {
	hexValues := make(map[string]struct{})
	for _, eventHeaderString := range eventHeaders {
		hexValue := convertToEventSignature(eventHeaderString)
		hexValues[hexValue] = struct{}{}
	}
	return hexValues
}

// convertTicksToRatios converts ticks received from "Mint" event to understandable token ratios.
// More info: http://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf
func convertTicksToRatios(position Position) (float64, float64, bool) {
	reverseOrder := false
	token0, token1 := position.Token0, position.Token1
	tokDec0, tokDec1 := token0.Decimals, token1.Decimals
	lowerTick, upperTick := position.LowerTick, position.UpperTick

	lowerTickRatio := math.Pow(1.0001, float64(lowerTick)) / math.Pow(10, float64(tokDec1-tokDec0))
	upperTickRatio := math.Pow(1.0001, float64(upperTick)) / math.Pow(10, float64(tokDec1-tokDec0))

	if isStableOrNativeInvolved(position) && isOrderCorrect(position) {
		lowerTickRatio = 1 / lowerTickRatio
		upperTickRatio = 1 / upperTickRatio
	} else if isStableOrNativeInvolved(position) {
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
func convertTransferAmount(amountHex string, decimals int) float64 {
	amount := convertHexToBigInt(amountHex)
	amountFloat := new(big.Float).SetInt(amount)
	scaleDecFactor := new(big.Float).SetFloat64(math.Pow10(decimals))
	amountScaled, _ := new(big.Float).Quo(amountFloat, scaleDecFactor).Float64() // amount / 10^decimals
	return amountScaled
}

func splitBurnDatatoHexStrings(data string) (string, string, string, error) {
	const (
		AmountOffset            = 2
		AmountToken0Size        = 64
		AmountToken1Size        = 64
		RequiredDataFieldLength = 194
	)

	if len(data) != RequiredDataFieldLength {
		return "", "", "", fmt.Errorf("the data field length is not of expected size, could not parse amount fields.")
	}
	amountHex := "0x" + data[AmountOffset:AmountOffset+AmountToken0Size]
	amountToken0Hex := "0x" + data[AmountOffset+AmountToken0Size:AmountOffset+AmountToken0Size+AmountToken0Size]
	amountToken1Hex := "0x" + data[AmountOffset+AmountToken0Size+AmountToken1Size:]

	return amountHex, amountToken0Hex, amountToken1Hex, nil
}

func splitMintDatatoHexFields(data string) (string, string, string, error) {
	const (
		AmountOffset            = 2
		AmountOwnerAddress      = 64
		AmountSize              = 64
		AmountToken0Size        = 64
		AmountToken1Size        = 64
		RequiredDataFieldLength = 258
		AmountSkip              = AmountOffset + AmountOwnerAddress
	)

	if len(data) != RequiredDataFieldLength {
		return "", "", "", fmt.Errorf("the data field length is not of expected size, could not parse amount fields.")
	}
	amountHex := "0x" + data[AmountSkip:AmountSkip+AmountSize]
	amountToken0Hex := "0x" + data[AmountSkip+AmountSize:AmountSkip+AmountSize+AmountToken0Size]
	amountToken1Hex := "0x" + data[AmountSkip+AmountSize+AmountToken0Size:]

	return amountHex, amountToken0Hex, amountToken1Hex, nil
}

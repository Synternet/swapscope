package ethereum

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

// convertTickToRatio converts ticks received from "Mint" event to understandable token ratios.
// More info: http://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf
func convertTickToRatio(tick, token0Decimal, token1Decimal int) float64 {
	return math.Pow(1.0001, float64(tick)) / math.Pow(10, float64(token1Decimal-token0Decimal))
}

func parseEventLogMessage(data []byte) (EventLog, error) {
	var eLog EventLog
	err := json.Unmarshal(data, &eLog)
	if err != nil {
		return EventLog{}, err
	}

	if !eLog.hasTopics() {
		return EventLog{}, fmt.Errorf("parsed event log has no topics.")
	}

	return eLog, nil
}

func parseJsonToAbi(jsonABI string) abi.ABI {
	abi, err := abi.JSON(strings.NewReader(jsonABI))
	if err != nil {
		log.Fatal(err)
	}
	return abi
}

// convertTransferAmount converts Transfer's hex amount into scaled actual amount of tokens
func convertTransferAmount(amountHex string, decimals int) float64 {
	amount := convertHexToBigInt(amountHex)
	amountFloat := new(big.Float).SetInt(amount)
	scaleDecFactor := new(big.Float).SetFloat64(math.Pow10(decimals))
	amountScaled, _ := new(big.Float).Quo(amountFloat, scaleDecFactor).Float64() // amount / 10^decimals
	return amountScaled
}

func convertLogDataToHexAmounts(rawData string, eventName string) (string, string, error) {
	var abiToUse abi.ABI
	switch {
	case eventName == collectEvent:
		abiToUse = uniswapLiqPositionsABI
	default:
		abiToUse = uniswapLiqPoolsABI
	}
	hexString := strings.TrimPrefix(rawData, "0x")
	data, err := hex.DecodeString(hexString)

	var args = make(map[string]interface{})
	err = abiToUse.UnpackIntoMap(args, eventName, []byte(data))
	if err != nil {
		return "", "", err
	}

	resAmount0Hex := "0x" + args["amount0"].(*big.Int).Text(16)
	resAmount1Hex := "0x" + args["amount1"].(*big.Int).Text(16)
	return resAmount0Hex, resAmount1Hex, nil
}

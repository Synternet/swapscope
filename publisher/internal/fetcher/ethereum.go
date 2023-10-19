package fetcher

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/SyntropyNet/swapscope/publisher/pkg/repository"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	//go:embed token_abi.json
	tokenABI        string
	tokenABIMethods = []string{"name", "symbol", "decimals"}
)

type EthereumFetcher struct {
	db     repository.Repository
	url    string
	client *ethclient.Client
	abi    abi.ABI
}

func NewEthereumFetcher(ctx context.Context, rpcURL string, db repository.Repository) (*EthereumFetcher, error) {
	ret := &EthereumFetcher{
		db:  db,
		url: rpcURL,
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to Ethereum rpc node: %w", err)
	}
	ret.client = client

	// Load the contract's ABI (Application Binary Interface)
	tokenAbi, err := abi.JSON(strings.NewReader(tokenABI))
	if err != nil {
		return nil, fmt.Errorf("failed to load token ABI: %w", err)
	}
	for _, method := range tokenABIMethods {
		if _, found := tokenAbi.Methods[method]; !found {
			return nil, fmt.Errorf("loaded ABI does not contain %s method", method)
		}
	}

	ret.abi = tokenAbi

	return ret, nil
}

func (a *EthereumFetcher) Token(address string) (repository.Token, error) {
	// Try to look at database
	token, found := a.db.GetToken(address)
	if found {
		return token, nil
	}
	log.Println("Token", address, "not in DB... Try to fetch it from ETH node.")

	// There is no such token in database fetch the token from ETH node and add it to the database
	token, err := a.fetchToken(address)
	if err != nil {
		return repository.Token{}, err
	}
	err = a.db.AddToken(token)
	return token, err
}

func (a *EthereumFetcher) fetchToken(address string) (repository.Token, error) {
	// Create a context with a timeout (adjust the timeout as needed)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the contract's name() function
	symbol, err := a.callContractFunc(ctx, address, "symbol")
	if err != nil {
		log.Println("Failed to fetch token", address, "data (symbol) from full node:", err)
	}

	name, err := a.callContractFunc(ctx, address, "name")
	if err != nil {
		log.Println("Failed to fetch token", address, "data (name) from full node:", err)
	}

	decimals, err := a.callContractFunc(ctx, address, "decimals")
	if err != nil {
		log.Println("Failed to fetch token", address, "data (decimals) from full node:", err)
	}

	decimalsInt, err := strconv.Atoi(decimals)
	if err != nil {
		log.Println("Failed to convert token", address, "decimals data:", err)
	}

	log.Println("Token info gathered from node:", address, symbol, name, decimalsInt)

	return repository.Token{
		Address:  address,
		Symbol:   symbol,
		Name:     name,
		Decimals: decimalsInt,
	}, nil
}

func (a *EthereumFetcher) callContractFunc(ctx context.Context, address, method string) (string, error) {
	// Replace with the contract address of the token you want to query
	contractAddress := common.HexToAddress(address)

	// Call the contract's method() function
	result, err := a.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddress,
		Data: a.abi.Methods[method].ID,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call %s method: %w", method, err)
	}

	resultField, err := a.abi.Unpack(method, result)
	if err != nil {
		return "", fmt.Errorf("failed unpacking method's %s result: %w", method, err)
	}

	strField := fmt.Sprintf("%v", resultField[0])

	return strField, err
}

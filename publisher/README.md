# SwapScope publisher

SwapScope publisher consumes Syntropy Data Layer's Ethereum event log stream. Liquidity addition (mint) and removal (burn) events are detected, decoded and expanded with additional information about involved tokens. Token information (symbol, decimals, prices) are received through CoinGecko API.
 
## Usage

1. Compile code.
```
make build
```
</br>

2. Set variables.</br>
* Using .env file. See [.example.env](https://github.com/SyntropyNet/swapscope/blob/main/publisher/.example.env)</br>
* OR Using flags:
<details>
<summary>
Table of available flags
</summary>

| Flag                   | Description                                                        |Default value                     |
| ---------------------- | ------------------------------------------------------------------ |--------------------------------- |
| nats                   | NATS servers URL                                                   | -                                |
| nats-sub-creds         | NATS Subscriber Credentials File path (combined JWT and NKey file) | -                                |
| nats-sub-jwt           | NATS Subscriber Credentials JWT string                             | -                                |
| nats-sub-nkey          | NATS Subscriber Credentials NKey string                            | -                                |
| pub-subject-prefix     | Subject prefix                                                     | syntropy.analytics               |
| nats-pub-creds         | NATS Publisher Credentials File path (combined JWT and NKey file)  | -                                |
| nats-pub-jwt           | NATS Publisher Credentials JWT string                              | -                                |
| nats-pub-nkey          | NATS Publisher Credentials NKey string                             | -                                |
| eth-node-address       | Ethereum Full Node address                                         | -                                |
| db-host                | Database host string                                               | -                                |
| db-port                | Database port                                                      | -                                |
| db-user                | Database User Name                                                 | -                                |
| db-passw               | Database Password                                                  | -                                |
| db-name                | Database Name                                                      | -                                |
| cache-logs-expire      | Log Cache Record Expiration Time                                   | 2m                               |
| cache-logs-purge       | Log Cache Record Purge Time                                        | 3m                               |
| cache-prices-expire    | Token Price Cache Record Expiration Time                           | 2m                               |
| cache-prices-purge     | Token Price Cache Record Purge Time                                | 3m                               |
| coingecko-api          | CoinGecko API url                                                  | https://api.coingecko.com/api/v3 |
| api-timeout            | API fetch timeout                                                  | 2m                               |
| api-ratelimit          | Conservative API Rate Limit(e.g. 10-30 calls per minute)           | 12                               |
</details>

*If `nats-sub-creds` (nats creds file location) is set, then `nats-sub-jwt` and `nats-sub-nkey` are not required. Otherwise `nats-sub-jwt` and `nats-sub-nkey` can be set and `nats-sub-creds` has to be empty. The same applies to `nats-pub-*`.
</br>

3. Run with golang (with flags if any).
```
go run ./cmd/swapscope [flags]
```

## Docker

1. Build image.
```
docker build -f ./docker/Dockerfile -t swapscope .
```

2. Run container with passed environment variables.
```
docker run -it --rm --env-file=.env swapscope
```
*Local database will not work with Docker

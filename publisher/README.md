# SwapScope publisher
![GitHub go.mod Go version (branch & subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/SyntropyNet/swapscope/main?filename=publisher%2Fgo.mod)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)

SwapScope publisher consumes Syntropy Data Layer's Ethereum event log stream. Liquidity addition (mint) and removal (burn) events are detected, decoded and expanded with additional information about involved tokens. Token information (symbol, decimals, prices) are received through CoinGecko API.
 
## Usage

1. Compile code.
```
make build
```
</br>

2. Set variables.</br>
* Using .env file. See [.example.env](https://github.com/SyntropyNet/swapscope/blob/main/publisher/.example.env) (works if running locally)</br>
* OR Using flags or environment variables:

| Flag                 | Environment Variables   | (Mandatory?) Description                                                  | Default value                    |
| -------------------- | ----------------------- | ------------------------------------------------------------------------- |--------------------------------- |
| nats                 | NATS_URL                | (Y) NATS servers URL                                                      | -                                |
| nats-sub-creds       | NATS_SUB_CREDS_FILE     | (Y/N[^1]) NATS Subscriber Credentials File path (combined JWT and NKey file) | -                                |
| nats-sub-jwt         | NATS_SUB_JWT            | (Y/N[^1]) NATS Subscriber Credentials JWT string                             | -                                |
| nats-sub-nkey        | NATS_SUB_NKEY           | (Y/N[^1]) NATS Subscriber Credentials NKey string                            | -                                |
| pub-subject-prefix   | SUBJECT_PREFIX          | (Y[^2]) Subject prefix                                                      | syntropy.analytics               |
| nats-pub-creds       | NATS_PUB_CREDS_FILE     | (Y/N[^1]) NATS Publisher Credentials File path (combined JWT and NKey file)  | -                                |
| nats-pub-jwt         | NATS_PUB_JWT            | (Y/N[^1]) NATS Publisher Credentials JWT string                              | -                                |
| nats-pub-nkey        | NATS_PUB_NKEY           | (Y/N[^1]) NATS Publisher Credentials NKey string                             | -                                |
| eth-node-address     | ETH_NODE                | (N) Ethereum Full Node address                                            | -                                |
| db-host              | DB_HOST                 | (Y) Database host string                                                  | -                                |
| db-port              | DB_PORT                 | (Y) Database port                                                         | -                                |
| db-user              | DB_USER                 | (Y) Database User Name                                                    | -                                |
| db-passw             | DB_PASSWORD             | (Y) Database Password                                                     | -                                |
| db-name              | DB_NAME                 | (Y) Database Name                                                         | -                                |
| cache-logs-expire    | LOG_CACHE_EXPIRY_TIME   | (N[^2]) Log Cache Record Expiration Time                                    | 2m                               |
| cache-logs-purge     | LOG_CACHE_PURGE_TIME    | (N[^2]) Log Cache Record Purge Time                                         | 3m                               |
| cache-prices-expire  | PRICE_CACHE_EXPIRY_TIME | (N[^2]) Token Price Cache Record Expiration Time                            | 2m                               |
| cache-prices-purge   | PRICE_CACHE_PURGE_TIME  | (N[^2]) Token Price Cache Record Purge Time                                 | 3m                               |
| coingecko-api        | COINGECKO_API_URL       | (N[^2]) CoinGecko API url                                                   | https://api.coingecko.com/api/v3 |
| api-timeout          | API_FETCH_TIMEOUT       | (N[^2]) API fetch timeout                                                   | 2m                               |
| api-ratelimit        | API_RATE_LIMIT          | (N[^2]) Conservative API Rate Limit (e.g. 10-30 calls per minute)           | 12                               |

[^1]: If `nats-sub-creds` (nats creds file location) is set, then `nats-sub-jwt` and `nats-sub-nkey` are not required. Otherwise `nats-sub-jwt` and `nats-sub-nkey` can be set and `nats-sub-creds` has to be empty. The same applies to `nats-pub-*`.</br>
[^2]: Default value is set by the app in `env.go`. For Cache timings if set to 0 - cache elements will never expire.
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

2. Run container.
```
docker run -it --rm --env-file=.env swapscope
```
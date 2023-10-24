package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	PublisherPrefixName          = "SUBJECT_PREFIX"
	LogCacheExpirationTimeName   = "LOG_CACHE_EXPIRY_TIME"
	LogCachePurgeTimeName        = "LOG_CACHE_PURGE_TIME"
	PriceCacheExpirationTimeName = "PRICE_CACHE_EXPIRY_TIME"
	PriceCachePurgeTimeName      = "PRICE_CACHE_PURGE_TIME"
	CoinGeckoApiUrl              = "COINGECKO_API_URL"
	ApiFetchTimeout              = "API_FETCH_TIMEOUT"
	ApiRateLimit                 = "API_RATE_LIMIT"
)

type ServiceConfig struct {
	natsUrls                 *string
	userSubCredsFile         *string
	userSubCredsJWT          *string
	userSubCredsSeed         *string
	userPubCredsFile         *string
	userPubCredsJWT          *string
	userPubCredsSeed         *string
	publisherPrefix          *string
	ethNodeAddress           *string
	dbHost                   *string
	dbPort                   *string
	dbUser                   *string
	dbPassword               *string
	dbName                   *string
	logCacheExpirationTime   *time.Duration
	logCachePurgeTime        *time.Duration
	priceCacheExpirationTime *time.Duration
	priceCachePurgeTime      *time.Duration
	coinGeckoApiUrl          *string
	apiFetchTimeout          *time.Duration
	apiRateLimit             *int
}

func setupDefaults() {
	setEnvDefaults(PublisherPrefixName, "syntropy.analytics")
	setEnvDefaults(LogCacheExpirationTimeName, "2m")
	setEnvDefaults(LogCachePurgeTimeName, "3m")
	setEnvDefaults(PriceCacheExpirationTimeName, "2m")
	setEnvDefaults(PriceCachePurgeTimeName, "3m")
	setEnvDefaults(ApiFetchTimeout, "2m")
	setEnvDefaults(ApiRateLimit, "12")
	setEnvDefaults(CoinGeckoApiUrl, "https://api.coingecko.com/api/v3")
}

func setEnvDefaults(field string, value string) {
	if os.Getenv(field) == "" {
		os.Setenv(field, value)
	}
}

func newServiceConfig() *ServiceConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load .env file")
	}

	setupDefaults()

	cfg := &ServiceConfig{
		natsUrls:                 flag.String("nats", os.Getenv("NATS_URL"), "NATS server URLs (separated by comma)"),
		userSubCredsFile:         flag.String("nats-sub-creds", os.Getenv("NATS_SUB_CREDS_FILE"), "NATS Subscriber Credentials File path (combined JWT and NKey file)"),
		userSubCredsJWT:          flag.String("nats-sub-jwt", os.Getenv("NATS_SUB_JWT"), "NATS Subscriber Credentials JWT string"),
		userSubCredsSeed:         flag.String("nats-sub-nkey", os.Getenv("NATS_SUB_NKEY"), "NATS Subscriber Credentials NKey string"),
		publisherPrefix:          flag.String("pub-subject-prefix", os.Getenv(PublisherPrefixName), "Subject prefix"),
		userPubCredsFile:         flag.String("nats-pub-creds", os.Getenv("NATS_PUB_CREDS_FILE"), "NATS Publisher Credentials File path (combined JWT and NKey file)"),
		userPubCredsJWT:          flag.String("nats-pub-jwt", os.Getenv("NATS_PUB_JWT"), "NATS Publisher Credentials JWT string"),
		userPubCredsSeed:         flag.String("nats-pub-nkey", os.Getenv("NATS_PUB_NKEY"), "NATS Publisher Credentials NKey string"),
		ethNodeAddress:           flag.String("eth-node-address", os.Getenv("ETH_NODE"), "Ethereum Full Node address"),
		dbHost:                   flag.String("db-host", os.Getenv("DB_HOST"), "Database host string"),
		dbPort:                   flag.String("db-port", os.Getenv("DB_PORT"), "Database port"),
		dbUser:                   flag.String("db-user", os.Getenv("DB_USER"), "Database User Name"),
		dbPassword:               flag.String("db-passw", os.Getenv("DB_PASSWORD"), "Database Password"),
		dbName:                   flag.String("db-name", os.Getenv("DB_NAME"), "Database Name"),
		logCacheExpirationTime:   flag.Duration("cache-logs-expire", stringToDuration(os.Getenv(LogCacheExpirationTimeName)), "Log Cache Record Expiration Time"),
		logCachePurgeTime:        flag.Duration("cache-logs-purge", stringToDuration(os.Getenv(LogCachePurgeTimeName)), "Log Cache Record Purge Time"),
		priceCacheExpirationTime: flag.Duration("cache-prices-expire", stringToDuration(os.Getenv(PriceCacheExpirationTimeName)), "Token Price Cache Record Expiration Time"),
		priceCachePurgeTime:      flag.Duration("cache-prices-purge", stringToDuration(os.Getenv(PriceCachePurgeTimeName)), "Token Price Cache Record Purge Time"),
		coinGeckoApiUrl:          flag.String("coingecko-api", os.Getenv(CoinGeckoApiUrl), "CoinGecko API url"),
		apiFetchTimeout:          flag.Duration("api-timeout", stringToDuration(os.Getenv(ApiFetchTimeout)), "API fetch timeout"),
		apiRateLimit:             flag.Int("api-ratelimit", stringToInt(os.Getenv(ApiRateLimit)), "Conservative API Rate Limit(e.g. 10-30 calls per minute)"),
	}

	flag.Parse()

	return cfg
}

func stringToDuration(stringDur string) time.Duration {
	duration, err := time.ParseDuration(stringDur)
	if err != nil && stringDur != "" {
		log.Panicln("Error converting string duration to time.Duration (perhaps add 's', 'm', or 'h'?) :", err)
	}
	return duration
}

func stringToInt(stringInt string) int {
	i, err := strconv.ParseInt(stringInt, 10, 64)
	if err != nil && stringInt != "" {
		log.Panicln("Error converting string to int:", err)
	}
	return int(i)
}

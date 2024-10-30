package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	svcnats "github.com/Synternet/pubsub-go/pubsub"
	"github.com/Synternet/swapscope/publisher/internal/analytics/ethereum"
	"github.com/Synternet/swapscope/publisher/internal/fetcher"
	"github.com/Synternet/swapscope/publisher/internal/repository/db"
	"github.com/Synternet/swapscope/publisher/internal/service"
	"github.com/nats-io/nats.go"
)

func connectNatsService(natsURL string, userCredsFile string, userCredsJWT string, userCredsSeed string) *svcnats.NatsService {
	opts := []nats.Option{}

	if userCredsFile != "" {
		opts = append(opts, nats.UserCredentials(userCredsFile))
	}

	if userCredsJWT != "" && userCredsSeed != "" {
		opts = append(opts, nats.UserJWTAndSeed(userCredsJWT, userCredsSeed))
	}

	svcnSub := svcnats.MustConnect(
		svcnats.Config{
			URI:  natsURL,
			Opts: opts,
		})
	log.Println("NATS server connected.")

	return svcnSub
}

func main() {
	cfg := newServiceConfig()

	svcnSub := connectNatsService(*cfg.natsUrls, *cfg.userSubCredsFile, *cfg.userSubCredsJWT, *cfg.userSubCredsSeed)
	svcnPub := connectNatsService(*cfg.natsUrls, *cfg.userPubCredsFile, *cfg.userPubCredsJWT, *cfg.userPubCredsSeed)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db, err := db.New(*cfg.dbHost, *cfg.dbPort, *cfg.dbUser, *cfg.dbPassword, *cfg.dbName)
	if err != nil {
		panic(err)
	}

	cgFetcher, err := fetcher.NewCoingeckoFetcher(
		ctx, db,
		*cfg.coinGeckoApiUrl,
		*cfg.priceCacheExpirationTime,
		*cfg.priceCachePurgeTime,
		*cfg.apiFetchTimeout,
		*cfg.apiRateLimit,
	)
	if err != nil {
		panic(err)
	}

	/* Possible to have Ethereum full node fetcher: */
	// ethFetcher, err := fetcher.NewEthereumFetcher(ctx, *cfg.ethNodeAddress, db)
	// if err != nil {
	// 	panic(err)
	// }

	a, err := ethereum.New(ctx, db,
		ethereum.WithEventLogCache(*cfg.logCacheExpirationTime, *cfg.logCachePurgeTime),
		ethereum.WithTokenPriceFetcher(cgFetcher),
		ethereum.WithTokenFetcher(cgFetcher),
	)
	if err != nil {
		panic(err)
	}

	s, err := service.New(ctx,
		service.WithNATS(svcnSub, svcnPub),
		service.WithAnalytics(a),
		service.WithPrefix(*cfg.publisherPrefix),
	)
	if err != nil {
		panic(err)
	}

	go s.Serve()

	<-ctx.Done()
	log.Println("Analytics service interrupted")

	shutdown, sdCancel := context.WithTimeout(s.Done(), time.Second*3)
	defer sdCancel()

	<-shutdown.Done()
	if shutdown.Err() != nil && shutdown.Err() != context.Canceled {
		panic(shutdown.Err())
	}
	log.Println("Done")
}

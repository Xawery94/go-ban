package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/jamiealquiza/envy"

	"github.com/pkg/errors"
)

var (
	redisURL     = flag.String("redisUrl", "redis://redis", "Redis Url")
	addr         = flag.String("addr", ":80", "Listen address")
	readTimeout  = flag.Duration("readTimeout", 10*time.Second, "TCP read timeout")
	writeTimeout = flag.Duration("writeTimeout", 10*time.Second, "TCP write timeout")
)

func parseFlags() {
	envy.Parse("AUTH")
	flag.Parse()
}

type config struct {
	rc *redis.Client
}

func newConfig() (*config, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to Redis")
	}

	return &config{
		rc: client,
	}, nil
}

// func (c *config) AuthnServer() authn.AuthnServer {
// 	svc, err := authn.NewPg(
// 		c.DB(),
// 		authn.IDKey(c.IDKey()),
// 		authn.AccessKey(c.AccessKey()),
// 		authn.Issuer(*issuer),
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize authn.Service: %s", err)
// 	}
// 	return svc
// }

// func (c *config) Server() *.Server {
// am := authz.NewMiddleware(
// authz.Key(c.AccessKey().Key),
// authz.Issuer(*issuer),
// )

// em := error.NewMiddleware(codes.Internal, codes.Unavailable, codes.DataLoss)

// opts = append(
// opts
// 	.UnaryInterceptor(_middleware.ChainUnaryServer(
// 		am.UnaryInterceptor(),
// 		em.UnaryInterceptor(),
// 	)),
// 	.StreamInterceptor(_middleware.ChainStreamServer(
// 		am.StreamInterceptor(),
// 		em.StreamInterceptor(),
// 	)),
// )

// 	if *certFile != "" && *keyFile != "" {
// 		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
// 		if err != nil {
// 			log.Fatalf("Failed to load TLS cert/key: %s", err)
// 		}
// 		opts = append(opts, .Creds(creds))
// 	}
//
// return .NewServer()
// }

func (c *config) Close() {
	if c.rc != nil {
		c.rc.Close()
	}
}

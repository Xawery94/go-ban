package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"

	"github.com/pkg/errors"
)

var (
	redisURL     = flag.String("redisUrl", "redis://redis", "Redis Url")
	addr         = flag.String("addr", ":8080", "Listen address")
	readTimeout  = flag.Duration("readTimeout", 10*time.Second, "TCP read timeout")
	writeTimeout = flag.Duration("writeTimeout", 10*time.Second, "TCP write timeout")
)

// func parseFlags() {
// 	envy.Parse("AUTH")
// 	flag.Parse()
// }

type config struct {
	rc redis.Client
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

// func (c *config) AccountsServer() account.AccountsServer {
// 	return account.NewPg(
// 		c.DB(),
// 		event.NewNatsEmitter(c.Nats(), "auth-service", "en"),
// 		account.Locales(locales.values...),
// 	)
// }

// func (c *config) OperatorsServer() operator.OperatorsServer {
// 	return operator.NewPg(c.DB())
// }

// func (c *config) HealthServer() health.HealthServer {
// 	return health.New(
// 		"auth-service",
// 		version,
// 		health.Downstream("db", func(ctx context.Context) (bool, error) {
// 			err := c.DB().PingContext(ctx)
// 			switch {
// 			case err == context.Canceled || err == context.DeadlineExceeded:
// 				return false, err
// 			default:
// 				return err == nil, nil
// 			}
// 		}),
// 		health.Downstream("nats", func(ctx context.Context) (bool, error) {
// 			return c.Nats().NatsConn().IsConnected(), nil
// 		}),
// 	)
// }

// func (c *config) Server(opts ....ServerOption) *.Server {
// 	am := authz.NewMiddleware(
// 		authz.Key(c.AccessKey().Key),
// 		authz.Issuer(*issuer),
// 	)

// 	em := error.NewMiddleware(codes.Internal, codes.Unavailable, codes.DataLoss)

// 	opts = append(
// 		opts,
// 		.UnaryInterceptor(_middleware.ChainUnaryServer(
// 			am.UnaryInterceptor(),
// 			em.UnaryInterceptor(),
// 		)),
// 		.StreamInterceptor(_middleware.ChainStreamServer(
// 			am.StreamInterceptor(),
// 			em.StreamInterceptor(),
// 		)),
// 	)

// 	if *certFile != "" && *keyFile != "" {
// 		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
// 		if err != nil {
// 			log.Fatalf("Failed to load TLS cert/key: %s", err)
// 		}
// 		opts = append(opts, .Creds(creds))
// 	}

// 	return .NewServer(opts...)
// }

// func (c *config) Close() {
// 	if c.db != nil {
// 		c.db.Close()
// 	}

// 	if c.nc != nil {
// 		c.nc.Close()
// 	}
// }

// type localesFlag struct {
// 	values []language.Tag
// }

// func (l *localesFlag) String() string {
// 	vs := make([]string, len(l.values))
// 	for i, loc := range l.values {
// 		vs[i] = loc.String()
// 	}

// 	return fmt.Sprint(vs)
// }

// func (l *localesFlag) Set(raw string) error {
// 	vs := strings.Split(raw, ",")
// 	for _, v := range vs {
// 		loc, err := language.Parse(v)
// 		if err != nil {
// 			return err
// 		}
// 		l.values = append(l.values, loc)
// 	}

// 	return nil
// }

package authz

import (
	"time"
)

var (
	DefaultKey = []byte("secret")
)

type config struct {
	key    interface{}
	issuer string
	leeway time.Duration
}

type Opt func(c *config)

func Key(key interface{}) Opt {
	return func(c *config) {
		c.key = key
	}
}

func Issuer(issuer string) Opt {
	return func(c *config) {
		c.issuer = issuer
	}
}

func Leeway(leeway time.Duration) Opt {
	return func(c *config) {
		c.leeway = leeway
	}
}

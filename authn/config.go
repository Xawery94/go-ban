package authn

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	jose "gopkg.in/square/go-jose.v2"
)

var (
	DefaultKey             = jose.SigningKey{Algorithm: jose.HS256, Key: []byte("secret")}
	DefaultCost            = bcrypt.DefaultCost
	DefaultAccessLifespan  = 5 * time.Minute
	DefaultRefreshLifespan = 24 * time.Hour
)

type config struct {
	accessKey, idKey jose.SigningKey
	issuer           string
	cost             int
	accessLifespan   time.Duration
	refreshLifespan  time.Duration
}

type Opt func(c *config)

func AccessKey(key jose.SigningKey) Opt {
	return func(c *config) {
		c.accessKey = key
	}
}

func IDKey(key jose.SigningKey) Opt {
	return func(c *config) {
		c.idKey = key
	}
}

func Issuer(iss string) Opt {
	return func(c *config) {
		c.issuer = iss
	}
}

func AccessLifespan(lifespan time.Duration) Opt {
	return func(c *config) {
		c.accessLifespan = lifespan
	}
}

func RefreshLifespan(lifespan time.Duration) Opt {
	return func(c *config) {
		c.refreshLifespan = lifespan
	}
}

func PasswordCost(cost int) Opt {
	return func(c *config) {
		c.cost = cost
	}
}

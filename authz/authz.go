package authz

import (
	"context"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	ErrInvalidAuthorization = status.Error(codes.Unauthenticated, "Invalid authorization")
	ErrInvalidToken         = status.Error(codes.Unauthenticated, "Invalid token")
	ErrExpiredToken         = status.Error(codes.Unauthenticated, "Token expired")
	ErrInvalidClaims        = status.Error(codes.Unauthenticated, "Invalid claims")
)

type contextKeyType string

var contextKey = contextKeyType("authz")

// Authz contains user authorization
type Authz struct {
	ID       int64
	Username string
	Scopes   []string
}

// HasScope checks if authorization contains given scope
func (a *Authz) HasScope(scope string) bool {
	for _, v := range a.Scopes {
		if v == scope {
			return true
		}
	}

	return false
}

// FromContext retrieves authorization from context
func FromContext(ctx context.Context) Authz {
	raw := ctx.Value(contextKey)
	if raw == nil {
		return Authz{}
	}

	return raw.(Authz)
}

// Middleware validates JWT token authorization
type Middleware interface {
	StreamInterceptor() grpc.StreamServerInterceptor
	UnaryInterceptor() grpc.UnaryServerInterceptor
}

func NewMiddleware(opts ...Opt) Middleware {
	conf := config{
		key: DefaultKey,
	}

	for _, opt := range opts {
		opt(&conf)
	}

	return &middleware{
		conf: conf,
	}
}

type middleware struct {
	conf config
}

func (m *middleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := m.auth(stream.Context())
		if err != nil {
			return err
		}

		return handler(srv, &wrappedStream{
			ServerStream: stream,
			ctx:          ctx,
		})
	}
}

func (m *middleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := m.auth(ctx)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (m *middleware) auth(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, nil
	}

	vs := md.Get("authorization")
	if len(vs) == 0 {
		return ctx, nil
	}

	auth := vs[0]
	raw := strings.TrimPrefix(auth, "Bearer ")
	if raw == auth {
		return nil, ErrInvalidAuthorization
	}

	tok, err := jwt.ParseSigned(raw)
	if err != nil {
		log.Printf("Token validation failed: %+v", err)
		return nil, ErrInvalidToken
	}

	cl := struct {
		jwt.Claims
		ID       int64    `json:"sub,string"`
		Username string   `json:"username"`
		Scopes   []string `json:"scopes"`
	}{}
	err = tok.Claims(m.conf.key, &cl)
	switch {
	case err == jose.ErrCryptoFailure:
		return nil, ErrInvalidToken
	case err != nil:
		log.Printf("Token validation failed: %+v", err)
		return nil, ErrInvalidToken
	}

	err = cl.ValidateWithLeeway(jwt.Expected{
		Issuer: m.conf.issuer,
		Time:   time.Now(),
	}, m.conf.leeway)
	switch {
	case err == jwt.ErrExpired || err == jwt.ErrNotValidYet:
		return nil, ErrExpiredToken
	case err != nil:
		return nil, ErrInvalidClaims
	}

	return context.WithValue(ctx, contextKey, Authz{
		ID:       cl.ID,
		Username: cl.Username,
		Scopes:   cl.Scopes,
	}), nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedStream) Context() context.Context {
	return s.ctx
}

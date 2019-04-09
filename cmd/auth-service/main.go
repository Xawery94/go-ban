package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Xawery/auth-service/account"
	"github.com/Xawery/auth-service/authn"
	_ "github.com/lib/pq"
)

var version = "unspecified"

func main() {
	parseFlags()

	c, err := newConfig()
	if err != nil {
		log.Printf("Failed to initialize: %s", err)
		return
	}
	defer c.Close()

	srv := c.Server()
	account.RegisterAccountsServer(srv, c.AccountsServer())
	authn.RegisterAuthnServer(srv, c.AuthnServer())
	health.RegisterHealthServer(srv, c.HealthServer())
	// reflection.Register(srv)
	operator.RegisterOperatorsServer(srv, c.OperatorsServer())

	go func() {
		lis, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		if err := srv.Serve(lis); err != http.ErrServerClosed && err != nil {
			log.Printf("Server error: %+v", err)
		}
	}()

	log.Printf("auth-service (%s) listening at %s", version, *addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	srv.GracefulStop()
}

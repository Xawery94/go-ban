package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Xawery/auth-service/handler"
	_ "github.com/lib/pq"
)

func main() {
	server := http.Server{
		Addr:    fmt.Sprint(*addr),
		Handler: handler.NewHandler(),
	}

	// *redisClient, _ = redis.GetRedisClient()

	// Run server
	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed !")
	}
}

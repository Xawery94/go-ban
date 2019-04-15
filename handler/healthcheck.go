package handler

import (
	"log"
	"net/http"

	"github.com/Xawery/auth-service/redis"
)

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get instance redis client
	redis, _ := redis.GetRedisClient()
	if err := redis.Ping().Err(); err != nil {
		log.Printf("redis unaccessible error: %v ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	return
}

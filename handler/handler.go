package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Xawery/auth-service/redis"
)

func addBan(w http.ResponseWriter, r *http.Request) {
	rClient, _ := redis.GetRedisClient()

	body := struct {
		IP  string        `json:"ip"`
		TTL time.Duration `json:"ttl"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("could not decode request: %v", err)
		http.Error(w, "could not decode request", http.StatusInternalServerError)
		return
	}

	//Add new key to bann list
	rClient.BanClient(body.IP, body.TTL)

	w.WriteHeader(http.StatusOK)
	return
}

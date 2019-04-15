package handler

import (
	"net/http"
)

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/add", addBan)

	return mux
}

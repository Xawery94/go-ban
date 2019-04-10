package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	parseFlags()

	c, err := newConfig()

	if err != nil {
		log.Printf("Failed to initialize: %s", err)
		return
	}
	defer c.Close()
	fmt.Println(*addr)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/todos", TodoIndex)
	router.HandleFunc("/todos/{todoId}", TodoShow)
	log.Fatal(http.ListenAndServe(*addr, router))
}

var (
	// NotFound indicates that given resource was not found in database
	NotFound = "Not found"
	// Unavailable indicates that downstream operation failed
	Unavailable = "Temporarily unavailable"
	// Internal indicates that internal error occured
	Internal = "Internal error"
)

//Account struct
type Account struct {
	ID        int64
	Username  string
	Timestamp *time.Time
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func TodoIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Todo Index!")
}

func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["todoId"]
	fmt.Fprintln(w, "Todo show:", todoId)
}

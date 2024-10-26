package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/challenge/{mail}", getChallengePage).Methods("GET")
	r.HandleFunc("/challenge/{mail}", saveChallengePage).Methods("POST")

	log.Println("http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/contact/get/{uid}", getContactHandler).Methods("GET")
	r.HandleFunc("/contact/accept", HelloHandler).Methods("POST")

	log.Println("http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

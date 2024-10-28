package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	if os.Getenv("RECAPTCHA_SITE") == "" || os.Getenv("RECAPTCHA_SECRET") == "" {
		log.Fatal("RECAPTCHA_SITE and RECAPTCHA_SECRET must be set")
	}

	r := mux.NewRouter()

	r.HandleFunc("/challenge/{mail}", getChallengePage).Methods("GET")
	r.HandleFunc("/challenge/{mail}", saveChallengePage).Methods("POST")
	r.HandleFunc("/success", getSuccessPage).Methods("GET")

	log.Println("http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

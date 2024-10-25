package main

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/polo2ro/mailinwhite/libs/contact"
)

func AcceptHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the POST request body
	var requestData struct {
		Mail         string `json:"mail"`
		CaptchaToken string `json:"captchaToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !verifyCaptcha(requestData.CaptchaToken) {
		http.Error(w, "Invalid captcha", http.StatusUnauthorized)
		return
	}

	// Set the confirmedHuman status in Redis
	rdb := contact.GetClient()
	defer rdb.Close()
	ctx := context.Background()

	err := rdb.Set(ctx, requestData.Mail, contact.StatusConfirmedHuman, 0).Err()
	if err != nil {
		http.Error(w, "Failed to set status in Redis", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]contact.Status{"status": contact.StatusConfirmedHuman}
	json.NewEncoder(w).Encode(response)
}

// Helper function to verify reCAPTCHA token
func verifyCaptcha(token string) bool {
	// TODO: Implement reCAPTCHA verification logic
	// This should make a request to the reCAPTCHA API and return true if valid
	return true // Placeholder
}

func getContactStatusHandler(w http.ResponseWriter, r *http.Request) {
	mail := mux.Vars(r)["mail"]

	rdb := contact.GetClient()
	defer rdb.Close()
	ctx := context.Background()

	var err error

	status, err := rdb.Get(ctx, mail).Result()
	if err == redis.Nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": status}
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		http.Error(w, "json encode error", http.StatusInternalServerError)
		return
	}
}

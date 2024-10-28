package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Helper function to verify reCAPTCHA token
func verifyCaptcha(token string) bool {
	url := fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s", os.Getenv("RECAPTCHA_SECRET"), token)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", nil)
	if err != nil {
		log.Printf("Error making request to reCAPTCHA: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success bool     `json:"success"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding reCAPTCHA response: %v", err)
		return false
	}

	return result.Success
}

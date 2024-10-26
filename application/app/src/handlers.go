package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"context"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/polo2ro/mailinwhite/libs/contact"
)

func sendPendingMails(ctx context.Context, senderEmail string) error {
	rdb := contact.GetMessagesClient()
	defer rdb.Close()

	senderKey := fmt.Sprintf("sender:%s", senderEmail)
	messageIDs, err := rdb.SMembers(ctx, senderKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get message IDs for sender %s: %w", senderEmail, err)
	}

	for _, messageID := range messageIDs {
		if err := contact.SendMessage(ctx, messageID); err != nil {
			log.Printf("Error sending message %s: %v", messageID, err)
			continue
		}

		if err := rdb.SRem(ctx, senderKey, messageID).Err(); err != nil {
			log.Printf("Error removing message ID %s from sender set: %v", messageID, err)
		}
	}

	return nil
}

func saveChallengePage(w http.ResponseWriter, r *http.Request) {
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
	rdb := contact.GetAddressesClient()
	defer rdb.Close()
	ctx := context.Background()

	err := rdb.Set(ctx, requestData.Mail, contact.StatusConfirmedHuman, 0).Err()
	if err != nil {
		http.Error(w, "Failed to set status in Redis", http.StatusInternalServerError)
		return
	}

	// Send the pending mails
	if err := sendPendingMails(ctx, requestData.Mail); err != nil {
		log.Printf("Error sending pending mails for %s: %v", requestData.Mail, err)
		// Note: We continue even if there's an error sending pending mails
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]int{"status": contact.StatusConfirmedHuman}
	json.NewEncoder(w).Encode(response)
}

// Helper function to verify reCAPTCHA token
func verifyCaptcha(token string) bool {
	// TODO: Implement reCAPTCHA verification logic
	// This should make a request to the reCAPTCHA API and return true if valid
	return true // Placeholder
}

func getMailStatus(mail string) (int, int, error) {
	rdb := contact.GetAddressesClient()
	defer rdb.Close()
	ctx := context.Background()

	status, err := rdb.Get(ctx, mail).Result()
	if err == redis.Nil {
		return 0, http.StatusNotFound, errors.New("contact not found")
	} else if err != nil {
		return 0, http.StatusInternalServerError, fmt.Errorf("redis: %w", err)
	}

	statusInt, err := strconv.Atoi(status)
	if err != nil {
		return 0, http.StatusInternalServerError, fmt.Errorf("invalid status format: %w", err)
	}

	return statusInt, 0, nil
}

func getChallengePage(w http.ResponseWriter, r *http.Request) {
	mail := mux.Vars(r)["mail"]

	contactStatus, httpCode, err := getMailStatus(mail)
	if err != nil {
		http.Error(w, err.Error(), httpCode)
		return
	}

	if contactStatus != contact.StatusPending {
		http.Error(w, "Invalid contact status", http.StatusBadRequest)
		return
	}

	// Load and execute the HTML template
	tmpl, err := template.ParseFiles("templates/captchaChallenge.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, struct{ Mail string }{Mail: mail})
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

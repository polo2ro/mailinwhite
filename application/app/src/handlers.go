package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
			return fmt.Errorf("error sending message %s: %w", messageID, err)
		}

		if err := rdb.SRem(ctx, senderKey, messageID).Err(); err != nil {
			return fmt.Errorf("error removing message ID %s from sender set: %v", messageID, err)
		}
	}

	return nil
}

func saveChallengePage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mail := r.FormValue("mail")

	if !verifyCaptcha(r.FormValue("g-recaptcha-response")) {
		http.Error(w, "Invalid captcha", http.StatusUnauthorized)
		return
	}

	rdb := contact.GetAddressesClient()
	defer rdb.Close()
	ctx := context.Background()

	err := rdb.Set(ctx, mail, contact.StatusConfirmedHuman, 0).Err()
	if err != nil {
		http.Error(w, "Failed to set status in Redis", http.StatusInternalServerError)
		return
	}

	if err := sendPendingMails(ctx, mail); err != nil {
		http.Error(w, fmt.Sprintf("Error sending pending mails for %s: %s", mail, err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app/success", http.StatusSeeOther)
}

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

	tmpl, err := template.ParseFiles("templates/captchaChallenge.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, struct {
		Mail          string
		RecaptchaSite string
	}{Mail: mail, RecaptchaSite: os.Getenv("RECAPTCHA_SITE")})
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func getSuccessPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/accepted.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

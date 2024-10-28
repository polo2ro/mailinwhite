package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"context"

	"github.com/gorilla/mux"

	"github.com/polo2ro/mailinwhite/libs/common"
)

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

	rdb := common.GetAddressesClient()
	defer rdb.Close()
	ctx := context.Background()

	err := rdb.Set(ctx, mail, common.StatusConfirmedHuman, 0).Err()
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

func getChallengePage(w http.ResponseWriter, r *http.Request) {
	mail := mux.Vars(r)["mail"]

	contactStatus, httpCode, err := getMailStatus(mail)
	if err != nil {
		http.Error(w, err.Error(), httpCode)
		return
	}

	if contactStatus != common.StatusPending {
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

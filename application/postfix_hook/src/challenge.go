package main

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strconv"

	"html/template"

	"github.com/joho/godotenv"
)

// sendChallengeRequestEmail sends an HTML email with a confirmation link to verify the sender is human
func sendChallengeRequestEmail(senderEmail string, recipientEmail string, confirmationLink string) error {
	envFile, err := godotenv.Read("/home/filter/.env")
	if err != nil {
		return fmt.Errorf("read env file: %w", err)
	}

	smtpPort, err := strconv.Atoi(envFile["SMTP_PORT"])
	if err != nil {
		return fmt.Errorf("port int conversion: %w", err)
	}

	if envFile["SMTP_HOST"] == "" || smtpPort == 0 {
		return fmt.Errorf("missing required SMTP configuration environment variables")
	}

	subject := "Confirm Your Humanity to Deliver Your Message"

	// Load the HTML template
	tmpl, err := template.ParseFiles("/home/filter/templates/challenge_request.html")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var htmlBody string
	data := struct {
		SenderEmail      string
		RecipientEmail   string
		ConfirmationLink string
	}{
		SenderEmail:      senderEmail,
		RecipientEmail:   recipientEmail,
		ConfirmationLink: confirmationLink,
	}

	// Execute the template with the data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	htmlBody = buf.String()

	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", senderEmail, subject, htmlBody))

	from := "mailinwhite@example.com"
	addr := fmt.Sprintf("%s:%d", envFile["SMTP_HOST"], smtpPort)
	var smtpErr error
	if envFile["SMTP_LOGIN"] != "" {
		auth := smtp.PlainAuth("", envFile["SMTP_LOGIN"], envFile["SMTP_PASSWORD"], envFile["SMTP_HOST"])
		smtpErr = smtp.SendMail(addr, auth, from, []string{senderEmail}, message)
	} else {
		smtpErr = smtp.SendMail(addr, nil, from, []string{senderEmail}, message)
	}

	if smtpErr != nil {
		return fmt.Errorf("failed to send confirmation email: %v", smtpErr)
	}

	return nil
}

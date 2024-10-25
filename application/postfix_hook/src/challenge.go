package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strconv"
)

// sendConfirmationEmail sends an HTML email with a confirmation link to verify the sender is human
func sendConfirmationEmail(senderEmail string, recipientEmail string, confirmationLink string) error {
	// SMTP server configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("port int conversion: %w", err)
	}
	smtpUser := os.Getenv("SMTP_LOGIN")
	password := os.Getenv("SMTP_PASSWORD")

	// Ensure required environment variables are set
	if smtpHost == "" || smtpPort == 0 || smtpUser == "" || password == "" {
		return fmt.Errorf("missing required SMTP configuration environment variables")
	}

	// Compose the email message
	subject := "Confirm Your Humanity to Deliver Your Message"
	htmlBody := fmt.Sprintf(`
		<html>
		<body>
		<p>You tried to submit an email to %s but your address %s has been declared unknown by the mail server.</p>
		<p>Please <a href="%s">click here</a> to verify that you are human and then proceed to the delivery of your message.</p>
		</body>
		</html>
	`, recipientEmail, senderEmail, confirmationLink)

	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", recipientEmail, subject, htmlBody))

	// Authentication and sending the email
	auth := smtp.PlainAuth("", smtpUser, password, smtpHost)
	err = smtp.SendMail(fmt.Sprintf("%s:%d", smtpHost, smtpPort), auth, smtpUser, []string{recipientEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email: %v", err)
	}

	return nil
}

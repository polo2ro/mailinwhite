package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"html/template"

	"github.com/polo2ro/mailinwhite/libs/common"
)

// sends an HTML email with a confirmation link to verify the sender is human
func sendChallengeRequestEmail(senderEmail string, recipients []string, confirmationLink string) error {
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
		RecipientEmail:   fmt.Sprintf("%v", common.GetValidRecipients(recipients)),
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

	args := append([]string{"-G", "-i", "-f", senderEmail}, recipients...)
	cmd := exec.Command("/usr/sbin/sendmail", args...)
	cmd.Stdin = bytes.NewReader(message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send challenge email: %v", err)
	}

	return nil
}

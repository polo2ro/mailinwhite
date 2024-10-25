package main

// receive mail from postfix
// if contact is approved in directory, forward the mail
// if contact does not exists or is not approved, send a challenge and postpone the mail in a queue

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/polo2ro/mailinwhite/libs/contact"
)

const (
	inspectDir = "/home/filter/spool"

	// Postfix exit codes
	EX_OK          = 0  // Successful termination
	EX_TEMPFAIL    = 75 // Temporary failure
	EX_UNAVAILABLE = 69 // Service unavailable
	EX_USAGE       = 64 // Command line usage error
)

// postfix mail filter
// https://www.postfix.org/FILTER_README.html
func filterContent(senderEmail string, recipientEmail string) error {
	ctx := context.Background()
	rdb := contact.GetClient()
	defer rdb.Close()

	// Check if the email exists in Redis
	exists, err := rdb.Exists(ctx, senderEmail).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	if exists == 0 {
		err := rdb.Set(ctx, senderEmail, contact.StatusPending, 6*time.Hour).Err()
		if err != nil {
			return fmt.Errorf("failed to create redis entry: %w", err)
		}

		err = sendConfirmationEmail(senderEmail, recipientEmail, "http://app/"+senderEmail)
		if err != nil {
			return fmt.Errorf("failed to send captcha challenge by mail: %w", err)
		}
	}

	return nil
}

func sendTmpFile(tmpFile *os.File, from string) error {
	args := append([]string{"-G", "-i", "-f", from}, os.Args[3:]...)
	cmd := exec.Command("/usr/sbin/sendmail", args...)
	var err error
	cmd.Stdin, err = os.Open(tmpFile.Name())
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running sendmail: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 4 || os.Args[1] != "-f" {
		fmt.Fprintln(os.Stderr, "Usage: script -f sender recipients...")
		os.Exit(EX_USAGE)
	}

	if err := os.Chdir(inspectDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s does not exist\n", inspectDir)
		os.Exit(EX_TEMPFAIL)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp(inspectDir, "in.")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create temporary file")
		os.Exit(EX_TEMPFAIL)
	}
	defer os.Remove(tmpFile.Name())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(tmpFile.Name())
		os.Exit(1)
	}()

	// Copy standard input to the temporary file
	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot save mail to file")
		os.Exit(EX_TEMPFAIL)
	}

	tmpFile.Close()

	from := os.Args[2]
	recipient := os.Args[3]

	// Ici, vous pouvez ajouter votre logique de filtrage personnalisée
	// Par exemple :
	if err := filterContent(from, recipient); err != nil {
		log.Println(err)
		fmt.Fprintf(os.Stderr, "Message rejected: %s\n", err)
		os.Exit(EX_TEMPFAIL)
	}

	if err := sendTmpFile(tmpFile, from); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(EX_TEMPFAIL)
	}

	os.Exit(EX_OK)
}

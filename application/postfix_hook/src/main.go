package main

// receive mail from postfix
// if contact is approved in directory, forward the mail
// if contact does not exists or is not approved, send a challenge and postpone the mail in a queue

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/polo2ro/mailinwhite/libs/common"
	"github.com/redis/go-redis/v9"
)

const (
	// Postfix exit codes
	EX_OK          = 0  // Successful termination
	EX_TEMPFAIL    = 75 // Temporary failure
	EX_UNAVAILABLE = 69 // Service unavailable
	EX_USAGE       = 64 // Command line usage error
)

// postfix mail filter
// https://www.postfix.org/FILTER_README.html
func getSenderAddressStatus(senderEmail string) (int, error) {
	ctx := context.Background()
	rdb := common.GetAddressesClient()
	defer rdb.Close()

	// Check if the email exists in Redis
	statusStr, err := rdb.Get(ctx, senderEmail).Result()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("redis error: %w", err)
	}

	var status int
	if err == redis.Nil {
		// Email doesn't exist, create new entry with StatusPending
		err = rdb.Set(ctx, senderEmail, common.StatusPending, 6*time.Hour).Err()
		if err != nil {
			return 0, fmt.Errorf("failed to create redis entry: %w", err)
		}

		status = common.StatusPending
	} else {
		status, err = strconv.Atoi(statusStr)
		if err != nil {
			return 0, fmt.Errorf("invalid status format: %w", err)
		}
	}

	return status, nil
}

func storeMessageInRedis(ctx context.Context, messageID string, from string, to []string, messageContent []byte) error {
	rdb := common.GetMessagesClient()
	defer rdb.Close()

	messageData := common.MessageData{
		Content: messageContent,
		From:    from,
		To:      to,
	}

	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	_, err = rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if err := pipe.Set(ctx, messageID, jsonData, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to store message in Redis: %w", err)
		}

		senderKey := fmt.Sprintf("sender:%s", from)
		if err := pipe.SAdd(ctx, senderKey, messageID).Err(); err != nil {
			return fmt.Errorf("failed to add message to sender's set: %w", err)
		}

		if err := pipe.Expire(ctx, senderKey, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to set expiration for sender's set: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to execute Redis transaction: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 4 || os.Args[1] != "-f" {
		fmt.Fprintln(os.Stderr, "Usage: script -f sender recipients...")
		os.Exit(EX_USAGE)
	}

	from := os.Args[2]
	recipients := os.Args[3:]

	ctx := context.Background()

	// Generate a unique message ID
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	// Read the entire message content
	messageContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot read message content")
		os.Exit(EX_TEMPFAIL)
	}

	// Store the message in Redis
	if err := storeMessageInRedis(ctx, messageID, from, recipients, messageContent); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(EX_TEMPFAIL)
	}

	addressStatus, err := getSenderAddressStatus(from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Message rejected: %s\n", err)
		os.Exit(EX_TEMPFAIL)
	}

	if addressStatus == common.StatusPending {
		err = sendChallengeRequestEmail(from, recipients, "http://localhost:8080/app/challenge/"+url.QueryEscape(from))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to send captcha challenge by mail: %s", err)
			os.Exit(EX_TEMPFAIL)
		}

		fmt.Fprintf(os.Stderr, "Sender address %s is pending\n", from)
		os.Exit(EX_OK)
	}

	if err := sendMessage(ctx, messageID); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(EX_TEMPFAIL)
	}

	os.Exit(EX_OK)
}

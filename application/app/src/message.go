package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"github.com/polo2ro/mailinwhite/libs/common"
	"github.com/redis/go-redis/v9"
)

func getSmtpAuth(smtpHost string) smtp.Auth {
	smtpLogin := os.Getenv("SMTP_LOGIN")

	if smtpLogin != "" {
		return smtp.PlainAuth("", smtpLogin, os.Getenv("SMTP_PASSWORD"), smtpHost)
	}

	return nil
}

func sendMessage(ctx context.Context, messageID string) error {
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("port int conversion: %w", err)
	}

	smtpHost := os.Getenv("SMTP_HOST")

	if smtpHost == "" || smtpPort == 0 {
		return fmt.Errorf("missing required SMTP configuration environment variables")
	}

	rdb := common.GetMessagesClient()
	defer rdb.Close()

	jsonData, err := rdb.Get(ctx, messageID).Bytes()
	if err != nil {
		return fmt.Errorf("failed to retrieve message from Redis: %w", err)
	}

	var messageData common.MessageData
	if err := json.Unmarshal(jsonData, &messageData); err != nil {
		return fmt.Errorf("failed to unmarshal message data: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	smtpErr := smtp.SendMail(addr, getSmtpAuth(smtpHost), messageData.From, common.GetValidRecipients(messageData.To), messageData.Content)

	if smtpErr != nil {
		return fmt.Errorf("failed to send confirmation email: %v", smtpErr)
	}

	return nil
}

func sendPendingMails(ctx context.Context, senderEmail string) error {
	rdb := common.GetMessagesClient()
	defer rdb.Close()

	senderKey := fmt.Sprintf("sender:%s", senderEmail)
	messageIDs, err := rdb.SMembers(ctx, senderKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get message IDs for sender %s: %w", senderEmail, err)
	}

	for _, messageID := range messageIDs {
		if err := sendMessage(ctx, messageID); err != nil {
			return fmt.Errorf("error sending message %s: %w", messageID, err)
		}

		if err := rdb.SRem(ctx, senderKey, messageID).Err(); err != nil {
			return fmt.Errorf("error removing message ID %s from sender set: %v", messageID, err)
		}
	}

	return nil
}

func getMailStatus(mail string) (int, int, error) {
	rdb := common.GetAddressesClient()
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

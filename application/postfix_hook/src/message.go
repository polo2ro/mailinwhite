package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/polo2ro/mailinwhite/libs/common"
)

func sendMessage(ctx context.Context, messageID string) error {
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

	args := append([]string{"-G", "-i", "-f", messageData.From}, messageData.To...)
	cmd := exec.Command("/usr/sbin/sendmail", args...)
	cmd.Stdin = bytes.NewReader(messageData.Content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running sendmail: %w", err)
	}

	// Delete the message from Redis after sending
	if err := rdb.Del(ctx, messageID).Err(); err != nil {
		return fmt.Errorf("warning: failed to delete message %s from Redis: %v", messageID, err)
	}

	return nil
}

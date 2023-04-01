package telegramnotifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"codeberg.org/cordalace/telepod/internal/workflow"
)

func NewTelegramNotifier(httpClient *http.Client, chatID string, botToken string) *TelegramNotifier {
	return &TelegramNotifier{httpClient: httpClient, chatID: chatID, botToken: botToken}
}

type TelegramNotifier struct {
	httpClient *http.Client
	chatID     string
	botToken   string
}

func (n *TelegramNotifier) CreateNotification(ctx context.Context, container *workflow.Container) error {
	data, err := json.Marshal(struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: n.chatID,
		Text:   fmt.Sprintf("%v was updated to %v", container.Name, container.ImageVersion),
	})
	if err != nil {
		return fmt.Errorf("error encoding telegram sendMessage json: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", n.botToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error building telegram request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making telegram sendMessage request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return &badStatusCodeError{StatusCode: resp.StatusCode}
	}

	return nil
}

type badStatusCodeError struct {
	StatusCode int
}

func (e *badStatusCodeError) Error() string {
	return fmt.Sprintf("telegram bad status code: %v", e.StatusCode)
}

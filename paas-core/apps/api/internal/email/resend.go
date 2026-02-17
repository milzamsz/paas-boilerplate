package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const resendAPIURL = "https://api.resend.com/emails"

// ResendProvider implements the email.Service interface using Resend.
// Docs: https://resend.com/docs/send-with-go
type ResendProvider struct {
	apiKey     string
	fromEmail  string // e.g. "PaaS <noreply@example.com>"
	httpClient *http.Client
}

// NewResendProvider creates a Resend email provider.
func NewResendProvider(apiKey, fromEmail string) *ResendProvider {
	return &ResendProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// resendPayload matches the Resend API request body.
type resendPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html,omitempty"`
	Text    string `json:"text,omitempty"`
}

// Send sends an email via the Resend API.
func (r *ResendProvider) Send(ctx context.Context, msg Message) error {
	payload := resendPayload{
		From:    r.fromEmail,
		To:      msg.To,
		Subject: msg.Subject,
		HTML:    msg.HTMLBody,
		Text:    msg.TextBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("Resend API error",
			"status", resp.StatusCode,
			"body", string(respBody),
			"to", msg.To,
			"subject", msg.Subject,
		)
		return fmt.Errorf("email: resend API returned %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("Email sent via Resend", "to", msg.To, "subject", msg.Subject)
	return nil
}

package email

import "context"

// Message represents an email to send.
type Message struct {
	To       string // recipient email address
	Subject  string
	HTMLBody string
	TextBody string // optional plain-text fallback
}

// Service defines a provider-agnostic email interface.
// Implementations: Resend (default), SMTP, SES, etc.
//
// Design follows the Goilerplate pattern — interface-based,
// easily swappable with zero handler changes.
type Service interface {
	// Send sends a single email message.
	Send(ctx context.Context, msg Message) error
}

// TemplateData holds common variables for email templates.
type TemplateData struct {
	AppName   string
	AppURL    string
	UserName  string
	UserEmail string
	Token     string
	Link      string
	ExpiresIn string // human-readable, e.g. "15 minutes"
}

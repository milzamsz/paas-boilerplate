package billing

import (
	"time"

	"github.com/google/uuid"
)

// CreateSubscriptionRequest is the DTO for creating/upgrading a subscription.
type CreateSubscriptionRequest struct {
	PlanID       uuid.UUID `json:"plan_id" binding:"required"`
	BillingCycle string    `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
}

// CancelSubscriptionRequest is the DTO for cancelling a subscription.
type CancelSubscriptionRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// PlanResponse is the public billing plan representation.
type PlanResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	PriceMonthly   int64     `json:"price_monthly"`
	PriceYearly    int64     `json:"price_yearly"`
	Currency       string    `json:"currency"`
	MaxProjects    int       `json:"max_projects"`
	MaxDeployments int       `json:"max_deployments"`
	MaxMembers     int       `json:"max_members"`
	Features       string    `json:"features,omitempty"`
	IsActive       bool      `json:"is_active"`
}

// SubscriptionResponse is the public subscription representation.
type SubscriptionResponse struct {
	ID                 uuid.UUID     `json:"id"`
	OrgID              uuid.UUID     `json:"org_id"`
	Plan               *PlanResponse `json:"plan,omitempty"`
	Status             string        `json:"status"`
	BillingCycle       string        `json:"billing_cycle"`
	CurrentPeriodStart time.Time     `json:"current_period_start"`
	CurrentPeriodEnd   time.Time     `json:"current_period_end"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
}

// InvoiceResponse is the public invoice representation.
type InvoiceResponse struct {
	ID               uuid.UUID  `json:"id"`
	OrgID            uuid.UUID  `json:"org_id"`
	SubscriptionID   uuid.UUID  `json:"subscription_id"`
	Amount           int64      `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	DueDate          time.Time  `json:"due_date"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	XenditPaymentURL string     `json:"xendit_payment_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// BillingOverview is a summary of the org's billing state.
type BillingOverview struct {
	Subscription *SubscriptionResponse `json:"subscription"`
	Usage        *UsageResponse        `json:"usage"`
}

// UsageResponse summarizes current resource usage vs limits.
type UsageResponse struct {
	ProjectsUsed    int `json:"projects_used"`
	ProjectsLimit   int `json:"projects_limit"`
	DeploymentsUsed int `json:"deployments_used"`
	DeploymentsLimit int `json:"deployments_limit"`
	MembersUsed     int `json:"members_used"`
	MembersLimit    int `json:"members_limit"`
}

// XenditWebhookPayload represents the incoming Xendit callback.
type XenditWebhookPayload struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
	Amount     int64  `json:"amount"`
	PaidAt     string `json:"paid_at,omitempty"`
}

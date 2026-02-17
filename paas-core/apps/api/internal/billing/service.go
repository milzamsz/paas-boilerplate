package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

// Service defines the billing service interface.
type Service interface {
	ListPlans(ctx context.Context) ([]PlanResponse, error)
	GetBillingOverview(ctx context.Context, orgID uuid.UUID) (*BillingOverview, error)

	CreateSubscription(ctx context.Context, orgID uuid.UUID, req CreateSubscriptionRequest) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, orgID uuid.UUID) error

	ListInvoices(ctx context.Context, orgID uuid.UUID) ([]InvoiceResponse, error)
	GetUsage(ctx context.Context, orgID uuid.UUID) (*UsageResponse, error)

	// Webhook processing
	ProcessXenditWebhook(ctx context.Context, payload XenditWebhookPayload) error
}

type service struct {
	repo Repository
}

// NewService creates a new billing service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// --- Plans ---

func (s *service) ListPlans(ctx context.Context) ([]PlanResponse, error) {
	plans, err := s.repo.ListActivePlans(ctx)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []PlanResponse
	for _, p := range plans {
		responses = append(responses, *toPlanResponse(&p))
	}
	return responses, nil
}

// --- Billing overview ---

func (s *service) GetBillingOverview(ctx context.Context, orgID uuid.UUID) (*BillingOverview, error) {
	sub, err := s.repo.FindActiveSubscription(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	usage, err := s.GetUsage(ctx, orgID)
	if err != nil {
		return nil, err
	}

	overview := &BillingOverview{Usage: usage}
	if sub != nil {
		overview.Subscription = toSubscriptionResponse(sub)
	}
	return overview, nil
}

// --- Subscriptions ---

func (s *service) CreateSubscription(ctx context.Context, orgID uuid.UUID, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	// Check for existing active subscription
	existing, err := s.repo.FindActiveSubscription(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if existing != nil {
		return nil, apiErrors.Conflict("Organization already has an active subscription")
	}

	plan, err := s.repo.FindPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	if plan == nil {
		return nil, apiErrors.NotFound("Billing plan not found")
	}

	now := time.Now()
	var periodEnd time.Time
	switch req.BillingCycle {
	case "monthly":
		periodEnd = now.AddDate(0, 1, 0)
	case "yearly":
		periodEnd = now.AddDate(1, 0, 0)
	default:
		return nil, apiErrors.BadRequest("Invalid billing cycle")
	}

	sub := &model.Subscription{
		OrgID:              orgID,
		PlanID:             plan.ID,
		Status:             "active",
		BillingCycle:       req.BillingCycle,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	// Create initial invoice
	amount := plan.PriceMonthly
	if req.BillingCycle == "yearly" {
		amount = plan.PriceYearly
	}

	invoice := &model.Invoice{
		OrgID:          orgID,
		SubscriptionID: sub.ID,
		Amount:         amount,
		Currency:       plan.Currency,
		Status:         "pending",
		DueDate:        now.AddDate(0, 0, 7), // 7 days to pay
	}
	if err := s.repo.CreateInvoice(ctx, invoice); err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	// [TODO] Call Xendit API to create invoice and get payment URL
	// For now, the invoice is created locally and Xendit integration is deferred.

	sub.Plan = *plan
	return toSubscriptionResponse(sub), nil
}

func (s *service) CancelSubscription(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.repo.FindActiveSubscription(ctx, orgID)
	if err != nil {
		return apiErrors.InternalServerError(err)
	}
	if sub == nil {
		return apiErrors.NotFound("No active subscription found")
	}

	now := time.Now()
	sub.Status = "cancelled"
	sub.CancelledAt = &now

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return apiErrors.InternalServerError(err)
	}

	// [TODO] Cancel subscription on Xendit side

	return nil
}

// --- Invoices ---

func (s *service) ListInvoices(ctx context.Context, orgID uuid.UUID) ([]InvoiceResponse, error) {
	invoices, err := s.repo.ListInvoicesByOrg(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	var responses []InvoiceResponse
	for _, inv := range invoices {
		responses = append(responses, *toInvoiceResponse(&inv))
	}
	return responses, nil
}

// --- Usage ---

func (s *service) GetUsage(ctx context.Context, orgID uuid.UUID) (*UsageResponse, error) {
	sub, err := s.repo.FindActiveSubscription(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	// Default limits for free tier
	limits := struct{ projects, deployments, members int }{1, 5, 1}
	if sub != nil {
		limits.projects = sub.Plan.MaxProjects
		limits.deployments = sub.Plan.MaxDeployments
		limits.members = sub.Plan.MaxMembers
	}

	projectCount, err := s.repo.CountProjectsByOrg(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	deploymentCount, err := s.repo.CountDeploymentsByOrg(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}
	memberCount, err := s.repo.CountMembersByOrg(ctx, orgID)
	if err != nil {
		return nil, apiErrors.InternalServerError(err)
	}

	return &UsageResponse{
		ProjectsUsed:     projectCount,
		ProjectsLimit:    limits.projects,
		DeploymentsUsed:  deploymentCount,
		DeploymentsLimit: limits.deployments,
		MembersUsed:      memberCount,
		MembersLimit:     limits.members,
	}, nil
}

// --- Xendit Webhook ---

func (s *service) ProcessXenditWebhook(ctx context.Context, payload XenditWebhookPayload) error {
	invoice, err := s.repo.FindInvoiceByXenditID(ctx, payload.ID)
	if err != nil {
		return apiErrors.InternalServerError(err)
	}
	if invoice == nil {
		return apiErrors.NotFound(fmt.Sprintf("Invoice not found for Xendit ID: %s", payload.ID))
	}

	switch payload.Status {
	case "PAID", "SETTLED":
		now := time.Now()
		invoice.Status = "paid"
		invoice.PaidAt = &now
	case "EXPIRED":
		invoice.Status = "failed"
	default:
		// Unknown status, just update as-is
		invoice.Status = payload.Status
	}

	if err := s.repo.UpdateInvoice(ctx, invoice); err != nil {
		return apiErrors.InternalServerError(err)
	}

	return nil
}

// --- Helpers ---

func toPlanResponse(p *model.BillingPlan) *PlanResponse {
	return &PlanResponse{
		ID:             p.ID,
		Name:           p.Name,
		Slug:           p.Slug,
		PriceMonthly:   p.PriceMonthly,
		PriceYearly:    p.PriceYearly,
		Currency:       p.Currency,
		MaxProjects:    p.MaxProjects,
		MaxDeployments: p.MaxDeployments,
		MaxMembers:     p.MaxMembers,
		Features:       p.Features,
		IsActive:       p.IsActive,
	}
}

func toSubscriptionResponse(sub *model.Subscription) *SubscriptionResponse {
	resp := &SubscriptionResponse{
		ID:                 sub.ID,
		OrgID:              sub.OrgID,
		Status:             sub.Status,
		BillingCycle:       sub.BillingCycle,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelledAt:        sub.CancelledAt,
		CreatedAt:          sub.CreatedAt,
	}
	if sub.Plan.ID != (uuid.UUID{}) {
		resp.Plan = toPlanResponse(&sub.Plan)
	}
	return resp
}

func toInvoiceResponse(inv *model.Invoice) *InvoiceResponse {
	return &InvoiceResponse{
		ID:               inv.ID,
		OrgID:            inv.OrgID,
		SubscriptionID:   inv.SubscriptionID,
		Amount:           inv.Amount,
		Currency:         inv.Currency,
		Status:           inv.Status,
		DueDate:          inv.DueDate,
		PaidAt:           inv.PaidAt,
		XenditPaymentURL: inv.XenditPaymentURL,
		CreatedAt:        inv.CreatedAt,
	}
}

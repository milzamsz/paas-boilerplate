package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
)

// Handler handles billing-related HTTP requests.
type Handler struct {
	billingService Service
	webhookSecret  string // Xendit webhook verification token
}

// NewHandler creates a new billing handler.
func NewHandler(billingService Service, webhookSecret string) *Handler {
	return &Handler{
		billingService: billingService,
		webhookSecret:  webhookSecret,
	}
}

// --- Plans ---

// ListPlans godoc
// @Summary List available billing plans
// @Tags billing
// @Success 200 {object} errors.Response{data=[]PlanResponse}
// @Router /api/v1/billing/plans [get]
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.billingService.ListPlans(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(plans))
}

// --- Billing Overview ---

// GetBillingOverview godoc
// @Summary Get billing overview for an organization
// @Tags billing
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=BillingOverview}
// @Router /api/v1/orgs/{orgId}/billing [get]
func (h *Handler) GetBillingOverview(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	overview, err := h.billingService.GetBillingOverview(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(overview))
}

// --- Subscriptions ---

// CreateSubscription godoc
// @Summary Create a subscription for an organization
// @Tags billing
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body CreateSubscriptionRequest true "Subscription request"
// @Success 201 {object} errors.Response{data=SubscriptionResponse}
// @Router /api/v1/orgs/{orgId}/billing/subscribe [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	sub, err := h.billingService.CreateSubscription(c.Request.Context(), orgID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, apiErrors.Success(sub))
}

// CancelSubscription godoc
// @Summary Cancel subscription for an organization
// @Tags billing
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response
// @Router /api/v1/orgs/{orgId}/billing/cancel [post]
func (h *Handler) CancelSubscription(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	if err := h.billingService.CancelSubscription(c.Request.Context(), orgID); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Subscription cancelled"}))
}

// --- Invoices ---

// ListInvoices godoc
// @Summary List invoices for an organization
// @Tags billing
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=[]InvoiceResponse}
// @Router /api/v1/orgs/{orgId}/billing/invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	invoices, err := h.billingService.ListInvoices(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(invoices))
}

// --- Usage ---

// GetUsage godoc
// @Summary Get resource usage for an organization
// @Tags billing
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} errors.Response{data=UsageResponse}
// @Router /api/v1/orgs/{orgId}/billing/usage [get]
func (h *Handler) GetUsage(c *gin.Context) {
	orgID := c.MustGet("org_id").(uuid.UUID)

	usage, err := h.billingService.GetUsage(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(usage))
}

// --- Xendit Webhook ---

// XenditWebhook godoc
// @Summary Handle Xendit payment webhook
// @Tags webhooks
// @Param x-callback-token header string true "Xendit callback verification token"
// @Success 200 {object} errors.Response
// @Router /api/v1/webhooks/xendit [post]
func (h *Handler) XenditWebhook(c *gin.Context) {
	// Verify callback token
	callbackToken := c.GetHeader("x-callback-token")
	if h.webhookSecret != "" && callbackToken != h.webhookSecret {
		// Also support HMAC signature verification
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			_ = c.Error(apiErrors.BadRequest("Failed to read request body"))
			return
		}
		sig := c.GetHeader("x-callback-signature")
		if !verifyHMAC(body, sig, h.webhookSecret) {
			_ = c.Error(apiErrors.Unauthorized("Invalid webhook signature"))
			return
		}

		// Re-bind from the body we already read
		var payload XenditWebhookPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			_ = c.Error(apiErrors.BadRequest("Invalid webhook payload"))
			return
		}

		if err := h.billingService.ProcessXenditWebhook(c.Request.Context(), payload); err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Webhook processed"}))
		return
	}

	// Simple token verification passed
	var payload XenditWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid webhook payload"))
		return
	}

	if err := h.billingService.ProcessXenditWebhook(c.Request.Context(), payload); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Webhook processed"}))
}

func verifyHMAC(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

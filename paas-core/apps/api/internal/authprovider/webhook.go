package authprovider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"

	"gorm.io/gorm"
)

// WebhookPayload represents the incoming Supabase auth webhook event.
type WebhookPayload struct {
	Type   string          `json:"type"`   // e.g. "INSERT", "UPDATE", "DELETE"
	Table  string          `json:"table"`  // e.g. "users"
	Schema string          `json:"schema"` // e.g. "auth"
	Record json.RawMessage `json:"record"`
}

// SupabaseAuthUser maps to auth.users from Supabase.
type SupabaseAuthUser struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	UserMetadata map[string]interface{} `json:"raw_user_meta_data"`
	CreatedAt    string                 `json:"created_at"`
}

// WebhookHandler processes Supabase auth webhook events to sync users
// into the local users table. Register at POST /api/v1/webhooks/supabase/auth.
type WebhookHandler struct {
	db            *gorm.DB
	webhookSecret string
}

// NewWebhookHandler creates a handler for Supabase auth webhooks.
func NewWebhookHandler(db *gorm.DB, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		db:            db,
		webhookSecret: webhookSecret,
	}
}

// HandleAuthWebhook processes POST /api/v1/webhooks/supabase/auth.
func (h *WebhookHandler) HandleAuthWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("failed to read request body"))
		return
	}
	defer c.Request.Body.Close()

	// Verify webhook signature if secret is configured
	if h.webhookSecret != "" {
		sig := c.GetHeader("X-Supabase-Webhook-Signature")
		if !verifyWebhookSignature(body, sig, h.webhookSecret) {
			_ = c.Error(apiErrors.Unauthorized("invalid webhook signature"))
			return
		}
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		_ = c.Error(apiErrors.BadRequest("invalid webhook payload"))
		return
	}

	// Only process auth.users events
	if payload.Schema != "auth" || payload.Table != "users" {
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "not auth.users"})
		return
	}

	var supaUser SupabaseAuthUser
	if err := json.Unmarshal(payload.Record, &supaUser); err != nil {
		_ = c.Error(apiErrors.BadRequest("invalid user record in payload"))
		return
	}

	switch payload.Type {
	case "INSERT":
		h.handleUserInsert(c, supaUser)
	case "UPDATE":
		h.handleUserUpdate(c, supaUser)
	case "DELETE":
		h.handleUserDelete(c, supaUser)
	default:
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "unknown event type"})
	}
}

func (h *WebhookHandler) handleUserInsert(c *gin.Context, supaUser SupabaseAuthUser) {
	userID, err := uuid.Parse(supaUser.ID)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("invalid user ID"))
		return
	}

	name := ""
	if supaUser.UserMetadata != nil {
		name, _ = supaUser.UserMetadata["name"].(string)
	}
	if name == "" {
		name = supaUser.Email
	}

	user := model.User{
		Name:  name,
		Email: supaUser.Email,
	}
	user.ID = userID

	if err := h.db.Where("id = ?", userID).FirstOrCreate(&user).Error; err != nil {
		slog.Error("Failed to sync Supabase user", "error", err, "supabase_id", supaUser.ID)
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	slog.Info("Synced Supabase user",
		"user_id", userID,
		"email", supaUser.Email,
		"action", "insert",
	)

	c.JSON(http.StatusOK, gin.H{"status": "synced", "user_id": userID.String()})
}

func (h *WebhookHandler) handleUserUpdate(c *gin.Context, supaUser SupabaseAuthUser) {
	userID, err := uuid.Parse(supaUser.ID)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("invalid user ID"))
		return
	}

	updates := map[string]interface{}{
		"email": supaUser.Email,
	}
	if supaUser.UserMetadata != nil {
		if name, ok := supaUser.UserMetadata["name"].(string); ok && name != "" {
			updates["name"] = name
		}
	}

	if err := h.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		slog.Error("Failed to update Supabase user", "error", err, "supabase_id", supaUser.ID)
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	slog.Info("Updated Supabase user",
		"user_id", userID,
		"email", supaUser.Email,
		"action", "update",
	)

	c.JSON(http.StatusOK, gin.H{"status": "updated", "user_id": userID.String()})
}

func (h *WebhookHandler) handleUserDelete(c *gin.Context, supaUser SupabaseAuthUser) {
	userID, err := uuid.Parse(supaUser.ID)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("invalid user ID"))
		return
	}

	// Soft-delete (GORM default with DeletedAt field)
	if err := h.db.Where("id = ?", userID).Delete(&model.User{}).Error; err != nil {
		slog.Error("Failed to delete Supabase user", "error", err, "supabase_id", supaUser.ID)
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	slog.Info("Deleted Supabase user",
		"user_id", userID,
		"action", "delete",
	)

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "user_id": userID.String()})
}

// verifyWebhookSignature verifies the HMAC-SHA256 signature of the webhook payload.
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

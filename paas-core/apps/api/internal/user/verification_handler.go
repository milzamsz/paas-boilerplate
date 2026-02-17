package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apiErrors "paas-core/apps/api/internal/errors"
)

// VerificationHandler handles email verification and password reset HTTP requests.
type VerificationHandler struct {
	verificationService *VerificationService
}

// NewVerificationHandler creates a new verification handler.
func NewVerificationHandler(vs *VerificationService) *VerificationHandler {
	return &VerificationHandler{verificationService: vs}
}

// VerifyEmailRequest is the DTO for email verification.
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// RequestPasswordResetRequest is the DTO for requesting a password reset.
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest is the DTO for resetting a password.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=12,max=128"`
}

// ResendVerificationRequest is the DTO for resending a verification email.
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify a user's email address with the token from the verification email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification request"
// @Success 200 {object} errors.Response "Email verified"
// @Failure 400 {object} errors.Response "Invalid or expired token"
// @Router /api/v1/auth/verify-email [post]
func (h *VerificationHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	if err := h.verificationService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send a password reset email to the specified address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "Password reset request"
// @Success 200 {object} errors.Response "Reset email sent (always returns success)"
// @Router /api/v1/auth/request-reset [post]
func (h *VerificationHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	// Send email in background — always return success to not reveal email existence
	if err := h.verificationService.SendPasswordResetEmail(c.Request.Context(), req.Email); err != nil {
		// Log but don't return error to user
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{"message": "If that email is registered, a password reset link has been sent."})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset a user's password using a valid reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} errors.Response "Password reset"
// @Failure 400 {object} errors.Response "Invalid token or weak password"
// @Router /api/v1/auth/reset-password [post]
func (h *VerificationHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	if err := h.verificationService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}

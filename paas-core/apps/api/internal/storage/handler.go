package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authPkg "paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
)

// Handler handles file upload HTTP requests.
type Handler struct {
	uploadService *UploadService
}

// NewHandler creates a new upload handler.
func NewHandler(uploadService *UploadService) *Handler {
	return &Handler{uploadService: uploadService}
}

// UploadUserAvatar godoc
// @Summary Upload user avatar
// @Description Upload an avatar image for the current user (max 5MB, JPEG/PNG/GIF/WebP)
// @Tags users
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formance file true "Avatar image file"
// @Success 200 {object} errors.Response "Avatar URL"
// @Failure 400 {object} errors.Response "Invalid file"
// @Router /api/v1/users/me/avatar [post]
func (h *Handler) UploadUserAvatar(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	authClaims := claims.(*authPkg.Claims)

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("No avatar file provided"))
		return
	}
	defer file.Close()

	url, err := h.uploadService.UploadAvatar(
		c.Request.Context(),
		authClaims.UserID,
		"user",
		file,
		header,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{
		"avatar_url": url,
	}))
}

// UploadOrgAvatar godoc
// @Summary Upload org avatar/logo
// @Description Upload an avatar image for an organization (max 5MB, JPEG/PNG/GIF/WebP)
// @Tags organizations
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param avatar formance file true "Avatar image file"
// @Success 200 {object} errors.Response "Logo URL"
// @Failure 400 {object} errors.Response "Invalid file"
// @Router /api/v1/orgs/{orgId}/avatar [post]
func (h *Handler) UploadOrgAvatar(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid organization ID"))
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("No avatar file provided"))
		return
	}
	defer file.Close()

	url, err := h.uploadService.UploadAvatar(
		c.Request.Context(),
		orgID,
		"org",
		file,
		header,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{
		"logo_url": url,
	}))
}

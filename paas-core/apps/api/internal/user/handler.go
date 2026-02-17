package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	authPkg "paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	userService Service
}

// NewHandler creates a new user handler.
func NewHandler(userService Service) *Handler {
	return &Handler{userService: userService}
}

// GetMe godoc
// @Summary Get current user profile
// @Tags users
// @Security BearerAuth
// @Success 200 {object} errors.Response{data=auth.UserResponse}
// @Router /api/v1/users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	authClaims := claims.(*authPkg.Claims)

	user, err := h.userService.GetUserByID(c.Request.Context(), authClaims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(user))
}

// UpdateMe godoc
// @Summary Update current user profile
// @Tags users
// @Security BearerAuth
// @Param request body UpdateUserRequest true "Update request"
// @Success 200 {object} errors.Response{data=auth.UserResponse}
// @Router /api/v1/users/me [put]
func (h *Handler) UpdateMe(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized(""))
		return
	}
	authClaims := claims.(*authPkg.Claims)

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), authClaims.UserID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(user))
}

// ListUsers godoc
// @Summary List all users (admin only)
// @Tags users
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param search query string false "Search by name or email"
// @Param sort query string false "Sort field" Enums(name, email, created_at)
// @Param order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} errors.Response{data=[]auth.UserResponse}
// @Router /api/v1/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filters := FilterParams{
		Search: c.Query("search"),
		Sort:   c.DefaultQuery("sort", "created_at"),
		Order:  c.DefaultQuery("order", "desc"),
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), filters, page, perPage)
	if err != nil {
		_ = c.Error(err)
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, apiErrors.SuccessWithMeta(users, &apiErrors.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}))
}

package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"
)

// OrgResolver extracts the orgId from the URL, verifies membership,
// and stores the membership info in the Gin context.
func OrgResolver(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDStr := c.Param("orgId")
		if orgIDStr == "" {
			_ = c.Error(apiErrors.BadRequest("Organization ID is required"))
			c.Abort()
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			_ = c.Error(apiErrors.BadRequest("Invalid organization ID"))
			c.Abort()
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			_ = c.Error(apiErrors.Unauthorized(""))
			c.Abort()
			return
		}
		userID := userIDVal.(uuid.UUID)

		var membership model.Membership
		result := db.Where("org_id = ? AND user_id = ?", orgID, userID).First(&membership)
		if result.Error != nil {
			_ = c.Error(apiErrors.Forbidden("You are not a member of this organization"))
			c.Abort()
			return
		}

		c.Set("org_id", orgID)
		c.Set("membership", membership)
		c.Set("org_role", membership.Role)
		c.Next()
	}
}

// RequireOrgRole checks that the user has at least the required role within the org.
func RequireOrgRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := c.Get("org_role")
		if !exists {
			_ = c.Error(apiErrors.Forbidden("Org role not resolved"))
			c.Abort()
			return
		}
		userRole := roleStr.(string)

		if !model.HasPermission(userRole, requiredRole) {
			_ = c.Error(apiErrors.Forbidden(fmt.Sprintf("Requires %s role or higher", requiredRole)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMaxAge helper to format max-age header.
func FormatMaxAge(maxAge int) string {
	return strconv.Itoa(maxAge)
}

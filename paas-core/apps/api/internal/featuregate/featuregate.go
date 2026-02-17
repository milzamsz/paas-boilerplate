package featuregate

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/model"

	"gorm.io/gorm"
)

// GateService checks subscription quotas and feature flags.
type GateService struct {
	db *gorm.DB
}

// NewGateService creates a new feature gate service.
func NewGateService(db *gorm.DB) *GateService {
	return &GateService{db: db}
}

// PlanLimits holds resolved limits for an org.
type PlanLimits struct {
	MaxProjects    int
	MaxDeployments int
	MaxMembers     int
	Features       []string
}

// GetPlanLimits resolves the active plan limits for an org.
// Falls back to free-tier defaults if no subscription exists.
func (g *GateService) GetPlanLimits(orgID uuid.UUID) (*PlanLimits, error) {
	var sub model.Subscription
	err := g.db.
		Preload("Plan").
		Where("org_id = ? AND status IN ?", orgID, []string{"active", "trialing"}).
		Order("created_at DESC").
		First(&sub).Error

	if err != nil {
		// No subscription → free tier
		return &PlanLimits{
			MaxProjects:    FreeTierLimits.MaxProjects,
			MaxDeployments: FreeTierLimits.MaxDeployments,
			MaxMembers:     FreeTierLimits.MaxMembers,
			Features:       nil,
		}, nil
	}

	return &PlanLimits{
		MaxProjects:    sub.Plan.MaxProjects,
		MaxDeployments: sub.Plan.MaxDeployments,
		MaxMembers:     sub.Plan.MaxMembers,
		Features:       UnmarshalFeatures(sub.Plan.Features),
	}, nil
}

// CheckQuota verifies that the org hasn't exceeded its plan limit for the given resource.
// Returns nil if within limits, or a 402 error if quota exceeded.
func (g *GateService) CheckQuota(orgID uuid.UUID, resource string) error {
	limits, err := g.GetPlanLimits(orgID)
	if err != nil {
		return apiErrors.InternalServerError(err)
	}

	var current int
	var max int

	switch resource {
	case "projects":
		max = limits.MaxProjects
		var count int64
		if err := g.db.Model(&model.Project{}).Where("org_id = ?", orgID).Count(&count).Error; err != nil {
			return apiErrors.InternalServerError(err)
		}
		current = int(count)

	case "deployments":
		max = limits.MaxDeployments
		var count int64
		if err := g.db.Model(&model.Deployment{}).
			Joins("JOIN projects ON projects.id = deployments.project_id").
			Where("projects.org_id = ? AND deployments.status = ?", orgID, "running").
			Count(&count).Error; err != nil {
			return apiErrors.InternalServerError(err)
		}
		current = int(count)

	case "members":
		max = limits.MaxMembers
		var count int64
		if err := g.db.Model(&model.Membership{}).Where("org_id = ?", orgID).Count(&count).Error; err != nil {
			return apiErrors.InternalServerError(err)
		}
		current = int(count)

	default:
		return apiErrors.InternalServerError(fmt.Errorf("unknown resource: %s", resource))
	}

	// -1 means unlimited
	if max == -1 {
		return nil
	}

	if current >= max {
		return &apiErrors.APIError{
			StatusCode: http.StatusPaymentRequired,
			Code:       "upgrade_required",
			Message: fmt.Sprintf(
				"You have reached the maximum number of %s (%d) for your current plan. Please upgrade to add more.",
				resource, max,
			),
		}
	}

	return nil
}

// HasFeature checks if the org's plan includes a specific feature flag.
func (g *GateService) HasFeature(orgID uuid.UUID, featureName string) (bool, error) {
	limits, err := g.GetPlanLimits(orgID)
	if err != nil {
		return false, err
	}

	for _, f := range limits.Features {
		if f == featureName {
			return true, nil
		}
	}
	return false, nil
}

// --- Gin Middleware ---

// RequireQuota returns a Gin middleware that checks quota limits before allowing
// resource creation. Must be used AFTER OrgResolver middleware.
func RequireQuota(gate *GateService, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDVal, exists := c.Get("org_id")
		if !exists {
			_ = c.Error(apiErrors.BadRequest("Org context not resolved"))
			c.Abort()
			return
		}
		orgID := orgIDVal.(uuid.UUID)

		if err := gate.CheckQuota(orgID, resource); err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireFeature returns a Gin middleware that checks if the org's plan includes
// a specific feature. Returns 402 if the feature is not available.
func RequireFeature(gate *GateService, featureName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDVal, exists := c.Get("org_id")
		if !exists {
			_ = c.Error(apiErrors.BadRequest("Org context not resolved"))
			c.Abort()
			return
		}
		orgID := orgIDVal.(uuid.UUID)

		has, err := gate.HasFeature(orgID, featureName)
		if err != nil {
			_ = c.Error(apiErrors.InternalServerError(err))
			c.Abort()
			return
		}

		if !has {
			_ = c.Error(&apiErrors.APIError{
				StatusCode: http.StatusPaymentRequired,
				Code:       "feature_not_available",
				Message: fmt.Sprintf(
					"The '%s' feature is not available on your current plan. Please upgrade to access this feature.",
					featureName,
				),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

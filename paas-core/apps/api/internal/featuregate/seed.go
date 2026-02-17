package featuregate

import (
	"log/slog"

	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// SeedDefaultPlans creates the default billing plans if they don't exist.
// This is idempotent — existing plans are left untouched.
func SeedDefaultPlans(db *gorm.DB) {
	for _, def := range DefaultPlans() {
		var existing model.BillingPlan
		result := db.Where("slug = ?", def.Slug).First(&existing)
		if result.Error == nil {
			// Plan already exists, skip
			slog.Debug("Billing plan already exists, skipping", "slug", def.Slug)
			continue
		}

		plan := model.BillingPlan{
			Name:           def.Name,
			Slug:           def.Slug,
			PriceMonthly:   def.PriceMonthly,
			PriceYearly:    def.PriceYearly,
			Currency:       def.Currency,
			MaxProjects:    def.MaxProjects,
			MaxDeployments: def.MaxDeployments,
			MaxMembers:     def.MaxMembers,
			Features:       MarshalFeatures(def.Features),
			IsActive:       true,
		}

		if err := db.Create(&plan).Error; err != nil {
			slog.Error("Failed to seed billing plan", "slug", def.Slug, "error", err)
		} else {
			slog.Info("Seeded billing plan", "slug", def.Slug, "name", def.Name)
		}
	}
}

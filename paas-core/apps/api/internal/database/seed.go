package database

import (
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"paas-core/apps/api/internal/model"
)

// SeedDevUsers creates default dev/test users if they don't already exist.
// Only call this in non-production environments.
func SeedDevUsers(db *gorm.DB) {
	// Ensure required roles exist
	seedRoles(db)

	type seedUser struct {
		Name     string
		Email    string
		Password string
		Roles    []string
	}

	users := []seedUser{
		{Name: "Admin", Email: "admin@paas.local", Password: "admin", Roles: []string{"super_admin", "user"}},
		{Name: "Demo User", Email: "demo@paas.local", Password: "demo", Roles: []string{"user"}},
	}

	for _, su := range users {
		var existing model.User
		if err := db.Where("email = ?", su.Email).First(&existing).Error; err == nil {
			slog.Debug("Dev user already exists, skipping", "email", su.Email)
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(su.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Failed to hash password for dev user", "email", su.Email, "error", err)
			continue
		}

		user := model.User{
			Name:          su.Name,
			Email:         su.Email,
			PasswordHash:  string(hash),
			EmailVerified: true,
		}

		if err := db.Create(&user).Error; err != nil {
			slog.Error("Failed to seed dev user", "email", su.Email, "error", err)
			continue
		}

		for _, roleName := range su.Roles {
			var role model.Role
			if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
				slog.Error("Role not found for dev user", "role", roleName, "error", err)
				continue
			}
			userRole := model.UserRole{UserID: user.ID, RoleID: role.ID}
			if err := db.Create(&userRole).Error; err != nil {
				slog.Error("Failed to assign role to dev user", "email", su.Email, "role", roleName, "error", err)
			}
		}

		slog.Info("Seeded dev user", "email", su.Email, "name", su.Name)
	}
}

func seedRoles(db *gorm.DB) {
	roles := []string{"super_admin", "admin", "user"}
	for _, name := range roles {
		var existing model.Role
		if err := db.Where("name = ?", name).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&model.Role{Name: name}).Error; err != nil {
			slog.Error("Failed to seed role", "name", name, "error", err)
		} else {
			slog.Info("Seeded role", "name", name)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"paas-core/apps/api/internal/auth"
	"paas-core/apps/api/internal/authprovider"
	"paas-core/apps/api/internal/billing"
	"paas-core/apps/api/internal/config"
	"paas-core/apps/api/internal/database"
	"paas-core/apps/api/internal/email"
	apiErrors "paas-core/apps/api/internal/errors"
	"paas-core/apps/api/internal/featuregate"
	"paas-core/apps/api/internal/middleware"
	"paas-core/apps/api/internal/model"
	"paas-core/apps/api/internal/oauth"
	"paas-core/apps/api/internal/org"
	"paas-core/apps/api/internal/project"
	"paas-core/apps/api/internal/storage"
	"paas-core/apps/api/internal/user"
)

func main() {
	// --- 1. Config ---
	cfg, err := config.LoadConfig("")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// --- 2. Logger ---
	logLevel := cfg.Logging.GetLogLevel()
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	cfg.LogSafeConfig(logger)

	// --- 3. Database ---
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// --- 3a. Auto-migrate new models (safe: only adds missing tables/columns) ---
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.UserRole{},
		&model.RefreshToken{},
		&model.EmailVerificationToken{},
		&model.PasswordResetToken{},
		&model.OAuthAccount{},
		&model.FileUpload{},
		&model.Org{},
		&model.Membership{},
		&model.Project{},
		&model.Deployment{},
		&model.EnvVar{},
		&model.BillingPlan{},
		&model.Subscription{},
		&model.Invoice{},
		&model.AuditLog{},
	); err != nil {
		slog.Error("AutoMigrate failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Database schema migrated")

	// --- 3b. Seed Default Plans ---
	featuregate.SeedDefaultPlans(db)

	// --- 3c. Seed Dev Users (non-production only) ---
	if strings.ToLower(cfg.App.Environment) != "production" {
		database.SeedDevUsers(db)
	}

	// --- 4. Repositories ---
	userRepo := user.NewRepository(db)
	orgRepo := org.NewRepository(db)
	projectRepo := project.NewRepository(db)
	billingRepo := billing.NewRepository(db)

	// --- 5. Services ---
	authService := auth.NewService(&cfg.JWT, db) // creates its own refresh token repo
	userService := user.NewService(userRepo)
	orgService := org.NewService(orgRepo)
	projectService := project.NewService(projectRepo)
	billingService := billing.NewService(billingRepo)
	gateService := featuregate.NewGateService(db)

	// --- 5b. Email Service ---
	emailService := email.NewResendProvider(cfg.Email.APIKey, cfg.Email.FromEmail)
	verificationService := user.NewVerificationService(db, emailService, cfg.App.Name, cfg.Email.AppURL)

	// --- 5c. Storage Service ---
	var uploadService *storage.UploadService
	s3Provider, err := storage.NewS3Provider(context.Background(), storage.S3Config{
		Endpoint:        cfg.Storage.Endpoint,
		Region:          cfg.Storage.Region,
		Bucket:          cfg.Storage.Bucket,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		UsePathStyle:    cfg.Storage.UsePathStyle,
		PublicURL:       cfg.Storage.PublicURL,
	})
	if err != nil {
		if cfg.App.Environment == "production" {
			slog.Error("Failed to initialize storage provider", "error", err)
			os.Exit(1)
		}
		slog.Warn("Storage provider not configured — file uploads disabled", "error", err)
	} else {
		uploadService = storage.NewUploadService(db, s3Provider)
	}

	// --- 5d. OAuth Providers ---
	oauthProviders := make(map[string]oauth.Provider)
	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	if cfg.App.Environment == "production" {
		baseURL = cfg.OAuth.FrontendURL // use the frontend URL for production redirect URIs
	}
	if cfg.OAuth.Google.Enabled {
		oauthProviders["google"] = oauth.NewGoogleProvider(cfg.OAuth.Google, baseURL)
		slog.Info("OAuth provider enabled", "provider", "google")
	}
	if cfg.OAuth.GitHub.Enabled {
		oauthProviders["github"] = oauth.NewGitHubProvider(cfg.OAuth.GitHub, baseURL)
		slog.Info("OAuth provider enabled", "provider", "github")
	}
	oauthService := oauth.NewOAuthService(db)
	oauthHandler := oauth.NewHandler(oauthProviders, oauthService, authService, cfg.OAuth.FrontendURL)

	// --- 5e. Auth Provider Selection ---
	var authProvider authprovider.AuthProvider
	if cfg.Supabase.Enabled {
		authProvider = authprovider.NewSupabaseProvider(cfg.Supabase)
		slog.Info("Auth provider: supabase", "url", cfg.Supabase.URL)
	} else {
		authProvider = authprovider.NewLocalProvider(authService, userService)
		slog.Info("Auth provider: local")
	}

	// --- 6. Handlers ---
	authHandler := auth.NewHandler(authProvider)
	userHandler := user.NewHandler(userService)
	orgHandler := org.NewHandler(orgService)
	projectHandler := project.NewHandler(projectService)
	billingHandler := billing.NewHandler(billingService, cfg.Xendit.WebhookToken)
	verificationHandler := user.NewVerificationHandler(verificationService)
	uploadHandler := storage.NewHandler(uploadService)

	// --- 7. Gin Router ---
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(apiErrors.ErrorHandler())

	// Security headers (X-Frame-Options, CSP, HSTS, etc.)
	r.Use(middleware.SecurityHeaders())

	// CORS
	allowedOrigins := cfg.CORS.AllowedOrigins
	if len(allowedOrigins) == 1 && strings.Contains(allowedOrigins[0], ",") {
		allowedOrigins = strings.Split(allowedOrigins[0], ",")
	}
	allowedHeaders := cfg.CORS.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-CSRF-Token", "X-Request-ID",
		}
	}
	corsMaxAge := cfg.CORS.MaxAge
	if corsMaxAge == 0 {
		corsMaxAge = 43200 // 12 hours
	}
	r.Use(middleware.CORS(
		allowedOrigins,
		allowedHeaders,
		cfg.CORS.AllowCredentials,
		corsMaxAge,
	))

	// CSRF (double-submit cookie, secure in production)
	isSecure := strings.ToLower(cfg.App.Environment) == "production"
	r.Use(middleware.CSRFProtection(isSecure))

	// --- 8. Health Checks ---
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if err := database.HealthCheck(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "unavailable",
				"error":    err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// --- 9. API v1 Routes ---
	v1 := r.Group("/api/v1")

	// Auth routes (public, rate-limited)
	authGroup := v1.Group("/auth")
	authLimiter := middleware.NewRateLimiter(5, 15*time.Minute) // 5 requests per 15 min per IP
	{
		authGroup.POST("/register", middleware.RateLimit(authLimiter), authHandler.Register)
		authGroup.POST("/login", middleware.RateLimit(authLimiter), authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/verify-email", verificationHandler.VerifyEmail)
		authGroup.POST("/request-reset", middleware.RateLimit(authLimiter), verificationHandler.RequestPasswordReset)
		authGroup.POST("/reset-password", middleware.RateLimit(authLimiter), verificationHandler.ResetPassword)
		authGroup.GET("/oauth/:provider", oauthHandler.Initiate)
		authGroup.GET("/oauth/:provider/callback", oauthHandler.Callback)
	}

	// Public billing plans
	v1.GET("/billing/plans", billingHandler.ListPlans)

	// Authenticated routes
	authed := v1.Group("")
	authed.Use(middleware.JWTAuth(authProvider))
	{
		// Auth (requires token)
		authed.POST("/auth/logout", authHandler.Logout)

		// Users
		authed.GET("/users/me", userHandler.GetMe)
		authed.PUT("/users/me", userHandler.UpdateMe)
		authed.POST("/users/me/avatar", uploadHandler.UploadUserAvatar)
		authed.GET("/users/me/oauth-accounts", oauthHandler.GetLinkedAccounts)
		authed.DELETE("/users/me/oauth-accounts/:provider", oauthHandler.UnlinkAccount)

		// Admin-only user listing
		admin := authed.Group("")
		admin.Use(middleware.RequireRole("super_admin", "admin"))
		{
			admin.GET("/users", userHandler.ListUsers)
		}

		// Orgs (top-level, no org context needed)
		authed.POST("/orgs", orgHandler.CreateOrg)
		authed.GET("/orgs", orgHandler.ListOrgs)

		// Invite acceptance (by token, no org context needed)
		authed.POST("/invites/:token/accept", orgHandler.AcceptInvite)

		// Org-scoped routes
		orgs := authed.Group("/orgs/:orgId")
		orgs.Use(middleware.OrgResolver(db))
		{
			// Org management
			orgs.GET("", orgHandler.GetOrg)
			orgs.PUT("", orgHandler.UpdateOrg)
			orgs.DELETE("", orgHandler.DeleteOrg)
			orgs.POST("/avatar", uploadHandler.UploadOrgAvatar)

			// Members
			orgs.GET("/members", orgHandler.ListMembers)
			orgs.PUT("/members/:memberId", orgHandler.UpdateMemberRole)
			orgs.DELETE("/members/:memberId", orgHandler.RemoveMember)

			// Invites
			orgs.POST("/invites", featuregate.RequireQuota(gateService, "members"), orgHandler.InviteMember)
			orgs.GET("/invites", orgHandler.ListInvites)
			orgs.DELETE("/invites/:inviteId", orgHandler.RevokeInvite)

			// Projects
			orgs.POST("/projects", featuregate.RequireQuota(gateService, "projects"), projectHandler.CreateProject)
			orgs.GET("/projects", projectHandler.ListProjects)
			orgs.GET("/projects/:projectId", projectHandler.GetProject)
			orgs.PUT("/projects/:projectId", projectHandler.UpdateProject)
			orgs.DELETE("/projects/:projectId", projectHandler.DeleteProject)

			// Deployments
			orgs.POST("/projects/:projectId/deployments", featuregate.RequireQuota(gateService, "deployments"), projectHandler.CreateDeployment)
			orgs.GET("/projects/:projectId/deployments", projectHandler.ListDeployments)

			// Env Vars
			orgs.POST("/projects/:projectId/env", projectHandler.SetEnvVar)
			orgs.GET("/projects/:projectId/env", projectHandler.ListEnvVars)
			orgs.DELETE("/projects/:projectId/env/:envVarId", projectHandler.DeleteEnvVar)

			// Billing
			orgs.GET("/billing", billingHandler.GetBillingOverview)
			orgs.POST("/billing/subscribe", billingHandler.CreateSubscription)
			orgs.POST("/billing/cancel", billingHandler.CancelSubscription)
			orgs.GET("/billing/invoices", billingHandler.ListInvoices)
			orgs.GET("/billing/usage", billingHandler.GetUsage)
		}
	}

	// Webhooks (no auth, verified by signature)
	webhooks := v1.Group("/webhooks")
	{
		webhooks.POST("/xendit", billingHandler.XenditWebhook)

		// Supabase auth webhook (syncs auth.users → local users table)
		if cfg.Supabase.Enabled {
			webhookHandler := authprovider.NewWebhookHandler(db, cfg.Supabase.WebhookSecret)
			webhooks.POST("/supabase/auth", webhookHandler.HandleAuthWebhook)
			slog.Info("Supabase auth webhook registered at /api/v1/webhooks/supabase/auth")
		}
	}

	// --- 10. Server ---
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// --- 11. Graceful Shutdown ---
	go func() {
		slog.Info("Server starting", "port", port, "environment", cfg.App.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutdown signal received")

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server stopped")
}

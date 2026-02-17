package middleware

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"paas-core/apps/api/internal/auth"
	apiErrors "paas-core/apps/api/internal/errors"
)

// TokenValidator is any component that can validate a JWT and return claims.
// Both auth.Service and authprovider.AuthProvider satisfy this interface.
type TokenValidator interface {
	ValidateToken(tokenString string) (*auth.Claims, error)
}

// RequestID injects a unique request-id header for tracing.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger is a structured slog request logger.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"ip", c.ClientIP(),
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}
		if reqID, exists := c.Get("request_id"); exists {
			attrs = append(attrs, "request_id", reqID)
		}

		switch {
		case status >= 500:
			slog.Error("Request", attrs...)
		case status >= 400:
			slog.Warn("Request", attrs...)
		default:
			slog.Info("Request", attrs...)
		}
	}
}

// JWTAuth validates the JWT access token from the Authorization header and
// stores the parsed Claims in the Gin context.
func JWTAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try Authorization header first
		header := c.GetHeader("Authorization")
		var tokenString string

		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				tokenString = parts[1]
			}
		}

		// Fall back to cookie
		if tokenString == "" {
			if cookie, err := c.Cookie("access_token"); err == nil {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			_ = c.Error(apiErrors.Unauthorized("Missing authorization token"))
			c.Abort()
			return
		}

		claims, err := validator.ValidateToken(tokenString)
		if err != nil {
			_ = c.Error(apiErrors.Unauthorized(err.Error()))
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// RequireRole checks that the authenticated user has at least one of the required roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get("claims")
		if !exists {
			_ = c.Error(apiErrors.Unauthorized(""))
			c.Abort()
			return
		}
		authClaims := claimsVal.(*auth.Claims)

		for _, userRole := range authClaims.Roles {
			for _, required := range roles {
				if userRole == required {
					c.Next()
					return
				}
			}
		}

		_ = c.Error(apiErrors.Forbidden("Insufficient permissions"))
		c.Abort()
	}
}

// CORS middleware for cross-origin requests.
func CORS(allowedOrigins []string, allowedHeaders []string, allowCredentials bool, maxAge int) gin.HandlerFunc {
	originsSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originsSet[o] = true
	}

	allowHeadersStr := strings.Join(allowedHeaders, ",")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if originsSet["*"] || originsSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", allowHeadersStr)
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")

		if allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if maxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(maxAge))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery wraps Gin's recovery with structured error responses.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		slog.Error("Panic recovered", "error", recovered, "path", c.Request.URL.Path)
		_ = c.Error(apiErrors.InternalServerError(nil))
		c.Abort()
	})
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds production-grade security headers to every response.
// Based on Goilerplate's security middleware:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - X-XSS-Protection: 0 (disabled; CSP handles this)
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Permissions-Policy: restrictive defaults
//   - Content-Security-Policy: restrictive defaults (API responses are JSON)
//   - Strict-Transport-Security: max-age 1 year, includeSubDomains
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()

		// Prevent MIME-type sniffing
		h.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		h.Set("X-Frame-Options", "DENY")

		// Disable legacy XSS filter (CSP is the modern approach)
		h.Set("X-XSS-Protection", "0")

		// Control referrer information
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// CSP for API (only serves JSON, no inline scripts/styles needed)
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// HSTS — enforce HTTPS for 1 year
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

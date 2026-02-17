package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apiErrors "paas-core/apps/api/internal/errors"
)

const (
	csrfCookieName = "__csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenLen   = 32
)

// CSRFProtection implements the double-submit cookie pattern.
// Safe methods (GET, HEAD, OPTIONS) are skipped.
// For state-changing methods the middleware checks that the header
// X-CSRF-Token matches the value in the __csrf_token cookie.
//
// Goilerplate pattern: https://goilerplate.com/docs/features/security
func CSRFProtection(secureCookie bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Always ensure a CSRF cookie exists (set on every response).
		token, err := c.Cookie(csrfCookieName)
		if err != nil || token == "" {
			token = generateCSRFToken()
		}
		sameSite := http.SameSiteLaxMode
		c.SetSameSite(sameSite)
		c.SetCookie(csrfCookieName, token, 86400, "/", "", secureCookie, false) // readable by JS

		// Safe methods — skip validation.
		method := strings.ToUpper(c.Request.Method)
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// State-changing method — validate double-submit.
		headerToken := c.GetHeader(csrfHeaderName)
		if headerToken == "" || headerToken != token {
			_ = c.Error(apiErrors.Forbidden("CSRF token mismatch"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLen)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

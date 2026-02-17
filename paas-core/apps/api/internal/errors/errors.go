package errors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// --- Structured Response Envelope ---

// Response is the standard API response envelope following JSend-inspired design.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains structured error details.
type ErrorInfo struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Path       string            `json:"path,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
	Timestamp  string            `json:"timestamp"`
	RetryAfter int               `json:"retry_after,omitempty"`
}

// Meta holds pagination and metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// --- APIError (implements error interface) ---

// APIError is a structured error that the error middleware can convert into a response.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]string
	RetryAfter int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// --- Factory Functions ---

func NotFound(msg string) *APIError {
	if msg == "" {
		msg = "Resource not found"
	}
	return &APIError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: msg}
}

func Unauthorized(msg string) *APIError {
	if msg == "" {
		msg = "Unauthorized"
	}
	return &APIError{StatusCode: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: msg}
}

func Forbidden(msg string) *APIError {
	if msg == "" {
		msg = "Forbidden"
	}
	return &APIError{StatusCode: http.StatusForbidden, Code: "FORBIDDEN", Message: msg}
}

func Conflict(msg string) *APIError {
	if msg == "" {
		msg = "Conflict"
	}
	return &APIError{StatusCode: http.StatusConflict, Code: "CONFLICT", Message: msg}
}

func BadRequest(msg string) *APIError {
	if msg == "" {
		msg = "Bad request"
	}
	return &APIError{StatusCode: http.StatusBadRequest, Code: "BAD_REQUEST", Message: msg}
}

func ValidationError(details map[string]string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Details:    details,
	}
}

func InternalServerError(err error) *APIError {
	msg := "Internal server error"
	if err != nil {
		msg = err.Error()
	}
	return &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: msg}
}

func RateLimitExceeded(retryAfter int) *APIError {
	return &APIError{
		StatusCode: http.StatusTooManyRequests,
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Too many requests",
		RetryAfter: retryAfter,
	}
}

// FromGinValidation converts Gin's binding validation errors into a structured APIError.
func FromGinValidation(err error) *APIError {
	details := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				details[field] = fmt.Sprintf("%s is required", field)
			case "email":
				details[field] = fmt.Sprintf("%s must be a valid email", field)
			case "min":
				details[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
			case "max":
				details[field] = fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
			default:
				details[field] = fmt.Sprintf("%s failed validation: %s", field, fe.Tag())
			}
		}
	} else {
		details["body"] = err.Error()
	}
	return ValidationError(details)
}

// --- Response Helpers ---

// Success creates a success envelope.
func Success(data interface{}) Response {
	return Response{Success: true, Data: data}
}

// SuccessWithMeta creates a success envelope with pagination metadata.
func SuccessWithMeta(data interface{}, meta *Meta) Response {
	return Response{Success: true, Data: data, Meta: meta}
}

// --- Gin Error Middleware ---

// ErrorHandler is a Gin middleware that catches errors added via c.Error() and
// returns a structured JSON response.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		lastErr := c.Errors.Last()
		apiErr, ok := lastErr.Err.(*APIError)
		if !ok {
			apiErr = InternalServerError(lastErr.Err)
		}

		requestID, _ := c.Get("request_id")
		reqIDStr, _ := requestID.(string)

		resp := Response{
			Success: false,
			Error: &ErrorInfo{
				Code:       apiErr.Code,
				Message:    apiErr.Message,
				Details:    apiErr.Details,
				Path:       c.Request.URL.Path,
				RequestID:  reqIDStr,
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				RetryAfter: apiErr.RetryAfter,
			},
		}

		c.JSON(apiErr.StatusCode, resp)
	}
}

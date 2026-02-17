package user

import (
	"errors"
	"strings"
	"unicode"
)

// Password validation errors.
var (
	ErrPasswordTooShort    = errors.New("password must be at least 12 characters")
	ErrPasswordNoUpper     = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower     = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit     = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial   = errors.New("password must contain at least one special character")
	ErrPasswordCommon      = errors.New("password is too common")
)

// commonPasswords is a minimal blocklist of extremely common passwords.
// In production, consider checking against the HaveIBeenPwned API.
var commonPasswords = map[string]bool{
	"password1234": true,
	"123456789012": true,
	"qwertyuiop12": true,
	"password1234!": true,
	"admin12345678": true,
	"letmein123456": true,
	"welcome12345":  true,
	"changeme1234":  true,
	"iloveyou1234":  true,
	"trustno1trust": true,
}

// ValidatePasswordNIST validates a password according to NIST SP 800-63B guidelines:
//   - Minimum 12 characters (Goilerplate standard)
//   - At least one uppercase letter
//   - At least one lowercase letter
//   - At least one digit
//   - At least one special character
//   - Not in the common password blocklist
//
// These rules are stricter than the bare NIST minimum (8 chars + blocklist)
// but aligned with Goilerplate's production hardening approach.
func ValidatePasswordNIST(password string) error {
	if len(password) < 12 {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	if commonPasswords[strings.ToLower(password)] {
		return ErrPasswordCommon
	}

	return nil
}

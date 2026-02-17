package email

import "fmt"

// RenderVerificationEmail returns the HTML body for an email verification message.
func RenderVerificationEmail(data TemplateData) Message {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px; color: #1a1a1a;">
  <h1 style="font-size: 24px; margin-bottom: 24px;">Verify your email</h1>
  <p>Hi %s,</p>
  <p>Welcome to <strong>%s</strong>! Please verify your email address by clicking the button below.</p>
  <div style="text-align: center; margin: 32px 0;">
    <a href="%s" style="display: inline-block; padding: 12px 32px; background: #0070f3; color: #fff; text-decoration: none; border-radius: 6px; font-weight: 600;">Verify Email</a>
  </div>
  <p style="font-size: 14px; color: #666;">This link expires in %s. If you didn't create an account, you can safely ignore this email.</p>
  <hr style="border: none; border-top: 1px solid #eee; margin: 32px 0;">
  <p style="font-size: 12px; color: #999;">%s</p>
</body>
</html>`, data.UserName, data.AppName, data.Link, data.ExpiresIn, data.AppName)

	text := fmt.Sprintf("Hi %s,\n\nWelcome to %s! Verify your email: %s\n\nThis link expires in %s.", data.UserName, data.AppName, data.Link, data.ExpiresIn)

	return Message{
		To:       data.UserEmail,
		Subject:  fmt.Sprintf("Verify your email — %s", data.AppName),
		HTMLBody: html,
		TextBody: text,
	}
}

// RenderPasswordResetEmail returns the HTML body for a password reset message.
func RenderPasswordResetEmail(data TemplateData) Message {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px; color: #1a1a1a;">
  <h1 style="font-size: 24px; margin-bottom: 24px;">Reset your password</h1>
  <p>Hi %s,</p>
  <p>We received a request to reset your password for <strong>%s</strong>. Click the button below to choose a new password.</p>
  <div style="text-align: center; margin: 32px 0;">
    <a href="%s" style="display: inline-block; padding: 12px 32px; background: #0070f3; color: #fff; text-decoration: none; border-radius: 6px; font-weight: 600;">Reset Password</a>
  </div>
  <p style="font-size: 14px; color: #666;">This link expires in %s. If you didn't request a password reset, you can safely ignore this email.</p>
  <hr style="border: none; border-top: 1px solid #eee; margin: 32px 0;">
  <p style="font-size: 12px; color: #999;">%s</p>
</body>
</html>`, data.UserName, data.AppName, data.Link, data.ExpiresIn, data.AppName)

	text := fmt.Sprintf("Hi %s,\n\nReset your password for %s: %s\n\nThis link expires in %s.", data.UserName, data.AppName, data.Link, data.ExpiresIn)

	return Message{
		To:       data.UserEmail,
		Subject:  fmt.Sprintf("Reset your password — %s", data.AppName),
		HTMLBody: html,
		TextBody: text,
	}
}

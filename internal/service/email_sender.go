package service

import (
	"context"
	"fmt"
	"html"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, token string) error
	SendPasswordChangedEmail(ctx context.Context, toEmail string) error
	SendAccountLockoutEmail(ctx context.Context, toEmail, lockedUntil string) error
}

type SendGridEmailSender struct {
	apiKey    string
	fromEmail string
	fromName  string
	baseURL   string
}

func NewSendGridEmailSender(frontendBaseURL string) *SendGridEmailSender {
	return &SendGridEmailSender{
		apiKey:    os.Getenv("SENDGRID_API_KEY"),
		fromEmail: getEnvOrDefault("EMAIL_FROM", "noreply@em2292.zeno-cy.com"),
		fromName:  getEnvOrDefault("EMAIL_FROM_NAME", "ZenoN Cloud"),
		baseURL:   frontendBaseURL,
	}
}

func (s *SendGridEmailSender) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	if s.apiKey == "" {
		log.Warn().Str("to", html.EscapeString(toEmail)).Msg("SendGrid API key not set, skipping email")
		return nil
	}
	log.Info().Str("to", html.EscapeString(toEmail)).Str("from", s.fromEmail).Str("api_key_set", "yes").Msg("Sending verification email")

	verifyURL := fmt.Sprintf("%s#/verify-email?token=%s", s.baseURL, token)

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("", toEmail)
	subject := "Verify your email address"

	plainTextContent := fmt.Sprintf(`Hello,

Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you did not create an account, please ignore this email.

Best regards,
ZenoN Cloud Team
`, verifyURL)

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2563eb;">Verify Your Email Address</h2>
        <p>Hello,</p>
        <p>Please verify your email address by clicking the button below:</p>
        <div style="margin: 30px 0;">
            <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
        </div>
        <p style="color: #666; font-size: 14px;">Or copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 14px; word-break: break-all;">%s</p>
        <p style="color: #666; font-size: 14px;">This link will expire in 24 hours.</p>
        <p style="color: #666; font-size: 14px;">If you did not create an account, please ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">Best regards,<br>ZenoN Cloud Team</p>
    </div>
</body>
</html>`, html.EscapeString(verifyURL), html.EscapeString(verifyURL))

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.apiKey)

	response, err := client.Send(message)
	if err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("Failed to send verification email")
		return err
	}
	if response.StatusCode >= 400 {
		log.Error().Int("status", response.StatusCode).Str("to", html.EscapeString(toEmail)).Msg("SendGrid returned error")
		return fmt.Errorf("sendgrid error: %d", response.StatusCode)
	}
	log.Info().Str("to", html.EscapeString(toEmail)).Int("status", response.StatusCode).Msg("Verification email sent")
	return nil
}

func (s *SendGridEmailSender) SendPasswordResetEmail(ctx context.Context, toEmail, token string) error {
	if s.apiKey == "" {
		log.Warn().Str("to", html.EscapeString(toEmail)).Msg("SendGrid API key not set, skipping email")
		return nil
	}

	resetURL := fmt.Sprintf("%s#/reset-password?token=%s", s.baseURL, token)

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("", toEmail)
	subject := "Reset your password"

	plainTextContent := fmt.Sprintf(`Hello,

You requested to reset your password. Click the link below to set a new password:

%s

This link will expire in 1 hour.

If you did not request this, please ignore this email and your password will remain unchanged.

Best regards,
ZenoN Cloud Team
`, resetURL)

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2563eb;">Reset Your Password</h2>
        <p>Hello,</p>
        <p>You requested to reset your password. Click the button below to set a new password:</p>
        <div style="margin: 30px 0;">
            <a href="%s" style="background-color: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p style="color: #666; font-size: 14px;">Or copy and paste this link into your browser:</p>
        <p style="color: #666; font-size: 14px; word-break: break-all;">%s</p>
        <p style="color: #666; font-size: 14px;">This link will expire in 1 hour.</p>
        <p style="color: #666; font-size: 14px;">If you did not request this, please ignore this email and your password will remain unchanged.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">Best regards,<br>ZenoN Cloud Team</p>
    </div>
</body>
</html>`, html.EscapeString(resetURL), html.EscapeString(resetURL))

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.apiKey)

	response, err := client.Send(message)
	if err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("Failed to send password reset email")
		return err
	}

	if response.StatusCode >= 400 {
		log.Error().Int("status", response.StatusCode).Str("to", html.EscapeString(toEmail)).Msg("SendGrid returned error")
		return fmt.Errorf("sendgrid error: %d", response.StatusCode)
	}

	log.Info().Str("to", html.EscapeString(toEmail)).Msg("Password reset email sent")
	return nil
}

func (s *SendGridEmailSender) SendPasswordChangedEmail(ctx context.Context, toEmail string) error {
	if s.apiKey == "" {
		log.Warn().Str("to", html.EscapeString(toEmail)).Msg("SendGrid API key not set, skipping email")
		return nil
	}

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("", toEmail)
	subject := "Your password has been changed"

	plainTextContent := `Hello,

Your password has been successfully changed.

If you did not make this change, please contact support immediately.

Best regards,
ZenoN Cloud Team
`

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2563eb;">Password Changed</h2>
        <p>Hello,</p>
        <p>Your password has been successfully changed.</p>
        <p style="color: #dc2626; font-weight: bold;">If you did not make this change, please contact support immediately.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">Best regards,<br>ZenoN Cloud Team</p>
    </div>
</body>
</html>`

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.apiKey)

	response, err := client.Send(message)
	if err != nil {
		log.Error().Err(err).Str("to", html.EscapeString(toEmail)).Msg("Failed to send password changed email")
		return err
	}

	if response.StatusCode >= 400 {
		log.Error().Int("status", response.StatusCode).Str("to", html.EscapeString(toEmail)).Msg("SendGrid returned error")
		return fmt.Errorf("sendgrid error: %d", response.StatusCode)
	}

	log.Info().Str("to", html.EscapeString(toEmail)).Msg("Password changed email sent")
	return nil
}

func (s *SendGridEmailSender) SendAccountLockoutEmail(ctx context.Context, toEmail, lockedUntil string) error {
	if s.apiKey == "" {
		log.Warn().Str("to", html.EscapeString(toEmail)).Msg("SendGrid API key not set, skipping email")
		return nil
	}

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("", toEmail)
	subject := "Account temporarily locked"

	plainTextContent := fmt.Sprintf(`Hello,

Your account has been temporarily locked due to multiple failed login attempts.

Your account will be automatically unlocked at: %s

If this was not you, please contact support immediately.

Best regards,
ZenoN Cloud Team
`, lockedUntil)

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #dc2626;">Account Temporarily Locked</h2>
        <p>Hello,</p>
        <p>Your account has been temporarily locked due to multiple failed login attempts.</p>
        <p><strong>Your account will be automatically unlocked at:</strong> %s</p>
        <p style="color: #dc2626; font-weight: bold;">If this was not you, please contact support immediately.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">Best regards,<br>ZenoN Cloud Team</p>
    </div>
</body>
</html>`, html.EscapeString(lockedUntil))

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.apiKey)

	response, err := client.Send(message)
	if err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("Failed to send lockout email")
		return err
	}

	if response.StatusCode >= 400 {
		log.Error().Int("status", response.StatusCode).Str("to", toEmail).Msg("SendGrid returned error")
		return fmt.Errorf("sendgrid error: %d", response.StatusCode)
	}

	log.Info().Str("to", toEmail).Msg("Account lockout email sent")
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

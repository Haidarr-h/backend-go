package otp

import (
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/pkg/logger"
	"github.com/resend/resend-go/v3"
)

func SendOTP(toEmail string, otp string, cfg *config.Config) error {
	client := resend.NewClient(cfg.ResendKey)

	html := `
	<div style="font-family:Arial,sans-serif;max-width:480px;margin:0 auto;padding:32px;background:#f9f9f9;border-radius:8px;">
		<h1 style="color:#1a1a1a;font-size:24px;margin-bottom:8px;">IronLog</h1>
		<p style="color:#444;font-size:16px;margin-bottom:24px;">Use the code below to verify your identity. It expires in <strong>5 minutes</strong>.</p>
		<div style="background:#1a1a1a;color:#ffffff;font-size:36px;font-weight:bold;letter-spacing:12px;text-align:center;padding:24px;border-radius:8px;">` + otp + `</div>
		<p style="color:#888;font-size:13px;margin-top:24px;">If you didn't request this, you can safely ignore this email.</p>
	</div>`

	params := &resend.SendEmailRequest{
		From:    "IronLog <noreply@liftlogs.my.id>",
		To:      []string{toEmail},
		Html:    html,
		Subject: "OTP Verification Code IronLog",
	}

	sent, err := client.Emails.Send(params)
	logger.Log.Info("send email code", "send", sent)

	if err != nil {
		logger.Log.Error("error sending otp by resend", "error", err)
		return err
	}

	return nil
}
